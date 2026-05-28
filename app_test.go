package main

import (
	"context"
	"testing"
)

func TestNewAppInitializesEmptyCalculatorState(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}
	if app.calc == nil {
		t.Fatal("NewApp() should initialize calc")
	}
	if got := app.calc.Count(); got != 0 {
		t.Fatalf("initial package count = %d, want 0", got)
	}
}

func TestStartupStoresContext(t *testing.T) {
	ctx := context.Background()
	app := NewApp()
	app.startup(ctx)
	if app.ctx != ctx {
		t.Fatal("startup() did not store the context")
	}
}
