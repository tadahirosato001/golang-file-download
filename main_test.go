package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_Proc(t *testing.T) {
	ginEngine := gin.Default()
	ginEngine.GET("/test", csvDL)

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := serveAndUnmarshal(ginEngine, req, nil)
	if err != nil {
		fmt.Printf("%v", err)
	}
	if resp.StatusCode == http.StatusOK {
		fmt.Println("OK")
		fmt.Println("--- response(http header) start---")
		for name, value := range resp.Header {
			fmt.Printf("%v: %v\n", name, value)
		}
		fmt.Println("--- response(http header) end ---")
	} else {
		fmt.Println("---Error---")
		fmt.Println("StatusCode:", string(rune(resp.StatusCode)))
	}
}

func serveAndUnmarshal(ginRuter *gin.Engine, req *http.Request, result interface{}) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	ginRuter.ServeHTTP(recorder, req)

	if recorder.Code-299 > 0 {
		err := fmt.Errorf("Request: %v, Body: %s, Failure code: %d", req, recorder.Body.Bytes(), recorder.Code)
		// still unmarshal the result in case an error was expected
		_ = json.Unmarshal(recorder.Body.Bytes(), result)
		return recorder.Result(), err
	}

	if result != nil {
		if err := json.Unmarshal(recorder.Body.Bytes(), result); err != nil {
			formattedErr := fmt.Errorf("Request: %v, Body: %s, Unmarshalling error: %s", req, recorder.Body.Bytes(), err)
			return recorder.Result(), formattedErr
		}
	}
	return recorder.Result(), nil
}
