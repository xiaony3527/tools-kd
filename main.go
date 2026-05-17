package main

import (
	"fmt"
	"log"

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
	case 1: // 尺寸
		if p.IsDirectVol {
			return "直接输入"
		}
		if p.Length == float64(int(p.Length)) && p.Width == float64(int(p.Width)) && p.Height == float64(int(p.Height)) {
			return fmt.Sprintf("%.0f×%.0f×%.0f", p.Length, p.Width, p.Height)
		}
		return fmt.Sprintf("%.1f×%.1f×%.1f", p.Length, p.Width, p.Height)
	case 2: // 实重
		if p.ActualWeight > 0 {
			return fmt.Sprintf("%.1f", p.ActualWeight)
		}
		return "—"
	case 3: // 体积
		return fmt.Sprintf("%.0f", p.Volume)
	case 4: // 申抛重
		return fmt.Sprintf("%d", StoVolWeight(p.Volume))
	case 5: // 百抛重
		return fmt.Sprintf("%d", BsVolWeight(p.Volume))
	case 6: // 申计费
		return fmt.Sprintf("%d", StoBillable(p.Volume, p.ActualWeight))
	case 7: // 百计费
		return fmt.Sprintf("%d", BsBillable(p.Volume, p.ActualWeight))
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
									{Title: "体积cm³", Width: 68, Alignment: AlignFar},
									{Title: "申抛重", Width: 52, Alignment: AlignFar},
									{Title: "百抛重", Width: 52, Alignment: AlignFar},
									{Title: "申计费", Width: 55, Alignment: AlignFar},
									{Title: "百计费", Width: 55, Alignment: AlignFar},
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
											Label{Text: "百世快运 · 计费重", TextColor: walk.RGB(255,255,255), Font: Font{PointSize: 8}},
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
							// Address
							GroupBox{
								Title:  "📍 收件地址",
								Layout: HBox{MarginsZero: true},
								Children: []Widget{
									LineEdit{CueBanner: "省份 城市 区县"},
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
		inpVolume, inpVolQty, inpVolActual,
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
		lblPreview.SetText("体积 — cm³ ｜ 实重 — kg ｜ 抛重 申—/百— ｜ 计费 申—/百— kg")
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
		"体积 %.0f cm³ ｜ 实重 %.1f kg ｜ 抛重 申%d/百%d kg ｜ 计费 申%d/百%d kg",
		volume, actual, stoVol, bsVol, stoBill, bsBill,
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
