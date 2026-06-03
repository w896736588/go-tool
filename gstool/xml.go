package gstool

import (
	"github.com/sbabiv/xml2map"
	"strings"
)

// XmlToMap xml转为map
func XmlToMap(xmlStr string) (map[string]interface{}, error) {
	decoder := xml2map.NewDecoder(strings.NewReader(xmlStr))
	return decoder.Decode()
}
