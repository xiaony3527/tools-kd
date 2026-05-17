package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/lxn/walk"
)

var inputMode InputMode = ModeDimWeight

func setupEvents() {
	// Focus select-all
	for _, inp := range []*walk.LineEdit{inpLength, inpWidth, inpHeight, inpQty, inpActual, inpVolume, inpVolQty, inpAddress} {
		inp.FocusedChanged().Attach(func() {
			txt := inp.Text()
			if len(txt) > 0 {
				inp.SetTextSelection(0, len(txt))
			}
		})
	}

	// Enter key navigation for dim+weight mode
	setupEnterKey(inpLength, inpWidth)
	setupEnterKey(inpWidth, inpHeight)
	setupEnterKey(inpHeight, inpQty)
	setupEnterKey(inpQty, inpActual)
	inpActual.KeyDown().Attach(func(key walk.Key) { if key == walk.KeyReturn { onAddPackage() } })

	// Input changes
	inpLength.TextChanged().Attach(func() { updateAll() })
	inpWidth.TextChanged().Attach(func() { updateAll() })
	inpHeight.TextChanged().Attach(func() { updateAll() })
	inpQty.TextChanged().Attach(func() { updateAll() })
	inpActual.TextChanged().Attach(func() { updateAll() })
	inpVolume.TextChanged().Attach(func() { updateAll() })
	inpVolQty.TextChanged().Attach(func() { updateAll() })

	// Table double-click delete
	tblPackages.ItemActivated().Attach(onDeleteSelected)

	// City change
	cmbCity.CurrentIndexChanged().Attach(func() { updatePriceCards() })

	// Close
	mw.Closing().Attach(func(cancel *bool, reason walk.CloseReason) { walk.App().Exit(0) })

	rbDim.SetChecked(true)
}

func setupEnterKey(from, to *walk.LineEdit) {
	from.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyReturn {
			to.SetFocus()
		}
	})
}

func switchMode(m int) {
	inputMode = InputMode(m)
	dimInputs.SetVisible(inputMode == ModeDimWeight)
	wtInputs.SetVisible(inputMode == ModeWeightOnly)
	volInputs.SetVisible(inputMode == ModeVolumeOnly)
	updateAll()
}

// ====== Input parsing ======
func parseFloat(s string) float64 { var f float64; fmt.Sscanf(s, "%f", &f); return f }
func parseInt(s string) int       { var n int; fmt.Sscanf(s, "%d", &n); return n }

func getInputData() (l, w, h, vol, actual float64, qty int, valid bool) {
	switch inputMode {
	case ModeWeightOnly:
		actual = parseFloat(inpActual.Text())
		if actual <= 0 {
			return
		}
		qty = parseInt(inpQty.Text())
	case ModeVolumeOnly:
		vol = parseFloat(inpVolume.Text())
		if vol <= 0 {
			return
		}
		qty = parseInt(inpVolQty.Text())
	default: // ModeDimWeight
		l = parseFloat(inpLength.Text())
		w = parseFloat(inpWidth.Text())
		h = parseFloat(inpHeight.Text())
		if l <= 0 || w <= 0 || h <= 0 {
			return
		}
		qty = parseInt(inpQty.Text())
		actual = parseFloat(inpActual.Text())
		vol = l * w * h
	}
	if qty < 1 {
		qty = 1
	}
	valid = true
	return
}

// ====== Update ======
func updateAll() {
	updatePreview()
}

func updatePreview() {
	_, _, _, vol, actual, _, valid := getInputData()
	if !valid {
		lblPreview.SetText("预览: —")
		return
	}
	var bill float64
	switch inputMode {
	case ModeWeightOnly:
		bill = actual
	case ModeVolumeOnly:
		bill = math.Ceil(vol / BsCoefficient)
	default:
		bill = math.Max(actual, math.Ceil(vol/BsCoefficient))
	}
	lblPreview.SetText(fmt.Sprintf("预览: 计费重 %.0f kg", bill))
}

func updatePackageList() {
	model.PublishRowsReset()
	cnt := calcInst.Count()
	totalSto := float64(calcInst.TotalSto())
	totalBs := float64(calcInst.TotalBs())
	tw := math.Max(totalSto, totalBs)
	lblCount.SetText(fmt.Sprintf("共 %d 件 · 总重 %.0f kg", cnt, tw))
	btnClear.SetEnabled(cnt > 0)
	updatePriceCards()
}

// ====== Price PK Cards ======
func updatePriceCards() {
	pi := cmbProvince.CurrentIndex()
	ci := cmbCity.CurrentIndex()
	if pi < 0 || ci < 0 || pi >= len(provinces) {
		return
	}
	province := provinces[pi]
	cities := cmbCity.Model().([]string)
	if ci >= len(cities) {
		return
	}
	city := cities[ci]

	// --- 百世 ---
	p := GetPriceDefault(province, city)
	totalBs := float64(calcInst.TotalBs())
	var costBs float64
	if totalBs > 0 {
		costBs = CalcBsCost(totalBs, p)
		lblBsCost.SetText(fmt.Sprintf("¥%.2f", costBs))
		lblBsDetail.SetText(fmt.Sprintf("计费重 %.0f kg", totalBs))
	} else {
		costBs = p.Base30
		lblBsCost.SetText(fmt.Sprintf("¥%.2f", costBs))
		lblBsDetail.SetText("(0kg 基础)")
	}

	// --- 申通 ---
	fk, ak, ok := GetStoPrice(province)
	totalSto := float64(calcInst.TotalSto())
	if !ok {
		lblStoCost.SetText("无报价")
		lblStoDetail.SetText("")
		lblStoNote.SetText("")
	} else if totalSto > 50 {
		cost, _ := CalcStoCost(totalSto, fk, ak)
		lblStoCost.SetText(fmt.Sprintf("¥%.2f", cost))
		lblStoDetail.SetText(fmt.Sprintf("计费重 %.0f kg (上限)", totalSto))
		lblStoNote.SetText("⚠ >50kg 仅发百世")
	} else if totalSto > 0 {
		cost, _ := CalcStoCost(totalSto, fk, ak)
		lblStoCost.SetText(fmt.Sprintf("¥%.2f", cost))
		lblStoDetail.SetText(fmt.Sprintf("计费重 %.0f kg", totalSto))
	} else {
		lblStoCost.SetText(fmt.Sprintf("¥%.2f", fk))
		lblStoDetail.SetText("(首重)")
	}

	// --- Recommend ---
	costSto, _ := CalcStoCost(totalSto, fk, ak)
	if totalBs > 0 && totalSto > 50 {
		lblRecommend.SetText("✅ 推荐 (仅百世)")
	} else if totalBs > 0 && totalSto > 0 && costBs < costSto {
		lblRecommend.SetText("✅ 推荐")
	} else if totalBs > 0 && totalSto > 0 {
		lblRecommend.SetText("")
	} else if totalBs > 0 {
		lblRecommend.SetText("✅ 推荐")
	}
}

// ====== Actions ======
func onAddPackage() {
	l, w, h, vol, actual, qty, valid := getInputData()
	if !valid {
		return
	}
	calcInst.AddPackage(l, w, h, vol, actual, qty, inputMode)
	updatePackageList()
	// Clear
	inpLength.SetText("")
	inpWidth.SetText("")
	inpHeight.SetText("")
	inpQty.SetText("1")
	inpActual.SetText("")
	inpVolume.SetText("")
	inpVolQty.SetText("1")
	updateAll()
}

func onDeleteSelected() {
	idx := tblPackages.CurrentIndex()
	if idx < 0 {
		return
	}
	pkgs := calcInst.GetPackages()
	if idx < len(pkgs) {
		calcInst.DeletePackage(pkgs[idx].ID)
	}
	updatePackageList()
	updateAll()
}

func onClearAll() {
	calcInst.ClearPackages()
	updatePackageList()
	updateAll()
}

// ====== Address ======
func onProvinceChanged() {
	idx := cmbProvince.CurrentIndex()
	if idx < 0 || idx >= len(provinces) {
		return
	}
	cities := GetCities(provinces[idx])
	cmbCity.SetModel(cities)
	if len(cities) > 0 {
		cmbCity.SetCurrentIndex(0)
	}
	updatePriceCards()
}

func onAnalyzeAddress() {
	addr := inpAddress.Text()
	if strings.TrimSpace(addr) == "" {
		walk.MsgBox(mw, "提示", "请先粘贴收件地址", walk.MsgBoxIconInformation)
		return
	}
	btnAnalyze.SetEnabled(false)
	btnAnalyze.SetText("解析中...")
	province, city, err := AnalyzeAddress(addr)
	btnAnalyze.SetText("🤖 AI 解析")
	btnAnalyze.SetEnabled(true)
	if err != nil {
		lblDest.SetText(fmt.Sprintf("解析失败: %v", err))
		return
	}
	if province == "" && city == "" {
		lblDest.SetText("未能识别，请手动选择")
		return
	}
	for i, p := range provinces {
		if strings.Contains(p, province) || strings.Contains(province, p) {
			cmbProvince.SetCurrentIndex(i)
			onProvinceChanged()
			if city != "" {
				cities := cmbCity.Model().([]string)
				for j, c := range cities {
					if strings.Contains(c, city) || strings.Contains(city, c) {
						cmbCity.SetCurrentIndex(j)
						break
					}
				}
			}
			lblDest.SetText(fmt.Sprintf("▸ %s %s", province, city))
			updatePriceCards()
			return
		}
	}
	lblDest.SetText(fmt.Sprintf("▸ %s %s (未匹配)", province, city))
}
