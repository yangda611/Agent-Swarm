package orchestrator

import (
	"testing"

	"maliangswarm/internal/domain"
)

func TestBuildProviderURLAvoidsDuplicatingBasePrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider domain.AIProviderConfig
		path     string
		want     string
	}{
		{
			name: "openai compatible base already includes version prefix",
			provider: domain.AIProviderConfig{
				BaseURL: "https://openclawroot.com/v1",
			},
			path: "/v1/responses",
			want: "https://openclawroot.com/v1/responses",
		},
		{
			name: "base with version prefix joins suffix path once",
			provider: domain.AIProviderConfig{
				BaseURL: "https://openclawroot.com/v1",
			},
			path: "/responses",
			want: "https://openclawroot.com/v1/responses",
		},
		{
			name: "plain base still supports legacy full path",
			provider: domain.AIProviderConfig{
				BaseURL: "https://api.openai.com",
			},
			path: "/v1/responses",
			want: "https://api.openai.com/v1/responses",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := buildProviderURL(tc.provider, tc.path)
			if err != nil {
				t.Fatalf("buildProviderURL() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("buildProviderURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeProviderInputLaneModels(t *testing.T) {
	t.Parallel()

	t.Run("keeps explicitly configured lane models", func(t *testing.T) {
		t.Parallel()

		got := normalizeProviderInput(AIProviderInput{
			Format:        "openai-compatible",
			BaseURL:       "https://openclawroot.com/v1",
			APIPath:       "/responses",
			DefaultModel:  "qwen3-max",
			PlannerModel:  "gpt-5.4",
			WorkerModel:   "gpt-5.4-mini",
			ReviewerModel: "gpt-5.4",
		})

		if got.PlannerModel != "gpt-5.4" {
			t.Fatalf("PlannerModel = %q, want %q", got.PlannerModel, "gpt-5.4")
		}
		if got.WorkerModel != "gpt-5.4-mini" {
			t.Fatalf("WorkerModel = %q, want %q", got.WorkerModel, "gpt-5.4-mini")
		}
		if got.ReviewerModel != "gpt-5.4" {
			t.Fatalf("ReviewerModel = %q, want %q", got.ReviewerModel, "gpt-5.4")
		}
	})

	t.Run("blank lane models inherit the default model", func(t *testing.T) {
		t.Parallel()

		got := normalizeProviderInput(AIProviderInput{
			Format:       "openai-compatible",
			BaseURL:      "https://openclawroot.com/v1",
			APIPath:      "/responses",
			DefaultModel: "qwen3-max",
		})

		if got.PlannerModel != "qwen3-max" || got.WorkerModel != "qwen3-max" || got.ReviewerModel != "qwen3-max" {
			t.Fatalf("lane models should inherit default model, got planner=%q worker=%q reviewer=%q", got.PlannerModel, got.WorkerModel, got.ReviewerModel)
		}
	})
}
