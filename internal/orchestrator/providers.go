package orchestrator

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"maliangswarm/internal/domain"
)

const productName = "maliang swarm"

func defaultAIProviders() []domain.AIProviderConfig {
	providers := []domain.AIProviderConfig{
		{
			ID:            "provider-openai-compatible",
			Name:          "OpenAI 兼容网关",
			Format:        "openai-compatible",
			BaseURL:       "https://api.openai.com/v1",
			APIPath:       "/chat/completions",
			DefaultModel:  "gpt-5.4",
			PlannerModel:  "gpt-5.4",
			WorkerModel:   "gpt-5.4-mini",
			ReviewerModel: "gpt-5.4",
			HeadersJSON:   "{}",
			Notes:         "主推荐路由，适用于主脑规划、审核 Agent 和执行 Agent。填入密钥后即可启用实时请求。",
			Enabled:       true,
			IsPrimary:     true,
			SortOrder:     1,
		},
		{
			ID:            "provider-anthropic",
			Name:          "Anthropic 消息接口",
			Format:        "anthropic",
			BaseURL:       "https://api.anthropic.com",
			APIPath:       "/v1/messages",
			APIVersion:    "2023-06-01",
			DefaultModel:  "claude-sonnet-4-5",
			PlannerModel:  "claude-opus-4-1",
			WorkerModel:   "claude-sonnet-4-5",
			ReviewerModel: "claude-sonnet-4-5",
			HeadersJSON:   "{\n  \"anthropic-version\": \"2023-06-01\"\n}",
			Notes:         "高推理备份路由，适合规划和最终审核类工作负载。",
			Enabled:       false,
			IsPrimary:     false,
			SortOrder:     2,
		},
		{
			ID:            "provider-ollama",
			Name:          "Ollama 本地路由",
			Format:        "ollama",
			BaseURL:       "http://127.0.0.1:11434",
			APIPath:       "/api/chat",
			DefaultModel:  "qwen2.5-coder:14b",
			PlannerModel:  "qwen2.5-coder:14b",
			WorkerModel:   "qwen2.5-coder:7b",
			ReviewerModel: "qwen2.5-coder:14b",
			HeadersJSON:   "{}",
			Notes:         "适合开发联调、冒烟验证和低成本批量任务的本地模型路由。",
			Enabled:       false,
			IsPrimary:     false,
			SortOrder:     3,
		},
	}

	return normalizeProviders(providers)
}

func modelProfilesForProviders(providers []domain.AIProviderConfig) []domain.ModelProfile {
	primary, ok := primaryProvider(providers)
	if !ok {
		return []domain.ModelProfile{
			{SortOrder: 1, Tier: "tier-strategic", Binding: "主脑路由待配置", Responsibility: "负责总编排、规划和关键复核"},
			{SortOrder: 2, Tier: "tier-review", Binding: "审核路由待配置", Responsibility: "负责同行评审、QA、安全与合规"},
			{SortOrder: 3, Tier: "tier-execution", Binding: "执行路由待配置", Responsibility: "负责实现、集成和交付产物"},
			{SortOrder: 4, Tier: "tier-routing", Binding: "路由层待配置", Responsibility: "负责需求接入和轻量级分类"},
		}
	}

	suffix := ""
	if strings.TrimSpace(primary.APIKey) == "" {
		suffix = "（密钥待配置）"
	}

	return []domain.ModelProfile{
		{SortOrder: 1, Tier: "tier-strategic", Binding: fmt.Sprintf("%s / %s%s", primary.Name, primary.PlannerModel, suffix), Responsibility: "负责总编排、规划和关键复核"},
		{SortOrder: 2, Tier: "tier-review", Binding: fmt.Sprintf("%s / %s%s", primary.Name, primary.ReviewerModel, suffix), Responsibility: "负责同行评审、QA、安全与合规"},
		{SortOrder: 3, Tier: "tier-execution", Binding: fmt.Sprintf("%s / %s%s", primary.Name, primary.WorkerModel, suffix), Responsibility: "负责实现、集成和交付产物"},
		{SortOrder: 4, Tier: "tier-routing", Binding: fmt.Sprintf("%s / %s%s", primary.Name, primary.DefaultModel, suffix), Responsibility: "负责需求接入和轻量级分类"},
	}
}

func primaryProvider(providers []domain.AIProviderConfig) (domain.AIProviderConfig, bool) {
	for _, provider := range providers {
		if provider.IsPrimary {
			return provider, true
		}
	}
	for _, provider := range providers {
		if provider.Enabled {
			return provider, true
		}
	}
	if len(providers) == 0 {
		return domain.AIProviderConfig{}, false
	}
	return providers[0], true
}

func primaryProviderLabel(providers []domain.AIProviderConfig) string {
	primary, ok := primaryProvider(providers)
	if !ok {
		return "尚未配置接口"
	}
	if strings.TrimSpace(primary.APIKey) == "" {
		return fmt.Sprintf("%s（%s，密钥待配置）", primary.Name, zhProviderFormat(primary.Format))
	}
	return fmt.Sprintf("%s（%s）", primary.Name, zhProviderFormat(primary.Format))
}

func enabledProviderCount(providers []domain.AIProviderConfig) int {
	count := 0
	for _, provider := range providers {
		if provider.Enabled {
			count++
		}
	}
	return count
}

func normalizeProviders(providers []domain.AIProviderConfig) []domain.AIProviderConfig {
	if len(providers) == 0 {
		return nil
	}

	slices.SortFunc(providers, func(left, right domain.AIProviderConfig) int {
		if left.SortOrder == right.SortOrder {
			return strings.Compare(left.Name, right.Name)
		}
		return left.SortOrder - right.SortOrder
	})

	primaryAssigned := false
	for index := range providers {
		provider := &providers[index]
		provider.ID = compact(provider.ID)
		provider.Name = compact(provider.Name)
		provider.Format = normalizeProviderFormat(provider.Format)
		provider.BaseURL = strings.TrimSpace(provider.BaseURL)
		provider.APIPath = strings.TrimSpace(provider.APIPath)
		provider.APIVersion = strings.TrimSpace(provider.APIVersion)
		provider.DefaultModel = compact(provider.DefaultModel)
		provider.PlannerModel = compact(provider.PlannerModel)
		provider.WorkerModel = compact(provider.WorkerModel)
		provider.ReviewerModel = compact(provider.ReviewerModel)
		provider.HeadersJSON = strings.TrimSpace(provider.HeadersJSON)
		provider.Notes = compact(provider.Notes)

		if provider.ID == "" {
			provider.ID = fmt.Sprintf("provider-%s-%d", slugify(provider.Name), index+1)
		}
		if provider.Name == "" {
			provider.Name = fmt.Sprintf("Provider %d", index+1)
		}
		if provider.BaseURL == "" {
			provider.BaseURL = defaultProviderBaseURL(provider.Format)
		}
		if provider.APIPath == "" {
			provider.APIPath = defaultProviderAPIPath(provider.Format)
		}
		if provider.APIVersion == "" {
			provider.APIVersion = defaultProviderVersion(provider.Format)
		}
		if provider.DefaultModel == "" {
			provider.DefaultModel = defaultProviderModel(provider.Format)
		}
		if provider.PlannerModel == "" {
			provider.PlannerModel = provider.DefaultModel
		}
		if provider.WorkerModel == "" {
			provider.WorkerModel = provider.DefaultModel
		}
		if provider.ReviewerModel == "" {
			provider.ReviewerModel = provider.DefaultModel
		}
		if provider.HeadersJSON == "" {
			provider.HeadersJSON = defaultProviderHeaders(provider.Format, provider.APIVersion)
		}
		if provider.SortOrder <= 0 {
			provider.SortOrder = index + 1
		}

		if provider.IsPrimary && !primaryAssigned {
			primaryAssigned = true
			continue
		}
		if provider.IsPrimary {
			provider.IsPrimary = false
		}
	}

	if !primaryAssigned {
		for index := range providers {
			if providers[index].Enabled {
				providers[index].IsPrimary = true
				primaryAssigned = true
				break
			}
		}
	}
	if !primaryAssigned && len(providers) > 0 {
		providers[0].IsPrimary = true
	}

	return providers
}

func upsertProviderConfig(providers []domain.AIProviderConfig, input AIProviderInput, now time.Time) []domain.AIProviderConfig {
	normalized := normalizeProviderInput(input)
	if normalized.ID == "" {
		normalized.ID = buildProviderID(normalized.Name, now)
	}

	replaced := false
	for index := range providers {
		if providers[index].ID != normalized.ID {
			continue
		}
		existing := providers[index]
		providers[index] = domain.AIProviderConfig{
			ID:            existing.ID,
			Name:          normalized.Name,
			Format:        normalized.Format,
			BaseURL:       normalized.BaseURL,
			APIPath:       normalized.APIPath,
			APIVersion:    normalized.APIVersion,
			APIKey:        keepProviderSecret(existing.APIKey, normalized.APIKey),
			DefaultModel:  normalized.DefaultModel,
			PlannerModel:  normalized.PlannerModel,
			WorkerModel:   normalized.WorkerModel,
			ReviewerModel: normalized.ReviewerModel,
			HeadersJSON:   normalized.HeadersJSON,
			Notes:         normalized.Notes,
			Enabled:       normalized.Enabled,
			IsPrimary:     normalized.IsPrimary,
			SortOrder:     existing.SortOrder,
		}
		replaced = true
		break
	}

	if !replaced {
		providers = append(providers, domain.AIProviderConfig{
			ID:            normalized.ID,
			Name:          normalized.Name,
			Format:        normalized.Format,
			BaseURL:       normalized.BaseURL,
			APIPath:       normalized.APIPath,
			APIVersion:    normalized.APIVersion,
			APIKey:        normalized.APIKey,
			DefaultModel:  normalized.DefaultModel,
			PlannerModel:  normalized.PlannerModel,
			WorkerModel:   normalized.WorkerModel,
			ReviewerModel: normalized.ReviewerModel,
			HeadersJSON:   normalized.HeadersJSON,
			Notes:         normalized.Notes,
			Enabled:       normalized.Enabled,
			IsPrimary:     normalized.IsPrimary,
			SortOrder:     maxProviderSortOrder(providers) + 1,
		})
	}

	return normalizeProviders(providers)
}

func deleteProviderConfig(providers []domain.AIProviderConfig, id string) []domain.AIProviderConfig {
	id = compact(id)
	if id == "" {
		return normalizeProviders(providers)
	}

	filtered := make([]domain.AIProviderConfig, 0, len(providers))
	for _, provider := range providers {
		if provider.ID == id {
			continue
		}
		filtered = append(filtered, provider)
	}

	return normalizeProviders(filtered)
}

func normalizeProviderInput(input AIProviderInput) AIProviderInput {
	input.ID = compact(input.ID)
	input.Name = compact(input.Name)
	input.Format = normalizeProviderFormat(input.Format)
	input.BaseURL = strings.TrimSpace(input.BaseURL)
	input.APIPath = normalizeProviderAPIPath(input.Format, input.APIPath)
	input.APIVersion = strings.TrimSpace(input.APIVersion)
	input.APIKey = strings.TrimSpace(input.APIKey)
	input.DefaultModel = compact(input.DefaultModel)
	input.PlannerModel = compact(input.PlannerModel)
	input.WorkerModel = compact(input.WorkerModel)
	input.ReviewerModel = compact(input.ReviewerModel)
	input.HeadersJSON = strings.TrimSpace(input.HeadersJSON)
	input.Notes = compact(input.Notes)

	if input.Name == "" {
		input.Name = "Untitled Provider"
	}
	if input.BaseURL == "" {
		input.BaseURL = defaultProviderBaseURL(input.Format)
	}
	if input.APIPath == "" {
		input.APIPath = defaultProviderAPIPath(input.Format)
	}
	if input.APIVersion == "" {
		input.APIVersion = defaultProviderVersion(input.Format)
	}
	if input.DefaultModel == "" {
		input.DefaultModel = defaultProviderModel(input.Format)
	}
	if input.PlannerModel == "" {
		input.PlannerModel = input.DefaultModel
	}
	if input.WorkerModel == "" {
		input.WorkerModel = input.DefaultModel
	}
	if input.ReviewerModel == "" {
		input.ReviewerModel = input.DefaultModel
	}
	if input.HeadersJSON == "" {
		input.HeadersJSON = defaultProviderHeaders(input.Format, input.APIVersion)
	}

	return input
}

func normalizeProviderAPIPath(format string, value string) string {
	path := strings.TrimSpace(value)
	if format != "openai-compatible" {
		return path
	}

	normalized := strings.ToLower(path)
	if normalized == "" || normalized == "/responses" || normalized == "responses" {
		return "/chat/completions"
	}

	return path
}

func providerState(provider domain.AIProviderConfig) AIProvider {
	return AIProvider{
		ID:               provider.ID,
		Name:             provider.Name,
		Format:           provider.Format,
		BaseURL:          provider.BaseURL,
		APIPath:          provider.APIPath,
		APIVersion:       provider.APIVersion,
		DefaultModel:     provider.DefaultModel,
		PlannerModel:     provider.PlannerModel,
		WorkerModel:      provider.WorkerModel,
		ReviewerModel:    provider.ReviewerModel,
		HeadersJSON:      provider.HeadersJSON,
		Notes:            provider.Notes,
		Enabled:          provider.Enabled,
		IsPrimary:        provider.IsPrimary,
		APIKeyConfigured: strings.TrimSpace(provider.APIKey) != "",
		APIKeyPreview:    maskAPIKey(provider.APIKey),
	}
}

func normalizeProviderFormat(value string) string {
	switch strings.ToLower(compact(value)) {
	case "anthropic", "gemini", "azure-openai", "ollama", "custom-http":
		return strings.ToLower(compact(value))
	default:
		return "openai-compatible"
	}
}

func defaultProviderBaseURL(format string) string {
	switch format {
	case "anthropic":
		return "https://api.anthropic.com"
	case "gemini":
		return "https://generativelanguage.googleapis.com"
	case "azure-openai":
		return "https://YOUR-RESOURCE.openai.azure.com"
	case "ollama":
		return "http://127.0.0.1:11434"
	case "custom-http":
		return "https://your-gateway.example.com"
	default:
		return "https://api.openai.com/v1"
	}
}

func defaultProviderAPIPath(format string) string {
	switch format {
	case "anthropic":
		return "/v1/messages"
	case "gemini":
		return "/v1beta/models"
	case "azure-openai":
		return "/openai/responses"
	case "ollama":
		return "/api/chat"
	case "custom-http":
		return "/v1/dispatch"
	default:
		return "/chat/completions"
	}
}

func defaultProviderVersion(format string) string {
	switch format {
	case "anthropic":
		return "2023-06-01"
	case "gemini":
		return "v1beta"
	case "azure-openai":
		return "2024-10-21"
	default:
		return ""
	}
}

func defaultProviderHeaders(format, apiVersion string) string {
	switch format {
	case "anthropic":
		version := apiVersion
		if version == "" {
			version = "2023-06-01"
		}
		return fmt.Sprintf("{\n  \"anthropic-version\": %q\n}", version)
	case "custom-http":
		return "{\n  \"content-type\": \"application/json\"\n}"
	default:
		return "{}"
	}
}

func defaultProviderModel(format string) string {
	switch format {
	case "anthropic":
		return "claude-sonnet-4-5"
	case "gemini":
		return "gemini-2.5-pro"
	case "azure-openai", "openai-compatible":
		return "gpt-5.4"
	case "ollama":
		return "qwen2.5-coder:14b"
	default:
		return "custom-model"
	}
}

func buildProviderID(name string, now time.Time) string {
	return fmt.Sprintf("provider-%s-%d", slugify(name), now.Unix())
}

func maxProviderSortOrder(providers []domain.AIProviderConfig) int {
	order := 0
	for _, provider := range providers {
		order = max(order, provider.SortOrder)
	}
	return order
}

func keepProviderSecret(existing, replacement string) string {
	if strings.TrimSpace(replacement) == "" {
		return existing
	}
	return strings.TrimSpace(replacement)
}

func maskAPIKey(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "Not configured"
	}
	if len(secret) <= 8 {
		return strings.Repeat("*", len(secret))
	}
	return fmt.Sprintf("%s...%s", secret[:4], secret[len(secret)-4:])
}
