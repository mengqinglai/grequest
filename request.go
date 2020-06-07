package grequest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const (
	POST    = "POST"
	GET     = "GET"
	HEAD    = "HEAD"
	PUT     = "PUT"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"

	TypeJSON       = "json"
	TypeXML        = "xml"
	TypeUrlencoded = "urlencoded"
	TypeForm       = "form"
	TypeFormData   = "form-data"
	TypeHTML       = "html"
	TypeText       = "text"
	TypeMultipart  = "multipart"
)
const (
	_MaxRetry   = 5
	_MaxLogBody = 1024
)

var Types = map[string]string{
	TypeJSON:       "application/json",
	TypeXML:        "application/xml",
	TypeForm:       "application/x-www-form-urlencoded",
	TypeFormData:   "application/x-www-form-urlencoded",
	TypeUrlencoded: "application/x-www-form-urlencoded",
	TypeHTML:       "text/html",
	TypeText:       "text/plain",
	TypeMultipart:  "multipart/form-data",
}

// A Request is a object storing all request data for client.
type Request struct {
	Client  *http.Client
	Url     string
	Method  string
	Header  http.Header
	Cookies []*http.Cookie

	FormData         url.Values
	QueryData        url.Values
	TargetType       string
	ForceType        string
	BasicAuth        struct{ Username, Password string }
	IsKeepAlive      bool
	ExpectRespStatus int
	RetryCfg         *RetryConfig
}

// Used to create a new Request object.
func NewRequest(client *http.Client) *Request {
	r := &Request{
		Header:      http.Header{},
		FormData:    url.Values{},
		QueryData:   url.Values{},
		Cookies:     make([]*http.Cookie, 0),
		BasicAuth:   struct{ Username, Password string }{},
		Client:      client,
		IsKeepAlive: true,
	}
	// default init
	return r
}

// Just a wrapper to initialize Request instance by method string
func (r *Request) CustomMethod(method, targetUrl string) *Request {
	switch method {
	case POST:
		return r.Post(targetUrl)
	case GET:
		return r.Get(targetUrl)
	case HEAD:
		return r.Head(targetUrl)
	case PUT:
		return r.Put(targetUrl)
	case DELETE:
		return r.Delete(targetUrl)
	case PATCH:
		return r.Patch(targetUrl)
	case OPTIONS:
		return r.Options(targetUrl)
	default:
		r.Method = method
		r.Url = targetUrl
		return r
	}
}

func (r *Request) EnableKeepAlive() *Request {
	r.IsKeepAlive = true
	return r
}

func (r *Request) DisableKeepAlive() *Request {
	r.IsKeepAlive = false
	return r
}

func (r *Request) Get(targetUrl string) *Request {
	r.Method = GET
	r.Url = targetUrl
	return r
}

func (r *Request) Post(targetUrl string) *Request {
	r.Method = POST
	r.Url = targetUrl
	return r
}

func (r *Request) Head(targetUrl string) *Request {
	r.Method = HEAD
	r.Url = targetUrl
	return r
}

func (r *Request) Put(targetUrl string) *Request {
	r.Method = PUT
	r.Url = targetUrl
	return r
}

func (r *Request) Delete(targetUrl string) *Request {
	r.Method = DELETE
	r.Url = targetUrl
	return r
}

func (r *Request) Patch(targetUrl string) *Request {
	r.Method = PATCH
	r.Url = targetUrl
	return r
}

func (r *Request) Options(targetUrl string) *Request {
	r.Method = OPTIONS
	r.Url = targetUrl
	return r
}

// header
func (r *Request) SetHeader(key string, value string) *Request {
	r.Header.Set(key, value)
	return r
}

func (r *Request) AddHeader(key string, value string) *Request {
	r.Header.Add(key, value)
	return r
}

func (r *Request) SetBasicAuth(username string, password string) *Request {
	r.BasicAuth = struct{ Username, Password string }{username, password}
	return r
}

// cookie
func (r *Request) AddCookie(c *http.Cookie) *Request {
	r.Cookies = append(r.Cookies, c)
	return r
}

func (r *Request) AddCookies(cookies []*http.Cookie) *Request {
	r.Cookies = append(r.Cookies, cookies...)
	return r
}

// Query
func (r *Request) SetQueryMap(content map[string]string) *Request {
	for k, v := range content {
		r.SetQueryParam(k, v)
	}
	return r
}

func (r *Request) AddQueryParam(key string, value string) *Request {
	r.QueryData.Add(key, value)
	return r
}

func (r *Request) SetQueryParam(key string, value string) *Request {
	r.QueryData.Set(key, value)
	return r
}

// struct to url.Values
func structToUrlValue(r interface{}) (url.Values, error) {
	m := structs.Map(r)
	values := url.Values{}
	for k, v := range m {
		vStr, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("error map value type:%v, key:%s", v, k)
		}
		values.Set(k, vStr)
	}
	return values, nil
}

func jsonToValues(j []byte) (url.Values, error) {
	m := make(map[string]interface{}, 0)
	values := url.Values{}

	err := json.Unmarshal(j, m)
	if err != nil {
		return nil, err
	}
	for k, v := range m {
		vStr, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("error map value type:%v, key:%s", v, k)
		}
		values.Set(k, vStr)
	}
	return values, nil
}

func mapToValues(m map[string]string) url.Values {
	values := url.Values{}
	for k, v := range m {
		values.Set(k, v)
	}
	return values
}

// retry
func (r *Request) SetRetry(maxRetry int, interval time.Duration, status []int, statusBegin int, statusEnd int) *Request {
	r.RetryCfg = NewRetryConfig(maxRetry, interval, status, statusBegin, statusEnd)
	return r
}

func (r *Request) SetRetryConfig(config *RetryConfig) *Request {
	r.RetryCfg = config
	return r
}

// response
// set a http status of response which you expected
func (r *Request) SetExpectRespStatus(httpStatus int) *Request {
	r.ExpectRespStatus = httpStatus
	return r
}

// send a request without 'body'
func (r *Request) Send() *Result {
	return r.sendRequest(nil)
}

// send file(s)
type File struct {
	Abs     string
	Name    string
	Field   string
	Content []byte
}

//send fileContent if it's not nil or read and send absFile
func (r *Request) SendFile(absFile, fileName, fieldName string, fileContent []byte) *Result {
	f := &File{Abs: absFile, Name: fileName, Field: fileName, Content: fileContent}
	return r.SendFiles([]*File{f})
}

func (r *Request) SendFiles(files []*File) *Result {
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	for _, f := range files {
		fw, err := mw.CreateFormFile(f.Field, f.Name)
		if err != nil {
			return &Result{Err: err}
		}
		if len(f.Content) > 0 {
			fw.Write(f.Content)
		} else {
			body, err := ioutil.ReadFile(f.Abs)
			if err != nil {
				return &Result{Err: err}
			}
			fw.Write(body)
		}
	}

	r.Header.Set("Content-Type", mw.FormDataContentType())
	return r.sendRequest(buf)
}

// send json
func (r *Request) SendJson(content interface{}) *Result {
	r.Header.Set("Content-Type", Types[TypeJSON])

	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		return r.SendString(v.String())
	case reflect.Map, reflect.Struct, reflect.Array, reflect.Slice, reflect.Ptr:
		buf, err := json.Marshal(content)
		if err != nil {
			return &Result{Err: err}
		}
		return r.SendBytes(buf)
	default:
		return &Result{Err: fmt.Errorf("doesn't support content type: %s ", v.Type().String())}
	}
}

// send form
func (r *Request) SendForm(content interface{}) *Result {
	var (
		values url.Values
		err    error
	)
	r.Header.Set("Content-Type", Types[TypeForm])

	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		return r.SendString(v.String())
	case reflect.Struct:
		values, err = structToUrlValue(content)
		if err != nil {
			return &Result{Err: err}
		}
	case reflect.Slice: // only support []byte
		j, ok := content.([]byte)
		if !ok {
			return &Result{Err: fmt.Errorf("doesn't support slice type: %s ", v.Type().String())}
		}
		values, err = jsonToValues(j)
		if err != nil {
			return &Result{Err: err}
		}
	case reflect.Map: // only support map[string]string
		m, ok := content.(map[string]string)
		if !ok {
			return &Result{Err: fmt.Errorf("doesn't support map type: %s ", v.Type().String())}
		}
		values = mapToValues(m)
	default:
		return &Result{Err: fmt.Errorf("doesn't support content type: %s ", v.Type().String())}
	}
	return r.SendString(values.Encode())
}

// send raw
func (r *Request) SendString(body string) *Result {
	return r.SendBytes([]byte(body))
}

func (r *Request) SendBytes(body []byte) *Result {
	bodyReader := bytes.NewReader(body)
	return r.sendRequest(bodyReader)
}

func (r *Request) sendRequest(bodyReader io.Reader) *Result {
	var (
		err error
		req *http.Request
	)

	// Make Request
	req, err = r.makeRequest(bodyReader)
	if err != nil {
		return nil
	}

	// Log details of this request
	// Send request
	result := r.retryRequest(req)

	return result
}

func (r *Request) makeRequest(contentReader io.Reader) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	if req, err = http.NewRequest(r.Method, r.Url, contentReader); err != nil {
		return nil, err
	}

	// set header
	for k, vals := range r.Header {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
		if strings.EqualFold(k, "Host") {
			req.Host = vals[0]
		}
	}

	// set connection
	if req.Header.Get("Connection") == "" {
		if !r.IsKeepAlive {
			req.Header.Set("Connection", "close")
		}
	}

	// Add all querystring from Query func
	q := req.URL.Query()
	for k, v := range r.QueryData {
		for _, vv := range v {
			q.Add(k, vv)
		}
	}
	req.URL.RawQuery = q.Encode()

	// Add basic auth
	if r.BasicAuth != struct{ Username, Password string }{} {
		req.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	// Add cookies
	for _, cookie := range r.Cookies {
		req.AddCookie(cookie)
	}

	return req, nil
}

func (r *Request) retryRequest(req *http.Request) *Result {
	var (
		err    error
		body   []byte
		resp   *http.Response
		result = new(Result)
	)
	aBegin := NowUnix()
	for i := 0; i < _MaxRetry; i++ {
		start := NowUnix()
		resp, err = r.Client.Do(req)
		end := NowUnix()
		traceLog := &TraceLog{RetryId: i, CostTimeMs: end - start, Err: err, Body: make([]byte, 0, 200)}
		if resp != nil {
			body, err = ioutil.ReadAll(resp.Body)
			if err == nil {
				bodyLen := len(body)
				if bodyLen > _MaxLogBody {
					bodyLen = _MaxLogBody
				}
				traceLog.Body = make([]byte, bodyLen)
				copy(traceLog.Body, body)

			}
			traceLog.Status = resp.StatusCode
			traceLog.Err = err
			resp.Body.Close()
		}
		// log request
		result.Logs = append(result.Logs, traceLog)

		if r.RetryCfg == nil || !r.RetryCfg.isContinueRetry(err, traceLog.Status) {
			break
		}
	}
	result.req = req
	result.Resp = resp
	result.Err = err
	result.Body = body
	result.TotalCostTime = NowUnix() - aBegin

	if result.Err == nil {
		if result.Resp == nil {
			result.Err = fmt.Errorf("resp is null")
		} else if r.ExpectRespStatus > 0 && result.Resp.StatusCode != r.ExpectRespStatus {
			result.Err = fmt.Errorf("resp status is: %d, expect is:%d", result.Resp.StatusCode, r.ExpectRespStatus)
		}
	}

	return result
}

func NowUnix() int64 {
	return time.Now().UnixNano()
}
