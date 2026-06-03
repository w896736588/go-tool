package gstool

import "github.com/spf13/cast"

type MapStringAny map[string]any

func (m MapStringAny) ToStringInt() map[string]int {
	ret := make(map[string]int, len(m))
	for k, v := range m {
		ret[k] = cast.ToInt(v)
	}
	return ret
}

func (m MapStringAny) ToStringInt64() map[string]int64 {
	ret := make(map[string]int64, len(m))
	for k, v := range m {
		ret[k] = cast.ToInt64(v)
	}
	return ret
}

func (m MapStringAny) ToStringFloat32() map[string]float32 {
	ret := make(map[string]float32, len(m))
	for k, v := range m {
		ret[k] = cast.ToFloat32(v)
	}
	return ret
}

func (m MapStringAny) ToStringFloat64() map[string]float64 {
	ret := make(map[string]float64, len(m))
	for k, v := range m {
		ret[k] = cast.ToFloat64(v)
	}
	return ret
}

func (m MapStringAny) ToStringString() map[string]string {
	ret := make(map[string]string, len(m))
	for k, v := range m {
		ret[k] = cast.ToString(v)
	}
	return ret
}

func (m MapStringAny) ToIntInt() map[int]int {
	ret := make(map[int]int, len(m))
	for k, v := range m {
		key := cast.ToInt(k)
		ret[key] = cast.ToInt(v)
	}
	return ret
}

func (m MapStringAny) ToIntInt64() map[int]int64 {
	ret := make(map[int]int64, len(m))
	for k, v := range m {
		key := cast.ToInt(k)
		ret[key] = cast.ToInt64(v)
	}
	return ret
}

func (m MapStringAny) ToIntFloat32() map[int]float32 {
	ret := make(map[int]float32, len(m))
	for k, v := range m {
		key := cast.ToInt(k)
		ret[key] = cast.ToFloat32(v)
	}
	return ret
}

func (m MapStringAny) ToIntFloat64() map[int]float64 {
	ret := make(map[int]float64, len(m))
	for k, v := range m {
		key := cast.ToInt(k)
		ret[key] = cast.ToFloat64(v)
	}
	return ret
}

func (m MapStringAny) ToIntString() map[int]string {
	ret := make(map[int]string, len(m))
	for k, v := range m {
		key := cast.ToInt(k)
		ret[key] = cast.ToString(v)
	}
	return ret
}

func (m MapStringAny) ToInt64Int() map[int64]int {
	ret := make(map[int64]int, len(m))
	for k, v := range m {
		key := cast.ToInt64(k)
		ret[key] = cast.ToInt(v)
	}
	return ret
}

func (m MapStringAny) ToInt64Int64() map[int64]int64 {
	ret := make(map[int64]int64, len(m))
	for k, v := range m {
		key := cast.ToInt64(k)
		ret[key] = cast.ToInt64(v)
	}
	return ret
}

func (m MapStringAny) ToInt64Float32() map[int64]float32 {
	ret := make(map[int64]float32, len(m))
	for k, v := range m {
		key := cast.ToInt64(k)
		ret[key] = cast.ToFloat32(v)
	}
	return ret
}

func (m MapStringAny) ToInt64Float64() map[int64]float64 {
	ret := make(map[int64]float64, len(m))
	for k, v := range m {
		key := cast.ToInt64(k)
		ret[key] = cast.ToFloat64(v)
	}
	return ret
}

func (m MapStringAny) ToInt64String() map[int64]string {
	ret := make(map[int64]string, len(m))
	for k, v := range m {
		key := cast.ToInt64(k)
		ret[key] = cast.ToString(v)
	}
	return ret
}
