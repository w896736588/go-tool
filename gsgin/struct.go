package gsgin

type Response struct {
	Errcode int         `json:"ErrCode"`
	Errmsg  string      `json:"ErrMsg"`
	Data    interface{} `json:"Data"`
}
