package grequest

import "time"

func (r *Request) SetDefaultRetry() *Request {
	r.RetryCfg = NewRetryConfig(1, time.Millisecond*50, nil, 500, 504)
	return r
}
