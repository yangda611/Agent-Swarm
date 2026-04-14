package orchestrator

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"maliangswarm/internal/eventbus"
)

func TestCreateRunCanReplaceExistingSnapshot(t *testing.T) {
	t.Parallel()

	service, err := NewService(t.TempDir(), eventbus.New())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	defer func() {
		if closeErr := service.Close(); closeErr != nil {
			t.Fatalf("service.Close() error = %v", closeErr)
		}
	}()

	firstInput := RunCreationInput{
		Title:                 "first run",
		Mission:               "create the first engineering mission for regression coverage",
		Deliverable:           "first delivery bundle",
		TaskType:              "engineering",
		Priority:              "high",
		MaxAgents:             12,
		RequiresQA:            true,
		RequiresSecurity:      true,
		RequiresHumanApproval: true,
	}

	secondInput := RunCreationInput{
		Title:                 "second run",
		Mission:               "create the second engineering mission and overwrite the previous snapshot cleanly",
		Deliverable:           "second delivery bundle",
		TaskType:              "engineering",
		Priority:              "medium",
		MaxAgents:             8,
		RequiresQA:            true,
		RequiresSecurity:      false,
		RequiresHumanApproval: false,
	}

	first, err := service.CreateRun(context.Background(), firstInput)
	if err != nil {
		t.Fatalf("first CreateRun() error = %v", err)
	}
	if first.Title != firstInput.Title {
		t.Fatalf("first CreateRun() title = %q, want %q", first.Title, firstInput.Title)
	}

	second, err := service.CreateRun(context.Background(), secondInput)
	if err != nil {
		t.Fatalf("second CreateRun() error = %v", err)
	}
	if second.Title != secondInput.Title {
		t.Fatalf("second CreateRun() title = %q, want %q", second.Title, secondInput.Title)
	}
	if len(second.Agents) == 0 {
		t.Fatalf("second CreateRun() produced no agents")
	}
}

func TestBlankWorkspaceCanPersistSettingsAndProvidersWithoutRun(t *testing.T) {
	t.Parallel()

	service, err := NewService(t.TempDir(), eventbus.New())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	defer func() {
		if closeErr := service.Close(); closeErr != nil {
			t.Fatalf("service.Close() error = %v", closeErr)
		}
	}()

	ctx := context.Background()

	initial, err := service.GetDashboardState(ctx)
	if err != nil {
		t.Fatalf("GetDashboardState() error = %v", err)
	}
	if initial.HasActiveRun {
		t.Fatalf("GetDashboardState() unexpectedly reported an active run")
	}
	if len(initial.Tasks) != 0 || len(initial.Agents) != 0 {
		t.Fatalf("blank workspace should not preload tasks or agents")
	}
	if len(initial.Settings.AIProviders) != 0 {
		t.Fatalf("blank workspace should not preload AI providers")
	}

	updatedSettings, err := service.UpdateRuntimeSettings(ctx, SettingsInput{
		ConcurrencyLimit: 9,
		ApprovalPolicy:   "human approval required for risky actions",
		Theme:            "pixel-hq-mint",
		BudgetMode:       "balanced",
	})
	if err != nil {
		t.Fatalf("UpdateRuntimeSettings() error = %v", err)
	}
	if updatedSettings.HasActiveRun {
		t.Fatalf("UpdateRuntimeSettings() should not create a run")
	}
	if updatedSettings.Settings.ConcurrencyLimit != 9 {
		t.Fatalf("UpdateRuntimeSettings() concurrency = %d, want 9", updatedSettings.Settings.ConcurrencyLimit)
	}

	providerState, err := service.UpsertAIProvider(ctx, AIProviderInput{
		Name:          "Test Custom Provider",
		Format:        "custom-http",
		BaseURL:       "http://127.0.0.1:8080",
		APIPath:       "/v1/dispatch",
		DefaultModel:  "test-model",
		PlannerModel:  "test-model",
		WorkerModel:   "test-model",
		ReviewerModel: "test-model",
		HeadersJSON:   "{}",
		Enabled:       true,
		IsPrimary:     true,
	})
	if err != nil {
		t.Fatalf("UpsertAIProvider() error = %v", err)
	}
	if providerState.HasActiveRun {
		t.Fatalf("UpsertAIProvider() should not create a run")
	}
	if len(providerState.Settings.AIProviders) != 1 {
		t.Fatalf("UpsertAIProvider() provider count = %d, want 1", len(providerState.Settings.AIProviders))
	}

	reloaded, err := service.GetDashboardState(ctx)
	if err != nil {
		t.Fatalf("GetDashboardState() reload error = %v", err)
	}
	if reloaded.HasActiveRun {
		t.Fatalf("reloaded workspace unexpectedly reported an active run")
	}
	if reloaded.Settings.ConcurrencyLimit != 9 {
		t.Fatalf("reloaded concurrency = %d, want 9", reloaded.Settings.ConcurrencyLimit)
	}
	if len(reloaded.Settings.AIProviders) != 1 {
		t.Fatalf("reloaded provider count = %d, want 1", len(reloaded.Settings.AIProviders))
	}
}

func TestServiceMethodsAreConnectedInOneChain(t *testing.T) {
	service, err := NewService(t.TempDir(), eventbus.New())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	defer func() {
		if closeErr := service.Close(); closeErr != nil {
			t.Fatalf("service.Close() error = %v", closeErr)
		}
	}()

	ctx := context.Background()

	state, err := service.GetDashboardState(ctx)
	if err != nil {
		t.Fatalf("GetDashboardState() error = %v", err)
	}
	if state.HasActiveRun {
		t.Fatalf("GetDashboardState() unexpectedly reported an active run")
	}
	if len(state.Settings.AIProviders) != 0 {
		t.Fatalf("GetDashboardState() unexpectedly preloaded AI providers")
	}

	created, err := service.CreateRun(ctx, RunCreationInput{
		Title:                 "end-to-end chain run",
		Mission:               "connect all service methods in one functional chain",
		Deliverable:           "chain verification bundle",
		TaskType:              "engineering",
		Priority:              "high",
		MaxAgents:             10,
		RequiresQA:            true,
		RequiresSecurity:      true,
		RequiresHumanApproval: false,
	})
	if err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}
	if created.RunStep < 1 {
		t.Fatalf("CreateRun() step = %d, want >= 1", created.RunStep)
	}

	updatedSettings, err := service.UpdateRuntimeSettings(ctx, SettingsInput{
		ConcurrencyLimit: 9,
		ApprovalPolicy:   "高风险动作需人工审批，普通动作双人复核",
		Theme:            "pixel-hq-mint",
		BudgetMode:       "balanced",
	})
	if err != nil {
		t.Fatalf("UpdateRuntimeSettings() error = %v", err)
	}
	if updatedSettings.Settings.ConcurrencyLimit != 9 {
		t.Fatalf("UpdateRuntimeSettings() concurrency = %d, want 9", updatedSettings.Settings.ConcurrencyLimit)
	}

	providerState, err := service.UpsertAIProvider(ctx, AIProviderInput{
		Name:          "Test Custom Provider",
		Format:        "custom-http",
		BaseURL:       "",
		APIPath:       "/v1/dispatch",
		DefaultModel:  "test-model",
		PlannerModel:  "test-model",
		WorkerModel:   "test-model",
		ReviewerModel: "test-model",
		HeadersJSON:   "{}",
		Enabled:       true,
		IsPrimary:     false,
	})
	if err != nil {
		t.Fatalf("UpsertAIProvider() error = %v", err)
	}
	if len(providerState.Settings.AIProviders) == 0 {
		t.Fatalf("UpsertAIProvider() returned empty provider list")
	}

	lastProvider := providerState.Settings.AIProviders[len(providerState.Settings.AIProviders)-1]
	report, err := service.ProbeAIProvider(ctx, AIProviderInput{
		ID:            lastProvider.ID,
		Name:          lastProvider.Name,
		Format:        "custom-http",
		BaseURL:       "",
		APIPath:       "/v1/dispatch",
		DefaultModel:  "test-model",
		PlannerModel:  "test-model",
		WorkerModel:   "test-model",
		ReviewerModel: "test-model",
		HeadersJSON:   "{}",
		Enabled:       true,
		IsPrimary:     false,
	})
	if err != nil {
		t.Fatalf("ProbeAIProvider() error = %v", err)
	}
	if report == "" {
		t.Fatalf("ProbeAIProvider() returned empty report")
	}

	advanced, err := service.AdvanceDemoRun(ctx)
	if err != nil {
		t.Fatalf("AdvanceDemoRun() error = %v", err)
	}
	if advanced.RunStep <= created.RunStep {
		t.Fatalf("AdvanceDemoRun() step = %d, want > %d", advanced.RunStep, created.RunStep)
	}

	reset, err := service.ResetDemoRun(ctx)
	if err != nil {
		t.Fatalf("ResetDemoRun() error = %v", err)
	}
	if reset.RunStep != 0 {
		t.Fatalf("ResetDemoRun() step = %d, want 0", reset.RunStep)
	}

	deleted, err := service.DeleteAIProvider(ctx, lastProvider.ID)
	if err != nil {
		t.Fatalf("DeleteAIProvider() error = %v", err)
	}
	for _, provider := range deleted.Settings.AIProviders {
		if provider.ID == lastProvider.ID {
			t.Fatalf("DeleteAIProvider() provider %q still exists", lastProvider.ID)
		}
	}
}

func TestAdvanceRunPersistsArtifactsAndTrace(t *testing.T) {
	t.Parallel()

	service, err := NewService(t.TempDir(), eventbus.New())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	defer func() {
		if closeErr := service.Close(); closeErr != nil {
			t.Fatalf("service.Close() error = %v", closeErr)
		}
	}()

	created, err := service.CreateRun(context.Background(), RunCreationInput{
		Title:                 "artifact run",
		Mission:               "build docs and keep handoff trace",
		Deliverable:           "docs package",
		TaskType:              "engineering",
		Priority:              "high",
		MaxAgents:             10,
		RequiresQA:            true,
		RequiresSecurity:      true,
		RequiresHumanApproval: true,
	})
	if err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}

	foundArtifactPath := ""
	for _, task := range created.Tasks {
		if strings.TrimSpace(task.OutputSummary) == "" {
			continue
		}
		if strings.TrimSpace(task.ArtifactPath) == "" {
			t.Fatalf("task %q has output summary but empty artifact path", task.ID)
		}
		foundArtifactPath = task.ArtifactPath
		break
	}
	if foundArtifactPath == "" {
		t.Fatalf("no task produced an artifact path after CreateRun()")
	}
	if _, statErr := os.Stat(foundArtifactPath); statErr != nil {
		t.Fatalf("artifact file does not exist: %v", statErr)
	}

	content, readErr := os.ReadFile(foundArtifactPath)
	if readErr != nil {
		t.Fatalf("os.ReadFile() error = %v", readErr)
	}
	if strings.TrimSpace(string(content)) == "" {
		t.Fatalf("artifact file is empty: %s", foundArtifactPath)
	}

	foundAgentTrace := false
	for _, agent := range created.Agents {
		if strings.TrimSpace(agent.LastOutput) != "" || strings.TrimSpace(agent.LastHandoff) != "" || strings.TrimSpace(agent.LastAction) != "" {
			foundAgentTrace = true
			break
		}
	}
	if !foundAgentTrace {
		t.Fatalf("no agent trace fields were populated")
	}
}

func TestStressRunsCanAdvanceToCompletionWithoutStall(t *testing.T) {
	t.Parallel()

	service, err := NewService(t.TempDir(), eventbus.New())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	defer func() {
		if closeErr := service.Close(); closeErr != nil {
			t.Fatalf("service.Close() error = %v", closeErr)
		}
	}()

	const rounds = 12
	for i := 1; i <= rounds; i++ {
		title := fmt.Sprintf("stress run #%d", i)
		deliverable := fmt.Sprintf("delivery-package-%d", i)
		created, createErr := service.CreateRun(context.Background(), RunCreationInput{
			Title:                 title,
			Mission:               "validate agent handoff continuity and output persistence under repeated runs",
			Deliverable:           deliverable,
			TaskType:              "engineering",
			Priority:              "high",
			MaxAgents:             16,
			RequiresQA:            true,
			RequiresSecurity:      true,
			RequiresHumanApproval: true,
		})
		if createErr != nil {
			t.Fatalf("round %d CreateRun() error = %v", i, createErr)
		}
		if created.RunStep < 1 {
			t.Fatalf("round %d CreateRun() step = %d, want >= 1", i, created.RunStep)
		}

		state := created
		guard := 0
		for state.RunStep < state.MaxStep {
			guard++
			if guard > state.MaxStep+3 {
				t.Fatalf("round %d stalled before completion at step %d", i, state.RunStep)
			}
			state, err = service.AdvanceDemoRun(context.Background())
			if err != nil {
				t.Fatalf("round %d AdvanceDemoRun() error = %v", i, err)
			}
		}

		if state.RunStep != state.MaxStep {
			t.Fatalf("round %d final step = %d, want %d", i, state.RunStep, state.MaxStep)
		}

		foundArtifact := false
		for _, task := range state.Tasks {
			if strings.TrimSpace(task.OutputSummary) == "" {
				continue
			}
			if strings.TrimSpace(task.ArtifactPath) == "" {
				t.Fatalf("round %d task %q missing artifactPath", i, task.ID)
			}
			if _, statErr := os.Stat(task.ArtifactPath); statErr != nil {
				t.Fatalf("round %d task %q artifact missing on disk: %v", i, task.ID, statErr)
			}
			foundArtifact = true
			break
		}
		if !foundArtifact {
			t.Fatalf("round %d produced no artifact files", i)
		}

		foundTrace := false
		for _, agent := range state.Agents {
			if strings.TrimSpace(agent.LastAction) != "" ||
				strings.TrimSpace(agent.LastOutput) != "" ||
				strings.TrimSpace(agent.LastHandoff) != "" ||
				strings.TrimSpace(agent.LastArtifactPath) != "" {
				foundTrace = true
				break
			}
		}
		if !foundTrace {
			t.Fatalf("round %d did not expose agent trace fields", i)
		}
	}
}
