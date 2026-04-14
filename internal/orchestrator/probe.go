package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"maliangswarm/internal/domain"
)

func previewProviderConfig(providers []domain.AIProviderConfig, input AIProviderInput, now time.Time) domain.AIProviderConfig {
	normalized := normalizeProviderInput(input)
	if normalized.ID == "" {
		normalized.ID = buildProviderID(normalized.Name, now)
	}

	sortOrder := maxProviderSortOrder(providers) + 1
	for _, provider := range providers {
		if provider.ID != normalized.ID {
			continue
		}
		normalized.APIKey = keepProviderSecret(provider.APIKey, normalized.APIKey)
		sortOrder = provider.SortOrder
		break
	}

	return domain.AIProviderConfig{
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
		Enabled:       true,
		IsPrimary:     normalized.IsPrimary,
		SortOrder:     sortOrder,
	}
}

func probeProvider(ctx context.Context, provider domain.AIProviderConfig) string {
	address := strings.TrimSpace(provider.BaseURL)
	if targetURL, err := buildProviderURL(provider, provider.APIPath); err == nil {
		address = targetURL
	} else if path := strings.TrimSpace(provider.APIPath); path != "" {
		address += path
	}

	lines := []string{
		fmt.Sprintf("接口：%s（%s）", provider.Name, zhProviderFormat(provider.Format)),
		fmt.Sprintf("地址：%s%s", provider.BaseURL, provider.APIPath),
	}
	lines[1] = fmt.Sprintf("地址：%s", address)

	if compact(provider.BaseURL) == "" {
		lines = append(lines, "总结果：失败。缺少基础 URL。")
		return strings.Join(lines, "\n")
	}
	if providerRequiresKey(provider) && compact(provider.APIKey) == "" {
		lines = append(lines, "总结果：失败。当前接口格式需要 API Key 或等效密钥。")
		return strings.Join(lines, "\n")
	}

	snapshot := domain.Snapshot{
		Run: domain.Run{
			Title:            "接口连通性测试",
			Mission:          "验证主脑、执行、审核三条模型路由",
			Deliverable:      "接口测试报告",
			TaskType:         "platform",
			Priority:         "medium",
			ActiveTemplate:   "接口探测模板",
			AtomicTaskPolicy: "每条链路只生成一个探测产物",
			ReviewMode:       "仅用于操作者验证",
		},
	}

	lanes := []struct {
		Label string
		Lane  string
		Model string
		Task  domain.Task
	}{
		{
			Label: "主脑",
			Lane:  "planning",
			Model: provider.PlannerModel,
			Task: domain.Task{
				ID:           "probe-planner",
				Title:        "生成编排蓝图预览",
				OwnerRole:    "Chief Orchestrator",
				ArtifactName: "probe-blueprint.json",
				Detail:       "概述该接口如何规划一个企业级多 Agent 运行。",
			},
		},
		{
			Label: "执行",
			Lane:  "execution",
			Model: provider.WorkerModel,
			Task: domain.Task{
				ID:           "probe-worker",
				Title:        "生成执行产物预览",
				OwnerRole:    "Implementation Agent",
				ArtifactName: "probe-worker-artifact.md",
				Detail:       "生成一段精简的执行产物预览。",
			},
		},
		{
			Label: "审核",
			Lane:  "review",
			Model: provider.ReviewerModel,
			Task: domain.Task{
				ID:           "probe-reviewer",
				Title:        "审核生成产物",
				OwnerRole:    "Peer Reviewer",
				ArtifactName: "probe-review-note.md",
				Detail:       "生成一段带企业门禁风格的审核意见。",
			},
		},
	}

	activeLaneCount := 0
	successCount := 0
	for _, lane := range lanes {
		if compact(lane.Model) == "" {
			lines = append(lines, fmt.Sprintf("%s：已跳过，未配置模型。", lane.Label))
			continue
		}
		activeLaneCount++

		route := providerRoute{
			Provider: provider,
			Model:    lane.Model,
			Source:   fmt.Sprintf("%s / %s", provider.Name, lane.Model),
		}

		laneCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		text, err := requestProviderText(laneCtx, route, snapshot, lane.Task, lane.Lane)
		cancel()
		if err != nil {
			lines = append(lines, fmt.Sprintf("%s：失败，模型 %s 未返回可用结果。", lane.Label, lane.Model))
			lines = append(lines, fmt.Sprintf("原因：%s", compact(err.Error())))
			continue
		}

		successCount++
		preview := compact(text)
		if len(preview) > 220 {
			preview = preview[:217] + "..."
		}
		lines = append(lines, fmt.Sprintf("%s：通过，模型 %s 已返回结果。", lane.Label, lane.Model))
		lines = append(lines, fmt.Sprintf("预览：%s", preview))
	}

	switch {
	case activeLaneCount == 0:
		lines = append(lines, "总结果：失败。主脑、执行、审核三个模型都尚未配置。")
	case successCount == activeLaneCount:
		lines = append(lines, fmt.Sprintf("总结果：成功。%d/%d 条已配置链路都返回了可用文本。", successCount, activeLaneCount))
	case successCount > 0:
		lines = append(lines, fmt.Sprintf("总结果：部分成功。%d/%d 条已配置链路返回了可用文本。", successCount, activeLaneCount))
	default:
		lines = append(lines, fmt.Sprintf("总结果：失败。%d 条已配置链路中没有一条返回可用文本。", activeLaneCount))
	}

	return strings.Join(lines, "\n")
}

func providerRequiresKey(provider domain.AIProviderConfig) bool {
	switch provider.Format {
	case "ollama", "custom-http":
		return false
	default:
		return true
	}
}
