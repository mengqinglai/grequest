package grequest

import (
	"net/http"

	//"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

var (
	urlStr = "http://your_test_url"
)

func TestGetRequest(t *testing.T) {
	// init global client
	SetDefaultClient(NewClient(time.Second * 30))
	query := make(map[string]string, 0)
	query["q1"] = "v1"
	query["q2"] = "v2"
	Convey("request_get", t, func() {
		resp, body, err := NewRequest(DefaultClient()).
			SetHeader("TestHeader", "header").
			AddHeader("TestHeader", "header2").
			SetQueryMap(query).
			SetRetryDefault().
			DisableKeepAlive().
			SetExpectRespStatus(http.StatusOK).
			Get(urlStr).
			Send().
			Response()
		So(err, ShouldBeNil)
		t.Logf("status:%d, body:%s", resp.StatusCode, string(body))
	})
}

func TestPostRequest(t *testing.T) {
	// init global client
	SetDefaultClient(NewClient(time.Second * 30))
	type Req struct {
		Name string
		Age  int
	}
	req := make(map[string]string, 0)
	req["key1"] = "value1"
	req["key2"] = "value2"
	m := make(map[string]string, 0)
	Convey("request_post_form", t, func() {
		resp, body, err := NewRequest(DefaultClient()).
			SetRetryDefault().
			Post(urlStr + "form").
			SendForm("fa=test&fb=test").
			Response()
		So(err, ShouldBeNil)
		t.Logf("status:%d, body:%s", resp.StatusCode, string(body))
	})

	Convey("request_post_json", t, func() {
		resp, body, err := NewRequest(DefaultClient()).
			SetHeader("TestHeader", "header").
			AddHeader("TestHeader", "header2").
			SetRetryDefault().
			Post(urlStr + "json").
			SendJson(&Req{Name: "test", Age: 10}).
			JsonTo(m)
		So(err, ShouldBeNil)
		t.Logf("status:%d, body:%s", resp.StatusCode, string(body))
	})

	Convey("request_post_file", t, func() {
		result := NewRequest(DefaultClient()).
			SetRetryDefault().
			DisableKeepAlive().
			Post(urlStr + "file").
			SendFile("./test.txt", "test.txt", "test_field", nil).
			LogInfo()
		resp, body, err := result.Response()
		So(err, ShouldBeNil)
		t.Logf("status:%d, body:%s", resp.StatusCode, string(body))
	})

}
