package main

import (
	"context"
	"log"
	"os"
	"sync"

	"maliangswarm/internal/eventbus"
	"maliangswarm/internal/orchestrator"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type closeAction string

const (
	closeActionMinimise closeAction = "minimise"
	closeActionExit     closeAction = "exit"
)

const (
	minimiseToBackgroundLabel = "最小化到后台"
	closeBackgroundLabel      = "关闭后台"
)

// App struct
type App struct {
	ctx               context.Context
	bus               *eventbus.Bus
	dashboard         *orchestrator.Service
	bootErr           error
	closeMu           sync.Mutex
	shutdownOnce      sync.Once
	closeApproved     bool
	promptCloseAction func(context.Context) (closeAction, error)
	minimiseWindow    func(context.Context)
	closeDashboard    func() error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		bus:               eventbus.New(),
		promptCloseAction: promptForCloseAction,
		minimiseWindow:    runtime.WindowMinimise,
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
	if a.bootErr == nil && a.dashboard != nil {
		a.closeDashboard = a.dashboard.Close
	}
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

func (a *App) beforeClose(ctx context.Context) bool {
	if a.isCloseApproved() {
		return false
	}

	action, err := a.resolveCloseAction(ctx)
	if err != nil {
		log.Printf("close confirmation failed, falling back to full shutdown: %v", err)
		a.approveClose()
		return false
	}

	if action == closeActionExit {
		a.approveClose()
		return false
	}

	if currentCtx := a.resolveContext(ctx); currentCtx != nil && a.minimiseWindow != nil {
		a.minimiseWindow(currentCtx)
	}

	return true
}

func (a *App) shutdown(_ context.Context) {
	a.approveClose()

	a.shutdownOnce.Do(func() {
		if a.closeDashboard == nil && a.dashboard != nil {
			a.closeDashboard = a.dashboard.Close
		}
		if a.closeDashboard == nil {
			return
		}
		if err := a.closeDashboard(); err != nil {
			log.Printf("failed to close background service cleanly: %v", err)
		}
	})
}

func (a *App) resolveCloseAction(ctx context.Context) (closeAction, error) {
	if a.promptCloseAction == nil {
		return closeActionExit, nil
	}
	currentCtx := a.resolveContext(ctx)
	if currentCtx == nil {
		return closeActionExit, nil
	}
	return a.promptCloseAction(currentCtx)
}

func (a *App) resolveContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return a.ctx
}

func (a *App) approveClose() {
	a.closeMu.Lock()
	defer a.closeMu.Unlock()

	a.closeApproved = true
}

func (a *App) isCloseApproved() bool {
	a.closeMu.Lock()
	defer a.closeMu.Unlock()

	return a.closeApproved
}

func promptForCloseAction(ctx context.Context) (closeAction, error) {
	selection, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "关闭 maliang swarm",
		Message:       "点击「是(Y)」将彻底关闭后台并释放当前占用。\n点击「否(N)」将最小化到后台继续运行。",
		Buttons:       []string{closeBackgroundLabel, minimiseToBackgroundLabel},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil {
		log.Printf("close dialog error: %v, default to exit", err)
		return closeActionExit, nil
	}

	log.Printf("close dialog selection: %q", selection)

	if selection == "Yes" || selection == closeBackgroundLabel {
		log.Println("user confirmed full shutdown")
		return closeActionExit, nil
	}

	log.Println("user chose minimise to background")
	return closeActionMinimise, nil
}
