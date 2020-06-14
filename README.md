# grequest
Make it easier to send HTTP requests in Golang

# features
* Method chaining
* Reuse client
* Get/Post/Put/Head/Delete/Patch/Options
* Set - set header and cookie
* JSON - Post JSON request, receive JSON response
* Timeout - set request timeout
* Retry - set retry HTTP status
* BasicAuth - setting authentication header

# Usage
## Get 
#### Example 1
```Golang
//new a client with 1s timeout
client := NewClient(time.Second)  

// send request
resp, body, err := NewRequest(client).Get("https://github.com/").send()
```
#### Example 2
```Golang
resp, body, err := NewRequest(client).
    SetHeader("TestHeader", "header").  //set a new header 
    AddHeader("TestHeader", "header2").  //add(not replace) a header which already exist 
    SetRetry(1, time.Millisecond*50, nil, 500, 504).  //using retry, try again 50ms later when last request return a http status between 500 and 504
    DisableKeepAlive().
    Get("http://test.com"). 
    Send().
    Response()
```

## Post
```Golang
type Req struct {
  Name string
  Age  int
}
respJson := make(map[string]string, 0)
resp, body, err := NewRequest(client).
    SetRetry(1, time.Millisecond*50, []int{500,502}, -1, -1).  //using retry,try again 50ms later when last request return a http status is 500 or 502
    Post("http://test.com").
    SendJson(&Req{Name:"test", Age:10}).
    JsonTo(respJson)
```

## Using a global default client
```Golang
//new a default client with 3s timeout
SetDefaultClient(NewClient(time.Second * 3))

// send request using default client
resp, body, err := NewRequest(DefaultClient()).Get("http://test.com").send()
```
