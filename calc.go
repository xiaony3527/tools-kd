package main

import "math"

// Package 表示一个包裹
type Package struct {
	ID          int
	Length      float64 // 长(cm)
	Width       float64 // 宽(cm)
	Height      float64 // 高(cm)
	Volume      float64 // 体积(cm³)
	Quantity    int     // 件数
	IsDirectVol bool    // true=直接体积模式
}

// Calc 包裹列表管理器
type Calc struct {
	packages []Package
	nextID   int
}

func NewCalc() *Calc {
	return &Calc{
		packages: make([]Package, 0),
		nextID:   1,
	}
}

// Coefficients
const (
	StoCoefficient = 8000.0
	BsCoefficient  = 5000.0
)

// StoWeight 申通体积重 = Ceil(体积/8000)
func StoWeight(volume float64) int {
	return int(math.Ceil(volume / StoCoefficient))
}

// BsWeight 百世体积重 = Ceil(体积/5000)
func BsWeight(volume float64) int {
	return int(math.Ceil(volume / BsCoefficient))
}

// CalcVolume 根据长宽高计算体积（长宽高模式）
func CalcVolume(l, w, h float64) float64 {
	return l * w * h
}

// AddPackage 添加包裹
func (c *Calc) AddPackage(l, w, h, vol float64, qty int, isDirect bool) []Package {
	p := Package{
		ID:          c.nextID,
		Length:      l,
		Width:       w,
		Height:      h,
		Volume:      vol,
		Quantity:    qty,
		IsDirectVol: isDirect,
	}
	c.packages = append(c.packages, p)
	c.nextID++
	return c.packages
}

// DeletePackage 删除包裹
func (c *Calc) DeletePackage(id int) []Package {
	idx := -1
	for i, p := range c.packages {
		if p.ID == id {
			idx = i
			break
		}
	}
	if idx >= 0 {
		c.packages = append(c.packages[:idx], c.packages[idx+1:]...)
	}
	return c.packages
}

// ClearPackages 清空所有包裹
func (c *Calc) ClearPackages() {
	c.packages = make([]Package, 0)
	c.nextID = 1
}

// GetPackages 获取所有包裹
func (c *Calc) GetPackages() []Package {
	return c.packages
}

// TotalSto 申通体积重汇总
func (c *Calc) TotalSto() int {
	total := 0
	for _, p := range c.packages {
		total += StoWeight(p.Volume) * p.Quantity
	}
	return total
}

// TotalBs 百世体积重汇总
func (c *Calc) TotalBs() int {
	total := 0
	for _, p := range c.packages {
		total += BsWeight(p.Volume) * p.Quantity
	}
	return total
}

// Count 包裹总件数
func (c *Calc) Count() int {
	n := 0
	for _, p := range c.packages {
		n += p.Quantity
	}
	return n
}
