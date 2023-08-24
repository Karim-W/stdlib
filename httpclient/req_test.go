package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetReq(t *testing.T) {
	// Successful Request
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp := interface{}(nil)
	res := Req(baseUrl + "/get").Get()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
	tr := res.GetTraceInfo()
	fmt.Println("trace: ", tr)
	fmt.Println("elaspsed: ", tr.TotalTime.Milliseconds())
}

func TestPostReq(t *testing.T) {
	// Successful Request
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp := interface{}(nil)
	body := map[string]interface{}{
		"foo": "bar",
	}
	res := Req(baseUrl + "/post").
		AddBody(body).
		Post()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
}

func TestPutReq(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	// Successful Request
	resp := interface{}(nil)
	body := map[string]interface{}{
		"foo": "bar",
	}
	res := Req(baseUrl + "/put").
		AddBody(body).
		Put()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
}

func TestPatchReq(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp := interface{}(nil)
	body := map[string]interface{}{
		"foo": "bar",
	}
	res := Req(baseUrl + "/patch").
		AddBody(body).
		Patch()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
}

func TestPatchAsyncReq(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp := interface{}(nil)
	body := map[string]interface{}{
		"foo": "bar",
	}
	resAsync := Req(baseUrl+"/patch").
		AddBody(body).
		AddQueryArray("foo", []string{"bar", "baz"}).
		AddQueryArray("doo", []string{"dar", "daz"}).
		Dev().
		AddHeader("X-Test", "test").
		Begin().
		WithRetries(CONSTANT_BACKOFF, 3, 1*time.Second).
		PatchAsync()
	res := <-resAsync
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
	curl := res.CURL()
	assert.NotNil(t, curl)
	assert.Equal(
		t,
		"curl -X PATCH '"+baseUrl+"/patch?foo=bar&foo=baz&doo=dar&doo=daz' -H 'X-Test: test' -d '{\"foo\":\"bar\"}'",
		curl,
	)
}

func TestDeleteReq(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp := interface{}(nil)
	res := Req(baseUrl+"/delete").
		WithContext(context.Background()).
		WithCookie(&http.Cookie{
			Name:  "test",
			Value: "test",
		}).
		WithLogger(&defaultLogger{}).
		DevFromEnv().
		WithRetries(CONSTANT_BACKOFF, 3, 1*time.Second).
		Del()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
	elapsed := res.GetElapsedTime()
	assert.NotNil(t, elapsed)
	url := res.GetUrl()
	assert.NotNil(t, url)
}

func TestDeleteAsyncReq(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	os.Setenv("DEV_MODE", "true")
	resp := interface{}(nil)
	resAsync := Req(baseUrl + "/delete").
		WithContext(context.Background()).
		WithCookie(&http.Cookie{
			Name:  "test",
			Value: "test",
		}).
		WithLogger(&defaultLogger{}).
		DevFromEnv().
		DelAsync()
	res := <-resAsync
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
	elapsed := res.GetElapsedTime()
	assert.NotNil(t, elapsed)
	url := res.GetUrl()
	assert.NotNil(t, url)
	cookies := res.GetCookies()
	assert.NotNil(t, cookies)
	assert.Contains(t, cookies, &http.Cookie{
		Name:  "test",
		Value: "test",
	})
}

func TestTransactions(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp1 := interface{}(nil)
	resp2 := interface{}(nil)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		res := Req(baseUrl + "/get").Get()
		assert.Equal(t, true, res.IsSuccess())
		assert.Equal(t, 200, res.GetStatusCode())
		err := res.SetResult(&resp1)
		assert.Nil(t, err)
		assert.NotNil(t, resp1)
		err = res.CatchError()
		assert.Nil(t, err)
	}()
	go func() {
		defer wg.Done()
		res := Req(baseUrl + "/get").Get()
		assert.Equal(t, true, res.IsSuccess())
		assert.Equal(t, 200, res.GetStatusCode())
		err := res.SetResult(&resp2)
		assert.Nil(t, err)
		assert.NotNil(t, resp2)
		err = res.CatchError()
		assert.Nil(t, err)
	}()
	wg.Wait()
}

func TestAfterHook(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	hook := func(req *http.Request, res *http.Response, meta HTTPMetadata, err error) {
		fmt.Println("hook called")
		fmt.Println("elapsed(ms): ", meta.EndTime.Sub(meta.StartTime).Milliseconds())
		assert.Nil(t, err)
	}
	resp := interface{}(nil)
	res := Req(baseUrl + "/get").
		AddAfterHook(hook).Get()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
}

// func TestErroneousHttpBeforeHook(t *testing.T) {
// 	hook := func(req *http.Request) error {
// 		return errors.New("error")
// 	}
// 	req := &_HttpRequest{
// 		readOnlyUrl: "https://httpbin.org/get",
// 		headers:     make(http.Header),
// 		traces:      &clientTrace{},
// 		client:      &http.Client{},
// 		ctx:         context.TODO(),
// 		httpHooks: &HTTPHook{
// 			Before: []func(*http.Request) error{hook},
// 			After:  make([]func(*http.Request, *http.Response, HTTPMetadata, error), 0, 2),
// 		},
// 	}
// 	assert.Equal(t, 1, len(req.httpHooks.Before))
// 	// req := Req("https://httpbin.org/get").AddBeforeHook(hook)
// 	// assert.Equal(t, false, req.
// 	res := req.Get()
// 	assert.Equal(t, false, res.IsSuccess())
// 	assert.Equal(t, -1, res.GetStatusCode())
// 	err := res.CatchError()
// 	assert.NotNil(t, err)
// 	assert.Equal(t, "error", err.Error())
// }

func TestGetResponseBody(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	res := Req(baseUrl + "/get").Get()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	byts := res.GetBody()
	assert.NotNil(t, byts)
}

func TestCatchErrorObject(t *testing.T) {
	bt := []byte(`{"message": "error"}`)
	res := &_HttpRequest{
		statusCode: 400,
		resBody:    bt,
	}
	resp := map[string]interface{}{}
	err := res.Catch(&resp)
	assert.Nil(t, err)
	assert.Equal(t, "error", resp["message"])
}

func TestMatchErrorTestMatchError(t *testing.T) {
	res := &_HttpRequest{
		statusCode: 400,
	}
	err := res.CatchError()
	assert.NotNil(t, err)
	assert.Equal(t, "request failed with status code 400", err.Error())
}

// func TestRetires(t *testing.T) {
// 	res := Req("https://httpbin.org/get").WithRetries(
// 		CONSTANT_BACKOFF, 3, time.Second).Get()
// 	assert.Equal(t, true, res.IsSuccess())
// 	assert.Equal(t, 200, res.GetStatusCode())
// }

func TestToCurlFunctionGet(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	res := Req(baseUrl + "/get").Get()
	curl := res.CURL()
	t.Log(curl)
	assert.Equal(t, "curl -X GET '"+baseUrl+"/get'", curl)
}

func TestToCurlFunctionPost(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	req := map[string]interface{}{
		"hello": "world",
	}
	res := Req(baseUrl + "/post").
		AddBody(req).Post()
	curl := res.CURL()
	t.Log(curl)
	assert.Equal(t, "curl -X POST '"+baseUrl+"/post' -d '{\"hello\":\"world\"}'", curl)
}

func TestNewFunc(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	req := map[string]interface{}{
		"hello": "world",
	}
	hook := func(req *http.Request, res *http.Response, meta HTTPMetadata, err error) {
		fmt.Println("hook called")
		fmt.Println("elapsed(ms): ", meta.EndTime.Sub(meta.StartTime).Milliseconds())
		assert.Nil(t, err)
	}
	client := Req(baseUrl + "/post").
		AddAfterHook(hook)
	res := client.AddBody(req).Post()
	curl := res.CURL()
	t.Log(curl)
	assert.Equal(t, "curl -X POST '"+baseUrl+"/post' -d '{\"hello\":\"world\"}'", curl)
	res = client.New(baseUrl + "/get").Get()
	curl = res.CURL()
	t.Log(curl)
	assert.Equal(t, "curl -X GET '"+baseUrl+"/get'", curl)
}

func TestGetAsync(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp := interface{}(nil)
	resAsync := Req(baseUrl + "/get").GetAsync()
	time.Sleep(1 * time.Second)
	res := <-resAsync
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
	tr := res.GetTraceInfo()
	fmt.Println("trace: ", tr)
	fmt.Println("elaspsed: ", tr.TotalTime.Milliseconds())
}

func TestGetAndPostAsync(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	respGet := interface{}(nil)
	respPost := interface{}(nil)
	postReq := map[string]interface{}{
		"hello": "world",
	}
	resAsync := Req(baseUrl+"/get").AddQuery("hello", "world").AddQuery("hello2", "world2").
		AddBearerAuth("hello").GetAsync()
	resAsync2 := Req(
		baseUrl+"/post",
	).AddQueryArray("hello", []string{"world", "world2"}).
		AddBody(postReq).
		AddHeader("hello", "world").
		AddHeaders(map[string]string{
			"hello2": "world2",
			"hello3": "world3",
		}).
		AddBasicAuth("hello", "world").
		PostAsync()
	time.Sleep(1 * time.Second)
	res := <-resAsync
	res2 := <-resAsync2
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&respGet)
	assert.Nil(t, err)
	err = res2.SetResult(&respPost)
	assert.Nil(t, err)
	assert.NotNil(t, respGet)
	assert.NotNil(t, respPost)
	err = res.CatchError()
	assert.Nil(t, err)
	err = res2.CatchError()
	assert.Nil(t, err)
	assert.NotEqual(t, respGet, respPost)
}

func TestPutAsync(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp := interface{}(nil)
	resAsync := ReqCtx(context.Background(), baseUrl+"/put").JSON().
		WithRetries(EXPONENTIAL_BACKOFF, 3, time.Second).
		Dev().
		Dev().
		PutAsync()
	time.Sleep(1 * time.Second)
	res := <-resAsync
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
	tr := res.GetTraceInfo()
	fmt.Println("trace: ", tr)
	fmt.Println("elaspsed: ", tr.TotalTime.Milliseconds())
}

func TestLogger(t *testing.T) {
	resp := _HttpRequest{
		logger: &defaultLogger{},
		err:    errors.New("test error"),
		httpHooks: &HTTPHook{
			After: make([]func(*http.Request, *http.Response, HTTPMetadata, error), 0, 3),
		},
	}
	res := resp.Dev().Get()
	assert.Equal(t, false, res.IsSuccess())
}

func TestGetWithRetries(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp := interface{}(nil)
	res := Req(baseUrl+"/get").
		WithRetries(EXPONENTIAL_BACKOFF, 3, time.Second).
		Get()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
	tr := res.GetTraceInfo()
	fmt.Println("trace: ", tr)
	fmt.Println("elaspsed: ", tr.TotalTime.Milliseconds())
}

func TestInvokePostWithRetires(t *testing.T) {
	baseUrl := os.Getenv("HTTPBIN_URL")
	if baseUrl == "" {
		t.Skip("HTTPBIN_URL not set")
	}
	resp := interface{}(nil)
	postReq := map[string]interface{}{
		"hello": "world",
	}
	resAsync := Req(baseUrl+"/post").
		WithRetries(EXPONENTIAL_BACKOFF, 3, time.Second).
		AddBody(postReq).DevFromEnv().DevFromEnv().
		InvokeAsync(context.Background(), "POST", nil, resp)
	res := <-resAsync
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
	tr := res.GetTraceInfo()
	fmt.Println("trace: ", tr)
	fmt.Println("elaspsed: ", tr.TotalTime.Milliseconds())
}
