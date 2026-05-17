package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"runtime"

	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"
)

//go:embed res/index.html
var indexHTML string

func main() {
	runtime.LockOSThread() // Sciter 要求

	// 创建窗口 (960x500, 位置自适应)
	w, err := window.New(
		sciter.SW_TITLEBAR|sciter.SW_RESIZEABLE|sciter.SW_CONTROLS|sciter.SW_MAIN|sciter.SW_ENABLE_DEBUG,
		&sciter.Rect{Left: 100, Top: 100, Right: 1060, Bottom: 600},
	)
	if err != nil {
		log.Fatal("创建窗口失败:", err)
	}

	w.SetTitle("快递体积重计算器")
	calc := NewCalc()

	// ---- 注册 Go 函数供 TIScript 调用 ----

	// 添加包裹: view.addPackage(length, width, height, volume, isDirect, qty)
	w.DefineFunction("addPackage", func(args ...*sciter.Value) *sciter.Value {
		if len(args) < 6 {
			return sciter.NewValue("error: addPackage needs 6 args")
		}
		l := args[0].Float()
		wV := args[1].Float()
		h := args[2].Float()
		vol := args[3].Float()
		isDirect := args[4].Int() == 1
		qty := args[5].Int()
		if qty < 1 {
			qty = 1
		}

		calc.AddPackage(l, wV, h, vol, qty, isDirect)
		return buildPackageListResult(calc)
	})

	// 删除包裹: view.deletePackage(id)
	w.DefineFunction("deletePackage", func(args ...*sciter.Value) *sciter.Value {
		if len(args) < 1 {
			return sciter.NewValue("error: deletePackage needs 1 arg")
		}
		id := args[0].Int()
		calc.DeletePackage(id)
		return buildPackageListResult(calc)
	})

	// 清空包裹: view.clearPackages()
	w.DefineFunction("clearPackages", func(args ...*sciter.Value) *sciter.Value {
		calc.ClearPackages()
		return buildPackageListResult(calc)
	})

	// 实时预览: view.preview(l, w, h, vol, isDirect)
	w.DefineFunction("preview", func(args ...*sciter.Value) *sciter.Value {
		if len(args) < 5 {
			return sciter.NewValue("error: preview needs 5 args")
		}
		l := args[0].Float()
		wV := args[1].Float()
		h := args[2].Float()
		vol := args[3].Float()
		isDirect := args[4].Int() == 1

		var volume float64
		if isDirect {
			volume = vol
		} else {
			volume = l * wV * h
		}

		result := sciter.NewValue()
		result.Set("volume", sciter.NewValue(volume))
		result.Set("stoWeight", sciter.NewValue(StoWeight(volume)))
		result.Set("bsWeight", sciter.NewValue(BsWeight(volume)))
		return result
	})

	// 加载 HTML（嵌入资源，无需外部文件）
	w.LoadHtml(indexHTML, "about:blank")

	w.Show()
	w.Run()
}

// buildPackageListResult 构造含包裹列表和汇总的返回值
func buildPackageListResult(calc *Calc) *sciter.Value {
	result := sciter.NewValue()

	// 包裹数组
	packages := sciter.NewValue()
	for _, p := range calc.GetPackages() {
		item := sciter.NewValue()
		item.Set("id", sciter.NewValue(p.ID))
		item.Set("length", sciter.NewValue(p.Length))
		item.Set("width", sciter.NewValue(p.Width))
		item.Set("height", sciter.NewValue(p.Height))
		item.Set("volume", sciter.NewValue(p.Volume))
		item.Set("quantity", sciter.NewValue(p.Quantity))
		item.Set("isDirectVol", sciter.NewValue(boolToInt(p.IsDirectVol)))
		item.Set("stoWeight", sciter.NewValue(StoWeight(p.Volume)))
		item.Set("bsWeight", sciter.NewValue(BsWeight(p.Volume)))
		packages.Append(item)
	}
	result.Set("packages", packages)

	// 汇总
	result.Set("totalSto", sciter.NewValue(calc.TotalSto()))
	result.Set("totalBs", sciter.NewValue(calc.TotalBs()))
	result.Set("count", sciter.NewValue(calc.Count()))

	// JSON 字符串（兼容备用）
	type PkgJSON struct {
		ID          int     `json:"id"`
		Length      float64 `json:"length"`
		Width       float64 `json:"width"`
		Height      float64 `json:"height"`
		Volume      float64 `json:"volume"`
		Quantity    int     `json:"quantity"`
		IsDirectVol bool    `json:"isDirectVol"`
		StoWeight   int     `json:"stoWeight"`
		BsWeight    int     `json:"bsWeight"`
	}
	type ResultJSON struct {
		Packages []PkgJSON `json:"packages"`
		TotalSto int       `json:"totalSto"`
		TotalBs  int       `json:"totalBs"`
		Count    int       `json:"count"`
	}
	rj := ResultJSON{TotalSto: calc.TotalSto(), TotalBs: calc.TotalBs(), Count: calc.Count()}
	for _, p := range calc.GetPackages() {
		rj.Packages = append(rj.Packages, PkgJSON{
			ID: p.ID, Length: p.Length, Width: p.Width, Height: p.Height,
			Volume: p.Volume, Quantity: p.Quantity, IsDirectVol: p.IsDirectVol,
			StoWeight: StoWeight(p.Volume), BsWeight: BsWeight(p.Volume),
		})
	}
	jsonBytes, _ := json.Marshal(rj)
	result.Set("json", sciter.NewValue(string(jsonBytes)))

	return result
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
