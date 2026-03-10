package define

const InfoCrawlTaskStatusDelete = 0
const InfoCrawlTaskStatusNormal = 1

const InfoCrawlPageLoginStatusNo = 0
const InfoCrawlPageLoginStatusOk = 1
const InfoCrawlPageLoginStatusExpired = 2

const InfoCrawlRunStatusRunning = `running`
const InfoCrawlRunStatusSuccess = `success`
const InfoCrawlRunStatusPartialFailed = `partial_failed`
const InfoCrawlRunStatusFailed = `failed`

const InfoCrawlRunPageStatusSuccess = `success`
const InfoCrawlRunPageStatusFailed = `failed`
const InfoCrawlRunPageStatusLoginRequired = `login_required`

const InfoCrawlPlannerActionWait = `wait`
const InfoCrawlPlannerActionClick = `click`
const InfoCrawlPlannerActionExistWait = `exist_wait`
const InfoCrawlPlannerActionNoExistWait = `no_exist_wait`
const InfoCrawlPlannerActionTextContent = `text_content`
const InfoCrawlPlannerActionBoolResult = `bool_result`

const InfoCrawlPlannerActionMaxCount = 8
const InfoCrawlPlannerWaitMaxMillis = 15000
const InfoCrawlPageTextMaxLength = 12000
const InfoCrawlSummaryInputMaxLength = 50000
