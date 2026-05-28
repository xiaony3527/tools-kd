package main

import "testing"

func TestAddPackageIgnoresEmptyInput(t *testing.T) {
	app := NewApp()
	app.SetDestination("浙江省", "杭州市")

	state := app.AddPackage(PackageInput{})
	if len(state.Packages) != 0 {
		t.Fatalf("package rows = %d, want 0 (empty input should be no-op)", len(state.Packages))
	}
	// Verify internal state also unchanged
	if app.calc.Count() != 0 {
		t.Fatalf("calc.Count() = %d, want 0", app.calc.Count())
	}
}

func TestAddPackageClampsQuantityToOne(t *testing.T) {
	app := NewApp()
	app.SetDestination("浙江省", "杭州市")

	state := app.AddPackage(PackageInput{ActualWeight: 2.5, Quantity: 0})
	if state.Summary.TotalCount != 1 {
		t.Fatalf("TotalCount = %d, want 1 (quantity 0 clamped to 1)", state.Summary.TotalCount)
	}
}

func TestAddPackagePrefersDimensionsOverVolumeAndWeight(t *testing.T) {
	app := NewApp()
	app.SetDestination("浙江省", "杭州市")

	// All three paths available: dims should win → ModeDimWeight
	state := app.AddPackage(PackageInput{
		Length: 10, Width: 20, Height: 30,
		Volume:       99999,
		ActualWeight: 5,
		Quantity:     1,
	})
	if len(state.Packages) != 1 {
		t.Fatalf("package rows = %d, want 1", len(state.Packages))
	}
	row := state.Packages[0]
	if row.ModeLabel != "长宽高+实重" {
		t.Fatalf("ModeLabel = %q, want %q", row.ModeLabel, "长宽高+实重")
	}
	if row.Volume != 10*20*30 {
		t.Fatalf("Volume = %v, want %v", row.Volume, float64(10*20*30))
	}
}

func TestAddPackageUsesVolumeWhenNoFullDimensions(t *testing.T) {
	app := NewApp()
	app.SetDestination("浙江省", "杭州市")

	// Volume-only with partial dims → ModeVolumeOnly
	state := app.AddPackage(PackageInput{
		Length: 10, Width: 20, // missing Height → not full dims
		Volume:       5000,
		ActualWeight: 3,
		Quantity:     1,
	})
	if len(state.Packages) != 1 {
		t.Fatalf("package rows = %d, want 1", len(state.Packages))
	}
	row := state.Packages[0]
	if row.ModeLabel != "仅体积" {
		t.Fatalf("ModeLabel = %q, want %q", row.ModeLabel, "仅体积")
	}
	if row.Volume != 5000 {
		t.Fatalf("Volume = %v, want 5000", row.Volume)
	}
}

func TestAddDeleteAndClearPackageReturnUpdatedQuoteState(t *testing.T) {
	app := NewApp()
	app.SetDestination("浙江省", "杭州市")

	state := app.AddPackage(PackageInput{ActualWeight: 2.5, Quantity: 2})
	if len(state.Packages) != 1 {
		t.Fatalf("package rows = %d, want 1", len(state.Packages))
	}

	id := state.Packages[0].ID
	state = app.DeletePackage(id)
	if len(state.Packages) != 0 {
		t.Fatalf("package rows after delete = %d, want 0", len(state.Packages))
	}

	state = app.AddPackage(PackageInput{ActualWeight: 3, Quantity: 1})
	state = app.ClearPackages()
	if state.Summary.TotalCount != 0 {
		t.Fatalf("TotalCount after clear = %d, want 0", state.Summary.TotalCount)
	}
}
