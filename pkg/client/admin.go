package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

/*
	向php后台发起请求的通用工具集合
*/

type AdminClient struct {
	adminHost string // 后台服务请求域名
	//payHost   string // 支付服务请求域名
}

// 公告
type Notice struct {
	Id           int       `json:"id"`            // id
	KindId       int       `json:"kind_id"`       // 关联子游戏
	ContentType  int       `json:"content_type"`  // 内容类型(文字 or 图片)
	Image        string    `json:"image"`         // 图片内容
	Content      string    `json:"content"`       // 文字内容
	PositionType int       `json:"position_type"` // 公告类型
	ShowType     int       `json:"show_type"`     // 展示类型(每天一次 or 登录一次)
	StartAt      string    `json:"start_at"`      // 开始时间
	EndAt        string    `json:"end_at"`        // 结束时间
	Start        time.Time `json:"-"`
	End          time.Time `json:"-"`
	Flag         bool      `json:"-"` // 是否广播标志位
}

var (
	adminClient *AdminClient
)

const (
	// 短信类型
	SmsTypeLogin         = 1 // 登录
	SmsTypeBind          = 2 // 绑定
	SmsTypeUnbind        = 3 // 解绑
	SmsTypeChangeBind    = 4 // 更换绑定
	SmsTypeResetPassword = 5 // 重置密码
	SmsTypeRegister      = 6 // 注册
)

// 初始化客户端
func InitAdminClient(c *CommonConfig) {
	adminClient = new(AdminClient)
	adminClient.adminHost = c.AdminHost
	//adminClient.adminHost = "http://jsgm.hhkin.com"
	//adminClient.payHost = "http://pay.hhkin.com"
	//adminClient.payAppId = "facai148849639"
	//adminClient.payAppToken = "05ffef168c5089c3fc633c022c5e99d2"
}

// 校验验证码类型
func (self *AdminClient) CheckSmsType(_type uint8) bool {
	if _type == SmsTypeLogin || _type == SmsTypeBind || _type == SmsTypeUnbind || _type == SmsTypeChangeBind || _type == SmsTypeResetPassword || _type == SmsTypeRegister {
		return true
	}
	return false
}

// 发送验证码
func (self *AdminClient) SendSmsCode(mobile string, _type uint8) error {
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	postBody := fmt.Sprintf(`{"data":"{\"mobile\":\"%s\",\"type\":%d}","sign":"","time":"","encrypt":false}`, mobile, _type)
	data, err := util.HttpPost(self.adminHost+"/api/sms/send", header, postBody)
	if err != nil {
		return err
	}

	type response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	resp := new(response)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		err = errors.New(resp.Msg)
		return err
	}
	return nil
}

// 校验验证码
func (self *AdminClient) CheckSmsCode(mobile string, code string, _type int) error {
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	postBody := fmt.Sprintf(`{"data":"{\"mobile\":\"%s\",\"type\":%d,\"code\":\"%s\"}","sign":"","time":"","encrypt":false}`, mobile, _type, code)
	data, err := util.HttpPost(self.adminHost+"/api/sms/check", header, postBody)
	if err != nil {
		return err
	}

	// 去掉bom头
	data = bytes.TrimPrefix(data, []byte{239, 187, 191})

	type response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	resp := new(response)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		err = errors.New(resp.Msg)
		return err
	}
	return nil
}

// // 获取指定类型公告
// func (self *AdminClient) GetNotices(_type int) ([]*Notice, error) {
// 	reqUrl := fmt.Sprintf(self.adminHost+"/api/notice/byType?type=%d", _type)
// 	data, err := util.HttpGet(reqUrl, nil)
// 	if err != nil {
// 		syslog.Logger().Errorln(err)
// 		return nil, err
// 	}
//
// 	// 去掉bom头
// 	data = bytes.TrimPrefix(data, []byte{239, 187, 191})
//
// 	type response struct {
// 		Code int       `json:"code"`
// 		Msg  string    `json:"msg"`
// 		Data []*Notice `json:"data"`
// 	}
//
// 	result := new(response)
// 	err = json.Unmarshal(data, result)
// 	if err != nil {
// 		syslog.Logger().Errorln(err)
// 		return nil, err
// 	}
//
// 	if result.Code != 0 {
// 		err = errors.New(result.Msg)
// 		syslog.Logger().Errorln(err)
// 		return nil, err
// 	}
//
// 	// 处理数据
// 	for _, notice := range result.Data {
// 		notice.Start, _ = time.ParseInLocation("2006-01-02 15:04:05", notice.StartAt, time.Local)
// 		notice.End, _ = time.ParseInLocation("2006-01-02 15:04:05", notice.EndAt, time.Local)
// 	}
//
// 	return result.Data, nil
// }

// 删除指定类型公告
func (self *AdminClient) DeleteNotices(_id int) error {
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	postBody := fmt.Sprintf(`{"data":"{\"id\":%d}","sign":"","time":"","encrypt":false}`, _id)

	data, err := util.HttpPost(self.adminHost+"/api/notice/delete", header, postBody)
	if err != nil {
		return err
	}

	type response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	resp := new(response)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		err = errors.New(resp.Msg)
		return err
	}

	return nil
}

// 获取指定类型公告
func (self *AdminClient) WriteBackResult(_id int, res int) error {

	reqUrl := fmt.Sprintf("%v/api/game/writeBackResult?id=%v&status=%v", self.adminHost, _id, res)
	data, err := util.HttpGet(reqUrl, nil)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	type response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	resp := new(response)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		err = errors.New(resp.Msg)
		return err
	}

	return nil
}

// 注册app
//func (self *AdminClient) ResgisterApp(granturl string) {
//	reqUrl := self.payHost + "/app/register"
//	params := fmt.Sprintf("name=%s&prefix=%s&granturl=%s&backurl=%s", "facai", "facai", granturl, "com.facai://")
//	data, err := util.HttpPost(reqUrl, nil, params)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//
//	type response struct {
//		Code int    `json:"code"`
//		Msg  string `json:"msg"`
//		Data *struct {
//			AppId string `json:"appid"`
//			Token string `json:"token"`
//		} `json:"data"`
//	}
//	resp := new(response)
//	err = json.Unmarshal(data, resp)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//}
//
//// 获取商品列表
//func (self *AdminClient) GetProductList(requestMap map[string]interface{}) (interface{}, error) {
//	header := make(map[string]string)
//	header["Content-Type"] = "application/x-www-form-urlencoded"
//
//	reqUrl := self.payHost + "/product/index"
//	data, err := util.HttpPost(reqUrl, header, self.encryptData(requestMap))
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//
//	type response struct {
//		Code int         `json:"code"`
//		Msg  string      `json:"msg"`
//		Data interface{} `json:"data"`
//	}
//	result := new(response)
//	err = json.Unmarshal(data, result)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	if result.Code != 0 {
//		return nil, errors.New(result.Msg)
//	}
//	return result.Data, nil
//}
//
//// 支付(无选择方式)
//func (self *AdminClient) Pay(requestMap map[string]interface{}) (interface{}, error) {
//	header := make(map[string]string)
//	header["Content-Type"] = "application/x-www-form-urlencoded"
//
//	reqUrl := self.payHost + "/pay/index"
//	data, err := util.HttpPost(reqUrl, header, self.encryptData(requestMap))
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//
//	type response struct {
//		Code int         `json:"code"`
//		Msg  string      `json:"msg"`
//		Data interface{} `json:"data"`
//	}
//	result := new(response)
//	err = json.Unmarshal(data, result)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	if result.Code != 0 {
//		return nil, errors.New(result.Msg)
//	}
//	return result.Data, nil
//}
//
//// 请求数据key-value
//type Pair struct {
//	Key   string
//	Value interface{}
//}
//
//type PairSlice []*Pair
//
//func (s PairSlice) Len() int {
//	return len(s)
//}
//
//func (s PairSlice) Less(i, j int) bool {
//	return s[i].Key < s[j].Key
//}
//
//func (s PairSlice) Swap(i, j int) {
//	s[i], s[j] = s[j], s[i]
//}
//
//// 为请求数据签名
//func (self *AdminClient) encryptData(requestMap map[string]interface{}) string {
//	// 添加随机字符串
//	requestMap["nonce_str"] = "ibuaiVcKdpRxkhJA"
//	// 添加appId
//	requestMap["appid"] = self.payAppId
//	// map转slice
//	var arr PairSlice
//	for k, v := range requestMap {
//		arr = append(arr, &Pair{
//			Key:   k,
//			Value: v,
//		})
//	}
//	// 排序
//	sort.Sort(arr)
//	// 获取排序字符串
//	params := ""
//	for i, p := range arr {
//		params = params + p.Key + `=` + fmt.Sprint(p.Value)
//		if i != len(arr)-1 {
//			params = params + "&"
//		}
//	}
//	// 计算签名
//	signStr := util.MD5(params + "&key=" + self.payAppToken)
//	return params + "&sign=" + signStr
//}
// 通知后台修改数据库
func (self *AdminClient) UpdateUserInfo(args ...interface{}) error {
	header := make(map[string]string)

	postBody := fmt.Sprintf(`circler_id=%v&go_user={"%v":"%v"}&php_circler={"%v":"%v"}`, args...)
	//postBody = fmt.Sprintf(`{\"circler_id\":\"%v\",\"go_user\":"{\"%v\":\"%v\"}",\"php_circler\":"{\"%v\":\"%v\"}`, args...)

	//fmt.Println(postBody)

	data, err := util.HttpPost(self.adminHost+"/api/circler/update", header, postBody)

	if err != nil {
		return err
	}

	fmt.Println(data)

	return nil
}
