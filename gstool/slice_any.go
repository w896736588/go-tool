package gstool

import (
	"github.com/spf13/cast"
)

type SliceAny []any

// ToInt 转为 []int，保留所有元素（包括0）
func (h SliceAny) ToInt() []int {
	ret := make([]int, len(h))
	for i, item := range h {
		ret[i] = cast.ToInt(item)
	}
	return ret
}

func (h SliceAny) ToIntFilter() []int {
	seen := make(map[int]bool)
	var ret []int
	for _, item := range h {
		val := cast.ToInt(item)
		if val == 0 {
			continue
		}
		if seen[val] {
			continue
		}
		seen[val] = true
		ret = append(ret, val)
	}
	return ret
}

func (h SliceAny) ToString() []string {
	ret := make([]string, len(h))
	for i, item := range h {
		ret[i] = cast.ToString(item)
	}
	return ret
}

func (h SliceAny) ToStringFilter() []string {
	seen := make(map[string]bool)
	var ret []string
	for _, item := range h {
		val := cast.ToString(item)
		if val == "" {
			continue
		}
		if seen[val] {
			continue
		}
		seen[val] = true
		ret = append(ret, val)
	}
	return ret
}

// ToInt64 转为 []int64
func (h SliceAny) ToInt64() []int64 {
	ret := make([]int64, len(h))
	for i, item := range h {
		ret[i] = cast.ToInt64(item)
	}
	return ret
}

// ToInt64Filter 转为 []int64，过滤0并去重
func (h SliceAny) ToInt64Filter() []int64 {
	seen := make(map[int64]bool)
	var ret []int64
	for _, item := range h {
		val := cast.ToInt64(item)
		if val == 0 {
			continue
		}
		if seen[val] {
			continue
		}
		seen[val] = true
		ret = append(ret, val)
	}
	return ret
}

// ToFloat32 转为 []float32
func (h SliceAny) ToFloat32() []float32 {
	ret := make([]float32, len(h))
	for i, item := range h {
		ret[i] = cast.ToFloat32(item)
	}
	return ret
}

// ToFloat32Filter 转为 []float32，过滤0.0并去重（注意浮点数去重需谨慎）
func (h SliceAny) ToFloat32Filter() []float32 {
	seen := make(map[float32]bool)
	var ret []float32
	for _, item := range h {
		val := cast.ToFloat32(item)
		if val == 0.0 {
			continue
		}
		if seen[val] {
			continue
		}
		seen[val] = true
		ret = append(ret, val)
	}
	return ret
}

// ToFloat64 转为 []float64
func (h SliceAny) ToFloat64() []float64 {
	ret := make([]float64, len(h))
	for i, item := range h {
		ret[i] = cast.ToFloat64(item)
	}
	return ret
}

// ToFloat64Filter 转为 []float64，过滤0.0并去重（同样注意浮点精度问题）
func (h SliceAny) ToFloat64Filter() []float64 {
	seen := make(map[float64]bool)
	var ret []float64
	for _, item := range h {
		val := cast.ToFloat64(item)
		if val == 0.0 {
			continue
		}
		if seen[val] {
			continue
		}
		seen[val] = true
		ret = append(ret, val)
	}
	return ret
}
