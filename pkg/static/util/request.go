package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// get请求
func HttpGet(reqUrl string, header map[string]string) ([]byte, error) {
	c := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	// 添加http header
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err = errors.New(fmt.Sprintf("get response failed: [status_code: %d]", resp.StatusCode))
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// post请求
func HttpPost(reqUrl string, header map[string]string, body string) ([]byte, error) {
	c := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 添加http header
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// fmt.Println(fmt.Sprintf("%+v",req.Form))

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err = errors.New(fmt.Sprintf("get response failed: [status_code: %d]", resp.StatusCode))
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

func HttpPostWithTimeOut(reqUrl string, header map[string]string, timeOutSec time.Duration, body string) ([]byte, error) {
	c := http.Client{Timeout: timeOutSec}
	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 添加http header
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// fmt.Println(fmt.Sprintf("%+v",req.Form))

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err = errors.New(fmt.Sprintf("get response failed: [status_code: %d]", resp.StatusCode))
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}
