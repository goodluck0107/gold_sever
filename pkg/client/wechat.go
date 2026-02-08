package client

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"net/url"
	"strings"
)

type WeixinClient struct {
	AppId     string
	AppSecret string
}

func NewWeixinClient(wxAppId, wxAppSecret string) *WeixinClient {
	weixinClient := new(WeixinClient)
	weixinClient.AppId = wxAppId
	weixinClient.AppSecret = wxAppSecret
	return weixinClient
}

type weixinAccessToken struct {
	ErrCode      int    `json:"errcode"` // errcode
	ErrMsg       string `json:"errmsg"`  // errmsg
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Openid       string `json:"openid"`
	Scope        string `json:"scope"`
}

type weixinAppletAccessToken struct {
	ErrCode    int    `json:"errcode"` // errcode
	ErrMsg     string `json:"errmsg"`  // errmsg
	OpenId     string `json:"openid"`  // 普通用户的标识，对当前开发者帐号唯一
	UnionId    string `json:"unionid"` // 用户统一标识。针对一个微信开放平台帐号下的应用，同一用户的unionid是唯一的。
	SessionKey string `json:"session_key"`
}

type weixinInfoResult struct {
	OpenId    string `json:"openId"`
	Nickname  string `json:"nickName"`
	Gender    int    `json:"gender"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	AvatarUrl string `json:"avatarUrl"`
	UnionId   string `json:"unionId"`
}

/*
	Error message:
	{
	  "errcode": 40029,
	  "errmsg": "invalid code, hints: [ req_id: Oq1BNa0722ns86 ]"
	}
*/
type WeixinUserInfo struct {
	ErrCode    int      `json:"errcode"`    // errcode
	ErrMsg     string   `json:"errmsg"`     // errmsg
	OpenId     string   `json:"openid"`     // 普通用户的标识，对当前开发者帐号唯一
	Nickname   string   `json:"nickname"`   // 普通用户昵称
	Sex        int      `json:"sex"`        // 普通用户性别，1为男性，2为女性
	Province   string   `json:"province"`   // 普通用户个人资料填写的省份
	City       string   `json:"city"`       // 普通用户个人资料填写的城市
	Country    string   `json:"country"`    // 国家，如中国为CN
	Headimgurl string   `json:"headimgurl"` // 用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0为默认值，代表640*640正方形头像），用户没有头像时该项为空
	Privilege  []string `json:"privilege"`  // 用户特权信息，json数组，如微信沃卡用户为（chinaunicom）
	UnionId    string   `json:"unionid"`    // 用户统一标识。针对一个微信开放平台帐号下的应用，同一用户的unionid是唯一的。
}

// 小程序微信用户信息
type WeiXinAppletUserInfo struct {
	Nickname  string `json:"nickName"`
	Gender    int    `json:"gender"`
	Language  string `json:"language"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	AvatarUrl string `json:"avatarUrl"`
}

// 获取access_token
func getWeixinAccessToken(code string, c *WeixinClient) (*weixinAccessToken, error) {
	values := url.Values{}
	values.Add("appid", c.AppId)
	values.Add("secret", c.AppSecret)
	values.Add("code", code)
	values.Add("grant_type", "authorization_code")
	reqUrl := "https://api.weixin.qq.com/sns/oauth2/access_token?" + values.Encode()

	response := new(weixinAccessToken)
	data, err := util.HttpGet(reqUrl, nil)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}
	err = json.Unmarshal(data, response)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}

	if response.ErrCode != 0 {
		err = errors.New(response.ErrMsg)
		xlog.Logger().Errorln(err)
		return nil, err
	}
	return response, nil
}

// 小程序获取unionid
func getWeixinAppletAccessToken(code string, c *WeixinClient) (*weixinAppletAccessToken, error) {
	values := url.Values{}
	values.Add("appid", c.AppId)
	values.Add("secret", c.AppSecret)
	values.Add("js_code", code)
	values.Add("grant_type", "authorization_code")
	reqUrl := "https://api.weixin.qq.com/sns/jscode2session?" + values.Encode()

	response := new(weixinAppletAccessToken)
	data, err := util.HttpGet(reqUrl, nil)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}
	err = json.Unmarshal(data, response)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}

	if response.ErrCode != 0 {
		err = errors.New(response.ErrMsg)
		xlog.Logger().Errorln(err)
		return nil, err
	}
	return response, nil
}

func (c *WeixinClient) GetWeixinUserInfoByAccessToken(openid, accessToken string) (*WeixinUserInfo, error) {
	values := url.Values{}
	values.Add("openid", openid)
	values.Add("access_token", accessToken)
	values.Add("lang", "zh_CN")
	reqUrl := "https://api.weixin.qq.com/sns/userinfo?" + values.Encode()

	response := new(WeixinUserInfo)
	data, err := util.HttpGet(reqUrl, nil)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}
	err = json.Unmarshal(data, response)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}

	if response.ErrCode != 0 {
		err = errors.New(response.ErrMsg)
		xlog.Logger().Errorln(err)
		return nil, err
	}
	return response, nil
}

func (c *WeixinClient) GetWeixinAppletUserInfoByAccessToken(openid, accessToken string) (*WeixinUserInfo, error) {
	values := url.Values{}
	values.Add("openid", openid)
	values.Add("access_token", accessToken)
	values.Add("lang", "zh_CN")
	reqUrl := "https://api.weixin.qq.com/sns/userinfo?" + values.Encode()

	response := new(WeixinUserInfo)
	data, err := util.HttpGet(reqUrl, nil)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}
	err = json.Unmarshal(data, response)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}

	if response.ErrCode != 0 {
		err = errors.New(response.ErrMsg)
		xlog.Logger().Errorln(err)
		return nil, err
	}
	return response, nil
}

// 应用授权
func (c *WeixinClient) GetWeixinUserInfo(code string) (*WeixinUserInfo, error) {
	accessToken, err := getWeixinAccessToken(code, c)
	if err != nil {
		return nil, err
	}

	return c.GetWeixinUserInfoByAccessToken(accessToken.Openid, accessToken.AccessToken)
}

// 小程序授权
func (c *WeixinClient) GetAppletWeixinUserInfo(code string, rawdata string) (*WeixinUserInfo, error) {
	accessToken, err := getWeixinAppletAccessToken(code, c)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}

	// 快速登录 只用openId信息
	_weixinUserInfo := WeixinUserInfo{
		OpenId:  accessToken.OpenId,
		UnionId: accessToken.UnionId,
	}

	// 解析rawdata
	if len(rawdata) > 0 {
		appletInfo := new(WeiXinAppletUserInfo)
		err = json.Unmarshal([]byte(rawdata), appletInfo)
		if err != nil {
			xlog.Logger().Error(fmt.Sprintf("rawData解析出错:%s", err))
			return nil, err
		}
		_weixinUserInfo.Nickname = appletInfo.Nickname
		_weixinUserInfo.Sex = appletInfo.Gender
		_weixinUserInfo.Province = appletInfo.Province
		_weixinUserInfo.City = appletInfo.City
		_weixinUserInfo.Country = appletInfo.Country
		_weixinUserInfo.Headimgurl = appletInfo.AvatarUrl
	}

	return &_weixinUserInfo, nil
}

// 需要处理空格的情况 替换成加号再转码
func DecryptWXOpenData(sessionKey, encryptData, iv string) (*weixinInfoResult, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(strings.Replace(encryptData, ` `, `+`, -1))
	if err != nil {
		xlog.Logger().Errorln("eve data err: ", err)
		return nil, err
	}
	sessionKeyBytes, err := base64.StdEncoding.DecodeString(strings.Replace(sessionKey, ` `, `+`, -1))
	if err != nil {
		xlog.Logger().Errorln("sessionKey err: ", err)
		return nil, err
	}
	ivBytes, err := base64.StdEncoding.DecodeString(strings.Replace(iv, ` `, `+`, -1))
	if err != nil {
		xlog.Logger().Errorln("iv err: ", err)
		return nil, err
	}
	dataBytes, err := AesDecrypt(decodeBytes, sessionKeyBytes, ivBytes)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}
	xlog.Logger().Info(dataBytes)
	xlog.Logger().Info(string(dataBytes))
	startByte := bytes.Index(dataBytes, []byte{123})
	endByte := bytes.LastIndex(dataBytes, []byte{125})
	if startByte == -1 || endByte == -1 || startByte >= endByte {
		err = errors.New("decrypt wechat eve data failed")
		xlog.Logger().Errorln(err)
		return nil, err
	}

	m := new(weixinInfoResult)
	err = json.Unmarshal(dataBytes[startByte:endByte+1], m)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}
	return m, nil
}

func AesDecrypt(crypted, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}
	//blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	//获取的数据尾端有'/x0e'占位符,去除它
	for i, ch := range origData {
		if ch == '\x0e' {
			origData[i] = ' '
		}
	}
	//{"phoneNumber":"15082726017","purePhoneNumber":"15082726017","countryCode":"86","watermark":{"timestamp":1539657521,"appid":"wx4c6c3ed14736228c"}}//<nil>
	return origData, nil
}
