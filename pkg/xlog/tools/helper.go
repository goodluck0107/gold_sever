package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	goSrcGitHubRegexp        = regexp.MustCompile(`sirupsen/logrus(@.*)?/.*.go`)
	goSrcRegexp              = regexp.MustCompile(`facai/log(@.*)?/.*.go`)
	goTestRegexp             = regexp.MustCompile(`facai/log(@.*)?/.*test.go`)
	NumericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
	SqlRegexp                = regexp.MustCompile(`\?`)
)

func ExeName() string {
	args := os.Args
	if len(args) > 0 {
		base := strings.TrimSpace(filepath.Base(os.Args[0]))
		dir, _ := filepath.Abs(filepath.Dir(base))
		exec := strings.Replace(strings.Replace(strings.Replace(strings.Replace(base, "/", "", -1), "\\", "", -1), dir, "", -1), ".exe", "", -1)
		if len(args) > 1 {
			return exec + args[1]
		}
		return exec
	}
	return ""
}

func RunInDocker() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	return true
}

func FileWithLineNum(maxDeep int) string {
	var buf bytes.Buffer
	if maxDeep < 1 {
		return ""
	}
	var deep int
	for i := 2; i < 15 && deep < maxDeep; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && ((!goSrcRegexp.MatchString(file) && !goSrcGitHubRegexp.MatchString(file)) || goTestRegexp.MatchString(file)) {
			deep++
			if deep > 1 {
				buf.WriteString(" \r\n ")
			}
			buf.WriteString(fmt.Sprintf("%v:%v", file, line))
		}
	}
	return buf.String()
}

// post请求
func httpPost(reqUrl string, header map[string]string, body string) ([]byte, error) {
	c := http.Client{}
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

	if resp.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("get response failed: [status_code: %d]", resp.StatusCode))
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// 调用钉钉机器人
func CallDingRobot(token string, info string, atAll bool, phones ...string) error {
	if token == "" {
		return fmt.Errorf("call ding talk error: invaild token")
	}
	var msg struct {
		MsgType string `json:"msgtype"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
		At struct {
			AtMobiles []string `json:"atMobiles"`
			IsAtAll   bool     `json:"isAtAll"`
		} `json:"at"`
	}
	msg.MsgType = "text"
	msg.Text.Content = info
	msg.At.IsAtAll = atAll
	msg.At.AtMobiles = phones

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["encoding"] = "utf-8"

	s, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	body, err := httpPost(fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", token), header, string(s))
	if err != nil {
		return fmt.Errorf("request ding talk api error: %v", err)
	}
	var response struct {
		ErrMag  string `json:"errmsg"`
		ErrCode int    `json:"errcode"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("unmarshal form ding talk response error: %v", err)
	}
	if response.ErrCode != 0 {
		return fmt.Errorf("ding talk api response a error code: %d, msg: %s", response.ErrCode, response.ErrMag)
	}
	return nil
}
