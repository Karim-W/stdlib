package httpclient

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReq(t *testing.T) {
	// Successful Request
	resp := interface{}(nil)
	res := Req("https://httpbin.org/get").Get()
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
	resp := interface{}(nil)
	body := map[string]interface{}{
		"foo": "bar",
	}
	res := Req("https://httpbin.org/post").
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
	// Successful Request
	resp := interface{}(nil)
	body := map[string]interface{}{
		"foo": "bar",
	}
	res := Req("https://httpbin.org/put").
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
	resp := interface{}(nil)
	body := map[string]interface{}{
		"foo": "bar",
	}
	res := Req("https://httpbin.org/patch").
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

func TestDeleteReq(t *testing.T) {
	resp := interface{}(nil)
	res := Req("https://httpbin.org/delete").Del()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
}

func TestTransactions(t *testing.T) {
	resp1 := interface{}(nil)
	resp2 := interface{}(nil)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		res := Req("https://httpbin.org/get").Get()
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
		res := Req("https://httpbin.org/get").Get()
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
	hook := func(req *http.Request, res *http.Response, err error) {
		fmt.Println("hook called")
		assert.Nil(t, err)
	}
	resp := interface{}(nil)
	res := Req("https://httpbin.org/get").AddAfterHook(hook).Get()
	assert.Equal(t, true, res.IsSuccess())
	assert.Equal(t, 200, res.GetStatusCode())
	err := res.SetResult(&resp)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	err = res.CatchError()
	assert.Nil(t, err)
}

func TestCatchErrorObject(t *testing.T) {
	bt := []byte(`{"message": "error"}`)
	res := &_HttpRequest{
		statusCode: 400,
		resBody:    &bt,
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
