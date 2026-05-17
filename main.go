package main

import (
	"fmt"
	"log"
	"math"

	"github.com/lxn/walk"
)

const winW, winH = 900, 580

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
	case 1: // 尺寸
		switch p.Mode {
		case ModeWeightOnly:
			if p.Quantity > 1 {
				return fmt.Sprintf("仅实重 ×%d件", p.Quantity)
			}
			return "仅实重"
		case ModeVolumeOnly:
			s := fmt.Sprintf("仅体积: %.0fcm³", p.Volume)
			if p.Quantity > 1 {
				s += fmt.Sprintf(" ×%d件", p.Quantity)
			}
			return s
		default: // ModeDimWeight
			dims := fmt.Sprintf("%.0f×%.0f×%.0f", p.Length, p.Width, p.Height)
			if p.Quantity > 1 {
				return dims + fmt.Sprintf(" ×%d件", p.Quantity)
			}
			return dims
		}
	case 2: // 实重
		if p.ActualWeight > 0 {
			return fmt.Sprintf("%.1f kg", p.ActualWeight)
		}
		return "—"
	case 3: // 计费重 (×件数)
		stoB := StoBillable(p.Volume, p.ActualWeight)
		bsB := BsBillable(p.Volume, p.ActualWeight)
		bill := int(math.Max(float64(stoB), float64(bsB))) * p.Quantity
		return fmt.Sprintf("%d kg", bill)
	}
	return ""
}

// ====== Global State ======

var (
	mw       *walk.MainWindow
	calcInst *Calc
	model    *PackageTableModel

	// --- Title ---
	titleBar *walk.Composite

	// --- Zone 1: Address ---
	inpAddress *walk.LineEdit
	btnAnalyze *walk.PushButton
	lblDest    *walk.Label

	// --- Zone 2: Package ---
	rbDim      *walk.RadioButton
	rbWeight   *walk.RadioButton
	rbVolume   *walk.RadioButton
	inpLength  *walk.LineEdit
	inpWidth   *walk.LineEdit
	inpHeight  *walk.LineEdit
	inpQty     *walk.LineEdit
	inpActual  *walk.LineEdit
	inpVolume  *walk.LineEdit
	inpVolQty  *walk.LineEdit
	dimInputs  *walk.Composite
	volInputs  *walk.Composite
	wtInputs   *walk.Composite
	lblPreview *walk.Label
	tblPackages *walk.TableView
	btnClear   *walk.PushButton
	lblCount   *walk.Label

	// --- Zone 3: Price cards ---
	lblStoCost      *walk.Label
	lblStoDetail    *walk.Label
	lblStoNote      *walk.Label
	lblStoRecommend *walk.Label
	lblBsCost       *walk.Label
	lblBsDetail     *walk.Label
	lblRecommend    *walk.Label

	// --- Destination dropdowns (used with AI) ---
	cmbProvince *walk.ComboBox
	cmbCity     *walk.ComboBox
)

func main() {
	calcInst = NewCalc()
	model = &PackageTableModel{calc: calcInst}

	if err := buildUI(); err != nil {
		log.Fatal(err)
	}

	bg, _ := walk.NewSolidColorBrush(walk.RGB(102, 126, 234))
	titleBar.SetBackground(bg)

	setupEvents()
	onProvinceChanged()
	mw.SetVisible(true)
	mw.Run()
}
