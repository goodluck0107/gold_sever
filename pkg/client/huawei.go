package client

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"math/big"
	"net/url"
	"strings"
	"time"
)

// id token 验证（以单独应用接入 仅接入华为账号登录） 官方文档:https://developer.huawei.com/consumer/cn/doc/development/HMSCore-Guides-V5/open-platform-oauth-0000001053629189-V5#ZH-CN_TOPIC_0000001053629189__section1991115414569
// access token 验证（以联运游戏接入） 官方文档:https://developer.huawei.com/consumer/cn/doc/development/HMSCore-Guides/game-login-0000001050121526
// access token 验证（以联运游戏接入） 官方文档:https://developer.huawei.com/consumer/cn/doc/development/HMSCore-References-V5/account-gettokeninfo-0000001050050585-V5

const (
	HW_PUBLIC_KEY_REQ_URL    = "https://oauth-login.cloud.huawei.com/.well-known/openid-configuration" // 从该地址获取公钥地址jwks_uri
	HW_URL                   = "https://accounts.huawei.com"
	HW_APPLICATION_CLIENT_ID = "104380729"
	HW_ACCESS_TOKEN_POST_URL = "https://oauth-api.cloud.huawei.com/rest.php?nsp_fmt=JSON&nsp_svc=huawei.oauth2.user.getTokenInfo"
)

type HuaWeiUserInfo struct {
	OpenId    string `json:"openid"`    // 普通用户的标识
	UnionId   string `json:"unionid"`   // 用户统一标识
	Nickname  string `json:"nickname"`  // 普通用户昵称
	AvatarUrl string `json:"avatarurl"` // 用户头像地址
	//Gender        	int      		`json:"gender"`       		// 普通用户性别 -1 未知 1为男性，2为女性
	AccountFlag int    `json:"accountflag"` // 账号类型 0华为账号 1 AppTouch账号
	IdToken     string `json:"idtoken"`     // 验证所需的id token
}

type HWAccessToken struct {
	ErrMsg   string      `json:"error"`     // 错误信息
	ClientId json.Number `json:"client_id"` // 应用App ID
	ExpireIn int         `json:"expire_in"` // Access Token的过期时间，单位为秒。默认为60分钟。
	UnionId  string      `json:"union_id"`  // 用户的union_id，由用户帐号和应用开发者帐号签名而成，需要应用ID包含com.huawei.android.hms.account.getUnionId权限时才会返回
	OpenId   string      `json:"open_id"`   // 用户的open_id，由用户帐号和应用ID加密生成的，当Access Token为用户级，且入参open_id为OPENID时才返回。
	Scope    string      `json:"scope"`     // 用户授权scope列表，当Access Token为用户级时才返回。
}

// 认证客户端传递过来的AccessToken
func VerifyHWAccessToken(accessToken string) (*HWAccessToken, error) {
	header := make(map[string]string)
	header["Content-Type"] = "application/x-www-form-urlencoded;charset=utf-8"
	body := url.Values{}
	body.Add("open_id", "OPENID")
	body.Add("access_token", accessToken)
	data, err := util.HttpPostWithTimeOut(HW_ACCESS_TOKEN_POST_URL, header, 10*time.Second, body.Encode())
	if err != nil {
		return nil, err
	}

	result := &HWAccessToken{}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	if len(result.ErrMsg) > 0 {
		xlog.Logger().Errorln(result.ErrMsg)
		return nil, errors.New(result.ErrMsg)
	}

	// 验证应用ID是否匹配
	if result.ClientId.String() != HW_APPLICATION_CLIENT_ID {
		errMsg := "verify token fail, client id is not match"
		xlog.Logger().Errorln(errMsg)
		return nil, errors.New(errMsg)
	}

	return result, nil
}

// 认证客户端传递过来的token是否有效
func VerifyHWIdentityToken(cliToken string, cliUserID string) error {
	// 数据由 头部、载荷、签名 三部分组成
	cliTokenArr := strings.Split(cliToken, ".")
	if len(cliTokenArr) < 3 {
		xlog.Logger().Errorln("cliToken Split err ! cliToken = ", cliToken)
		return errors.New("cliToken Split err")
	}

	// 解析cliToken的header获取kid
	cliHeader, err := jwt.DecodeSegment(cliTokenArr[0])
	if err != nil {
		xlog.Logger().Errorln(err.Error())
		return err
	}

	var jHeader JwtHeader
	err = json.Unmarshal(cliHeader, &jHeader)
	if err != nil {
		xlog.Logger().Errorln(err.Error())
		return err
	}

	// 效验pubKey 及 token
	token, err := jwt.ParseWithClaims(cliToken, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return GetHWRSAPublicKey(jHeader.Kid), nil
	})

	if err != nil {
		xlog.Logger().Errorln(err.Error())
		return err
	}

	// 信息验证
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		if claims.StandardClaims.Issuer != HW_URL || claims.Audience != HW_APPLICATION_CLIENT_ID || claims.Subject != cliUserID {
			xlog.Logger().Errorln("verify token eve fail, eve is not match")
			return errors.New("verify token eve fail, eve is not match")
		}
		// here is verify ok !
	} else {
		return errors.New("token claims parse fail")
	}

	return nil
}

// 获取公钥的请求地址
func GetHWPublicKeyUrl() (string, error) {
	response, err := util.HttpGet(HW_PUBLIC_KEY_REQ_URL, nil)
	if err != nil {
		xlog.Logger().Errorln(err.Error())
		return "", err
	}
	hwAuthCfg := make(map[string]interface{})
	err = json.Unmarshal(response, &hwAuthCfg)
	if err != nil {
		xlog.Logger().Errorln(err.Error())
		return "", err
	}

	url, ok := hwAuthCfg["jwks_uri"]
	if !ok {
		return "", errors.New("get public key url fail")
	} else {
		return url.(string), nil
	}
}

// 向华为服务器获取解密signature所需要用的publicKey
func GetHWRSAPublicKey(kid string) *rsa.PublicKey {
	keyUrl, err := GetHWPublicKeyUrl()
	if err != nil {
		xlog.Logger().Errorln(err.Error())
		return nil
	}

	response, err := util.HttpGet(keyUrl, nil)
	if err != nil {
		xlog.Logger().Errorln(err.Error())
		return nil
	}

	var jKeys map[string][]JwtKeys
	err = json.Unmarshal(response, &jKeys)
	if err != nil {
		xlog.Logger().Errorln(err.Error())
		return nil
	}

	// 获取验证所需的公钥
	var pubKey rsa.PublicKey
	// 通过cliHeader的kid比对获取n和e值 构造公钥
	for _, data := range jKeys {
		for _, val := range data {
			if val.Kid == kid {
				n_bin, _ := base64.RawURLEncoding.DecodeString(val.N)
				n_data := new(big.Int).SetBytes(n_bin)

				e_bin, _ := base64.RawURLEncoding.DecodeString(val.E)
				e_data := new(big.Int).SetBytes(e_bin)

				pubKey.N = n_data
				pubKey.E = int(e_data.Uint64())
				break
			}
		}
	}

	if pubKey.E <= 0 {
		xlog.Logger().Errorln("rsa.PublicKey get fail !")
		return nil
	}

	return &pubKey
}
