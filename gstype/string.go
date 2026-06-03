package gstype

import "github.com/spf13/cast"

type String string

func (h *String) ToInt() int {
	return cast.ToInt(h)
}

func (h *String) ToStr() string {
	return cast.ToString(h)
}

func (h *String) ToFloat() float32 {
	return cast.ToFloat32(h)
}

func (h *String) ToInt64() int64 {
	return cast.ToInt64(h)
}

func (h *String) ToFloat64() float64 {
	return cast.ToFloat64(h)
}

func (h *String) ToBytes() []byte {
	return []byte(cast.ToString(h))
}

func (h *String) IsEmpty() bool {
	return cast.ToString(h) == ``
}

func (h *String) IsZero() bool {
	return cast.ToInt(h) == 0
}
