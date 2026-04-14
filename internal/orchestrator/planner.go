package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"maliangswarm/internal/domain"
)

var plannerHTTPClient = &http.Client{
	Timeout: 45 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("stopped after %d redirects", len(via))
		}
		if len(via) > 0 && via[0].Method == http.MethodPost {
			req.Method = http.MethodPost
			if via[0].Body != nil {
				req.Body = via[0].Body
				req.GetBody = via[0].GetBody
				req.ContentLength = via[0].ContentLength
			}
			for key, values := range via[0].Header {
				if key != "Content-Length" {
					req.Header[key] = values
				}
			}
		}
		return nil
	},
}

type plannerOutput struct {
	Title                 *string             `json:"title"`
	Deliverable           *string             `json:"deliverable"`
	TaskType              *string             `json:"taskType"`
	Priority              *string             `json:"priority"`
	MaxAgents             *int                `json:"maxAgents"`
	RequiresQA            *bool               `json:"requiresQA"`
	RequiresSecurity      *bool               `json:"requiresSecurity"`
	RequiresHumanApproval *bool               `json:"requiresHumanApproval"`
	ActiveTemplate        *string             `json:"activeTemplate"`
	AtomicTaskPolicy      *string             `json:"atomicTaskPolicy"`
	ReviewMode            *string             `json:"reviewMode"`
	Workstreams           []plannerWorkstream `json:"workstreams"`
}

type plannerWorkstream struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Artifact string `json:"artifact"`
	Detail   string `json:"detail"`
}

func buildPlanningBlueprint(ctx context.Context, input RunCreationInput, providers []domain.AIProviderConfig) planningBlueprint {
	fallback := buildFallbackBlueprint(input)
	provider, ok := selectPlanningProvider(providers)
	if !ok {
		fallback.PlanningSource = "rules fallback / no live planner route"
		return fallback
	}

	output, err := requestPlanningOutput(ctx, provider, fallback.Input)
	if err != nil {
		fallback.PlanningSource = fmt.Sprintf("rules fallback / %s unavailable", provider.Name)
		return fallback
	}

	return mergePlanningBlueprint(fallback, output, provider.Name)
}

func selectPlanningProvider(providers []domain.AIProviderConfig) (domain.AIProviderConfig, bool) {
	providers = normalizeProviders(providers)
	if len(providers) == 0 {
		return domain.AIProviderConfig{}, false
	}

	if primary, ok := primaryProvider(providers); ok && providerReadyForPlanning(primary) {
		return primary, true
	}

	for _, provider := range providers {
		if providerReadyForPlanning(provider) {
			return provider, true
		}
	}

	return domain.AIProviderConfig{}, false
}

func providerReadyForPlanning(provider domain.AIProviderConfig) bool {
	if !provider.Enabled {
		return false
	}
	if provider.PlannerModel == "" {
		return false
	}

	switch provider.Format {
	case "ollama", "custom-http":
		return strings.TrimSpace(provider.BaseURL) != ""
	default:
		return strings.TrimSpace(provider.BaseURL) != "" && strings.TrimSpace(provider.APIKey) != ""
	}
}

func mergePlanningBlueprint(fallback planningBlueprint, output plannerOutput, providerName string) planningBlueprint {
	result := fallback

	if output.Title != nil && compact(*output.Title) != "" {
		result.Input.Title = compact(*output.Title)
	}
	if output.Deliverable != nil && compact(*output.Deliverable) != "" {
		result.Input.Deliverable = compact(*output.Deliverable)
	}
	if output.TaskType != nil && compact(*output.TaskType) != "" {
		result.Input.TaskType = strings.ToLower(compact(*output.TaskType))
	}
	if output.Priority != nil && compact(*output.Priority) != "" {
		result.Input.Priority = strings.ToLower(compact(*output.Priority))
	}
	if output.MaxAgents != nil && *output.MaxAgents > 0 {
		result.Input.MaxAgents = min(max(*output.MaxAgents, 1), 100)
	}
	if output.RequiresQA != nil {
		result.Input.RequiresQA = *output.RequiresQA
	}
	if output.RequiresSecurity != nil {
		result.Input.RequiresSecurity = *output.RequiresSecurity
	}
	if output.RequiresHumanApproval != nil {
		result.Input.RequiresHumanApproval = *output.RequiresHumanApproval
	}
	if output.ActiveTemplate != nil && compact(*output.ActiveTemplate) != "" {
		result.ActiveTemplate = compact(*output.ActiveTemplate)
	}
	if output.AtomicTaskPolicy != nil && compact(*output.AtomicTaskPolicy) != "" {
		result.AtomicTaskPolicy = compact(*output.AtomicTaskPolicy)
	}
	if output.ReviewMode != nil && compact(*output.ReviewMode) != "" {
		result.ReviewMode = compact(*output.ReviewMode)
	}
	if len(output.Workstreams) > 0 {
		streams := make([]workstreamSpec, 0, len(output.Workstreams))
		for _, stream := range output.Workstreams {
			streams = append(streams, workstreamSpec{
				id:       compact(stream.ID),
				name:     compact(stream.Name),
				artifact: compact(stream.Artifact),
				detail:   compact(stream.Detail),
			})
		}
		result.Workstreams = streams
	}

	result.Input = normalizeRunInputForCompiler(result.Input)
	result.Workstreams = normalizeBlueprintWorkstreams(result.Workstreams, result.Input)
	result.PlanningSource = fmt.Sprintf("live via %s", providerName)
	return result
}

func requestPlanningOutput(ctx context.Context, provider domain.AIProviderConfig, input RunCreationInput) (plannerOutput, error) {
	systemPrompt, userPrompt := buildPlanningPrompts(input)

	switch provider.Format {
	case "anthropic":
		return requestAnthropicPlan(ctx, provider, systemPrompt, userPrompt)
	case "gemini":
		return requestGeminiPlan(ctx, provider, systemPrompt, userPrompt)
	case "azure-openai":
		return requestAzureOpenAIPlan(ctx, provider, systemPrompt, userPrompt)
	case "ollama":
		return requestOllamaPlan(ctx, provider, systemPrompt, userPrompt)
	case "custom-http":
		return requestCustomHTTPPlan(ctx, provider, systemPrompt, userPrompt)
	default:
		return requestOpenAICompatiblePlan(ctx, provider, systemPrompt, userPrompt)
	}
}

func buildPlanningPrompts(input RunCreationInput) (string, string) {
	systemPrompt := strings.TrimSpace(`你是 maliang swarm 的主脑规划器。
只返回一个 JSON 对象。
你的任务是把用户提交的任务简报转换成一个适合大型公司流程的执行蓝图。
规则：
- 方案必须适配大型团队的工作流。
- 每条执行链路都要足够原子，便于干净交接。
- 必须遵守 Agent 上限，绝不能超过 100。
- 仅在必要时加入 QA、安全审核和人工审批。
- 优先输出 3 到 8 条工作流，并给出明确产物。
- 措辞要简洁、可执行、适合直接落地。`)

	userPrompt := fmt.Sprintf(`请为下面这个任务生成规划蓝图。

标题：%s
任务目标：%s
期望交付物：%s
任务类型提示：%s
优先级提示：%s
期望最大 Agent 数：%d
是否建议 QA：%t
是否建议安全审核：%t
是否建议人工审批：%t

请返回 JSON，字段包括：
- title
- deliverable
- taskType
- priority
- maxAgents
- requiresQA
- requiresSecurity
- requiresHumanApproval
- activeTemplate
- atomicTaskPolicy
- reviewMode
- workstreams：对象数组，每项包含 id、name、artifact、detail

这些工作流必须体现办公室会真实配置的主要执行链路。`,
		input.Title,
		input.Mission,
		input.Deliverable,
		input.TaskType,
		input.Priority,
		input.MaxAgents,
		input.RequiresQA,
		input.RequiresSecurity,
		input.RequiresHumanApproval,
	)

	return systemPrompt, userPrompt
}

func requestOpenAICompatiblePlan(ctx context.Context, provider domain.AIProviderConfig, systemPrompt string, userPrompt string) (plannerOutput, error) {
	targetURL, err := buildProviderURL(provider, provider.APIPath)
	if err != nil {
		return plannerOutput{}, err
	}

	var payload map[string]any
	if strings.Contains(strings.ToLower(provider.APIPath), "chat/completions") {
		payload = map[string]any{
			"model": provider.PlannerModel,
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": userPrompt},
			},
			"temperature": 0.2,
			"response_format": map[string]any{
				"type": "json_schema",
				"json_schema": map[string]any{
					"name":   "maliang_swarm_plan",
					"strict": true,
					"schema": planningSchema(),
				},
			},
		}
	} else {
		payload = map[string]any{
			"model": provider.PlannerModel,
			"input": []map[string]any{
				{"role": "system", "content": []map[string]string{{"type": "input_text", "text": systemPrompt}}},
				{"role": "user", "content": []map[string]string{{"type": "input_text", "text": userPrompt}}},
			},
			"temperature": 0.2,
			"text": map[string]any{
				"format": map[string]any{
					"type":   "json_schema",
					"name":   "maliang_swarm_plan",
					"strict": true,
					"schema": planningSchema(),
				},
			},
		}
	}

	body, err := postProviderJSON(ctx, provider, targetURL, payload)
	if err != nil {
		return plannerOutput{}, err
	}
	return decodePlannerOutput(body)
}

func requestAnthropicPlan(ctx context.Context, provider domain.AIProviderConfig, systemPrompt string, userPrompt string) (plannerOutput, error) {
	targetURL, err := buildProviderURL(provider, provider.APIPath)
	if err != nil {
		return plannerOutput{}, err
	}

	payload := map[string]any{
		"model":       provider.PlannerModel,
		"max_tokens":  1800,
		"temperature": 0.2,
		"system":      systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	body, err := postProviderJSON(ctx, provider, targetURL, payload)
	if err != nil {
		return plannerOutput{}, err
	}
	return decodePlannerOutput(body)
}

func requestGeminiPlan(ctx context.Context, provider domain.AIProviderConfig, systemPrompt string, userPrompt string) (plannerOutput, error) {
	targetURL, err := buildGeminiURL(provider)
	if err != nil {
		return plannerOutput{}, err
	}

	payload := map[string]any{
		"system_instruction": map[string]any{
			"parts": []map[string]string{{"text": systemPrompt}},
		},
		"contents": []map[string]any{
			{"parts": []map[string]string{{"text": userPrompt}}},
		},
		"generationConfig": map[string]any{
			"temperature":      0.2,
			"responseMimeType": "application/json",
			"responseSchema":   planningSchema(),
		},
	}

	body, err := postProviderJSON(ctx, provider, targetURL, payload)
	if err != nil {
		return plannerOutput{}, err
	}
	return decodePlannerOutput(body)
}

func requestAzureOpenAIPlan(ctx context.Context, provider domain.AIProviderConfig, systemPrompt string, userPrompt string) (plannerOutput, error) {
	targetURL, err := buildProviderURL(provider, provider.APIPath)
	if err != nil {
		return plannerOutput{}, err
	}
	if provider.APIVersion != "" && !strings.Contains(targetURL, "api-version=") {
		targetURL, err = addQueryValue(targetURL, "api-version", provider.APIVersion)
		if err != nil {
			return plannerOutput{}, err
		}
	}

	payload := map[string]any{
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
	}
	if !strings.Contains(strings.ToLower(provider.APIPath), "/deployments/") {
		payload["model"] = provider.PlannerModel
	}

	body, err := postProviderJSON(ctx, provider, targetURL, payload)
	if err != nil {
		return plannerOutput{}, err
	}
	return decodePlannerOutput(body)
}

func requestOllamaPlan(ctx context.Context, provider domain.AIProviderConfig, systemPrompt string, userPrompt string) (plannerOutput, error) {
	targetURL, err := buildProviderURL(provider, provider.APIPath)
	if err != nil {
		return plannerOutput{}, err
	}

	payload := map[string]any{
		"model":  provider.PlannerModel,
		"stream": false,
		"format": planningSchema(),
	}

	if strings.Contains(strings.ToLower(provider.APIPath), "/api/generate") {
		payload["prompt"] = systemPrompt + "\n\n" + userPrompt
	} else {
		payload["messages"] = []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		}
	}

	body, err := postProviderJSON(ctx, provider, targetURL, payload)
	if err != nil {
		return plannerOutput{}, err
	}
	return decodePlannerOutput(body)
}

func requestCustomHTTPPlan(ctx context.Context, provider domain.AIProviderConfig, systemPrompt string, userPrompt string) (plannerOutput, error) {
	targetURL, err := buildProviderURL(provider, provider.APIPath)
	if err != nil {
		return plannerOutput{}, err
	}

	payload := map[string]any{
		"model":       provider.PlannerModel,
		"system":      systemPrompt,
		"input":       userPrompt,
		"temperature": 0.2,
		"schema":      planningSchema(),
	}

	body, err := postProviderJSON(ctx, provider, targetURL, payload)
	if err != nil {
		return plannerOutput{}, err
	}
	return decodePlannerOutput(body)
}

func postProviderJSON(ctx context.Context, provider domain.AIProviderConfig, targetURL string, payload map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")

	applyProviderAuthHeaders(req, provider)
	if err := applyProviderCustomHeaders(req, provider.HeadersJSON); err != nil {
		return nil, err
	}

	resp, err := plannerHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		detail := compact(string(responseBody))
		if detail == "" {
			return nil, fmt.Errorf("planner request failed with %s", resp.Status)
		}
		if len(detail) > 240 {
			detail = detail[:237] + "..."
		}
		return nil, fmt.Errorf("planner request failed with %s: %s", resp.Status, detail)
	}

	return responseBody, nil
}

func applyProviderAuthHeaders(req *http.Request, provider domain.AIProviderConfig) {
	switch provider.Format {
	case "anthropic":
		if provider.APIKey != "" {
			req.Header.Set("x-api-key", provider.APIKey)
		}
		if provider.APIVersion != "" {
			req.Header.Set("anthropic-version", provider.APIVersion)
		}
	case "azure-openai":
		if provider.APIKey != "" {
			req.Header.Set("api-key", provider.APIKey)
		}
	case "gemini":
		if provider.APIKey != "" {
			query := req.URL.Query()
			query.Set("key", provider.APIKey)
			req.URL.RawQuery = query.Encode()
		}
	default:
		if provider.APIKey != "" {
			req.Header.Set("authorization", "Bearer "+provider.APIKey)
		}
	}
}

func applyProviderCustomHeaders(req *http.Request, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return nil
	}

	headers := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &headers); err != nil {
		return err
	}
	for key, value := range headers {
		req.Header.Set(key, fmt.Sprint(value))
	}
	return nil
}

func buildGeminiURL(provider domain.AIProviderConfig) (string, error) {
	path := strings.TrimSpace(provider.APIPath)
	if path == "" {
		path = "/v1beta/models"
	}
	if strings.Contains(path, "{model}") {
		path = strings.ReplaceAll(path, "{model}", provider.PlannerModel)
	} else if !strings.Contains(path, ":generateContent") {
		path = strings.TrimRight(path, "/") + "/" + provider.PlannerModel + ":generateContent"
	}

	return buildProviderURL(provider, path)
}

func buildProviderURL(provider domain.AIProviderConfig, path string) (string, error) {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}

	baseURL := strings.TrimSpace(provider.BaseURL)
	if baseURL == "" {
		return "", errors.New("provider base URL is empty")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid provider base URL: %w", err)
	}
	if path == "" {
		return strings.TrimRight(base.String(), "/"), nil
	}

	relative, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid provider API path: %w", err)
	}

	normalizedPath := relative.Path
	if normalizedPath == "" {
		normalizedPath = "/"
	}
	if !strings.HasPrefix(normalizedPath, "/") {
		normalizedPath = "/" + normalizedPath
	}

	basePath := strings.TrimRight(base.Path, "/")
	if basePath != "" && (normalizedPath == basePath || strings.HasPrefix(normalizedPath, basePath+"/")) {
		normalizedPath = strings.TrimPrefix(normalizedPath, basePath)
	}

	base.Path = strings.TrimRight(basePath, "/") + normalizedPath
	query := base.Query()
	for key, values := range relative.Query() {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func addQueryValue(targetURL string, key string, value string) (string, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set(key, value)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func decodePlannerOutput(body []byte) (plannerOutput, error) {
	var direct plannerOutput
	if err := json.Unmarshal(body, &direct); err == nil && direct.hasMeaningfulContent() {
		return direct, nil
	}

	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return plannerOutput{}, err
	}

	if object, ok := locatePlannerObject(payload); ok {
		raw, err := json.Marshal(object)
		if err == nil {
			if err := json.Unmarshal(raw, &direct); err == nil && direct.hasMeaningfulContent() {
				return direct, nil
			}
		}
	}

	text := locatePlannerText(payload)
	if text == "" {
		return plannerOutput{}, errors.New("planner response did not contain text or JSON")
	}

	return decodePlannerText(text)
}

func decodePlannerText(text string) (plannerOutput, error) {
	jsonBlob := extractJSONObject(text)
	if jsonBlob == "" {
		return plannerOutput{}, errors.New("planner text did not contain a JSON object")
	}

	var output plannerOutput
	if err := json.Unmarshal([]byte(jsonBlob), &output); err != nil {
		return plannerOutput{}, err
	}
	if !output.hasMeaningfulContent() {
		return plannerOutput{}, errors.New("planner JSON was empty")
	}
	return output, nil
}

func (output plannerOutput) hasMeaningfulContent() bool {
	return output.Title != nil ||
		output.Deliverable != nil ||
		output.TaskType != nil ||
		output.Priority != nil ||
		output.MaxAgents != nil ||
		output.RequiresQA != nil ||
		output.RequiresSecurity != nil ||
		output.RequiresHumanApproval != nil ||
		output.ActiveTemplate != nil ||
		output.AtomicTaskPolicy != nil ||
		output.ReviewMode != nil ||
		len(output.Workstreams) > 0
}

func locatePlannerObject(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"plan", "planning_blueprint", "blueprint"} {
			if nested, ok := typed[key].(map[string]any); ok {
				return nested, true
			}
		}
	case []any:
		for _, item := range typed {
			if nested, ok := locatePlannerObject(item); ok {
				return nested, true
			}
		}
	}
	return nil, false
}

func locatePlannerText(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		chunks := make([]string, 0, len(typed))
		for _, item := range typed {
			if chunk := locatePlannerText(item); strings.TrimSpace(chunk) != "" {
				chunks = append(chunks, chunk)
			}
		}
		return strings.Join(chunks, "\n")
	case map[string]any:
		for _, key := range []string{"output_text", "text", "response"} {
			if chunk, ok := typed[key].(string); ok && strings.TrimSpace(chunk) != "" {
				return chunk
			}
		}
		for _, key := range []string{"message", "content", "parts", "choices", "candidates", "output"} {
			if nested, ok := typed[key]; ok {
				if chunk := locatePlannerText(nested); strings.TrimSpace(chunk) != "" {
					return chunk
				}
			}
		}
	}
	return ""
}

func extractJSONObject(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	if json.Valid([]byte(text)) {
		return text
	}

	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}

	depth := 0
	inString := false
	escaped := false
	for index := start; index < len(text); index++ {
		char := text[index]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if char == '\\' {
				escaped = true
				continue
			}
			if char == '"' {
				inString = false
			}
			continue
		}

		switch char {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				candidate := strings.TrimSpace(text[start : index+1])
				if json.Valid([]byte(candidate)) {
					return candidate
				}
				return ""
			}
		}
	}

	return ""
}

func planningSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"title":                 map[string]any{"type": "string"},
			"deliverable":           map[string]any{"type": "string"},
			"taskType":              map[string]any{"type": "string", "enum": []string{"engineering", "product", "research", "content"}},
			"priority":              map[string]any{"type": "string", "enum": []string{"low", "medium", "high", "critical"}},
			"maxAgents":             map[string]any{"type": "integer", "minimum": 1, "maximum": 100},
			"requiresQA":            map[string]any{"type": "boolean"},
			"requiresSecurity":      map[string]any{"type": "boolean"},
			"requiresHumanApproval": map[string]any{"type": "boolean"},
			"activeTemplate":        map[string]any{"type": "string"},
			"atomicTaskPolicy":      map[string]any{"type": "string"},
			"reviewMode":            map[string]any{"type": "string"},
			"workstreams": map[string]any{
				"type":     "array",
				"minItems": 3,
				"maxItems": 12,
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"id":       map[string]any{"type": "string"},
						"name":     map[string]any{"type": "string"},
						"artifact": map[string]any{"type": "string"},
						"detail":   map[string]any{"type": "string"},
					},
					"required": []string{"name", "artifact", "detail"},
				},
			},
		},
		"required": []string{
			"title",
			"deliverable",
			"taskType",
			"priority",
			"maxAgents",
			"requiresQA",
			"requiresSecurity",
			"requiresHumanApproval",
			"activeTemplate",
			"atomicTaskPolicy",
			"reviewMode",
			"workstreams",
		},
	}
}
