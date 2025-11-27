package p_api

const (
	FieldTypeString = "string"
	FieldTypeInt    = "int"
	FieldTypeFloat  = "float"
	FieldTypeBool   = "bool"
	FieldTypeFile   = "file"
)

type ApiDefine struct {
	Method      string            `json:"method"`
	Url         string            `json:"url"`
	Protocol    string            `json:"protocol"`
	Desc        string            `json:"desc"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	QueryParams string            `json:"query_params"`
	BodyForm    []KeyValue        `json:"body_form"` // application/x-www-form-urlencoded
	BodyJson    string            `json:"body_json"` // application/json
}

type KeyValue struct {
	Description string `json:"description"`
	Field       string `json:"field"`
	Type        string `json:"type"`
	Value       string `json:"value"`
}
