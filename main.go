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
	case 0:
		return p.ID
	case 1:
		if p.IsDirectVol {
			return "直接输入"
		}
		// Need to handle integer dimensions display cleanly
		if p.Length == float64(int(p.Length)) && p.Width == float64(int(p.Width)) && p.Height == float64(int(p.Height)) {
			return fmt.Sprintf("%.0f×%.0f×%.0f", p.Length, p.Width, p.Height)
		}
		return fmt.Sprintf("%.1f×%.1f×%.1f", p.Length, p.Width, p.Height)
	case 2:
		return fmt.Sprintf("%.0f", p.Volume)
	case 3:
		return fmt.Sprintf("%d kg", StoWeight(p.Volume))
	case 4:
		return fmt.Sprintf("%d kg", BsWeight(p.Volume))
	}
	return ""
}

// ====== Global State ======

var (
	mw           *walk.MainWindow
	calc         *Calc
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

	inpLength *walk.LineEdit
	inpWidth  *walk.LineEdit
	inpHeight *walk.LineEdit
	inpDimQty *walk.LineEdit
	inpVolume *walk.LineEdit
	inpVolQty *walk.LineEdit

	dimGroup *walk.Composite
	volGroup *walk.Composite
	lnkToggle *walk.LinkLabel

	titleBar *walk.Composite
	summaryBar *walk.Composite
)

func main() {
	calc = NewCalc()
	model = &PackageTableModel{calc: calc}
	isDirectMode = false

	if err := buildUI(); err != nil {
		log.Fatal(err)
	}

	// Imperative setup after window creation
	applyStyling()
	setupEvents()
	mw.SetVisible(true)
	updateAll()

	mw.Run()
}

// ====== UI Construction ======

func buildUI() error {
	return MainWindow{
		AssignTo: &mw,
		Title:    "快递体积重计算器",
		MinSize:  Size{800, 450},
		Size:     Size{960, 500},
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
						Layout: VBox{Margins: Margins{12, 12, 6, 12}},
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
									{Title: "#", Width: 36},
									{Title: "尺寸 (cm)", Width: 120},
									{Title: "体积 cm³", Width: 80, Alignment: AlignFar},
									{Title: "申通", Width: 72, Alignment: AlignFar},
									{Title: "百世", Width: 72, Alignment: AlignFar},
								},
							},
							// Summary bar
							Composite{
								AssignTo: &summaryBar,
								Layout:   HBox{Margins: Margins{10, 8, 10, 8}},
								MinSize:  Size{0, 52},
								Children: []Widget{
									Composite{
										Layout: VBox{MarginsZero: true, SpacingZero: true},
										Children: []Widget{
											Label{Text: "申通快递", TextColor: walk.RGB(255,255,255), Font: Font{PointSize: 8}},
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
						Layout: VBox{Margins: Margins{6, 12, 12, 12}, Spacing: 10},
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
								Title: "📝 包裹尺寸",
								Layout: VBox{Margins: Margins{10, 10, 10, 10}, Spacing: 6},
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
										},
									},
									Label{
										Text: "⏎ Enter 跳转 ｜ 聚焦全选 ｜ 件数处回车添加",
										Font: Font{PointSize: 7},
									},
									// Preview
									Label{
										AssignTo: &lblPreview,
										Text:     "体积 — cm³ ｜ 申通 — kg ｜ 百世 — kg",
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
	// Focus select-all on all input fields
	setupSelectAll(inpLength)
	setupSelectAll(inpWidth)
	setupSelectAll(inpHeight)
	setupSelectAll(inpDimQty)
	setupSelectAll(inpVolume)
	setupSelectAll(inpVolQty)

	// Enter key: tab between fields, submit on qty
	setupEnterKey(inpLength, inpWidth)
	setupEnterKey(inpWidth, inpHeight)
	setupEnterKey(inpHeight, inpDimQty)
	inpDimQty.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyReturn {
			onAddPackage()
		}
	})
	inpVolume.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyReturn {
			inpVolQty.SetFocus()
		}
	})
	inpVolQty.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyReturn {
			onAddPackage()
		}
	})

	// Input change → update preview + validate
	inpLength.TextChanged().Attach(func() { updateAll() })
	inpWidth.TextChanged().Attach(func() { updateAll() })
	inpHeight.TextChanged().Attach(func() { updateAll() })
	inpDimQty.TextChanged().Attach(func() { updateAll() })
	inpVolume.TextChanged().Attach(func() { updateAll() })
	inpVolQty.TextChanged().Attach(func() { updateAll() })

	// Table double-click = delete row
	tblPackages.ItemActivated().Attach(onDeleteSelected)

	// Window close handler
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

func getInputData() (l, w, h, vol float64, qty int, isDirect bool, valid bool) {
	isDirect = isDirectMode

	if isDirect {
		vol = parseFloat(inpVolume.Text())
		if vol <= 0 {
			return
		}
		qty = parseInt(inpVolQty.Text())
	} else {
		l = parseFloat(inpLength.Text())
		w = parseFloat(inpWidth.Text())
		h = parseFloat(inpHeight.Text())
		if l <= 0 || w <= 0 || h <= 0 {
			return
		}
		qty = parseInt(inpDimQty.Text())
		vol = l * w * h
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
	l, w, h, vol, _, isDirect, valid := getInputData()
	if !valid {
		lblPreview.SetText("体积 — cm³ ｜ 申通 — kg ｜ 百世 — kg")
		return
	}

	var volume float64
	if isDirect {
		volume = vol
	} else {
		volume = l * w * h
	}

	sto := StoWeight(volume)
	bs := BsWeight(volume)
	lblPreview.SetText(fmt.Sprintf("体积 %.0f cm³ ｜ 申通 %d kg ｜ 百世 %d kg", volume, sto, bs))
}

func updateAddButtonState() {
	_, _, _, _, _, _, valid := getInputData()
	btnAdd.SetEnabled(valid)
}

func updatePackageList() {
	model.PublishRowsReset()
	lblCount.SetText(fmt.Sprintf("共 %d 件", calc.Count()))
	lblSto.SetText(fmt.Sprintf("%d kg", calc.TotalSto()))
	lblBs.SetText(fmt.Sprintf("%d kg", calc.TotalBs()))
	btnClear.SetEnabled(calc.Count() > 0)
}

// ====== Actions ======

func onAddPackage() {
	l, w, h, vol, qty, isDirect, valid := getInputData()
	if !valid {
		return
	}

	calc.AddPackage(l, w, h, vol, qty, isDirect)
	updatePackageList()

	// Clear inputs, focus first field
	if isDirect {
		inpVolume.SetText("")
		inpVolQty.SetText("1")
		inpVolume.SetFocus()
	} else {
		inpLength.SetText("")
		inpWidth.SetText("")
		inpHeight.SetText("")
		inpDimQty.SetText("1")
		inpLength.SetFocus()
	}

	updateAll()
}

func onDeleteSelected() {
	idx := tblPackages.CurrentIndex()
	if idx < 0 {
		return
	}
	pkgs := calc.GetPackages()
	if idx < len(pkgs) {
		calc.DeletePackage(pkgs[idx].ID)
	}
	updatePackageList()
	updateAll()
}

func onClearAll() {
	calc.ClearPackages()
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
