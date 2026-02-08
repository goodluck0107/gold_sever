package authentication

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"math/rand"
	"reflect"
	"sort"
	"time"
)

const (
	authenticationCheckUrl      = "https://api.wlc.nppa.gov.cn/idcard/authentication/check"
	authenticationQueryUrl      = "http://api2.wlc.nppa.gov.cn/idcard/authentication/query"
	authenticationCollectionUrl = "http://api2.wlc.nppa.gov.cn/behavior/collection/loginout"

	appSecretKey = "009d175b40e0d2f5d62cca732e2e3ec1"
	appId        = "790434f691a5491e80b9f9cb1e3a9d45"
	bizId        = "1104000485"
)

type AuthenticationStatus int

const (
	//认证成功
	AuthenticationCheckStatusSuccess AuthenticationStatus = 0
	//认证中
	AuthenticationCheckStatusIng AuthenticationStatus = 1
	//认证失败
	AuthenticationCheckStatusFail AuthenticationStatus = 2
)

const (
	AuthErrCode_IllegalIDCard = 2001
	AuthErrCode_ResourceLimit = 2002
	AuthErrCode_NoAuthRecord  = 2003
	AuthErrCode_RepeatAuth    = 2004
)

type OnlineStatus int

const (
	//用户离线
	AuthenticationOnlineStatusOffline OnlineStatus = 0
	用户在线
	AuthenticationOnlineStatusOnline OnlineStatus = 1
)

type AuthenticationUsr struct {
	Uid         int64
	Token       string
	Pi          string
	MachineCode string
	Status      OnlineStatus
	TimeStamp   int64
}

var (
	AuthenticationChan = make(chan *AuthenticationUsr, 2000)
)

type AuthenticationCheckRequire struct {
	Ai    string `json:"ai"`
	Name  string `json:"name"`
	IdNum string `json:"idNum"`
}

type AuthenticationCResponse struct {
	Errcode int                         `json:"errcode"`
	Errmsg  string                      `json:"errmsg"`
	Data    AuthenticationRequireResult `json:"data"`
}

type AuthenticationRequireResult struct {
	Result CheckRequireResult `json:"result"`
}

type CheckRequireResult struct {
	Status AuthenticationStatus `json:"status"`
	Pi     string               `json:"pi"`
}

type CheckRequire struct {
	Status int    `json:"status"`
	Pi     string `json:"pi"`
}

type AuthenticationQueryRequire struct {
	Ai string `json:"ai"`
}

type Authenticationcollections struct {
	Collections []Authenticationcollection `json:"collections"`
}

type Authenticationcollection struct {
	No int          `json:"no"` //在批量模式中标识一条行为数 据，取值范围 1-128
	Si string       `json:"si"` //一个会话标识只能对应唯一的 实名用户，一个实名用户可以 拥有多个会话标识
	Bt OnlineStatus `json:"bt"` //游戏用户行为类型 0：下线 1：上线
	Ot int64        `json:"ot"` //行为发生时间戳，单位秒
	Ct int          `json:"ct"` //用户行为数据上报类型 0：已认证通过用户 2：游客用户
	Di string       `json:"di"` //游客模式设备标识，由游戏运 营单位生成，游客用户下必填
	Pi string       `json:"pi"` //已通过实名认证用户的唯一标 识，已认证通过用户必填
}

//! 瓶装实名认证头
func AuthenticationHeaderGet(data string, dataMap map[string]string) map[string]string {
	header := make(map[string]string)
	header["Content-Type"] = "application/json; charset=utf-8"
	header["appId"] = appId
	header["bizId"] = bizId
	timestamps := time.Now().UnixNano() / 1e6
	fmt.Sprintf("%d", timestamps)
	header["timestamps"] = fmt.Sprintf("%d", timestamps)

	signStr := appSecretKey

	if dataMap != nil {
		var keys []string
		for k := range dataMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			signStr += k
			signStr += dataMap[k]
		}
	}

	signStr += "appId"
	signStr += appId

	signStr += "bizId"
	signStr += bizId

	signStr += "timestamps"
	signStr += fmt.Sprintf("%d", timestamps)

	if data != "" {
		signStr += data
	}

	fmt.Printf("%s \n", signStr)

	header["sign"] = Sha256HashCode([]byte(signStr))

	headerstr, _ := json.Marshal(header)
	fmt.Printf("%s \n", headerstr)
	return header
}

//! 实名认证接口
func AuthenticationCheck(ai, name, idcard string) (AuthenticationStatus, string, int) {
	req := AuthenticationCheckRequire{
		Ai:    ai,
		Name:  name,
		IdNum: idcard,
	}
	data, _ := json.Marshal(&req)

	encodeStr := Encode(string(data))

	encodeStr = fmt.Sprintf(`{"data":"%s"}`, encodeStr)

	bytes, _ := util.HttpPost(authenticationCheckUrl, AuthenticationHeaderGet(encodeStr, nil), encodeStr)

	result := &AuthenticationCResponse{}
	json.Unmarshal(bytes, &result)

	if result.Errcode == 0 {
		return result.Data.Result.Status, result.Data.Result.Pi, 0
	} else {
		if result.Data.Result.Status == AuthenticationCheckStatusSuccess {
			result.Data.Result.Status = AuthenticationCheckStatusFail
		}
		return result.Data.Result.Status, result.Data.Result.Pi, result.Errcode
	}
}

//! 实名认证查询接口
func AuthenticationQuery(ai string) (AuthenticationStatus, string, error) {
	req := AuthenticationQueryRequire{
		Ai: ai,
	}

	url := fmt.Sprintf("%s?ai=%s", authenticationQueryUrl, ai)
	bytes, _ := util.HttpGet(url, AuthenticationHeaderGet("", Struct2Map(req)))

	result := &AuthenticationCResponse{}
	json.Unmarshal(bytes, &result)

	if result.Errcode == 0 {
		return result.Data.Result.Status, result.Data.Result.Pi, nil
	} else {
		if result.Data.Result.Status == AuthenticationCheckStatusSuccess {
			result.Data.Result.Status = AuthenticationCheckStatusFail
		}
		return result.Data.Result.Status, result.Data.Result.Pi, errors.New(result.Errmsg)
	}
}

//! 实名认证用户行为上报
func AuthenticationcollectionsUpload(uid int64, token string, pi string, machineCode string, status OnlineStatus) {
	AuthenticationChan <- &AuthenticationUsr{
		Uid:         uid,
		Token:       token,
		Pi:          pi,
		MachineCode: machineCode,
		Status:      status,
		TimeStamp:   time.Now().Unix(),
	}
	/*
		ct := 0
		if len(pi) == 0 {
			ct = 2	//游客
		}
		req := Authenticationcollections{
			Collections: []Authenticationcollection{
				{
					No: 1,
					Si: token,
					Bt: status,
					Ot: time.Now().Unix(),
					Ct: ct,
					Di: machineCode,
					Pi: pi,
				},
			},
		}

		data, _ := json.Marshal(&req)

		encodeStr := Encode(string(data))

		encodeStr = fmt.Sprintf(`{"data":"%s"}`, encodeStr)

		return util.HttpPost(authenticationCollectionUrl, AuthenticationHeaderGet(encodeStr, nil), encodeStr)
	*/
}

//! 统一加密
func Encode(msg string) string {
	res := AesGcmEncrypt(msg, appSecretKey)

	l := base64.StdEncoding.EncodeToString([]byte(res))

	return l
}

//! 统一解密
func Decode(msg string) string {
	data, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		xlog.Logger().Error(err)
	}
	res := AesGcmDecrypt(string(data), appSecretKey)
	return res
}

//! aes gcm加密
func AesGcmEncrypt(v string, aesKey string) string {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	// key := []byte("AES256Key-32Characters1234567890")
	// plaintext := []byte("exampleplaintext")
	key, _ := hex.DecodeString(aesKey)

	plaintext := []byte(v)

	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := GetRandomByte(12)

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	return fmt.Sprintf("%s%s", nonce, ciphertext)
}

//! aes gcm解密
func AesGcmDecrypt(ciphertextV, aesKey string) string {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	key, _ := hex.DecodeString(aesKey)

	nonce := []byte(ciphertextV[0:12])

	ciphertext := []byte(ciphertextV[12:])

	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ""
	}

	return string(plaintext)
}

//! SHA256生成哈希值
func Sha256HashCode(message []byte) string {
	//创建一个基于SHA256算法的hash.Hash接口的对象
	hash := sha256.New()
	//输入数据
	hash.Write(message)
	//计算哈希值
	bytes := hash.Sum(nil)
	//将字符串编码为16进制格式,返回字符串
	hashCode := hex.EncodeToString(bytes)
	//返回哈希值
	return hashCode
}

//! 随机生成
func GetRandomByte(l int) []byte {
	str := "0123456789abcdef"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return result
}

//! 结构体转map
func Struct2Map(obj interface{}) map[string]string {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Tag.Get("json")] = v.Field(i).String()
	}
	return data
}

func AuthenticationThread() {
	AuthenticationMap := make(map[int64]*AuthenticationUsr)
	ticker := time.NewTicker(60 * time.Second)
	submit := func() {
		if len(AuthenticationMap) == 0 {
			return
		}
		req := Authenticationcollections{}
		i := 1
		for k, v := range AuthenticationMap {
			ct := 0
			if len(v.Pi) == 0 {
				ct = 2 //游客
			}
			req.Collections = append(req.Collections, Authenticationcollection{
				No: i,
				Si: v.Token,
				Bt: v.Status,
				Ot: v.TimeStamp,
				Ct: ct,
				Di: v.MachineCode,
				Pi: v.Pi,
			})
			delete(AuthenticationMap, k)
			i++
			if i >= 100 {
				break
			}
		}
		if len(req.Collections) != 0 {
			data, _ := json.Marshal(&req)
			encodeStr := Encode(string(data))
			encodeStr = fmt.Sprintf(`{"data":"%s"}`, encodeStr)
			go func(encodeStr string) {
				if b, err := util.HttpPostWithTimeOut(authenticationCollectionUrl, AuthenticationHeaderGet(encodeStr, nil), 10*time.Second, encodeStr); err != nil {
					xlog.Logger().Errorln(b, err)
				}
			}(encodeStr)
		}
	}

	static.GOFOR(func() {
		select {
		case e := <-AuthenticationChan:
			if p := AuthenticationMap[e.Uid]; p != nil {
				//上线后5秒内下线了
				if p.Status != e.Status {
					delete(AuthenticationMap, e.Uid)
					return
				}
			} else {
				AuthenticationMap[e.Uid] = e
			}
			//批量上报
			if len(AuthenticationMap) >= 100 {
				submit()
			}
		case <-ticker.C:
			//批量上报
			submit()
		}
	})
}
