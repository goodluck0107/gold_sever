package client

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// facai php web 支付中心
// http://update.hhkin.com/web/#/item/password/10?page_id=231&redirect=%2F10%3Fpage_id%3D231
// 访问密码: 123456
type PayClient struct {
	host, appId, token string
}

// func GetPayClient(pCfg *model.ConfigWebPay) *PayClient {
// 	return &PayClient{
// 		host:  pCfg.Host,
// 		appId: pCfg.AppId,
// 		token: pCfg.Host,
// 	}
// }

// http://update.hhkin.com/web/#/20?page_id=2030
// 商店信息
type shopInfo struct {
	Status int    `json:"status"`
	Data   data   `json:"data"`
	Msg    string `json:"message"`
}

type data struct {
	MaxGoldRatio    int         `json:"maxjbbili"`
	MaxDiamondRatio int         `json:"maxzsbili"`
	Binding         bool        `json:"binding"`
	Gold            []*goldInfo `json:"jinbi"`
	// Diamond    接口中含有钻石字段，因业务逻辑不需要 所以这里不解析/不处理。
}

func (inf *shopInfo) GoldInfoById(id int) *goldInfo {
	for i, l := 0, len(inf.Data.Gold); i < l; i++ {
		if inf.Data.Gold[i].Id == id {
			return inf.Data.Gold[i]
		}
	}
	return nil
}

// func (inf *shopInfo) DiamondInfoById(id int) *goldInfo {
// 	for i, l := 0, len(inf.Data.Gold); i < l; i++ {
// 		if inf.Data.Gold[i].Id == id {
// 			return inf.Data.Gold[i]
// 		}
// 	}
// 	return nil
// }

type goldInfo struct {
	Id      int `json:"id"`
	Price   int `json:"price"`
	Num     int `json:"num"`
	Status  int `json:"status"`
	Sort    int `json:"sort"`
	NumAdd  int `json:"numAdd"`
	Percent int `json:"percent"`
	Type    int `json:"type"`
}

// 得到一个随机数字字符串
func randomNumber(max int) string {
	var letters = []rune("0123456789")
	result := make([]rune, max)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < max; i++ {
		result[i] = letters[r.Intn(len(letters))]
	}
	return string(result)
}

// GetUserShopInfo 得到用户的商城信息
func (pc *PayClient) GetUserShopInfo(uid int64, platform int) (*shopInfo, error) {
	nonce := strconv.FormatInt(static.HF_Atoi64(randomNumber(10)), 16)
	baseParam := fmt.Sprintf("appid=%s&nonce_str=%s&platform=%d&userid=%d", pc.appId, nonce, platform, uid)
	param := fmt.Sprintf("%s&sign=%s", baseParam, strings.ToLower(util.MD5(fmt.Sprintf("%s&key=%s", baseParam, pc.token))))
	header := make(map[string]string)
	header["Content-Type"] = "application/x-www-form-urlencoded"
	data, err := util.HttpPost(fmt.Sprintf("%s/circler/api/goods", pc.host), header, param)
	if err != nil {
		return nil, err
	}
	var resp shopInfo
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Status != 0 {
		return nil, fmt.Errorf("php responsed error: code:%d, msg:%s", resp.Status, resp.Msg)
	}
	return &resp, nil
}
