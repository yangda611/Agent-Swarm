package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"maliangswarm/internal/domain"
)

type providerRoute struct {
	Provider domain.AIProviderConfig
	Model    string
	Source   string
}

func advanceRunSnapshot(ctx context.Context, snapshot domain.Snapshot, nextStep int, now time.Time) domain.Snapshot {
	updated := applyProgression(snapshot, nextStep, now)

	switch nextStep {
	case 1:
		updated = executePlanningStage(ctx, updated)
	case 2:
		updated = executeExecutionStage(ctx, updated)
	case 3:
		updated = executeReviewStage(ctx, updated)
	case 4:
		updated = executeGovernanceStage(ctx, updated)
	case 5:
		updated = executeDeliveryStage(ctx, updated)
	}

	updated.Agents = hydrateAgentArtifacts(updated.Agents, updated.Tasks, updated.Handoffs)
	updated.Timeline = decorateTimeline(updated, nextStep, now)
	return updated
}

func clearRunOutputs(snapshot domain.Snapshot, now time.Time) domain.Snapshot {
	for index := range snapshot.Tasks {
		snapshot.Tasks[index].OutputSummary = ""
		snapshot.Tasks[index].ArtifactPath = ""
	}
	snapshot = applyProgression(snapshot, 0, now)
	snapshot.Run.PlannerSource = pickBlueprintString(snapshot.Run.PlannerSource, "rules fallback")
	return snapshot
}

func executePlanningStage(ctx context.Context, snapshot domain.Snapshot) domain.Snapshot {
	for index := range snapshot.Tasks {
		task := &snapshot.Tasks[index]
		if task.Stage != "intake" && task.Stage != "planning" {
			continue
		}
		if task.OutputSummary != "" {
			continue
		}
		task.OutputSummary = generateTaskSummary(ctx, snapshot, *task, "planning")
	}
	return snapshot
}

func executeExecutionStage(ctx context.Context, snapshot domain.Snapshot) domain.Snapshot {
	for index := range snapshot.Tasks {
		task := &snapshot.Tasks[index]
		if task.Stage != "execution" {
			continue
		}
		if task.OutputSummary != "" {
			continue
		}
		task.OutputSummary = generateTaskSummary(ctx, snapshot, *task, "execution")
	}
	return snapshot
}

func executeReviewStage(ctx context.Context, snapshot domain.Snapshot) domain.Snapshot {
	for index := range snapshot.Tasks {
		task := &snapshot.Tasks[index]
		if task.Stage != "review" {
			continue
		}
		if task.ID == "task-review-security" || task.ID == "task-review-approval" {
			continue
		}
		if task.OutputSummary != "" {
			continue
		}
		task.OutputSummary = generateTaskSummary(ctx, snapshot, *task, "review")
	}
	return snapshot
}

func executeGovernanceStage(ctx context.Context, snapshot domain.Snapshot) domain.Snapshot {
	for index := range snapshot.Tasks {
		task := &snapshot.Tasks[index]
		if task.ID != "task-review-security" && task.ID != "task-review-approval" {
			continue
		}
		task.OutputSummary = generateTaskSummary(ctx, snapshot, *task, "governance")
	}
	return snapshot
}

func executeDeliveryStage(ctx context.Context, snapshot domain.Snapshot) domain.Snapshot {
	for index := range snapshot.Tasks {
		task := &snapshot.Tasks[index]
		if task.Stage != "delivery" {
			continue
		}
		task.OutputSummary = generateTaskSummary(ctx, snapshot, *task, "delivery")
	}
	return snapshot
}

func generateTaskSummary(ctx context.Context, snapshot domain.Snapshot, task domain.Task, lane string) string {
	route, ok := selectRoute(snapshot.AIProviders, lane)
	if ok {
		text, err := requestProviderText(ctx, route, snapshot, task, lane)
		if err == nil && compact(text) != "" {
			return compact(text)
		}
	}
	return fallbackTaskSummary(snapshot, task, lane)
}

func selectRoute(providers []domain.AIProviderConfig, lane string) (providerRoute, bool) {
	primary, ok := selectPlanningProvider(providers)
	if !ok {
		return providerRoute{}, false
	}

	model := primary.PlannerModel
	switch lane {
	case "execution":
		model = primary.WorkerModel
	case "review", "governance":
		model = primary.ReviewerModel
	case "delivery":
		if primary.PlannerModel != "" {
			model = primary.PlannerModel
		}
	}

	if compact(model) == "" {
		return providerRoute{}, false
	}

	return providerRoute{
		Provider: primary,
		Model:    model,
		Source:   fmt.Sprintf("%s / %s", primary.Name, model),
	}, true
}

func requestProviderText(ctx context.Context, route providerRoute, snapshot domain.Snapshot, task domain.Task, lane string) (string, error) {
	systemPrompt, userPrompt := buildExecutionPrompts(snapshot, task, lane)
	switch route.Provider.Format {
	case "anthropic":
		return requestAnthropicText(ctx, route, systemPrompt, userPrompt)
	case "gemini":
		return requestGeminiText(ctx, route, systemPrompt, userPrompt)
	case "azure-openai":
		return requestAzureOpenAIText(ctx, route, systemPrompt, userPrompt)
	case "ollama":
		return requestOllamaText(ctx, route, systemPrompt, userPrompt)
	case "custom-http":
		return requestCustomHTTPText(ctx, route, systemPrompt, userPrompt)
	default:
		return requestOpenAICompatibleText(ctx, route, systemPrompt, userPrompt)
	}
}

func buildExecutionPrompts(snapshot domain.Snapshot, task domain.Task, lane string) (string, string) {
	systemPrompt := "你是 maliang swarm 中的企业级 Agent。请输出简洁、具体、面向操作者的中文产物内容，只返回纯文本。"
	userPrompt := fmt.Sprintf(`运行标题：%s
任务目标：%s
目标交付物：%s
任务类型：%s
优先级：%s
流程模板：%s
原子任务策略：%s
审核模式：%s
当前链路：%s
任务标题：%s
任务负责角色：%s
产物名称：%s
任务说明：%s

请输出这个 Agent 针对当前产物会写出的内容。
内容要简洁但具体，适合展示在抽屉面板或任务板预览中。`,
		snapshot.Run.Title,
		snapshot.Run.Mission,
		snapshot.Run.Deliverable,
		snapshot.Run.TaskType,
		snapshot.Run.Priority,
		snapshot.Run.ActiveTemplate,
		snapshot.Run.AtomicTaskPolicy,
		snapshot.Run.ReviewMode,
		lane,
		task.Title,
		task.OwnerRole,
		task.ArtifactName,
		task.Detail,
	)
	return systemPrompt, userPrompt
}

func requestOpenAICompatibleText(ctx context.Context, route providerRoute, systemPrompt string, userPrompt string) (string, error) {
	targetURL, err := buildProviderURL(route.Provider, route.Provider.APIPath)
	if err != nil {
		return "", err
	}

	var payload map[string]any
	if strings.Contains(strings.ToLower(route.Provider.APIPath), "chat/completions") {
		payload = map[string]any{
			"model": route.Model,
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": userPrompt},
			},
			"temperature": 0.2,
		}
	} else {
		payload = map[string]any{
			"model": route.Model,
			"input": []map[string]any{
				{"role": "system", "content": []map[string]string{{"type": "input_text", "text": systemPrompt}}},
				{"role": "user", "content": []map[string]string{{"type": "input_text", "text": userPrompt}}},
			},
			"temperature": 0.2,
		}
	}

	body, err := postProviderJSON(ctx, route.Provider, targetURL, payload)
	if err != nil {
		return "", err
	}
	text := locatePlannerTextFromBody(body)
	if compact(text) == "" {
		return "", fmt.Errorf("empty text response")
	}
	return text, nil
}

func requestAnthropicText(ctx context.Context, route providerRoute, systemPrompt string, userPrompt string) (string, error) {
	targetURL, err := buildProviderURL(route.Provider, route.Provider.APIPath)
	if err != nil {
		return "", err
	}
	payload := map[string]any{
		"model":       route.Model,
		"max_tokens":  1800,
		"temperature": 0.2,
		"system":      systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}
	body, err := postProviderJSON(ctx, route.Provider, targetURL, payload)
	if err != nil {
		return "", err
	}
	text := locatePlannerTextFromBody(body)
	if compact(text) == "" {
		return "", fmt.Errorf("empty text response")
	}
	return text, nil
}

func requestGeminiText(ctx context.Context, route providerRoute, systemPrompt string, userPrompt string) (string, error) {
	targetURL, err := buildGeminiURL(domain.AIProviderConfig{
		BaseURL:      route.Provider.BaseURL,
		APIPath:      route.Provider.APIPath,
		PlannerModel: route.Model,
	})
	if err != nil {
		return "", err
	}
	payload := map[string]any{
		"system_instruction": map[string]any{
			"parts": []map[string]string{{"text": systemPrompt}},
		},
		"contents": []map[string]any{
			{"parts": []map[string]string{{"text": userPrompt}}},
		},
		"generationConfig": map[string]any{
			"temperature": 0.2,
		},
	}
	body, err := postProviderJSON(ctx, route.Provider, targetURL, payload)
	if err != nil {
		return "", err
	}
	text := locatePlannerTextFromBody(body)
	if compact(text) == "" {
		return "", fmt.Errorf("empty text response")
	}
	return text, nil
}

func requestAzureOpenAIText(ctx context.Context, route providerRoute, systemPrompt string, userPrompt string) (string, error) {
	targetURL, err := buildProviderURL(route.Provider, route.Provider.APIPath)
	if err != nil {
		return "", err
	}
	if route.Provider.APIVersion != "" && !strings.Contains(targetURL, "api-version=") {
		targetURL, err = addQueryValue(targetURL, "api-version", route.Provider.APIVersion)
		if err != nil {
			return "", err
		}
	}

	payload := map[string]any{
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
	}
	if !strings.Contains(strings.ToLower(route.Provider.APIPath), "/deployments/") {
		payload["model"] = route.Model
	}

	body, err := postProviderJSON(ctx, route.Provider, targetURL, payload)
	if err != nil {
		return "", err
	}
	text := locatePlannerTextFromBody(body)
	if compact(text) == "" {
		return "", fmt.Errorf("empty text response")
	}
	return text, nil
}

func requestOllamaText(ctx context.Context, route providerRoute, systemPrompt string, userPrompt string) (string, error) {
	targetURL, err := buildProviderURL(route.Provider, route.Provider.APIPath)
	if err != nil {
		return "", err
	}
	payload := map[string]any{
		"model":  route.Model,
		"stream": false,
	}
	if strings.Contains(strings.ToLower(route.Provider.APIPath), "/api/generate") {
		payload["prompt"] = systemPrompt + "\n\n" + userPrompt
	} else {
		payload["messages"] = []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		}
	}
	body, err := postProviderJSON(ctx, route.Provider, targetURL, payload)
	if err != nil {
		return "", err
	}
	text := locatePlannerTextFromBody(body)
	if compact(text) == "" {
		return "", fmt.Errorf("empty text response")
	}
	return text, nil
}

func requestCustomHTTPText(ctx context.Context, route providerRoute, systemPrompt string, userPrompt string) (string, error) {
	targetURL, err := buildProviderURL(route.Provider, route.Provider.APIPath)
	if err != nil {
		return "", err
	}
	payload := map[string]any{
		"model":       route.Model,
		"system":      systemPrompt,
		"input":       userPrompt,
		"temperature": 0.2,
	}
	body, err := postProviderJSON(ctx, route.Provider, targetURL, payload)
	if err != nil {
		return "", err
	}
	text := locatePlannerTextFromBody(body)
	if compact(text) == "" {
		return "", fmt.Errorf("empty text response")
	}
	return text, nil
}

func locatePlannerTextFromBody(body []byte) string {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return string(body)
	}
	return locatePlannerText(payload)
}

func fallbackTaskSummary(snapshot domain.Snapshot, task domain.Task, lane string) string {
	switch lane {
	case "planning":
		return fmt.Sprintf("%s 已为 %s 产出 %s，当前方案遵循 %s，并保持审核链路清晰可见。", task.OwnerRole, snapshot.Run.Title, task.ArtifactName, snapshot.Run.ActiveTemplate)
	case "execution":
		return fmt.Sprintf("%s 已完成 %s，对应内容覆盖：%s。该产物已可送入同行评审和下游交接。", task.OwnerRole, task.ArtifactName, task.Detail)
	case "review":
		return fmt.Sprintf("%s 已审核执行产物包，并为 %s 记录了复核结论，责任边界和交付准备度均已检查。", task.OwnerRole, snapshot.Run.Title)
	case "governance":
		if task.ID == "task-review-approval" {
			return fmt.Sprintf("%s 已创建 %s，并在正式放行前把当前运行切换到明确审批跟踪状态。", task.OwnerRole, task.ArtifactName)
		}
		return fmt.Sprintf("%s 已在 %s 中完成治理分析，并记录了当前外发链路的风险姿态。", task.OwnerRole, task.ArtifactName)
	default:
		return fmt.Sprintf("%s 已为 %s 整理 %s，并准备好最终企业级交付包。", task.OwnerRole, snapshot.Run.Deliverable, task.ArtifactName)
	}
}

func hydrateAgentArtifacts(agents []domain.Agent, tasks []domain.Task, handoffs []domain.Handoff) []domain.Agent {
	updated := make([]domain.Agent, 0, len(agents))
	agentLabels := make(map[string]string, len(agents))
	for _, agent := range agents {
		agentLabels[agent.ID] = fmt.Sprintf("%s / %s", agent.Name, agent.Role)
	}

	for _, agent := range agents {
		next := agent
		next.LastAction = ""
		next.LastOutput = ""
		next.LastHandoff = ""
		next.LastReceiver = ""
		next.LastArtifactPath = ""

		owned := completedTasksOwnedBy(tasks, agent.ID)
		if len(owned) > 0 {
			latest := owned[len(owned)-1]
			if compact(latest.ArtifactName) != "" {
				next.Artifact = latest.ArtifactName
			}
			if compact(latest.Title) != "" {
				next.LastAction = latest.Title
			}
			if compact(latest.OutputSummary) != "" {
				next.LastOutput = latest.OutputSummary
			}
			if compact(latest.ArtifactPath) != "" {
				next.LastArtifactPath = latest.ArtifactPath
			}
			if compact(latest.OutputSummary) != "" && (next.Status == "done" || next.Status == "reviewing" || next.Status == "typing" || next.Status == "blocked") {
				next.Detail = latest.OutputSummary
			}
		}

		if handoff, ok := latestHandoffBySender(handoffs, agent.ID); ok {
			next.LastHandoff = handoff.ArtifactName
			target, exists := agentLabels[handoff.ToAgentID]
			if exists {
				next.LastReceiver = target
			} else {
				next.LastReceiver = handoff.ToAgentID
			}
		}

		if compact(next.LastAction) == "" {
			next.LastAction = next.CurrentTask
		}
		if compact(next.LastReceiver) == "" {
			next.LastReceiver = next.NextTarget
		}
		updated = append(updated, next)
	}
	return updated
}

func latestHandoffBySender(handoffs []domain.Handoff, senderID string) (domain.Handoff, bool) {
	var candidate domain.Handoff
	found := false
	for _, handoff := range handoffs {
		if handoff.FromAgentID != senderID {
			continue
		}
		if !found || handoff.SortOrder >= candidate.SortOrder {
			candidate = handoff
			found = true
		}
	}
	return candidate, found
}

func completedTasksOwnedBy(tasks []domain.Task, agentID string) []domain.Task {
	owned := make([]domain.Task, 0)
	for _, task := range tasks {
		if task.OwnerAgentID == agentID && compact(task.OutputSummary) != "" {
			owned = append(owned, task)
		}
	}
	return owned
}

func decorateTimeline(snapshot domain.Snapshot, step int, now time.Time) []domain.TimelineEvent {
	timeline := progressedTimeline(snapshot, step, now)
	liveTasks := make([]domain.Task, 0)
	for _, task := range snapshot.Tasks {
		if task.OutputSummary != "" {
			liveTasks = append(liveTasks, task)
		}
	}
	if len(liveTasks) == 0 {
		return timeline
	}

	latest := liveTasks[len(liveTasks)-1]
	return append(timeline, domain.TimelineEvent{
		SortOrder: len(timeline) + 1,
		TimeLabel: stamp(now, 20*time.Second),
		Kind:      "artifact",
		Title:     fmt.Sprintf("%s 更新了 %s", latest.OwnerRole, latest.ArtifactName),
		Detail:    latest.OutputSummary,
	})
}
