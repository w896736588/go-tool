package gstool

import "github.com/spf13/cast"

// GsConsMap 好用的不需要加锁的map；方便的基础类型转换
type GsConsMap struct {
	mapCons *HighMap
	keyList []string
}

func GsConsMapNew(shareNum int) *GsConsMap {
	return &GsConsMap{mapCons: HighMapCreate(shareNum), keyList: make([]string, 0)}
}

func (h *GsConsMap) G(key string) *GsCons {
	value, boolRet := h.mapCons.Get(key)
	if boolRet == false {
		return nil
	} else {
		switch value.(type) {
		case GsCons:
			return value.(*GsCons)
		default:
			return ConsNew(value)
		}
	}
}

func (h *GsConsMap) S(key string, value any) {
	h.mapCons.Set(key, value)
	ArrayAppendNotExist(&h.keyList, key)
}

func (h *GsConsMap) D(key string) {
	h.mapCons.Del(key)
	ArrayDeleteValue(&h.keyList, key)
}

func (h *GsConsMap) Each(call func(string, *GsCons)) {
	for _, value := range h.keyList {
		call(cast.ToString(value), h.G(cast.ToString(value)))
	}
}

func (h *GsConsMap) IsZeroLen() bool {
	if len(h.keyList) == 0 {
		return true
	}
	return false
}
