package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func buildUI() error {
	fs := Size{winW, winH}
	return MainWindow{
		AssignTo: &mw,
		Title:    "快递运费报价",
		MinSize:  Size{800, 500},
		Size:     fs,
		Layout:   VBox{MarginsZero: true, SpacingZero: true},
		Children: []Widget{
			// === Title bar ===
			Composite{
				AssignTo: &titleBar,
				Layout:   HBox{Margins: Margins{12, 8, 12, 8}},
				Children: []Widget{
					Label{Text: "📦 快递运费报价", Font: Font{PointSize: 13, Bold: true}, TextColor: walk.RGB(255, 255, 255)},
					HSpacer{},
					Label{Text: "v2.0", Font: Font{PointSize: 9}, TextColor: walk.RGB(255, 255, 255)},
				},
			},
			// === Zone 1: Address ===
			Composite{
				Layout: HBox{Margins: Margins{10, 8, 10, 0}, Spacing: 6},
				Children: []Widget{
					LineEdit{AssignTo: &inpAddress, CueBanner: "粘贴收件地址，点击AI解析", StretchFactor: 2},
					PushButton{AssignTo: &btnAnalyze, Text: "🤖 AI 解析", OnClicked: onAnalyzeAddress},
					Label{AssignTo: &lblDest, Text: "▸ —", Font: Font{Bold: true}},
				},
			},
			// === Zone 2: Package ===
			Composite{
				Layout: VBox{Margins: Margins{10, 6, 10, 6}, Spacing: 4},
				Children: []Widget{
					// Mode radio buttons
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 12},
						Children: []Widget{
							RadioButton{AssignTo: &rbDim, Text: "长宽高+实重", OnClicked: func() { switchMode(0) }},
							RadioButton{AssignTo: &rbWeight, Text: "仅实重", OnClicked: func() { switchMode(1) }},
							RadioButton{AssignTo: &rbVolume, Text: "仅体积", OnClicked: func() { switchMode(2) }},
						},
					},
					// Dim+Weight inputs (default visible)
					Composite{
						AssignTo: &dimInputs,
						Layout:   HBox{MarginsZero: true, Spacing: 4},
						Children: []Widget{
							Label{Text: "长(cm)"}, LineEdit{AssignTo: &inpLength},
							Label{Text: "宽(cm)"}, LineEdit{AssignTo: &inpWidth},
							Label{Text: "高(cm)"}, LineEdit{AssignTo: &inpHeight},
							Label{Text: "件数"}, LineEdit{AssignTo: &inpQty, Text: "1"},
							Label{Text: "实重(kg)"}, LineEdit{AssignTo: &inpActual},
							PushButton{Text: "+ 添加", OnClicked: onAddPackage},
						},
					},
					// Weight-only inputs (default hidden)
					Composite{
						AssignTo: &wtInputs, Visible: false,
						Layout:   HBox{MarginsZero: true, Spacing: 4},
						Children: []Widget{
							Label{Text: "实重(kg)"}, LineEdit{AssignTo: &inpActual},
							Label{Text: "件数"}, LineEdit{AssignTo: &inpQty, Text: "1"},
							PushButton{Text: "+ 添加", OnClicked: onAddPackage},
						},
					},
					// Volume-only inputs (default hidden)
					Composite{
						AssignTo: &volInputs, Visible: false,
						Layout:   HBox{MarginsZero: true, Spacing: 4},
						Children: []Widget{
							Label{Text: "体积(cm³)"}, LineEdit{AssignTo: &inpVolume},
							Label{Text: "件数"}, LineEdit{AssignTo: &inpVolQty, Text: "1"},
							PushButton{Text: "+ 添加", OnClicked: onAddPackage},
						},
					},
					Label{AssignTo: &lblPreview, Text: "预览: —", Font: Font{PointSize: 9}},
					// Table
					TableView{
						AssignTo: &tblPackages, Model: model,
						AlternatingRowBG: true, LastColumnStretched: true, StretchFactor: 3,
						Columns: []TableViewColumn{
							{Title: "#", Width: 28, Alignment: AlignCenter},
							{Title: "尺寸", Width: 180},
							{Title: "实重", Width: 62, Alignment: AlignFar},
							{Title: "计费重", Width: 72, Alignment: AlignFar},
						},
					},
					// Bottom bar
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 6},
						Children: []Widget{
							PushButton{AssignTo: &btnClear, Text: "清空", Enabled: false, OnClicked: onClearAll},
							Label{AssignTo: &lblCount, Text: "共 0 件 · 总重 0 kg"},
						},
					},
				},
			},
			// === Zone 3: Price PK cards ===
			Composite{
				Layout: HBox{Margins: Margins{10, 6, 10, 10}, Spacing: 12},
				Children: []Widget{
					// STO card
					GroupBox{
						Title:  "🚚 申通快递",
						Layout: VBox{Margins: Margins{12, 8, 12, 8}, Spacing: 4},
						Children: []Widget{
							Label{AssignTo: &lblStoCost, Text: "¥—", Font: Font{PointSize: 26, Bold: true}, TextColor: walk.RGB(102, 126, 234)},
							Label{AssignTo: &lblStoDetail, Text: "计费重 — kg", Font: Font{PointSize: 9}},
							Label{AssignTo: &lblStoNote, Text: "", Font: Font{PointSize: 9}, TextColor: walk.RGB(230, 126, 34)},
						},
					},
					Label{Text: "vs", Font: Font{PointSize: 18, Bold: true}},
					// BS card
					GroupBox{
						Title:  "📦 百世快运",
						Layout: VBox{Margins: Margins{12, 8, 12, 8}, Spacing: 4},
						Children: []Widget{
							Label{AssignTo: &lblBsCost, Text: "¥—", Font: Font{PointSize: 26, Bold: true}, TextColor: walk.RGB(118, 75, 162)},
							Label{AssignTo: &lblBsDetail, Text: "计费重 — kg", Font: Font{PointSize: 9}},
							Label{AssignTo: &lblRecommend, Text: "", Font: Font{PointSize: 10, Bold: true}, TextColor: walk.RGB(46, 139, 87)},
						},
					},
				},
			},
		},
	}.Create()
}
