package main

import (
	"encoding/json"
	"testing"
)

func TestCalculateQuoteReturnsRoundedDisplayWeightsAndRecommendation(t *testing.T) {
	app := NewApp()
	app.province = "浙江省"
	app.city = "杭州市"
	app.calc.AddPackage(0, 0, 0, 0, 1.2, 3, ModeWeightOnly)

	state := app.CalculateQuote()
	if state.Summary.TotalCount != 3 {
		t.Fatalf("TotalCount = %d, want 3", state.Summary.TotalCount)
	}
	if state.Sto.BillableWeight != 4 {
		t.Fatalf("Sto BillableWeight = %v, want 4", state.Sto.BillableWeight)
	}
	if state.Bs.BillableWeight != 4 {
		t.Fatalf("Bs BillableWeight = %v, want 4", state.Bs.BillableWeight)
	}
	if state.Sto.Price <= 0 || state.Bs.Price <= 0 {
		t.Fatal("expected both carrier prices to be calculated")
	}
	if !state.Sto.Recommended {
		t.Fatal("expected Sto to be recommended (cheaper)")
	}
}

func TestCalculateQuoteEmptyStateHasNoRecommendationAndEmptyArrays(t *testing.T) {
	app := NewApp()
	state := app.CalculateQuote()

	if state.Sto.Recommended {
		t.Fatal("Sto.Recommended should be false in empty state")
	}
	if state.Bs.Recommended {
		t.Fatal("Bs.Recommended should be false in empty state")
	}
	if len(state.Packages) != 0 {
		t.Fatalf("Packages len = %d, want 0", len(state.Packages))
	}
	if state.Packages == nil {
		t.Fatal("Packages should be non-nil empty slice, not nil")
	}
	if state.Destination.Cities == nil {
		t.Fatal("Destination.Cities should be non-nil empty slice, not nil")
	}
	if len(state.Destination.Cities) != 0 {
		t.Fatalf("Destination.Cities len = %d, want 0", len(state.Destination.Cities))
	}

	// JSON marshal should produce "packages":[],"cities":[] not null
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	// Check packages field is array not null
	var pkgs []json.RawMessage
	if err := json.Unmarshal(raw["packages"], &pkgs); err != nil {
		t.Fatalf("packages should be a JSON array, got error: %v\nraw: %s", err, string(raw["packages"]))
	}
	// Check destination.cities field is array not null
	var dest map[string]json.RawMessage
	if err := json.Unmarshal(raw["destination"], &dest); err != nil {
		t.Fatalf("destination should be a JSON object, got error: %v", err)
	}
	var cities []json.RawMessage
	if err := json.Unmarshal(dest["cities"], &cities); err != nil {
		t.Fatalf("destination.cities should be a JSON array, got error: %v\nraw: %s", err, string(dest["cities"]))
	}
}

func TestCalculateQuoteNilAppDoesNotPanic(t *testing.T) {
	var app *App
	state := app.CalculateQuote()
	if state.Packages == nil || len(state.Packages) != 0 {
		t.Fatal("nil app should return empty Packages slice")
	}
	if state.Destination.Cities == nil || len(state.Destination.Cities) != 0 {
		t.Fatal("nil app should return empty Cities slice")
	}
}

func TestCalculateQuoteNilCalcDoesNotPanic(t *testing.T) {
	app := &App{}
	state := app.CalculateQuote()
	if state.Packages == nil || len(state.Packages) != 0 {
		t.Fatal("nil calc should return empty Packages slice")
	}
	if state.Destination.Cities == nil || len(state.Destination.Cities) != 0 {
		t.Fatal("nil calc should return empty Cities slice")
	}
}

func TestCalculateQuoteUsesDefaultCityWhenCityEmpty(t *testing.T) {
	app := NewApp()
	app.province = "浙江省"
	app.city = "" // explicitly empty
	app.calc.AddPackage(0, 0, 0, 0, 1.2, 3, ModeWeightOnly)

	state := app.CalculateQuote()

	if state.Destination.City != "默认" {
		t.Fatalf("Destination.City = %q, want %q", state.Destination.City, "默认")
	}
	if !state.Bs.Available {
		t.Fatal("Bs should be available with default city pricing")
	}
	if state.Bs.Price <= 0 {
		t.Fatal("Bs.Price should be > 0 with valid default pricing")
	}
}

func TestCalculateQuotePackageRowsIncludeDimensionsAndModeLabel(t *testing.T) {
	app := NewApp()
	app.province = "浙江省"
	app.city = "杭州市"
	app.calc.AddPackage(10, 20, 30, 0, 1, 2, ModeDimWeight)

	state := app.CalculateQuote()

	if len(state.Packages) != 1 {
		t.Fatalf("expected 1 package row, got %d", len(state.Packages))
	}
	row := state.Packages[0]
	if row.Length != 10 {
		t.Fatalf("Length = %v, want 10", row.Length)
	}
	if row.Width != 20 {
		t.Fatalf("Width = %v, want 20", row.Width)
	}
	if row.Height != 30 {
		t.Fatalf("Height = %v, want 30", row.Height)
	}
	if row.ModeLabel != "长宽高+实重" {
		t.Fatalf("ModeLabel = %q, want %q", row.ModeLabel, "长宽高+实重")
	}
	if row.Mode != int(ModeDimWeight) {
		t.Fatalf("Mode = %v, want %v", row.Mode, int(ModeDimWeight))
	}
	if row.Volume <= 0 {
		t.Fatal("Volume should be > 0 for ModeDimWeight")
	}
}

func TestCalculateQuoteStoUnavailableOver50kg(t *testing.T) {
	app := NewApp()
	app.province = "浙江省"
	app.city = "杭州市"
	// Single package 51kg => total > 50, Sto unavailable
	app.calc.AddPackage(0, 0, 0, 0, 51, 1, ModeWeightOnly)

	state := app.CalculateQuote()

	if state.Sto.Available {
		t.Fatal("Sto should be unavailable for weight > 50kg")
	}
	if state.Sto.Note == "" {
		t.Fatal("Sto should have a non-empty note when unavailable")
	}
	if !state.Bs.Available {
		t.Fatal("Bs should still be available when Sto is unavailable")
	}
	if !state.Bs.Recommended {
		t.Fatal("Bs should be recommended when Sto is unavailable")
	}
	if state.Sto.Recommended {
		t.Fatal("Sto should not be recommended when unavailable")
	}
	if state.Summary.DisplayWeight <= 0 {
		t.Fatal("DisplayWeight should be > 0 even when Sto unavailable")
	}
}

func TestCalculateQuoteWeightOnlyModeLabel(t *testing.T) {
	app := NewApp()
	app.province = "浙江省"
	app.city = "杭州市"
	app.calc.AddPackage(0, 0, 0, 0, 2, 1, ModeWeightOnly)

	state := app.CalculateQuote()
	if len(state.Packages) != 1 {
		t.Fatalf("expected 1 package row, got %d", len(state.Packages))
	}
	if state.Packages[0].ModeLabel != "仅实重" {
		t.Fatalf("ModeLabel = %q, want %q", state.Packages[0].ModeLabel, "仅实重")
	}
}

func TestCalculateQuoteVolumeOnlyModeLabel(t *testing.T) {
	app := NewApp()
	app.province = "浙江省"
	app.city = "杭州市"
	app.calc.AddPackage(10, 20, 30, 0, 0, 1, ModeVolumeOnly)

	state := app.CalculateQuote()
	if len(state.Packages) != 1 {
		t.Fatalf("expected 1 package row, got %d", len(state.Packages))
	}
	if state.Packages[0].ModeLabel != "仅体积" {
		t.Fatalf("ModeLabel = %q, want %q", state.Packages[0].ModeLabel, "仅体积")
	}
}
