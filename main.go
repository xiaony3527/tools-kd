package main

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// ====== Table Model ======

type PackageTableModel struct {
	walk.TableModelBase
	calc *Calc
}

func (m *PackageTableModel) RowCount() int { return len(m.calc.GetPackages()) }

func (m *PackageTableModel) Value(row, col int) interface{} {
	p := m.calc.GetPackages()[row]
	switch col {
	case 0: // #
		return p.ID
	case 1: // 尺寸 (含件数)
		if p.IsDirectVol {
			if p.Quantity > 1 {
				return fmt.Sprintf("直接输入 ×%d件", p.Quantity)
			}
			return "直接输入"
		}
		var dims string
		if p.Length == float64(int(p.Length)) && p.Width == float64(int(p.Width)) && p.Height == float64(int(p.Height)) {
			dims = fmt.Sprintf("%.0f×%.0f×%.0f", p.Length, p.Width, p.Height)
		} else {
			dims = fmt.Sprintf("%.1f×%.1f×%.1f", p.Length, p.Width, p.Height)
		}
		if p.Quantity > 1 {
			return dims + fmt.Sprintf(" ×%d件", p.Quantity)
		}
		return dims
	case 2: // 实重
		if p.ActualWeight > 0 {
			return fmt.Sprintf("%.1f", p.ActualWeight)
		}
		return "—"
	case 3: // 体积(m³)
		return fmt.Sprintf("%.3f", p.Volume/1000000)

	case 4: // 申抛重 (×件数)
		v := StoVolWeight(p.Volume) * p.Quantity
		return fmt.Sprintf("%d", v)
	case 5: // 百抛重 (×件数)
		v := BsVolWeight(p.Volume) * p.Quantity
		return fmt.Sprintf("%d", v)
	case 6: // 申计费 (×件数)
		v := StoBillable(p.Volume, p.ActualWeight) * p.Quantity
		return fmt.Sprintf("%d", v)
	case 7: // 百计费 (×件数)
		v := BsBillable(p.Volume, p.ActualWeight) * p.Quantity
		return fmt.Sprintf("%d", v)
	}
	return ""
}

// ====== Global State ======

var (
	mw           *walk.MainWindow
	calcInst     *Calc
	model        *PackageTableModel
	isDirectMode bool

	// Widgets
	tblPackages *walk.TableView
	lblCount    *walk.Label
	lblSto      *walk.Label
	lblBs       *walk.Label
	lblPreview  *walk.Label
	btnAdd      *walk.PushButton
	btnClear    *walk.PushButton

	inpLength   *walk.LineEdit
	inpWidth    *walk.LineEdit
	inpHeight   *walk.LineEdit
	inpDimQty   *walk.LineEdit
	inpDimActual *walk.LineEdit
	inpVolume   *walk.LineEdit
	inpVolQty   *walk.LineEdit
	inpVolActual *walk.LineEdit

	dimGroup   *walk.Composite
	volGroup   *walk.Composite
	lnkToggle  *walk.LinkLabel

	cmbProvince *walk.ComboBox
	cmbCity     *walk.ComboBox
	lblBsCost   *walk.Label

	inpAddress    *walk.LineEdit
	btnAnalyze    *walk.PushButton
	lblDest       *walk.Label
	lblCostDetail *walk.Label
	lblStoCost    *walk.Label

	titleBar   *walk.Composite
	summaryBar *walk.Composite
)

func main() {
	calcInst = NewCalc()
	model = &PackageTableModel{calc: calcInst}
	isDirectMode = false

	if err := buildUI(); err != nil {
		log.Fatal(err)
	}

	applyStyling()
	setupEvents()
	mw.SetVisible(true)

	// Initialize city list for default province
	onProvinceChanged()

	updateAll()

	mw.Run()
}

// ====== UI Construction ======

const winW, winH = 1100, 520

func buildUI() error {
	fs := Size{winW, winH}
	return MainWindow{
		AssignTo: &mw,
		Title:    "快递体积重计算器",
		MinSize:  fs,
		MaxSize:  fs,
		Size:     fs,
		Layout:   VBox{MarginsZero: true, SpacingZero: true},
		Children: []Widget{
			// Title bar
			Composite{
				AssignTo: &titleBar,
				Layout:   HBox{Margins: Margins{12, 8, 12, 8}},
				Children: []Widget{
					Label{
						Text:      "📦 快递体积重计算器",
						Font:      Font{PointSize: 13, Bold: true},
						TextColor: walk.RGB(255, 255, 255),
					},
					HSpacer{},
					Label{
						Text:      "v1.0",
						Font:      Font{PointSize: 9},
						TextColor: walk.RGB(255, 255, 255),
					},
				},
			},
			// Main dual panel
			HSplitter{
				Children: []Widget{
					// Left panel: package list
					Composite{
						Layout: VBox{Margins: Margins{10, 10, 6, 10}},
						Children: []Widget{
							Composite{
								Layout: HBox{MarginsZero: true},
								Children: []Widget{
									Label{Text: "📋 包裹清单", Font: Font{Bold: true}},
									HSpacer{},
									Label{AssignTo: &lblCount, Text: "共 0 件"},
								},
							},
							TableView{
								AssignTo:            &tblPackages,
								Model:               model,
								AlternatingRowBG:    true,
								LastColumnStretched: true,
								StretchFactor:       3,
								Columns: []TableViewColumn{
									{Title: "#", Width: 28, Alignment: AlignCenter},
									{Title: "尺寸(cm)", Width: 110},
									{Title: "实重kg", Width: 52, Alignment: AlignFar},
									{Title: "体积m³", Width: 72, Alignment: AlignFar},
									{Title: "申抛重", Width: 50, Alignment: AlignFar},
									{Title: "百抛重", Width: 50, Alignment: AlignFar},
									{Title: "申计费", Width: 58, Alignment: AlignFar},
									{Title: "百计费", Width: 58, Alignment: AlignFar},
								},
							},
							// Summary bar
							Composite{
								AssignTo: &summaryBar,
								Layout:   HBox{Margins: Margins{10, 8, 10, 8}},
								MinSize:  Size{0, 50},
								Children: []Widget{
									Composite{
										Layout: VBox{MarginsZero: true, SpacingZero: true},
										Children: []Widget{
											Label{Text: "申通快递 · 计费重", TextColor: walk.RGB(255,255,255), Font: Font{PointSize: 8}},
											Label{AssignTo: &lblSto, Text: "0 kg", TextColor: walk.RGB(255,255,255), Font: Font{PointSize: 18, Bold: true}},
											Label{Text: "系数 8000", TextColor: walk.RGB(255,255,255), Font: Font{PointSize: 7}},
										},
									},
									Composite{
										Layout:  VBox{MarginsZero: true, SpacingZero: true},
										MinSize: Size{2, 0},
									},
									Composite{
										Layout: VBox{MarginsZero: true, SpacingZero: true},
										Children: []Widget{
											Label{Text: "百世快运", TextColor: walk.RGB(255,255,255), Font: Font{PointSize: 8}},
											Label{AssignTo: &lblBs, Text: "0 kg", TextColor: walk.RGB(255,255,255), Font: Font{PointSize: 18, Bold: true}},
											Label{Text: "系数 5000", TextColor: walk.RGB(255,255,255), Font: Font{PointSize: 7}},
										},
									},
								},
							},
							PushButton{
								AssignTo:  &btnClear,
								Text:      "清空全部包裹",
								Enabled:   false,
								OnClicked: onClearAll,
							},
						},
					},
					// Right panel: input
					Composite{
						Layout: VBox{Margins: Margins{6, 10, 10, 10}, Spacing: 8},
						Children: []Widget{
							// AI Address analysis
							GroupBox{
								Title:  "📍 收件地址 (粘贴后点击AI解析)",
								Layout: VBox{Margins: Margins{8, 6, 8, 6}, Spacing: 4},
								Children: []Widget{
									LineEdit{
										AssignTo: &inpAddress,
										CueBanner: "粘贴完整地址，如：浙江省杭州市余杭区仓前街道",
									},
									PushButton{
										AssignTo:  &btnAnalyze,
										Text:      "🤖 AI 解析地址",
										OnClicked: onAnalyzeAddress,
									},
								},
							},
							// Manual destination selection
							GroupBox{
								Title:  "📍 目的地 (手动选择)",
								Layout: VBox{Margins: Margins{8, 6, 8, 6}, Spacing: 4},
								Children: []Widget{
									ComboBox{
										AssignTo:              &cmbProvince,
										Model:                 provinces,
										CurrentIndex:          0,
										OnCurrentIndexChanged: onProvinceChanged,
									},
									ComboBox{
										AssignTo: &cmbCity,
										Model:    []string{},
									},
								},
							},
							// Shipping cost display
							GroupBox{
								Title:  "💰 运费估算",
								Layout: VBox{Margins: Margins{8, 6, 8, 6}, Spacing: 2},
								Children: []Widget{
									Label{AssignTo: &lblDest, Text: "目的地: —", Font: Font{PointSize: 10, Bold: true}},
									Label{AssignTo: &lblStoCost, Text: "申通: —", Font: Font{PointSize: 12, Bold: true}},
									Label{AssignTo: &lblBsCost, Text: "百世: —", Font: Font{PointSize: 12, Bold: true}},
									Label{AssignTo: &lblCostDetail, Text: "", Font: Font{PointSize: 9}},
								},
							},
							// Package input
							GroupBox{
								Title:  "📝 包裹尺寸",
								Layout: VBox{Margins: Margins{10, 10, 10, 10}, Spacing: 5},
								Children: []Widget{
									LinkLabel{
										AssignTo:        &lnkToggle,
										Text:            `<a>切换为体积输入</a>`,
										OnLinkActivated: onToggleMode,
									},
									// Dimension inputs (default visible)
									Composite{
										AssignTo: &dimGroup,
										Layout:   Grid{Columns: 2, Spacing: 6},
										Children: []Widget{
											Label{Text: "长 (cm):"}, LineEdit{AssignTo: &inpLength},
											Label{Text: "宽 (cm):"}, LineEdit{AssignTo: &inpWidth},
											Label{Text: "高 (cm):"}, LineEdit{AssignTo: &inpHeight},
											Label{Text: "件数:"}, LineEdit{AssignTo: &inpDimQty, Text: "1"},
											Label{Text: "实重(kg):"}, LineEdit{AssignTo: &inpDimActual, CueBanner: "可选"},
										},
									},
									// Volume input (hidden by default)
									Composite{
										AssignTo: &volGroup,
										Visible:  false,
										Layout:   Grid{Columns: 2, Spacing: 6},
										Children: []Widget{
											Label{Text: "体积 (cm³):"}, LineEdit{AssignTo: &inpVolume},
											Label{Text: "件数:"}, LineEdit{AssignTo: &inpVolQty, Text: "1"},
											Label{Text: "实重(kg):"}, LineEdit{AssignTo: &inpVolActual, CueBanner: "可选"},
										},
									},
									Label{
										Text: "⏎ Enter 跳转 ｜ 聚焦全选 ｜ 最后回车添加",
										Font: Font{PointSize: 7},
									},
									// Preview
									Label{
										AssignTo: &lblPreview,
										Text:     "体积 — cm³ ｜ 实重 — kg ｜ 抛重 申—/百— ｜ 计费 申—/百— kg",
										Font:     Font{PointSize: 9},
									},
									PushButton{
										AssignTo:  &btnAdd,
										Text:      "+ 添加包裹",
										Enabled:   false,
										Font:      Font{Bold: true},
										OnClicked: onAddPackage,
									},
								},
							},
						},
					},
				},
			},
		},
	}.Create()
}

// ====== Styling ======

func applyStyling() {
	bg, err := walk.NewSolidColorBrush(walk.RGB(102, 126, 234))
	if err != nil {
		log.Fatal(err)
	}
	titleBar.SetBackground(bg)
	summaryBar.SetBackground(bg)
}

// ====== Event Setup ======

func setupEvents() {
	// Focus select-all
	for _, inp := range []*walk.LineEdit{
		inpLength, inpWidth, inpHeight, inpDimQty, inpDimActual,
		inpVolume, inpVolQty, inpVolActual, inpAddress,
	} {
		setupSelectAll(inp)
	}

	// Enter key: tab between fields
	setupEnterKey(inpLength, inpWidth)
	setupEnterKey(inpWidth, inpHeight)
	setupEnterKey(inpHeight, inpDimQty)
	setupEnterKey(inpDimQty, inpDimActual)
	inpDimActual.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyReturn {
			onAddPackage()
		}
	})

	setupEnterKey(inpVolume, inpVolQty)
	setupEnterKey(inpVolQty, inpVolActual)
	inpVolActual.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyReturn {
			onAddPackage()
		}
	})

	// Input change → update preview + validate
	inpLength.TextChanged().Attach(func() { updateAll() })
	inpWidth.TextChanged().Attach(func() { updateAll() })
	inpHeight.TextChanged().Attach(func() { updateAll() })
	inpDimQty.TextChanged().Attach(func() { updateAll() })
	inpDimActual.TextChanged().Attach(func() { updateAll() })
	inpVolume.TextChanged().Attach(func() { updateAll() })
	inpVolQty.TextChanged().Attach(func() { updateAll() })
	inpVolActual.TextChanged().Attach(func() { updateAll() })

	// Table double-click = delete row
	tblPackages.ItemActivated().Attach(onDeleteSelected)

	mw.Closing().Attach(func(cancel *bool, reason walk.CloseReason) {
		walk.App().Exit(0)
	})

	// City selection change
	cmbCity.CurrentIndexChanged().Attach(func() { updateShippingCost() })
}

func setupSelectAll(inp *walk.LineEdit) {
	inp.FocusedChanged().Attach(func() {
		txt := inp.Text()
		if len(txt) > 0 {
			inp.SetTextSelection(0, len(txt))
		}
	})
}

func setupEnterKey(from, to *walk.LineEdit) {
	from.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyReturn {
			to.SetFocus()
		}
	})
}

// ====== Input Helpers ======

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func getInputData() (l, w, h, vol, actual float64, qty int, isDirect bool, valid bool) {
	isDirect = isDirectMode

	if isDirect {
		vol = parseFloat(inpVolume.Text())
		if vol <= 0 {
			return
		}
		qty = parseInt(inpVolQty.Text())
		actual = parseFloat(inpVolActual.Text())
	} else {
		l = parseFloat(inpLength.Text())
		w = parseFloat(inpWidth.Text())
		h = parseFloat(inpHeight.Text())
		if l <= 0 || w <= 0 || h <= 0 {
			return
		}
		qty = parseInt(inpDimQty.Text())
		vol = l * w * h
		actual = parseFloat(inpDimActual.Text())
	}

	if qty < 1 {
		qty = 1
	}
	valid = true
	return
}

// ====== Update Functions ======

func updateAll() {
	updatePreview()
	updateAddButtonState()
}

func updatePreview() {
	l, w, h, vol, actual, _, isDirect, valid := getInputData()
	if !valid {
		lblPreview.SetText("体积 — m³ ｜ 实重 — kg ｜ 抛重 申—/百— ｜ 计费 申—/百— kg")
		return
	}

	var volume float64
	if isDirect {
		volume = vol
	} else {
		volume = l * w * h
	}

	stoVol := StoVolWeight(volume)
	bsVol := BsVolWeight(volume)
	stoBill := StoBillable(volume, actual)
	bsBill := BsBillable(volume, actual)

	lblPreview.SetText(fmt.Sprintf(
		"体积 %.3f m³ ｜ 实重 %.1f kg ｜ 抛重 申%d/百%d kg ｜ 计费 申%d/百%d kg",
		volume/1000000, actual, stoVol, bsVol, stoBill, bsBill,
	))
}

func updateAddButtonState() {
	_, _, _, _, _, _, _, valid := getInputData()
	btnAdd.SetEnabled(valid)
}

func updatePackageList() {
	model.PublishRowsReset()
	lblCount.SetText(fmt.Sprintf("共 %d 件", calcInst.Count()))
	lblSto.SetText(fmt.Sprintf("%d kg", calcInst.TotalSto()))
	lblBs.SetText(fmt.Sprintf("%d kg", calcInst.TotalBs()))
	btnClear.SetEnabled(calcInst.Count() > 0)
	updateShippingCost()
}

// ====== Actions ======

func onAddPackage() {
	l, w, h, vol, actual, qty, isDirect, valid := getInputData()
	if !valid {
		return
	}

	calcInst.AddPackage(l, w, h, vol, actual, qty, isDirect)
	updatePackageList()

	// Clear inputs, focus first field
	if isDirect {
		inpVolume.SetText("")
		inpVolQty.SetText("1")
		inpVolActual.SetText("")
		inpVolume.SetFocus()
	} else {
		inpLength.SetText("")
		inpWidth.SetText("")
		inpHeight.SetText("")
		inpDimQty.SetText("1")
		inpDimActual.SetText("")
		inpLength.SetFocus()
	}

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
	updateShippingCost()
}

func onCityChanged() {
	updateShippingCost()
}

func updateShippingCost() {
	pi := cmbProvince.CurrentIndex()
	ci := cmbCity.CurrentIndex()
	if pi < 0 || ci < 0 || pi >= len(provinces) {
		lblBsCost.SetText("百世: —")
		lblStoCost.SetText("申通: —")
		lblDest.SetText("目的地: —")
		lblCostDetail.SetText("")
		return
	}
	province := provinces[pi]
	cities := cmbCity.Model().([]string)
	if ci >= len(cities) {
		return
	}
	city := cities[ci]

	lblDest.SetText(fmt.Sprintf("目的地: %s %s", province, city))

	// 百世运费 (全重量段)
	p := GetPriceDefault(province, city)
	totalBs := float64(calcInst.TotalBs())
	if totalBs <= 0 {
		lblBsCost.SetText(fmt.Sprintf("百世: ¥%.2f (0kg基础)", p.Base50))
	} else {
		cost := CalcBsCost(totalBs, p)
		lblBsCost.SetText(fmt.Sprintf("百世: ¥%.2f (%.0fkg)", cost, totalBs))
	}

	// 申通运费 (≤50kg)
	totalSto := float64(calcInst.TotalSto())
	firstKg, addKg, ok := GetStoPrice(province)
	if !ok {
		lblStoCost.SetText("申通: 无报价")
	} else if totalSto > 50 {
		lblStoCost.SetText(fmt.Sprintf("申通: >50kg，仅发百世"))
	} else if totalSto <= 0 {
		lblStoCost.SetText(fmt.Sprintf("申通: ¥%.2f (首重)", firstKg))
	} else {
		cost, _ := CalcStoCost(totalSto, firstKg, addKg)
		lblStoCost.SetText(fmt.Sprintf("申通: ¥%.2f (%.0fkg)", cost, totalSto))
	}

	// 单行详情
	tw := math.Max(totalBs, totalSto)
	if tw > 0 {
		lblCostDetail.SetText(fmt.Sprintf("计费重 %.0f kg", tw))
	} else {
		lblCostDetail.SetText("")
	}
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

	btnAnalyze.SetText("🤖 AI 解析地址")
	btnAnalyze.SetEnabled(true)

	if err != nil {
		lblDest.SetText(fmt.Sprintf("解析失败: %v", err))
		return
	}

	if province == "" && city == "" {
		lblDest.SetText("未能识别省份/城市，请手动选择")
		return
	}

	// Auto-select province in dropdown
	for i, p := range provinces {
		if strings.Contains(p, province) || strings.Contains(province, p) {
			cmbProvince.SetCurrentIndex(i)
			onProvinceChanged()

			// Auto-select city
			if city != "" {
				cities := cmbCity.Model().([]string)
				for j, c := range cities {
					if strings.Contains(c, city) || strings.Contains(city, c) {
						cmbCity.SetCurrentIndex(j)
						break
					}
				}
			}
			updateShippingCost()
			return
		}
	}

	lblDest.SetText(fmt.Sprintf("识别: %s %s (未匹配到下拉选项)", province, city))
}

func onToggleMode(link *walk.LinkLabelLink) {
	isDirectMode = !isDirectMode
	if isDirectMode {
		dimGroup.SetVisible(false)
		volGroup.SetVisible(true)
		lnkToggle.SetText(`<a>切换为长宽高输入</a>`)
		inpVolume.SetFocus()
	} else {
		dimGroup.SetVisible(true)
		volGroup.SetVisible(false)
		lnkToggle.SetText(`<a>切换为体积输入</a>`)
		inpLength.SetFocus()
	}
	updateAll()
}
