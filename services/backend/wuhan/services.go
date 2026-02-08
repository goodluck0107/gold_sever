package wuhan

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	public "github.com/open-source/game/chess.git/pkg/static"
	syslog "github.com/open-source/game/chess.git/pkg/xlog"
	"net/http"
	"runtime/debug"
	"slices"
	"strconv"
	"time"
)

func SuperOpt(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                               // 允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") // header的类型
	w.Header().Set("content-type", "application/json")
	defer func() {
		x := recover()
		if x != nil {
			syslog.Logger().Errorln(x, string(debug.Stack()))
		}
	}() // 返回数据格式是json

	data := req.FormValue("token")
	if data != "d16783f09ea7" {
		w.Write([]byte("你在干吗"))
	} else {
		superId := public.HF_Atoi(req.FormValue("super"))
		if superId <= 0 {
			w.Write([]byte("你在干吗"))
		}
		opt := req.FormValue("opt")
		if opt == "1" {
			rate := public.HF_Atoi(req.FormValue("rate"))
			if rate < 0 || rate > 100 {
				w.Write([]byte("你在干吗"))
			}
			err := GetDBMgr().GetDBrControl().RedisV2.HSet("superman", req.FormValue("super"), req.FormValue("rate")).Err()
			if err != nil {
				syslog.Logger().Error(err)
				w.Write([]byte("数据库异常"))
			} else {
				w.Write([]byte("添加成功"))
			}
		} else if opt == "0" {
			err := GetDBMgr().GetDBrControl().RedisV2.HDel("superman", req.FormValue("super")).Err()
			if err != nil {
				syslog.Logger().Error(err)
				w.Write([]byte("数据库异常"))
			} else {
				w.Write([]byte("删除成功"))
			}
		} else {
			w.Write([]byte("你在干吗"))
		}
	}
}

func FakerOpt(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                               // 允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") // header的类型
	w.Header().Set("content-type", "application/json")
	defer func() {
		x := recover()
		if x != nil {
			syslog.Logger().Errorln(x, string(debug.Stack()))
		}
	}() // 返回数据格式是json

	data := req.FormValue("token")
	if data != "d16783f09ea7" {
		w.Write([]byte("你在干吗"))
	} else {
		fakerId := public.HF_Atoi(req.FormValue("faker"))
		if fakerId <= 0 {
			w.Write([]byte("你在干吗"))
		}
		opt := req.FormValue("opt")
		if opt == "1" {
			err := GetDBMgr().GetDBrControl().RedisV2.SAdd("faker_admin", fakerId).Err()
			if err != nil {
				syslog.Logger().Error(err)
				w.Write([]byte("数据库异常"))
			} else {
				w.Write([]byte("添加战绩查看权限成功"))
			}
		} else if opt == "0" {
			err := GetDBMgr().GetDBrControl().RedisV2.SRem("faker_admin", fakerId).Err()
			if err != nil {
				syslog.Logger().Error(err)
				w.Write([]byte("数据库异常"))
			} else {
				w.Write([]byte("删除战绩查看权限成功"))
			}
		} else {
			w.Write([]byte("你在干吗"))
		}
	}
}

func OwnerOpt(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                               // 允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") // header的类型
	w.Header().Set("content-type", "application/json")
	defer func() {
		x := recover()
		if x != nil {
			syslog.Logger().Errorln(x, string(debug.Stack()))
		}
	}() // 返回数据格式是json

	data := req.FormValue("token")
	if data != "d16783f09ea7" {
		w.Write([]byte("你在干吗?"))
	} else {
		hid := public.HF_Atoi64(req.FormValue("hid"))
		if hid <= 0 {
			w.Write([]byte("你在干吗??"))
			return
		}

		var house models.House
		err := GetDBMgr().GetDBmControl().First(&house, "hid = ?", hid).Error
		if err != nil {
			syslog.Logger().Error(err)
			w.Write([]byte("数据库异常:" + err.Error()))
			return
		}

		key := fmt.Sprintf("houseOwner:%d:%d", house.Id, house.UId)
		cli := GetDBMgr().GetDBrControl().RedisV2
		date := req.FormValue("date")
		res := make(map[string]string)
		if date == "" {
			res, err = cli.HGetAll(key).Result()
		} else {
			res[date], err = cli.HGet(key, date).Result()
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			syslog.Logger().Error(err)
			w.Write([]byte("数据库异常:" + err.Error()))
			return
		}

		type Res struct {
			Date   string `json:"date"`
			Income string `json:"income"`
		}

		resSlice := make([]Res, 0, len(res))
		for k, v := range res {
			resSlice = append(resSlice, Res{
				Date:   k,
				Income: v,
			})
		}
		slices.SortFunc(resSlice, func(a, b Res) int {
			return cmp.Compare(a.Date, b.Date)
		})

		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("当前圈号: %d\n", house.HId))
		buf.WriteString(fmt.Sprintf("当前圈名: %s\n", house.Name))
		buf.WriteString(fmt.Sprintf("当前房主: %d\n", house.UId))
		buf.WriteString("收益情况: \n")
		for _, v := range resSlice {
			buf.WriteString(fmt.Sprintf("\t日期:%s => 收益:%s\n", v.Date, v.Income))
		}
		w.Write([]byte(buf.String()))
		return
	}
}

func ChangeOwner(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                               // 允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") // header的类型
	w.Header().Set("content-type", "application/json")
	defer func() {
		x := recover()
		if x != nil {
			syslog.Logger().Errorln(x, string(debug.Stack()))
		}
	}() // 返回数据格式是json

	data := req.FormValue("token")
	if data != "d16783f09ea7" {
		w.Write([]byte("你在干吗?"))
	} else {
		hid := public.HF_Atoi(req.FormValue("hid"))
		if hid <= 0 {
			w.Write([]byte("你在干吗??"))
			return
		}
		uid := public.HF_Atoi64(req.FormValue("uid"))
		if uid <= 0 {
			w.Write([]byte("你在干吗???"))
			return
		}

		callData := &public.MsgHouseOwnerChange{
			Hid:      hid,
			Uid:      uid,
			PassCode: "",
		}
		// 通知客户端
		value, err := GetServer().CallHall("NewServerMsg", consts.MsgHouseOwnerChange, callData)
		if err != nil {
			syslog.Logger().Errorln("house revoke failed:", err.Error())
			w.Write([]byte("操作失败，请稍后重试1"))
			return
		}
		strVal := fmt.Sprintf("%s", value)
		if strVal != "true" && strVal != "SUC" {
			syslog.Logger().Errorln("house revoke failed:", strVal)
			w.Write([]byte("操作失败，请稍后重试2"))
			return
		}
		w.Write([]byte("OK"))
		return
	}
}

func SpinRecord(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                               // 允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") // header的类型
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	defer func() {
		x := recover()
		if x != nil {
			syslog.Logger().Errorln(x, string(debug.Stack()))
		}
	}() // 返回数据格式是json

	data := req.FormValue("token")
	if data != "d16783f09ea7" {
		w.Write([]byte("你在干吗?"))
	} else {
		uid := public.HF_Atoi64(req.FormValue("uid"))
		var records []models.RecordSpinAward
		db := GetDBMgr().GetDBmControl().Model(&models.RecordSpinAward{})
		if uid > 0 {
			db.Where("uid = ?", uid)
		}
		err := db.Order("id desc").Find(&records).Error
		if err != nil {
			syslog.Logger().Error(err)
			w.Write([]byte("数据库异常:" + err.Error()))
			return
		}
		// 生成HTML表格
		html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>奖励记录</title>
		<style>
			table {
				border-collapse: collapse;
				width: 100%;
			}
			th, td {
				border: 1px solid #ddd;
				padding: 8px;
				text-align: left;
			}
			th {
				background-color: #f2f2f2;
			}
			tr:nth-child(even) {
				background-color: #f9f9f9;
			}
		</style>
	</head>
	<body>
		<h1>奖励记录</h1>
		<table>
			<tr>
				<th>ID</th>
				<th>奖励序号</th>
				<th>奖励描述</th>
				<th>奖励类型</th>
				<th>奖励数量</th>
				<th>用户ID</th>
				<th>获奖时间</th>
			</tr>`

		for _, record := range records {
			html += `
			<tr>
				<td>` + strconv.FormatInt(record.Id, 10) + `</td>
				<td>` + strconv.Itoa(record.Seq) + `</td>
				<td>` + record.Desc + `</td>
				<td>` + record.TypeDesc() + `</td>
				<td>` + strconv.FormatFloat(float64(record.Count)/100, 'f', 2, 64) + `</td>
				<td>` + strconv.FormatInt(record.Uid, 10) + `</td>
				<td>` + record.CreatedAt.Format("2006-01-02 15:04:05") + `</td>
			</tr>`
		}

		html += `
		</table>
	</body>
	</html>`

		// 写入响应
		w.Write([]byte(html))

		return
	}
}

type ChargeRes struct {
	DateStr       string  `json:"date_str" gorm:"column:date_str"`
	SuccessPrice  float64 `json:"success_price" gorm:"column:success_price"`
	SuccessNum    int64   `json:"success_num" gorm:"column:success_num"`
	SuccessPlayer int64   `json:"success_player" gorm:"column:success_player"`
	TotalPrice    float64 `json:"total_price" gorm:"column:total_price"`
	TotalNum      int64   `json:"total_num" gorm:"column:total_num"`
	TotalPlayer   int64   `json:"total_player" gorm:"column:total_player"`
	LoginNum      int64   `json:"login_num" gorm:"column:login_num"`
}

type LoginRes struct {
	DateStr  string `json:"date_str" gorm:"column:date_str"`
	LoginNum int64  `json:"login_num" gorm:"column:login_num"`
}

func ChargeStat(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                               // 允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") // header的类型
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	defer func() {
		x := recover()
		if x != nil {
			syslog.Logger().Errorln(x, string(debug.Stack()))
		}
	}() // 返回数据格式是json

	data := req.FormValue("token")
	if data != "d16783f09ea7" {
		w.Write([]byte("你在干吗?"))
	} else {
		dateStr := req.FormValue("date")
		if dateStr != "" {
			// 尝试解析日期
			_, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				http.Error(w, "无效的日期格式. 期望格式: YYYY-MM-DD", http.StatusBadRequest)
				return
			}
		}

		var records []*ChargeRes

		db := GetDBMgr().GetDBmControl().Model(&models.PaymentOrder{})

		db = db.Select("DATE(created_at) AS date_str, SUM(CASE WHEN status='amount_arrived' THEN price/100 ELSE 0 END) AS success_price, COUNT(CASE WHEN status='amount_arrived' THEN 1 END) AS success_num, COUNT(DISTINCT CASE WHEN status='amount_arrived' THEN user_id END) AS success_player, SUM(price/100) AS total_price, COUNT(*) AS total_num, COUNT(DISTINCT user_id) AS total_player")

		if dateStr != "" {
			db = db.Where("created_at >= ?", dateStr+" 00:00:00")
		}

		err := db.Group("date_str").Order("date_str asc").Scan(&records).Error
		if err != nil {
			syslog.Logger().Error(err)
			w.Write([]byte("数据库异常:" + err.Error()))
			return
		}

		loginDb := GetDBMgr().GetDBmControl().Model(&models.UserLoginRecord{})
		loginDb = loginDb.Select("DATE(created_at) AS date_str, COUNT(DISTINCT uid) AS login_num")
		if dateStr != "" {
			loginDb = loginDb.Where("created_at >= ?", dateStr+" 00:00:00")
		}
		var loginRecords []*LoginRes
		err = loginDb.Group("date_str").Order("date_str asc").Scan(&loginRecords).Error
		if err != nil {
			syslog.Logger().Error(err)
			w.Write([]byte("数据库异常:" + err.Error()))
			return
		}

		// 合并数据
		for _, record := range records {
			for _, loginRecord := range loginRecords {
				if record.DateStr == loginRecord.DateStr {
					record.LoginNum = loginRecord.LoginNum
					break
				}
			}
		}

		// 生成HTML表格
		html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>充值统计</title>
		<style>
			table {
				border-collapse: collapse;
				width: 100%;
			}
			th, td {
				border: 1px solid #ddd;
				padding: 8px;
				text-align: left;
			}
			th {
				background-color: #f2f2f2;
			}
			tr:nth-child(even) {
				background-color: #f9f9f9;
			}
		</style>
	</head>
	<body>
		<h1>充值统计</h1>
		<table>
			<tr>
				<th>日期</th>
				<th>到账金额</th>
				<th>到账订单数</th>
				<th>到账人数</th>
				<th>发起总金额</th>
				<th>发起订单数</th>
				<th>发起总人数</th>
				<th>日活人数</th>
			</tr>`

		for _, record := range records {
			html += `
			<tr>
				<td>` + record.DateStr[:10] + `</td>
				<td>` + strconv.FormatFloat(record.SuccessPrice, 'f', 2, 64) + `</td>
				<td>` + strconv.FormatInt(record.SuccessNum, 10) + `</td>
				<td>` + strconv.FormatInt(record.SuccessPlayer, 10) + `</td>
				<td>` + strconv.FormatFloat(record.TotalPrice, 'f', 2, 64) + `</td>
				<td>` + strconv.FormatInt(record.TotalNum, 10) + `</td>
				<td>` + strconv.FormatInt(record.TotalPlayer, 10) + `</td>
				<td>` + strconv.FormatInt(record.LoginNum, 10) + `</td>
			</tr>`
		}

		html += `
		</table>
	</body>
	</html>`

		// 写入响应
		w.Write([]byte(html))
		return
	}
}
