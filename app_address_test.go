package main

import "testing"

func TestSetDestinationReturnsMatchingCityList(t *testing.T) {
	app := NewApp()
	state := app.SetDestination("浙江省", "杭州市")
	if state.Destination.Province != "浙江省" {
		t.Fatalf("Province = %q, want 浙江省", state.Destination.Province)
	}
	if len(state.Destination.Cities) == 0 {
		t.Fatal("expected city list for province")
	}
}

func TestSetDestinationFallsBackToFirstCityWhenUnknown(t *testing.T) {
	app := NewApp()
	state := app.SetDestination("浙江省", "不存在的城市")
	if state.Destination.City != "默认" {
		t.Fatalf("City = %q, want %q (first city in list)", state.Destination.City, "默认")
	}
}

func TestSetDestinationCanonicalizesTrimmedProvinceAndCity(t *testing.T) {
	app := NewApp()
	state := app.SetDestination("浙江", "杭州")
	if state.Destination.Province != "浙江省" {
		t.Fatalf("Province = %q, want 浙江省", state.Destination.Province)
	}
	if state.Destination.City != "杭州市" {
		t.Fatalf("City = %q, want 杭州市", state.Destination.City)
	}
	if len(state.Destination.Cities) == 0 {
		t.Fatal("expected city list for province")
	}
}

func TestSetDestinationCanonicalizesDirectAdminCity(t *testing.T) {
	app := NewApp()
	state := app.SetDestination("北京", "北京")
	if state.Destination.Province != "北京市" {
		t.Fatalf("Province = %q, want 北京市", state.Destination.Province)
	}
	if state.Destination.City != "北京市" {
		t.Fatalf("City = %q, want 北京市", state.Destination.City)
	}
	if len(state.Destination.Cities) == 0 {
		t.Fatal("expected city list for Beijing")
	}

	// Also test full canonical form works
	state = app.SetDestination("北京市", "北京市")
	if state.Destination.Province != "北京市" {
		t.Fatalf("Province = %q, want 北京市", state.Destination.Province)
	}
	if state.Destination.City != "北京市" {
		t.Fatalf("City = %q, want 北京市", state.Destination.City)
	}
}

func TestSetDestinationClearsUnknownProvinceAndCity(t *testing.T) {
	app := NewApp()
	state := app.SetDestination("不存在", "任意")
	if state.Destination.Province != "" {
		t.Fatalf("Province = %q, want empty", state.Destination.Province)
	}
	if state.Destination.City != "" {
		t.Fatalf("City = %q, want empty", state.Destination.City)
	}
	if len(state.Destination.Cities) != 0 {
		t.Fatalf("Cities len = %d, want 0", len(state.Destination.Cities))
	}
}

func TestAnalyzeAddressReturnsEmptyResultForBlankInput(t *testing.T) {
	app := NewApp()

	// Empty string
	result, err := app.AnalyzeAddress("")
	if err != nil {
		t.Fatalf("unexpected error for empty input: %v", err)
	}
	if result.Province != "" || result.City != "" {
		t.Fatalf("expected zero AddressResult, got %+v", result)
	}

	// Whitespace-only
	result, err = app.AnalyzeAddress("  \t  ")
	if err != nil {
		t.Fatalf("unexpected error for whitespace input: %v", err)
	}
	if result.Province != "" || result.City != "" {
		t.Fatalf("expected zero AddressResult for whitespace, got %+v", result)
	}
}
