package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/lxn/walk"
)

var (
	lastAddrChange time.Time
	pendingAnalyze bool
)

func setupEvents() {
	// Focus select-all
	for _, inp := range []*walk.LineEdit{inpLength, inpWidth, inpHeight, inpVolume, inpActual, inpQty, inpAddress} {
		inp := inp
		inp.FocusedChanged().Attach(func() {
			txt := inp.Text()
			if len(txt) > 0 {
				inp.SetTextSelection(0, len(txt))
			}
		})
	}

	// Enter key nav: 长→宽→高→体积→实重→件数→添加
	setupEnterKey(inpLength, inpWidth)
	setupEnterKey(inpWidth, inpHeight)
	setupEnterKey(inpHeight, inpVolume)
	setupEnterKey(inpVolume, inpActual)
	setupEnterKey(inpActual, inpQty)
	inpQty.KeyDown().Attach(func(key walk.Key) { if key == walk.KeyReturn { onAddPackage() } })

	// Input changes → update preview
	inpLength.TextChanged().Attach(func() { updateAll() })
	inpWidth.TextChanged().Attach(func() { updateAll() })
	inpHeight.TextChanged().Attach(func() { updateAll() })
	inpVolume.TextChanged().Attach(func() { updateAll() })
	inpActual.TextChanged().Attach(func() { updateAll() })
	inpQty.TextChanged().Attach(func() { updateAll() })

	// Address: mark for auto-analyze on input change (triggered by other input activity)
	inpAddress.TextChanged().Attach(func() {
		lastAddrChange = time.Now()
		if len(strings.TrimSpace(inpAddress.Text())) >= 5 {
			pendingAnalyze = true
		}
	})

	// Table double-click delete
	tblPackages.ItemActivated().Attach(onDeleteSelected)

	// City change
	cmbCity.CurrentIndexChanged().Attach(func() { updatePriceCards() })

	// Close
	mw.Closing().Attach(func(cancel *bool, reason walk.CloseReason) { walk.App().Exit(0) })

	// Click-to-copy price labels
	makeCopyable(lblStoCost)
	makeCopyable(lblBsCost)
}

func setupEnterKey(from, to *walk.LineEdit) {
	from.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyReturn {
			to.SetFocus()
		}
	})
}

func makeCopyable(lbl *walk.Label) {
	lbl.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		txt := lbl.Text()
		if txt != "" {
			walk.Clipboard().SetText(strings.TrimPrefix(txt, "¥"))
		}
	})
}

// ====== Input parsing — auto-detect mode ======
func parseFloat(s string) float64 { var f float64; fmt.Sscanf(s, "%f", &f); return f }
func parseInt(s string) int       { var n int; fmt.Sscanf(s, "%d", &n); return n }

func getInputData() (l, w, h, vol, actual float64, qty int, mode InputMode, valid bool) {
	l = parseFloat(inpLength.Text())
	w = parseFloat(inpWidth.Text())
	h = parseFloat(inpHeight.Text())
	actual = parseFloat(inpActual.Text())
	vol = parseFloat(inpVolume.Text())
	qty = parseInt(inpQty.Text())
	if qty < 1 {
		qty = 1
	}

	hasDim := l > 0 && w > 0 && h > 0
	hasVol := vol > 0
	hasActual := actual > 0

	switch {
	case hasVol && !hasDim:
		mode = ModeVolumeOnly
		valid = true
	case hasActual && !hasDim && !hasVol:
		mode = ModeWeightOnly
		valid = true
	case hasDim:
		mode = ModeDimWeight
		vol = l * w * h
		valid = true
	default:
		valid = false
	}
	return
}

// ====== Update ======
func updateAll() {
	updatePreview()
	if pendingAnalyze && time.Since(lastAddrChange) > 1500*time.Millisecond {
		pendingAnalyze = false
		onAnalyzeAddressSilent()
	}
}

func updatePreview() {
	l, w, h, vol, actual, qty, mode, valid := getInputData()
	if !valid {
		lblPreview.SetText("预览: 输入长宽高/体积/实重任一组合")
		return
	}
	_ = l; _ = w; _ = h
	var bill float64
	switch mode {
	case ModeWeightOnly:
		bill = actual * float64(qty)
	case ModeVolumeOnly:
		bill = math.Ceil(vol/BsCoefficient) * float64(qty)
	default:
		vw := math.Ceil(vol / BsCoefficient)
		bill = math.Max(actual, vw) * float64(qty)
	}
	lblPreview.SetText(fmt.Sprintf("预览: 模式=%s 计费重 %.0f kg", modeName(mode), bill))
}

func modeName(m InputMode) string {
	switch m {
	case ModeWeightOnly:
		return "仅实重"
	case ModeVolumeOnly:
		return "仅体积"
	default:
		return "长宽高+实重"
	}
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
	lblStoRecommend.SetText("")
	lblRecommend.SetText("")
	costSto, _ := CalcStoCost(totalSto, fk, ak)
	if totalBs > 0 && totalSto > 50 {
		lblRecommend.SetText("✅ 推荐 (仅百世)")
	} else if totalSto > 0 && totalBs > 0 {
		if costSto < costBs {
			lblStoRecommend.SetText("✅ 推荐")
		} else {
			lblRecommend.SetText("✅ 推荐")
		}
	}
}

// ====== Actions ======
func onAddPackage() {
	l, w, h, vol, actual, qty, mode, valid := getInputData()
	if !valid {
		return
	}
	calcInst.AddPackage(l, w, h, vol, actual, qty, mode)
	updatePackageList()
	// Clear
	inpLength.SetText("")
	inpWidth.SetText("")
	inpHeight.SetText("")
	inpVolume.SetText("")
	inpActual.SetText("")
	inpQty.SetText("1")
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

func onAnalyzeAddressSilent() {
	addr := inpAddress.Text()
	if strings.TrimSpace(addr) == "" {
		return
	}
	province, city, err := AnalyzeAddress(addr)
	if err != nil {
		return
	}
	if province == "" && city == "" {
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
	lblDest.SetText(fmt.Sprintf("▸ %s %s", province, city))
}

func onAnalyzeAddress() {
	btnAnalyze.SetEnabled(false)
	btnAnalyze.SetText("解析中...")
	onAnalyzeAddressSilent()
	btnAnalyze.SetText("🤖 AI 解析")
	btnAnalyze.SetEnabled(true)
}
