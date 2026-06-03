package gshttp

import "net/http"

type StreamInterface interface {
	ReceiveSplit(response *http.Response, responseByte *[]byte)
}
