package client

// 存储相关功能的引入包只有这两个，后面不再赘述
import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"time"
)

var qinNiuSingleton *QiNiuClient = nil

type QiNiuClient struct {
	cfg           models.ConfigQiNiu     // 自定义配置项
	mac           *auth.Credentials      // 七牛鉴权类
	bucketManager *storage.BucketManager // 提供了对资源进行管理的操作
}

func GetQiNiuClient(qnCfg *models.ConfigQiNiu) *QiNiuClient {
	if qnCfg == nil {
		qnCfg = &models.ConfigQiNiu{}
	}

	if qinNiuSingleton == nil {
		qinNiuSingleton = new(QiNiuClient)
		qinNiuSingleton.readConfig(qnCfg)
	}

	if qinNiuSingleton.checkRecreate(qnCfg) {
		qinNiuSingleton.readConfig(qnCfg)
		qinNiuSingleton.initBucket()
	}

	return qinNiuSingleton
}

func (qn *QiNiuClient) readConfig(cfg *models.ConfigQiNiu) {
	qn.cfg.Able = cfg.Able
	qn.cfg.Public = cfg.Public
	qn.cfg.Domain = cfg.Domain
	qn.cfg.Bucket = cfg.Bucket
	qn.cfg.SecretKey = cfg.SecretKey
	qn.cfg.AccessKey = cfg.AccessKey
}

func (qn *QiNiuClient) initMac() {
	qn.mac = qbox.NewMac(qn.cfg.AccessKey, qn.cfg.SecretKey)
}

func (qn *QiNiuClient) initBucket() {
	qn.initMac()
	cfg := storage.Config{
		// 是否使用https域名进行资源管理
		UseHTTPS: false,
	}
	qn.bucketManager = storage.NewBucketManager(qn.mac, &cfg)
}

func (qn *QiNiuClient) checkRecreate(newCfg *models.ConfigQiNiu) bool {
	return !qn.cfg.Verification(newCfg) || qn.mac == nil || qn.bucketManager == nil
}

func (qn *QiNiuClient) key(uid int64) string {
	return fmt.Sprintf("dqUser_%d", uid)
}

func (qn *QiNiuClient) FetchUserImg(uid int64, resURL string) (string, error) {
	if !qn.cfg.Able {
		return resURL, nil
	}
	if resURL == "" {
		return resURL, nil
	}
	fetchRet, err := qn.bucketManager.Fetch(resURL, qn.cfg.Bucket, qn.key(uid))
	if err != nil {
		return resURL, err
	}
	var accessURL string
	if qn.cfg.Public {
		accessURL = storage.MakePublicURL(qn.cfg.Domain, fetchRet.Key)
	} else {
		deadline := time.Now().Add(time.Hour * 72).Unix()
		accessURL = storage.MakePrivateURL(qn.mac, qn.cfg.Domain, fetchRet.Key, deadline)
	}
	return accessURL, nil
}

func (qn *QiNiuClient) CheckUserImg(uid int64, resURL string) bool {
	if !qn.cfg.Able {
		return true
	}
	if resURL == "" {
		return true
	}

	// 请求参数
	body := fmt.Sprintf(`{"data":{"uri":"%s"},"params":{"scenes": ["pulp","terror","politician"]}}`, resURL)

	// header
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Host"] = "ai.qiniuapi.com"
	dataTemp := "POST " + "/v3/image/censor" + "\nHost: " + "ai.qiniuapi.com" + "\nContent-Type: " + "application/json" + "\n\n" + body
	token := qn.mac.Sign([]byte(dataTemp))
	qiniuToken := "Qiniu " + token
	header["Authorization"] = qiniuToken
	// 请求地址
	reqUrl := "http://ai.qiniuapi.com/v3/image/censor"

	// 发送请求
	bytes, err := util.HttpPost(reqUrl, header, body)
	if err != nil {
		xlog.Logger().Errorln(err)
		return true
	}

	// 解析审核结果
	var rsp map[string]interface{}
	err = json.Unmarshal(bytes, &rsp)
	if err != nil {
		xlog.Logger().Errorln(err)
		return true
	}

	// 解析审核结果
	if code, ok := rsp["code"]; ok && int(code.(float64)) == 200 {
		ret := rsp["result"].(map[string]interface{})
		if val, ok := ret["suggestion"]; ok && val.(string) == "pass" {
			return true
		} else {
			xlog.Logger().Errorf("用户:%d 头像 被检测违规 ", uid, resURL)
			return false
		}
	}

	return true
}

func (qn *QiNiuClient) DeleteUserImg(uid int64) error {
	if !qn.cfg.Able {
		return nil
	}
	return qn.bucketManager.Delete(qn.cfg.Bucket, qn.key(uid))
}
