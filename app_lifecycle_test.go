package main

import (
	"context"
	"testing"
)

func TestBeforeCloseMinimisesToBackground(t *testing.T) {
	t.Parallel()

	minimised := false
	app := &App{
		promptCloseAction: func(context.Context) (closeAction, error) {
			return closeActionMinimise, nil
		},
		minimiseWindow: func(context.Context) {
			minimised = true
		},
	}

	preventClose := app.beforeClose(context.Background())

	if !preventClose {
		t.Fatalf("beforeClose() should prevent shutdown when minimising")
	}
	if !minimised {
		t.Fatalf("beforeClose() should minimise the window when the user chooses background mode")
	}
}

func TestBeforeCloseApprovesFullShutdown(t *testing.T) {
	t.Parallel()

	promptCalls := 0
	app := &App{
		promptCloseAction: func(context.Context) (closeAction, error) {
			promptCalls++
			return closeActionExit, nil
		},
		minimiseWindow: func(context.Context) {
			t.Fatalf("beforeClose() should not minimise on full shutdown")
		},
	}

	preventClose := app.beforeClose(context.Background())

	if preventClose {
		t.Fatalf("beforeClose() should allow shutdown when the user chooses to close the background service")
	}
	if !app.isCloseApproved() {
		t.Fatalf("beforeClose() should approve closing after the user confirms shutdown")
	}
	if promptCalls != 1 {
		t.Fatalf("beforeClose() should prompt exactly once, got %d", promptCalls)
	}
}

func TestBeforeCloseSkipsPromptWhenCloseAlreadyApproved(t *testing.T) {
	t.Parallel()

	promptCalls := 0
	app := &App{
		closeApproved: true,
		promptCloseAction: func(context.Context) (closeAction, error) {
			promptCalls++
			return closeActionExit, nil
		},
	}

	preventClose := app.beforeClose(context.Background())

	if preventClose {
		t.Fatalf("beforeClose() should not block shutdown once closing has already been approved")
	}
	if promptCalls != 0 {
		t.Fatalf("beforeClose() should skip prompting when shutdown is already approved")
	}
}

func TestShutdownClosesBackgroundServiceOnce(t *testing.T) {
	t.Parallel()

	closeCalls := 0
	app := &App{
		closeDashboard: func() error {
			closeCalls++
			return nil
		},
	}

	app.shutdown(context.Background())
	app.shutdown(context.Background())

	if closeCalls != 1 {
		t.Fatalf("shutdown() should close background services once, got %d calls", closeCalls)
	}
	if !app.isCloseApproved() {
		t.Fatalf("shutdown() should approve closing")
	}
}
