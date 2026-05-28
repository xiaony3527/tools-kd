package main

import (
	"testing"
)

func TestTotalRaw(t *testing.T) {
	c := NewCalc()
	// 1.2kg actual, ModeWeightOnly => volume=0
	c.AddPackage(0, 0, 0, 0, 1.2, 3, ModeWeightOnly)

	totalSto := c.TotalSto()
	totalBs := c.TotalBs()

	if totalSto < 3.59 || totalSto > 3.61 {
		t.Errorf("TotalSto expected ~3.6, got %v", totalSto)
	}
	if totalBs < 3.59 || totalBs > 3.61 {
		t.Errorf("TotalBs expected ~3.6, got %v", totalBs)
	}
}

func TestCalcBsCostCeil(t *testing.T) {
	// 30.2 ceil => 31 => rate 2 => 31*2 = 62
	got := CalcBsCost(30.2, PriceEntry{Base30: 0, R30_70: 2})
	if got != 62 {
		t.Errorf("CalcBsCost(30.2, {Base30:0,R30_70:2}) = %v, want 62", got)
	}
}

func TestCalcStoCostCeil(t *testing.T) {
	// 3.6 ceil => 4 => 3.5 + (4-1)*1.8 = 3.5+5.4 = 8.9
	got, ok := CalcStoCost(3.6, 3.5, 1.8)
	want := 3.5 + 3*1.8
	if got != want || !ok {
		t.Errorf("CalcStoCost(3.6,3.5,1.8) = (%v, %v), want (%v, true)", got, ok, want)
	}
}
