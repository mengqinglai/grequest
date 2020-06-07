package grequest

import (
	"encoding/json"
	"net/http"
)

type TraceLog struct {
	RetryId    int
	CostTimeMs int64
	Status     int
	Body       []byte
	Err        error
}

type Result struct {
	req           *http.Request
	Resp          *http.Response
	Body          []byte
	Err           error
	Logs          []*TraceLog
	TotalCostTime int64
}

func (r *Result) LogInfo() *Result {
	//TODO: implement your log here
	return r
}

func (r *Result) Response() (*http.Response, []byte, error) {
	return r.Resp, r.Body, r.Err
}

func (r *Result) JsonTo(v interface{}) (*http.Response, []byte, error) {
	if r.Err != nil {
		return r.Resp, r.Body, r.Err
	}
	err := json.Unmarshal(r.Body, &v)
	return r.Resp, r.Body, err
}
