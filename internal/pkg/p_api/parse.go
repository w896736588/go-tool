package p_api

// CurlStruct 表示解析后的 curl 命令结构
type CurlStruct struct {
	Method        string            `json:"method"`
	Url           string            `json:"url"`
	Protocol      string            `json:"protocol"`
	Desc          string            `json:"desc"`
	ContentType   string            `json:"content_type"`
	Headers       map[string]string `json:"headers"`
	QueryParams   string            `json:"query_params"`
	BodyForm      []KeyValue        `json:"body_form"`       // application/x-www-form-urlencoded
	BodyJson      string            `json:"body_json"`       // application/json
	BodyMultiForm string            `json:"body_multi_form"` // multipart/form-data
}

type ParseCurl struct {
	CurlStruct CurlStruct
	Curl       string
}

// NewParseCurl 解析curl命令字符串
func NewParseCurl(curl string) ParseCurl {
	return ParseCurl{
		Curl: curl,
	}
}

func (p *ParseCurl) ParseCurl() error {

	return nil
}
