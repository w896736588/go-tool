package define

const (
	SseAiCode = `ai_code` //废弃
	SseGitLab = `gitlab`  //固定全局唯一
)

const (
	SseEventClean = `[CLEAN]`                   //清除前端的数据
	SseEventLogin = `[LOGIN_USERNAME_PASSWORD]` //通知前端弹窗输入账号密码
	SseDown       = `[DONE]`                    //前端换个行
	SseConnect    = `[CONNECT]`                 //链接已建立
)

type SseData struct {
	SseDistributeId string `json:"sse_distribute_id"` //具体接收业务的id 因为公用一个链接
	Data            any    `json:"data"`              //发送的数据
	Type            string `json:"type"`              //数据类型
}

type SseEvent string

const (
	SseContentTypeMsg        = `msg`         //消息
	SseContentTypeErrorList  = `error_list`  //错误列表
	SseContentTypeFilterList = `filter_list` //拦截数量
	SseContentTypeFilter     = `filter`      //拦截
	SseContentTypeError      = `error`       //错误
)
