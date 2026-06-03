package gstool

import "github.com/spf13/cast"

type GsCons struct {
	cons any
}

// ConsNewMap 将一个map[string]interface{}转为map[string]*GsCons
func ConsNewMap(source map[string]interface{}) map[string]*GsCons {
	target := make(map[string]*GsCons)
	for k, v := range source {
		target[k] = ConsNew(v)
	}
	return target
}

// ConsNewInterface 将一个interface{} 转为map[string]GsCons
func ConsNewInterface(source interface{}) map[string]*GsCons {
	target := make(map[string]*GsCons)
	mapSource := make(map[string]interface{})
	switch source.(type) {
	case string:
		_ = JsonDecode(cast.ToString(source), &mapSource)
		return ConsNewMap(mapSource)
	case map[string]interface{}:
		mapSource, ok := source.(map[string]interface{})
		if !ok {
			return target
		}
		return ConsNewMap(mapSource)
	default:
		return target
	}
}

func ConsNew(num any) *GsCons {
	return &GsCons{cons: num}
}

func (h *GsCons) ToInt() int {
	if h == nil {
		return 0
	}
	return cast.ToInt(h.cons)
}

func (h *GsCons) ToStr() string {
	if h == nil {
		return ``
	}
	return cast.ToString(h.cons)
}

func (h *GsCons) ToBool() bool {
	if h == nil {
		return false
	}
	return cast.ToBool(h.cons)
}

func (h *GsCons) ToFloat() float32 {
	if h == nil {
		return 0
	}
	return cast.ToFloat32(h.cons)
}

func (h *GsCons) ToInt64() int64 {
	if h == nil {
		return 0
	}
	return cast.ToInt64(h.cons)
}

func (h *GsCons) ToFloat64() float64 {
	if h == nil {
		return 0
	}
	return cast.ToFloat64(h.cons)
}

func (h *GsCons) ToByte() []byte {
	if h == nil {
		return []byte{}
	}
	return []byte(h.ToStr())
}

func (h *GsCons) Value() any {
	if h == nil {
		return nil
	}
	return h.cons
}

func (h *GsCons) IsEmpty() bool {
	if h == nil {
		return true
	}
	switch h.Value().(type) {
	case string:
		if h.ToStr() == `` {
			return true
		}
		break
	case map[string]interface{}:
		mapSource, ok := h.Value().(map[string]interface{})
		if !ok {
			return true
		}
		if len(mapSource) == 0 {
			return true
		}
		break
	case []interface{}:
		source, ok := h.Value().([]interface{})
		if !ok {
			return true
		}
		if len(source) == 0 {
			return true
		}
		break
	default:
		return false
	}

	return false
}

func (h *GsCons) IsZero() bool {
	if h == nil {
		return true
	}
	if h.ToInt() == 0 {
		return true
	}
	return false
}
