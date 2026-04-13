package main

import (
	"context"
	"os"

	"maliangswarm/internal/eventbus"
	"maliangswarm/internal/orchestrator"
)

// App struct
type App struct {
	ctx       context.Context
	bus       *eventbus.Bus
	dashboard *orchestrator.Service
	bootErr   error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		bus: eventbus.New(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	workdir, err := os.Getwd()
	if err != nil {
		a.bootErr = err
		return
	}

	a.dashboard, a.bootErr = orchestrator.NewService(workdir, a.bus)
}

// GetDashboardState returns the current enterprise swarm snapshot used by the UI.
func (a *App) GetDashboardState() (orchestrator.DashboardState, error) {
	if a.bootErr != nil {
		return orchestrator.DashboardState{}, a.bootErr
	}
	return a.dashboard.GetDashboardState(a.ctx)
}

// AdvanceDemoRun moves the persisted workflow forward by one stage.
func (a *App) AdvanceDemoRun() (orchestrator.DashboardState, error) {
	if a.bootErr != nil {
		return orchestrator.DashboardState{}, a.bootErr
	}
	return a.dashboard.AdvanceDemoRun(a.ctx)
}

// ResetDemoRun rewinds the persisted demo workflow while preserving settings.
func (a *App) ResetDemoRun() (orchestrator.DashboardState, error) {
	if a.bootErr != nil {
		return orchestrator.DashboardState{}, a.bootErr
	}
	return a.dashboard.ResetDemoRun(a.ctx)
}

// UpdateRuntimeSettings saves control-desk settings to the local runtime store.
func (a *App) UpdateRuntimeSettings(input orchestrator.SettingsInput) (orchestrator.DashboardState, error) {
	if a.bootErr != nil {
		return orchestrator.DashboardState{}, a.bootErr
	}
	return a.dashboard.UpdateRuntimeSettings(a.ctx, input)
}

// CreateRun creates a new orchestration run from a submitted task brief.
func (a *App) CreateRun(input orchestrator.RunCreationInput) (orchestrator.DashboardState, error) {
	if a.bootErr != nil {
		return orchestrator.DashboardState{}, a.bootErr
	}
	return a.dashboard.CreateRun(a.ctx, input)
}

// UpsertAIProvider saves one AI API configuration into the local registry.
func (a *App) UpsertAIProvider(input orchestrator.AIProviderInput) (orchestrator.DashboardState, error) {
	if a.bootErr != nil {
		return orchestrator.DashboardState{}, a.bootErr
	}
	return a.dashboard.UpsertAIProvider(a.ctx, input)
}

// DeleteAIProvider removes one AI API configuration from the local registry.
func (a *App) DeleteAIProvider(id string) (orchestrator.DashboardState, error) {
	if a.bootErr != nil {
		return orchestrator.DashboardState{}, a.bootErr
	}
	return a.dashboard.DeleteAIProvider(a.ctx, id)
}

// ProbeAIProvider runs a live connectivity and routing probe against the current provider form.
func (a *App) ProbeAIProvider(input orchestrator.AIProviderInput) (string, error) {
	if a.bootErr != nil {
		return "", a.bootErr
	}
	return a.dashboard.ProbeAIProvider(a.ctx, input)
}
