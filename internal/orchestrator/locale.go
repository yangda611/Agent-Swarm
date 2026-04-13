package orchestrator

import (
	"fmt"
	"strings"
)

func zhTaskType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "engineering":
		return "工程研发"
	case "product":
		return "产品设计"
	case "research":
		return "研究分析"
	case "content":
		return "内容生产"
	default:
		return "综合任务"
	}
}

func zhPriority(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical":
		return "紧急"
	case "high":
		return "高"
	case "medium":
		return "中"
	case "low":
		return "低"
	default:
		return value
	}
}

func zhPlannerSource(value string) string {
	value = strings.TrimSpace(value)
	switch {
	case value == "":
		return ""
	case value == "rules fallback":
		return "规则回退"
	case value == "rules fallback / no live planner route":
		return "规则回退 / 未找到可用主脑路由"
	case strings.HasPrefix(value, "rules fallback / "):
		return "规则回退 / " + strings.TrimPrefix(value, "rules fallback / ")
	case strings.HasPrefix(value, "live via "):
		return "实时规划 / " + strings.TrimPrefix(value, "live via ")
	default:
		return value
	}
}

func zhProviderFormat(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "openai-compatible":
		return "OpenAI 兼容"
	case "anthropic":
		return "Anthropic"
	case "gemini":
		return "Gemini"
	case "azure-openai":
		return "Azure OpenAI"
	case "ollama":
		return "Ollama"
	case "custom-http":
		return "自定义 HTTP"
	default:
		return value
	}
}

func localizedTemplateName(taskType string) string {
	return fmt.Sprintf("%s企业流程", zhTaskType(taskType))
}
