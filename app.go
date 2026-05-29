package main

import (
	"context"
	"math"
	"strings"
	"sync"
)

// App is the main application struct for Wails.
type App struct {
	ctx      context.Context
	calc     *Calc
	province string
	city     string
}

// NewApp creates a new App with initialized calculator state.
func NewApp() *App {
	return &App{
		calc: NewCalc(),
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ====== Province/city canonicalization ======

var (
	provinceAliases     map[string]string
	provinceAliasesOnce sync.Once
)

func initProvinceAliases() {
	provinceAliases = make(map[string]string, len(provinces)*2)
	for _, p := range provinces {
		provinceAliases[p] = p
		// Support suffix-trimmed names: 浙江 → 浙江省, 北京 → 北京市
		trimmed := strings.TrimSuffix(strings.TrimSuffix(p, "省"), "市")
		if trimmed != p {
			provinceAliases[trimmed] = p
		}
	}
}

// canonicalProvince resolves a possibly suffix-stripped province name (e.g. 浙江, 北京)
// to the canonical form used in pricing data. Returns "" when unknown.
func canonicalProvince(name string) string {
	provinceAliasesOnce.Do(initProvinceAliases)
	return provinceAliases[name]
}

// canonicalCity resolves a possibly suffix-stripped city name to the canonical form
// found in GetCities(province). Returns "" when the city cannot be matched.
func canonicalCity(province, city string) string {
	cities := GetCities(province)
	if len(cities) == 0 {
		return ""
	}
	// Exact match
	for _, c := range cities {
		if c == city {
			return city
		}
	}
	// Try appending 市 (e.g. 杭州 → 杭州市)
	if !strings.HasSuffix(city, "市") {
		candidate := city + "市"
		for _, c := range cities {
			if c == candidate {
				return c
			}
		}
	}
	// Try stripping 市 (e.g. 杭州市 → 杭州 — matches "恩施" style entries)
	trimmed := strings.TrimSuffix(city, "市")
	if trimmed != city {
		for _, c := range cities {
			if c == trimmed {
				return c
			}
		}
	}
	return ""
}

// ====== DTOs for frontend state ======

// QuoteState is the full quote response sent to the frontend.
type QuoteState struct {
	Destination DestinationState `json:"destination"`
	Packages    []PackageRow     `json:"packages"`
	Summary     SummaryState     `json:"summary"`
	Sto         CarrierQuote     `json:"sto"`
	Bs          CarrierQuote     `json:"bs"`
}

// DestinationState holds the selected destination and available cities.
type DestinationState struct {
	Province string   `json:"province"`
	City     string   `json:"city"`
	Cities   []string `json:"cities"`
}

// SummaryState holds aggregate counts for the quote summary.
type SummaryState struct {
	TotalCount    int     `json:"totalCount"`
	DisplayWeight float64 `json:"displayWeight"`
}

// CarrierQuote holds the calculated quote for a single carrier.
type CarrierQuote struct {
	Price          float64 `json:"price"`
	BillableWeight float64 `json:"billableWeight"`
	Available      bool    `json:"available"`
	Recommended    bool    `json:"recommended"`
	Note           string  `json:"note"`
}

// PackageRow is a serializable row for each package in the current list.
type PackageRow struct {
	ID           int     `json:"id"`
	Length       float64 `json:"length"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	Quantity     int     `json:"quantity"`
	Mode         int     `json:"mode"`
	ModeLabel    string  `json:"modeLabel"`
	ActualWeight float64 `json:"actualWeight"`
	Volume       float64 `json:"volume"`
	StoBillable  float64 `json:"stoBillable"`
	BsBillable   float64 `json:"bsBillable"`
}

// modeLabel converts an InputMode to a human-readable Chinese label.
func modeLabel(m InputMode) string {
	switch m {
	case ModeDimWeight:
		return "长宽高+实重"
	case ModeWeightOnly:
		return "仅实重"
	case ModeVolumeOnly:
		return "仅体积"
	default:
		return "长宽高+实重"
	}
}

// AddressResult is the result of address analysis via AI.
type AddressResult struct {
	Province string `json:"province"`
	City     string `json:"city"`
}

// PackageInput is the input DTO for adding a package from the frontend.
type PackageInput struct {
	Length       float64 `json:"length"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	Volume       float64 `json:"volume"`
	ActualWeight float64 `json:"actualWeight"`
	Quantity     int     `json:"quantity"`
}

// emptyQuoteState returns a QuoteState with initialized empty slices
// so that JSON serialization produces [] instead of null.
func emptyQuoteState() QuoteState {
	return QuoteState{
		Destination: DestinationState{Cities: []string{}},
		Packages:    []PackageRow{},
	}
}

// CalculateQuote computes the full quote state from current calculator data.
func (a *App) CalculateQuote() QuoteState {
	// Guard against nil receiver or nil calc.
	if a == nil || a.calc == nil {
		return emptyQuoteState()
	}

	totalSto := a.calc.TotalSto()
	totalBs := a.calc.TotalBs()

	stoBillable := math.Ceil(totalSto)
	bsBillable := math.Ceil(totalBs)

	displayWeight := stoBillable
	if bsBillable > displayWeight {
		displayWeight = bsBillable
	}

	// City fallback: when province is set and city is empty, treat as "默认".
	city := a.city
	if a.province != "" && city == "" {
		city = "默认"
	}

	// Build package rows
	var pkgs []PackageRow
	for _, p := range a.calc.GetPackages() {
		pkgs = append(pkgs, PackageRow{
			ID:           p.ID,
			Length:       p.Length,
			Width:        p.Width,
			Height:       p.Height,
			Quantity:     p.Quantity,
			Mode:         int(p.Mode),
			ModeLabel:    modeLabel(p.Mode),
			ActualWeight: p.ActualWeight,
			Volume:       p.Volume,
			StoBillable:  math.Ceil(StoBillableRaw(p.Volume, p.ActualWeight) * float64(p.Quantity)),
			BsBillable:   math.Ceil(BsBillableRaw(p.Volume, p.ActualWeight) * float64(p.Quantity)),
		})
	}
	if pkgs == nil {
		pkgs = []PackageRow{}
	}

	// Destination state — ensure Cities is never nil.
	cities := GetCities(a.province)
	if cities == nil {
		cities = []string{}
	}
	dest := DestinationState{
		Province: a.province,
		City:     city,
		Cities:   cities,
	}

	// -------- BS (百世) quote --------
	bsPrice := 0.0
	bsAvailable := true
	bsNote := ""
	priceEntry := GetPriceDefault(a.province, city)
	if priceEntry.Base30 == 0 && priceEntry.R30_70 == 0 {
		bsAvailable = false
		bsNote = "无百世报价"
	} else {
		bsPrice = CalcBsCost(totalBs, priceEntry)
	}

	// -------- STO (申通) quote --------
	stoPrice := 0.0
	stoAvailable := true
	stoNote := ""
	firstKg, addKg, ok := GetStoPrice(a.province)
	if !ok {
		stoAvailable = false
		stoNote = "无申通报价"
	} else if totalSto > 50 {
		stoAvailable = false
		stoNote = "申通限重50kg"
	} else {
		cost, valid := CalcStoCost(totalSto, firstKg, addKg)
		if !valid {
			stoAvailable = false
			stoNote = "申通限重50kg"
		} else {
			stoPrice = cost
		}
	}

	// -------- Recommendation (only when there are items to ship) --------
	totalCount := a.calc.Count()
	stoRec := false
	bsRec := false
	if totalCount > 0 && displayWeight > 0 {
		if stoAvailable && bsAvailable {
			if stoPrice <= bsPrice {
				stoRec = true
			} else {
				bsRec = true
			}
		} else if stoAvailable {
			stoRec = true
		} else if bsAvailable {
			bsRec = true
		}
	}

	return QuoteState{
		Destination: dest,
		Packages:    pkgs,
		Summary: SummaryState{
			TotalCount:    totalCount,
			DisplayWeight: displayWeight,
		},
		Sto: CarrierQuote{
			Price:          stoPrice,
			BillableWeight: stoBillable,
			Available:      stoAvailable,
			Recommended:    stoRec,
			Note:           stoNote,
		},
		Bs: CarrierQuote{
			Price:          bsPrice,
			BillableWeight: bsBillable,
			Available:      bsAvailable,
			Recommended:    bsRec,
			Note:           bsNote,
		},
	}
}

// GetProvinces returns the list of all known provinces (for frontend dropdown).
func (a *App) GetProvinces() []string {
	return provinces
}

// GetInitialState returns the current quote state (no side effects).
func (a *App) GetInitialState() QuoteState {
	return a.CalculateQuote()
}

// SetDestination updates the province/city and returns the new quote state.
// Inputs are canonicalized so suffix-trimmed names (浙江, 北京) are resolved.
// When city is empty or not found in the city list, falls back to the first city.
// When province is unknown, both province and city are cleared.
func (a *App) SetDestination(province, city string) QuoteState {
	province = canonicalProvince(province)
	if province == "" {
		a.province = ""
		a.city = ""
		return a.CalculateQuote()
	}
	a.province = province

	cities := GetCities(province)
	if len(cities) == 0 {
		a.city = ""
		return a.CalculateQuote()
	}

	a.city = canonicalCity(province, city)
	if a.city == "" {
		a.city = cities[0]
	}
	return a.CalculateQuote()
}

// AnalyzeAddress wraps the top-level AnalyzeAddress function from aianalyze.go
// as an App binding for the frontend. Trims whitespace and returns an empty result
// for blank input without making a network call. Canonicalizes the AI output to
// match pricing data conventions (e.g. 浙江 → 浙江省, 杭州 → 杭州市).
func (a *App) AnalyzeAddress(address string) (AddressResult, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return AddressResult{}, nil
	}
	province, city, err := AnalyzeAddress(address)
	if err != nil {
		return AddressResult{}, err
	}
	canonProvince := canonicalProvince(province)
	canonCity := canonicalCity(canonProvince, city)
	return AddressResult{Province: canonProvince, City: canonCity}, nil
}

// AddPackage adds a package using the input and returns the updated quote state.
// Returns the current state unchanged (no mutation) when the input has no valid data:
// at least full dimensions, a positive volume, or a positive actual weight is required.
func (a *App) AddPackage(input PackageInput) QuoteState {
	hasFullDims := input.Length > 0 && input.Width > 0 && input.Height > 0
	hasVolume := input.Volume > 0
	hasWeight := input.ActualWeight > 0
	if !hasFullDims && !hasVolume && !hasWeight {
		return a.CalculateQuote()
	}

	qty := input.Quantity
	if qty < 1 {
		qty = 1
	}

	var mode InputMode
	switch {
	case hasFullDims:
		mode = ModeDimWeight
	case hasVolume:
		mode = ModeVolumeOnly
	default:
		mode = ModeWeightOnly
	}

	a.calc.AddPackage(input.Length, input.Width, input.Height, input.Volume, input.ActualWeight, qty, mode)
	return a.CalculateQuote()
}

// DeletePackage removes a package by ID and returns the updated quote state.
func (a *App) DeletePackage(id int) QuoteState {
	a.calc.DeletePackage(id)
	return a.CalculateQuote()
}

// ClearPackages removes all packages and returns the updated quote state.
func (a *App) ClearPackages() QuoteState {
	a.calc.ClearPackages()
	return a.CalculateQuote()
}
