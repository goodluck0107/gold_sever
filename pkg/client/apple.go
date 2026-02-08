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
	"strings"
)

// 官方文档:https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_rest_api/verifying_a_user

const (
	APPLE_PUBLIC_KEY_REQ_URL    = "https://appleid.apple.com/auth/keys"
	APPLE_URL                   = "https://appleid.apple.com"
	APPLE_APPLICATION_CLIENT_ID = "com.facai.chongyangmajiang"
)

// 苹果账号信息
type AppleUserInfo struct {
	UnionId  string `json:"unionid"`  // 用户统一标识 同一开发者账号下的不用应用，同一用户的unionid一样的
	Nickname string `json:"nickname"` // 普通用户昵称
}

// 认证客户端传递过来的token是否有效
func VerifyAppleIdentityToken(cliToken string, cliUserID string) error {
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
		return GetAppleRSAPublicKey(jHeader.Kid), nil
	})

	if err != nil {
		xlog.Logger().Errorln(err.Error())
		return err
	}

	// 信息验证
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		if claims.Issuer != APPLE_URL || claims.Audience != APPLE_APPLICATION_CLIENT_ID || claims.Subject != cliUserID {
			errMsg := "verify token fail, eve is not match"
			xlog.Logger().Errorln(errMsg)
			return errors.New(errMsg)
		}
		// here is verify ok !
	} else {
		return errors.New("token claims parse fail")
	}

	return nil
}

// 向苹果服务器获取解密signature所需要用的publicKey
func GetAppleRSAPublicKey(kid string) *rsa.PublicKey {
	response, err := util.HttpGet(APPLE_PUBLIC_KEY_REQ_URL, nil)
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
