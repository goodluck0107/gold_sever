package user

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/goEncrypt"
	"github.com/open-source/game/chess.git/pkg/static/wealthtalk"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/open-source/game/chess.git/pkg/static/util"
	common "github.com/open-source/game/chess.git/pkg/xerrors"

	"github.com/jinzhu/gorm"
)

// 登录消息返回
type Msg_Login_MsgData_Return struct {
	Api            string `json:"api"`              // api服务器地址
	Hall           string `json:"hall"`             // 大厅地址
	Gate           string `json:"gate"`             // 网关地址
	Uid            int64  `json:"uid"`              // 用户id
	Area           string `json:"area"`             // 玩家所在区域
	Token          string `json:"token"`            // 登录令牌
	NeedBindWechat bool   `json:"need_bind_wechat"` // 是否需要绑定微信
	IsNew          bool   `json:"isnew"`            // 是否是新用户
	AppId          string `json:"appid"`            // 支付appid
	AppToken       string `json:"apptoken"`         // 支付apptoken
	InsureendLine  int    `json:"insureendline"`    // 保险箱最低限制
	BroadcastCost  int    `json:"broadcast_cost"`   // 用户区域广播消耗
	PlayTime       int    `json:"playtime"`         // 未实名的累计游戏时长/已实名的未成人每日游戏时长
}

type ServerMethod int

func (self *ServerMethod) ServerMsg(ctx context.Context, args *static.Rpc_Args, reply *[]byte) error {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	if args == nil || reply == nil {
		return errors.New("nil paramters !")
	}

	head, _, _, data, ok, _, _ := static.HF_DecodeMsg(args.MsgData)
	if !ok {
		return errors.New("args err !")
	}
	xlog.Logger().WithFields(logrus.Fields{
		"@head": head,
		"@data": string(data),
	}).Infoln("【RECEIVED RPC】")
	switch head {
	case consts.MsgTypeReloadConfig: // 重新加载配置
		var _msg static.Msg_Null
		json.Unmarshal(data, &_msg)
		if err := reloadConfigs(0, []int{}); err != nil {
			xlog.Logger().Errorln("reloadconfig failed: ", err)
		}
	case consts.MsgTypeSetLogFileLevel:
		{
			req := new(static.MsgSetLogFileLevel)
			err := json.Unmarshal(data, &req)
			if err != nil {
				xlog.Logger().Errorln("rpc argue err :", err.Error())
				*reply = []byte("false")
				return err
			}
			xlog.SetFileLevel(req.Level)
		}
	}

	return nil
}

func (self *ServerMethod) NewServerMsg(ctx context.Context, args *[]byte, reply *[]byte) error {
	return self.ServerMsg(ctx, &static.Rpc_Args{MsgData: *args}, reply)
}

// 客户端登录相关接口
func Service(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                               // 允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") // header的类型
	w.Header().Set("content-type", "application/json")                                               // 返回数据格式是json
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	data := req.FormValue("msgdata")
	clientip := static.HF_GetHttpIP(req)
	xlog.Logger().WithFields(logrus.Fields{
		"clientIp": clientip,
		"data":     data,
	}).Infoln("【RECEIVED HTTP】")
	var msg static.Msg_Header
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		xlog.Logger().Errorln("Service err2:", err.Error())
		w.Write(getErrReturn("", xerrors.ArgumentError))
		return
	}
	// 如果使用加密了但客户端消息未加密则直接报错
	if GetServer().Con.Encode != consts.EncodeNone && msg.Sign.Encode == consts.EncodeNone {
		w.Write(getErrReturn("", xerrors.ArgumentError))
		return
	}
	// 消息解密处理
	err = getCommonMsg(&msg)
	if err != nil {
		xlog.Logger().Errorln("Service decrypt err:", err.Error())
		w.Write(getErrReturn("", xerrors.ArgumentError))
		return
	}
	switch msg.Header {
	case consts.MsgTypeCheckYK:
		{ // 检测是否是游客账号
			var msgdata static.Msg_LoginYK
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service loginYK Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 设备码不能为空
			if msgdata.MachineCode == "" {
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 判断用户是否存在
			var user models.User
			if err = GetDBMgr().db_M.Model(&user).Where("guest_id = ? and user_type = ? and account_type = 0", msgdata.MachineCode, consts.UserTypeYk).First(&user).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					// 异常报错
					xlog.Logger().Errorln("query user data failed: ", err.Error())
					w.Write(getErrReturn(msg.Header, xerrors.DBExecError))
					return
				}
				// 查不到游客账号直接返回false
				w.Write(static.HF_JtoB(getCommonResp(msg.Header, &static.Msg_S2C_CheckYK{ShowYK: false})))
				return
			}

			w.Write(static.HF_JtoB(getCommonResp(msg.Header, &static.Msg_S2C_CheckYK{ShowYK: true})))
		}
	case consts.MsgTypeLoginYK:
		{
			// 游客账号登录
			var msgdata static.Msg_LoginYK
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service loginYK Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验平台类型
			if !consts.CheckPlatform(msgdata.Platform) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无效的平台类型")))
				return
			}

			code := msgdata.MachineCode
			xlog.Logger().Warningln("loginYK,code:", code)
			if code == "" {
				// 不允许随机生成设备码
				// code = getNewGustCode()
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无法获取您的设备编号, 请检查授权是否开启。")))
				return
			}

			user, person, isNew, err := GetServer().loginByYK(code, msgdata.Platform, clientip, msgdata.MachineCode)
			if err != nil {
				xlog.Logger().Errorln("loginYK can not creater person err: ", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(fmt.Sprintf("登录失败, 请联系客服处理, 错误信息: %s", err.Error()))))
			} else {
				if lastError := OnUserLogin(isNew, user, msgdata.Platform, clientip, msgdata.MachineCode); lastError != nil {
					xlog.Logger().Error("OnUserLogin error:", lastError)
					w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				} else {
					resp, err := getLoginResp(consts.MsgTypeLoginYK, person, isNew)
					if err != nil {
						w.Write(getErrReturn(msg.Header, err))
					} else {
						w.Write(static.HF_JtoB(resp))
					}
				}
			}
		}
		return
	case consts.MsgTypeSmscode:
		{
			// 短信验证码
			var msgdata static.Msg_Smscode
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service smscode Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验手机号
			if !static.CheckMobile(msgdata.Mobile) {
				w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
				return
			}
			// 校验登录类型
			if !service2.GetAdminClient().CheckSmsType(msgdata.Type) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("验证码类型错误")))
				return
			}

			if msgdata.Type == service2.SmsTypeResetPassword { // 如果是找回密码的场景时, 需要判断用户是否存在
				var user models.User
				if err = GetDBMgr().db_M.Where("tel = ? and (user_type = ? or user_type = ?) and account_type = 0", msgdata.Mobile, consts.UserTypeMobile, consts.UserTypeMobile2).First(&user).Error; err != nil {
					// 若用户不存在则报错
					w.Write(getErrReturn(msg.Header, xerrors.NewXError("当前手机号未注册")))
					return
				}
			} else if msgdata.Type == service2.SmsTypeRegister {
				var user models.User
				if err = GetDBMgr().db_M.Where("tel = ? and (user_type = ? or user_type = ?) and account_type = 0", msgdata.Mobile, consts.UserTypeMobile, consts.UserTypeMobile2).First(&user).Error; err == nil {
					// 若用户存在则报错
					w.Write(getErrReturn(msg.Header, xerrors.NewXError("当前手机号已被注册过")))
					return
				}
			}

			// 检查是否在60s内发送过短信
			if GetDBMgr().db_R.CheckSmsCodeSend(msgdata.Mobile, msgdata.Type) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("请求过于频繁，请稍后再试")))
				return
			}

			// 发送短信验证码
			err = service2.GetAdminClient().SendSmsCode(msgdata.Mobile, msgdata.Type)
			if err != nil {
				xlog.Logger().Errorf("send sms code err:%s, molile:%s, ip:%s", err.Error(), msgdata.Mobile, clientip)
				w.Write(getErrReturn(msg.Header, xerrors.SendSmsError))
				return
			}
			// 设置发送过短信
			GetDBMgr().db_R.SetSmsCodeSend(msgdata.Mobile, msgdata.Type)

			resultHead := new(static.Msg_Header)
			resultHead.Header = consts.MsgTypeSmscode
			resultHead.Sign.Encode = 0
			resultHead.Sign.Time = time.Now().Unix()
			w.Write(static.HF_JtoB(resultHead))
		}
	case consts.MsgTypeLoginMobile:
		{ // 手机登录(手机号+验证码)
			var msgdata static.Msg_LoginMobile
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service loginMobile Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验平台类型
			if !consts.CheckPlatform(msgdata.Platform) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无效的平台类型")))
				return
			}

			// 校验手机号
			if !static.CheckMobile(msgdata.Mobile) {
				w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
				return
			}
			if msgdata.Code == "" {
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			if len(msgdata.Code) > 6 {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("验证码必须小于6位")))
				return
			}

			if len(msgdata.Mobile) > 11 {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("验证码必须小于11位")))
				return
			}

			// 校验验证码
			//err = service2.GetAdminClient().CheckSmsCode(msgdata.Mobile, msgdata.Code, service2.SmsTypeLogin)
			//if err != nil {
			//	xlog.Logger().Errorln(err)
			//	w.Write(getErrReturn(msg.Header, xerrors.SmsCodeError))
			//	return
			//}

			// 手机号码加密处理
			//msgdata.Mobile, err = static.HF_EncodeStr(static.HF_Atobytes(msgdata.Mobile), static.UserEncodeKey)
			//if err != nil {
			//	w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
			//	return
			//}

			msgdata.Mobile = fmt.Sprintf("%s_%s", msgdata.Mobile, msgdata.Code)

			user, person, isNew, err := GetServer().loginByMobile(msgdata.Mobile, msgdata.Platform, clientip, msgdata.MachineCode, true)
			if err != nil {
				xlog.Logger().Errorln("loginMobile can not creater person err ")
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(fmt.Sprintf("手机号登录失败, 请联系客服处理, 错误信息: %s", err.Error()))))
			} else {
				if person == nil {
					// 提示绑定微信
					result := new(Msg_Login_MsgData_Return)
					result.Uid = 0
					if GetServer().Con.UseSafeIp == 0 {
						result.Hall = GetServer().Con.Center
					} else {
						result.Hall = GetServer().Con.SafeHall
					}
					result.Gate = GetServer().Con.Gate
					result.Token = ""
					result.NeedBindWechat = true
					result.IsNew = true
					w.Write(static.HF_JtoB(getCommonResp(msg.Header, result)))
				} else {
					if lastError := OnUserLogin(isNew, user, msgdata.Platform, clientip, msgdata.MachineCode); lastError != nil {
						xlog.Logger().Error("OnUserLogin error:", lastError)
						w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
					} else {
						resp, err := getLoginResp(consts.MsgTypeLoginMobile, person, isNew)
						if err != nil {
							w.Write(getErrReturn(msg.Header, err))
						} else {
							w.Write(static.HF_JtoB(resp))
						}
					}
				}
			}
		}
	case consts.MsgTypeLoginMobileV2:
		{ // 手机登录(手机号+密码)
			var msgdata static.Msg_LoginMobilev2
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service loginMobile Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验平台类型
			if !consts.CheckPlatform(msgdata.Platform) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无效的平台类型")))
				return
			}

			// 校验手机号
			if !static.CheckMobile(msgdata.Mobile) {
				w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
				return
			}

			// 校验密码
			if len(msgdata.Password) < 6 || len(msgdata.Password) > 16 {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("账号或密码错误")))
				return
			}

			// 手机号码加密处理
			msgdata.Mobile, err = static.HF_EncodeStr(static.HF_Atobytes(msgdata.Mobile), static.UserEncodeKey)
			if err != nil {
				w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
				return
			}

			user, person, err := GetServer().loginByMobilev2(msgdata.Mobile, msgdata.Password, msgdata.Platform)
			if err != nil {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(fmt.Sprintf("手机号登录失败, 请联系客服处理, 错误信息: %s", err.Error()))))
			} else {
				if lastError := OnUserLogin(false, user, msgdata.Platform, clientip, msgdata.MachineCode); lastError != nil {
					xlog.Logger().Error("OnUserLogin error:", lastError)
					w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				} else {
					resp, err := getLoginResp(consts.MsgTypeLoginMobile, person, false)
					if err != nil {
						w.Write(getErrReturn(msg.Header, err))
					} else {
						w.Write(static.HF_JtoB(resp))
					}
				}
			}
		}
	case consts.MsgTypeMobileRegister:
		{ // 手机注册
			var msgdata static.Msg_MobileResgister
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service mobileRegister Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验平台类型
			if !consts.CheckPlatform(msgdata.Platform) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无效的平台类型")))
				return
			}

			// 校验昵称
			if len(msgdata.Nickname) > 21 {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("昵称不能超过14个英文或者7个汉字")))
				return
			}

			// 校验密码
			if len(msgdata.Password) < 6 || len(msgdata.Password) > 16 {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("密码长度应在6-16个字符之间")))
				return
			}

			// 校验手机号
			if !static.CheckMobile(msgdata.Mobile) {
				w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
				return
			}

			if msgdata.Code == "" {
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验验证码
			err = service2.GetAdminClient().CheckSmsCode(msgdata.Mobile, msgdata.Code, service2.SmsTypeRegister)
			if err != nil {
				xlog.Logger().Errorln(err)
				w.Write(getErrReturn(msg.Header, xerrors.SmsCodeError))
				return
			}

			// 手机号码加密处理
			msgdata.Mobile, err = static.HF_EncodeStr(static.HF_Atobytes(msgdata.Mobile), static.UserEncodeKey)
			if err != nil {
				w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
				return
			}

			// 提前校验, 避免验证码校验后失效
			var user models.User
			if err = GetDBMgr().db_M.Model(user).Where("tel = ? and account_type = 0", msgdata.Mobile).First(&user).Error; err == nil {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("手机账号已存在")))
				return
			}

			person, err := GetServer().registerByMobile(msgdata.Mobile, msgdata.Platform, clientip, msgdata.MachineCode, msgdata.Password, msgdata.Nickname)
			if err != nil {
				xlog.Logger().Errorln("mobileRegister can not creater person err ")
				w.Write(getErrReturn(msg.Header, err))
			} else {
				if lastError := OnUserLogin(true, &user, msgdata.Platform, clientip, msgdata.MachineCode); lastError != nil {
					xlog.Logger().Error("OnUserLogin error:", lastError)
					w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				} else {
					resp, err := getLoginResp(consts.MsgTypeLoginMobile, person, true)
					if err != nil {
						w.Write(getErrReturn(msg.Header, err))
					} else {
						w.Write(static.HF_JtoB(resp))
					}
				}
			}
		}
	case consts.MsgTypeResetPassword: // 修改密码
		{ // 重置密码
			var msgdata static.Msg_ResetPassword
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service ResetPassword Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验密码
			if len(msgdata.Password) < 6 || len(msgdata.Password) > 16 {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("密码长度应在6-16个字符之间")))
				return
			}

			// 校验手机号
			if !static.CheckMobile(msgdata.Mobile) {
				w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
				return
			}
			if msgdata.Code == "" {
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验验证码
			err = service2.GetAdminClient().CheckSmsCode(msgdata.Mobile, msgdata.Code, service2.SmsTypeResetPassword)
			if err != nil {
				xlog.Logger().Errorln(err)
				w.Write(getErrReturn(msg.Header, xerrors.SmsCodeError))
				return
			}

			// 获取用户
			var user models.User
			if err = GetDBMgr().db_M.Model(user).Where("tel = ? and account_type = 0", msgdata.Mobile).First(&user).Error; err != nil {
				w.Write(getErrReturn(msg.Header, xerrors.UserNotExistError))
				return
			}

			// 修改密码
			if err = GetDBMgr().db_M.Model(&user).Update("password", util.MD5(msgdata.Password)).Error; err != nil {
				w.Write(getErrReturn(msg.Header, xerrors.DBExecError))
				return
			}
			w.Write(static.HF_JtoB(getCommonResp(msg.Header, "")))
		}
	case consts.MsgTypeLoginWechat:
		{ // 微信登录
			var msgdata static.Msg_LoginWechat
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service loginWechat Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验平台类型
			if !consts.CheckPlatform(msgdata.Platform) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无效的平台类型")))
				return
			}

			if msgdata.Code == "" {
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}
			// 校验手机号
			if msgdata.Mobile != "" {
				if !static.CheckMobile(msgdata.Mobile) {
					w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
					return
				}
				// 手机号码加密处理
				msgdata.Mobile, err = static.HF_EncodeStr(static.HF_Atobytes(msgdata.Mobile), static.UserEncodeKey)
				if err != nil {
					w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
					return
				}
			}

			// 默认app
			if msgdata.AppId == "" {
				msgdata.AppId = consts.AppDefault
			}

			config := GetServer().GetAppConfig(msgdata.AppId)
			if config == nil {
				xlog.Logger().Errorln("wx login nil config app")
				w.Write(getErrReturn(msg.Header, xerrors.ThirdpartyError))
				return
			}

			info, err := service2.NewWeixinClient(config.WxAppId, config.WxAppSecret).GetWeixinUserInfo(msgdata.Code)
			if err != nil {
				xlog.Logger().WithFields(map[string]interface{}{
					"WxAppId":     config.WxAppId,
					"WxAppSecret": config.WxAppSecret,
					"Code":        msgdata.Code,
				}).Errorln("wx api err:", err)
				w.Write(getErrReturn(msg.Header, xerrors.ThirdpartyError))
				return
			}

			user, person, isNew, err := GetServer().loginByWechat(info, msgdata.Mobile, msgdata.Platform, clientip, msgdata.MachineCode)
			if err != nil {
				xlog.Logger().Errorln("loginWechat can not creater person err: ", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(fmt.Sprintf("微信登录失败, 请联系客服处理, 错误信息: %s", err.Error()))))
			} else {
				if lastError := OnUserLogin(isNew, user, msgdata.Platform, clientip, msgdata.MachineCode); lastError != nil {
					xlog.Logger().Error("OnUserLogin error:", lastError)
					w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				} else {
					resp, err := getLoginResp(consts.MsgTypeLoginWechat, person, isNew)
					if err != nil {
						w.Write(getErrReturn(msg.Header, err))
					} else {
						w.Write(static.HF_JtoB(resp))
					}
				}
			}
		}
	case consts.MsgTypeLoginWechatV2:
		{ // 微信小程序登录
			var msgdata static.Msg_LoginWechatV2
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service loginWechatV2 Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验平台类型
			if !consts.CheckPlatform(msgdata.Platform) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无效的平台类型")))
				return
			}

			if msgdata.Code == "" {
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验手机号
			if msgdata.Mobile != "" {
				if !static.CheckMobile(msgdata.Mobile) {
					w.Write(getErrReturn(msg.Header, xerrors.MobileInvalidError))
					return
				}
			}

			// 默认app
			if msgdata.AppId == "" {
				msgdata.AppId = consts.AppDefault
			}

			config := GetServer().GetAppConfig(msgdata.AppId)
			if config == nil {
				xlog.Logger().Errorln("wx applet login nil config app")
				w.Write(getErrReturn(msg.Header, xerrors.ThirdpartyError))
				return
			}

			info, err := service2.NewWeixinClient(config.WxAppId, config.WxAppSecret).GetAppletWeixinUserInfo(msgdata.Code, msgdata.RawData)
			if err != nil {
				xlog.Logger().Errorln("wx applet login api err :", err)
				w.Write(getErrReturn(msg.Header, xerrors.ThirdpartyError))
				return
			}

			user, person, isNew, err := GetServer().loginByWechat(info, msgdata.Mobile, msgdata.Platform, clientip, msgdata.MachineCode)
			if err != nil {
				xlog.Logger().Errorln("loginWechatV2 can not creater person err: ", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(fmt.Sprintf("微信小程序登录失败, 请联系客服处理, 错误信息: %s", err.Error()))))
			} else {
				if lastError := OnUserLogin(isNew, user, msgdata.Platform, clientip, msgdata.MachineCode); lastError != nil {
					xlog.Logger().Error("OnUserLogin error:", lastError)
					w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				} else {
					resp, err := getLoginResp(consts.MsgTypeLoginWechatV2, person, isNew)
					if err != nil {
						w.Write(getErrReturn(msg.Header, err))
					} else {
						w.Write(static.HF_JtoB(resp))
					}
				}
			}
		}
	case consts.MsgTypeLoginApple:
		{
			// ios账号登录
			var msgdata static.Msg_LoginApple
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service loginIOS Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验平台类型
			if !consts.CheckPlatform(msgdata.Platform) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无效的平台类型")))
				return
			}

			if msgdata.IdentityToken == "" {
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 默认app
			if msgdata.AppId == "" {
				msgdata.AppId = consts.AppDefault
			}

			config := GetServer().GetAppConfig(msgdata.AppId)
			if config == nil {
				xlog.Logger().Errorln("ios login nil config app")
				w.Write(getErrReturn(msg.Header, xerrors.ThirdpartyError))
				return
			}

			// 认证IdentityToken
			if err = service2.VerifyAppleIdentityToken(msgdata.IdentityToken, msgdata.UserUnionId); err != nil {
				xlog.Logger().Errorln(err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(err.Error())))
				return
			}

			// 构造苹果账号的用户信息
			var appleUser service2.AppleUserInfo
			appleUser.Nickname = msgdata.Nickname
			appleUser.UnionId = msgdata.UserUnionId

			// 认证成功 开始登陆
			user, person, isNew, err := GetServer().loginByApple(&appleUser, &msgdata, clientip)
			if err != nil {
				xlog.Logger().Errorln("loginApple can not creater person err: ", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(fmt.Sprintf("苹果账号登录失败, 请联系客服处理, 错误信息: %s", err.Error()))))
			} else {
				if lastError := OnUserLogin(isNew, user, msgdata.Platform, clientip, msgdata.MachineCode); lastError != nil {
					xlog.Logger().Error("OnUserLoginApple error:", lastError)
					w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				} else {
					resp, err := getLoginResp(consts.MsgTypeLoginApple, person, isNew)
					if err != nil {
						w.Write(getErrReturn(msg.Header, err))
					} else {
						w.Write(static.HF_JtoB(resp))
					}
				}
			}
		}
	case consts.MsgTypeLoginHW:
		{
			// 华为账号登录
			var msgdata static.Msg_LoginHW
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service loginHW Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验平台类型
			if !consts.CheckPlatform(msgdata.Platform) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无效的平台类型")))
				return
			}

			if msgdata.AccessToken == "" {
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 默认app
			if msgdata.AppId == "" {
				msgdata.AppId = consts.AppDefault
			}

			config := GetServer().GetAppConfig(msgdata.AppId)
			if config == nil {
				xlog.Logger().Errorln("HW login nil config app")
				w.Write(getErrReturn(msg.Header, xerrors.ThirdpartyError))
				return
			}

			// 认证AccessToken
			tokenInfo, err := service2.VerifyHWAccessToken(msgdata.AccessToken)
			if err != nil {
				xlog.Logger().Errorln(err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(err.Error())))
				return
			}

			// 验证用户信息是否匹配
			if tokenInfo.OpenId != msgdata.OpenId || tokenInfo.UnionId != msgdata.UnionId {
				errMsg := "verify token fail, user eve is not match"
				xlog.Logger().Errorln(errMsg)
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(errMsg)))
				return
			}

			// 构造华为账号的用户信息
			var hwUser service2.HuaWeiUserInfo
			hwUser.OpenId = msgdata.OpenId
			hwUser.UnionId = msgdata.UnionId
			hwUser.Nickname = msgdata.Nickname
			hwUser.AvatarUrl = msgdata.AvatarUrl
			hwUser.AccountFlag = msgdata.AccountFlag

			// 认证成功 开始登陆
			user, person, isNew, err := GetServer().loginByHW(&hwUser, &msgdata, clientip)
			if err != nil {
				xlog.Logger().Errorln("loginHW can not creater person err: ", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError(fmt.Sprintf("华为账号登录失败, 请联系客服处理, 错误信息: %s", err.Error()))))
			} else {
				if lastError := OnUserLogin(isNew, user, msgdata.Platform, clientip, msgdata.MachineCode); lastError != nil {
					xlog.Logger().Error("OnUserLoginHW error:", lastError)
					w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				} else {
					resp, err := getLoginResp(consts.MsgTypeLoginHW, person, isNew)
					if err != nil {
						w.Write(getErrReturn(msg.Header, err))
					} else {
						w.Write(static.HF_JtoB(resp))
					}
				}
			}
		}
	case consts.MsgTypeLoginToken:
		{ // 令牌直接登录
			var msgdata static.Msg_LoginToken
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Errorln("Service loginToken Unmarshal err:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			// 校验平台类型
			if !consts.CheckPlatform(msgdata.Platform) {
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("无效的平台类型")))
				return
			}

			if msgdata.Token == "" || msgdata.Uid == 0 {
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}

			user := new(models.User)

			if err := GetDBMgr().db_M.Where("id = ? and account_type = 0", msgdata.Uid).Find(user).Error; err != nil {
				xlog.Logger().Errorln("get user from mysql failed: ", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				return
			} else {
				if !static.HF_IsValidMachineCode(user.MachineCode) && static.HF_IsValidMachineCode(msgdata.MachineCode) {
					user.MachineCode = msgdata.MachineCode
					if err = GetDBMgr().db_M.Save(&user).Error; err != nil {
						xlog.Logger().Errorln("update user data failed: ", err.Error())
					}
				}
			}

			// 获取person
			person, err := GetDBMgr().db_R.GetPerson(msgdata.Uid)
			if err != nil {
				xlog.Logger().Errorln("get user from redis failed: ", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				return
			}

			if person.Token != msgdata.Token {
				xlog.Logger().Errorf("token 不一致, msgdata.Uid = %d, RealToken = %s, LoginToken = %s", msgdata.Uid, person.Token, msgdata.Token)
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
				return
			}
			if lastError := OnUserLogin(false, user, msgdata.Platform, clientip, msgdata.MachineCode); lastError != nil {
				xlog.Logger().Error("OnUserLogin error:", lastError)
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("登录失败")))
			} else {
				resp, cerr := getLoginResp(consts.MsgTypeLoginToken, person, false)
				if cerr != nil {
					w.Write(getErrReturn(msg.Header, cerr))
				} else {
					w.Write(static.HF_JtoB(resp))
				}
			}
		}
	case consts.MsgTypeGetReplay:
		{
			xlog.Logger().Debug("constant.MsgTypeGetReplay")
			var msgdata static.Msg_GetReplayInfo
			// syslog.Logger().Debug("解析消息")
			err = json.Unmarshal([]byte(msg.Data), &msgdata)
			if err != nil {
				xlog.Logger().Debug("Unmarshal:", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.ArgumentError))
				return
			}
			// syslog.Logger().Debug("得到回访数据")
			// 获取回放记录
			var gameReplay models.RecordGameReplay
			if err = GetDBMgr().db_M.Model(&gameReplay).Where("id = ?", msgdata.ReplayId).First(&gameReplay).Error; err != nil {
				xlog.Logger().Errorln("replay:", err.Error())
				cuserror := xerrors.NewXError("回放记录不存在")
				w.Write(getErrReturn(msg.Header, cuserror))
				return
			}

			replay := gameReplay.ConvertModel()

			gameInfo, err := GetDBMgr().SelectGameRecordInfo(gameReplay.GameNum)
			if err != nil {
				xlog.Logger().Errorln("replay:", err.Error())
				cuserror := xerrors.NewXError("回放记录不存在")
				w.Write(getErrReturn(msg.Header, cuserror))
				return
			}
			// 更新头像跟性别
			userAttrMap := make(map[int64]*static.Msg_S2C_GameRecordInfoUser)
			for index, userinfo := range gameInfo.UserArr {
				p, err := GetDBMgr().GetDBrControl().GetPerson(userinfo.Uid)
				if err == nil && p != nil {
					gameInfo.UserArr[index].Imgurl = p.Imgurl
					gameInfo.UserArr[index].Sex = p.Sex
				}
				userAttrMap[userinfo.Uid] = userinfo
			}

			// syslog.Logger().Debug("获取牌桌信息")
			// 获取牌桌信息
			var costRecord models.RecordGameCost
			if err = GetDBMgr().db_M.Where("game_num = ?", gameReplay.GameNum).First(&costRecord).Error; err != nil {
				xlog.Logger().Errorln("costRecord:", err.Error())
				cuserror := xerrors.NewXError("获取牌桌信息失败")
				w.Write(getErrReturn(msg.Header, cuserror))
				return
			}
			// syslog.Logger().Debug("获取游戏配置")
			// 获取游戏配置
			gameConfig := make(map[string]interface{})
			err = json.Unmarshal([]byte(costRecord.GameConfig), &gameConfig)
			if err != nil {
				xlog.Logger().Errorln("gameConfig", err.Error())
				w.Write(getErrReturn(msg.Header, xerrors.NewXError("获取牌桌信息失败")))
				return
			}

			result := new(static.Msg_S2C_ReplayInfo)
			result.HandCard = gameReplay.HandCard
			result.OutCard = gameReplay.OutCard
			result.EndInfo = gameReplay.EndInfo
			result.Table = &static.Msg_S2C_ReplayInfoTable{
				HId:  int(gameConfig["hid"].(float64)),
				TId:  costRecord.TId,
				FId:  int(gameConfig["nfid"].(float64)),
				NTId: int(gameConfig["ntid"].(float64)),

				KindId:     gameReplay.KindID,
				GameConfig: costRecord.GameConfig,
				CardsNum:   gameReplay.CardsNum,
			}

			isvitamin, ok := gameConfig["isvitamin"]
			if ok {
				result.Table.IsVitamin = isvitamin.(bool)
			} else {
				result.Table.IsVitamin = false
			}

			ishidHide, ok := gameConfig["ishidhide"]
			if ok {
				result.Table.IsHidHide = ishidHide.(bool)
			} else {
				result.Table.IsHidHide = false
			}

			// 计算用户上一局分数
			for userIndex, _ := range gameInfo.UserArr {
				score := float64(0)

				curUid := int64(0)

				if replay.PlayNum-1 >= 0 && replay.PlayNum-1 < len(gameInfo.ScoreArr) {
					curUid = gameInfo.ScoreArr[replay.PlayNum-1].Uids[userIndex]
				}

				for i := 0; i < replay.PlayNum-1; i++ {
					for userIndex2, _ := range gameInfo.UserArr {
						if gameInfo.ScoreArr[i].Uids[userIndex2] == curUid {
							score = score + gameInfo.ScoreArr[i].Score[userIndex2]
						}
					}
				}

				userAttrInfo, ok := userAttrMap[curUid]
				if ok {
					result.UserArr = append(result.UserArr, &static.Msg_S2C_ReplayInfoUser{
						Uid:            userAttrInfo.Uid,
						Nickname:       userAttrInfo.Nickname,
						Imgurl:         userAttrInfo.Imgurl,
						Sex:            userAttrInfo.Sex,
						LastRoundScore: score,
						Vitamin:        gameReplay.UVitaminMap[userAttrInfo.Uid],
					})
				}

			}
			// syslog.Logger().Debug("玩家战绩数据返回：", public.HF_JtoB(getCommonResp(constant.MsgTypeGetReplay, result)))
			w.Write(static.HF_JtoB(getCommonResp(consts.MsgTypeGetReplay, result)))
		}
	default:
		w.Write(getErrReturn(msg.Header, xerrors.NewXError("unknown head")))
		return
	}
}

func getNewGustCode() string {
	code := ""
	count := 0
	for {
		count++
		if count > 100 {
			xlog.Logger().Errorln("getNewGustCode err")
			break
		}
		codeS := static.HF_GetRandomString(5)
		code = fmt.Sprintf("guest%s", codeS)

		person := GetPersonMgr().GetPersonbyGuestid(code)
		if person != nil {
			continue
		} else {
			break
		}
	}
	return string(code)
}

// 获取登录返回结果
func getLoginResp(msgType string, person *static.Person, isNew bool) (*static.Msg_Header, *xerrors.XError) {
	if person.IsBlack == consts.BlackStatusForbiddenLogin {
		msg := fmt.Sprintf("您的ID:%d由于长时间未登录已封号，请联系客服", person.Uid)
		err := common.XError{
			Code: common.BlackListErrorCode,
			Msg:  msg,
		}

		return nil, &err // xerrors.BlackListError
	} else if person.IsBlack == consts.BlackStatusAccountAbnormal {
		return nil, xerrors.NewXError("账号异常，无法登录")
	}

	// syslog.Logger().Errorln("loginMsg person:", person)

	result := new(Msg_Login_MsgData_Return)
	result.Uid = person.Uid
	result.Area = person.Area
	result.PlayTime = person.PlayTime
	result.Hall = GetServer().Con.Center
	result.Gate = GetServer().Con.Gate
	result.Api = GetServer().Con.Api
	result.Token = person.Token
	result.NeedBindWechat = false
	result.IsNew = isNew
	if person.UserType == consts.UserTypeHW {
		appCfg := GetServer().GetAppConfig(consts.APPHuaWei)
		if appCfg != nil {
			result.AppId = appCfg.PayId
			result.AppToken = appCfg.PayToken
		}
	} else {
		result.AppId = GetServer().ConServers.AppId
		result.AppToken = GetServer().ConServers.AppToken
	}
	result.InsureendLine = GetServer().ConServers.InsureendLine
	result.BroadcastCost = GetServer().ConServers.BroadcastCost
	// 获取维护公告
	notice := CheckServerMaintainWithWhite(person.Uid, static.NoticeMaintainServerAllServer)
	if notice != nil {
		// 判断用户是否在游戏
		if person.TableId == 0 {
			// 不在游戏返回维护公告
			maintainError := xerrors.NewXError(static.HF_JtoA(notice))
			maintainError.Code = xerrors.ServerMaintainError.Code
			return nil, maintainError
		}
	}
	return getCommonResp(msgType, result), nil
}

// 获取错误返回结果
func getErrReturn(msgType string, err *xerrors.XError) []byte {
	resultHead := new(static.Msg_Header)
	resultHead.Header = msgType
	// resultHead.Data = err.Error()
	resultHead.ErrCode = err.Code
	resultHead.Sign.Encode = GetServer().Con.Encode
	resultHead.Sign.Time = time.Now().Unix()

	xlog.Logger().WithFields(logrus.Fields{
		"errhead": msgType,
		"errmsg":  err.Error(),
		"errcode": err.Code,
	}).Infoln("【SEND ERROR HTTP RESP】")

	switch resultHead.Sign.Encode {
	case consts.EncodeNone: // 不加密
		resultHead.Data = err.Error()
	case consts.EncodeAes: // aes + base64
		bytes, _ := goEncrypt.AesCTR_Encrypt([]byte(err.Error()), []byte(GetServer().Con.EncodeClientKey))
		resultHead.Data = base64.URLEncoding.EncodeToString(bytes)
	}

	return static.HF_JtoB(resultHead)
}

// 获取通用返回结果
func getCommonResp(msgType string, v interface{}) *static.Msg_Header {
	resp := new(static.Msg_Header)
	resp.Header = msgType
	resp.ErrCode = xerrors.SuccessCode
	resp.Sign.Encode = GetServer().Con.Encode
	resp.Sign.Time = time.Now().Unix()
	xlog.Logger().WithFields(logrus.Fields{
		"head": msgType,
		"data": static.HF_JtoA(v),
	}).Infoln("【SEND HTTP RESP】")
	switch resp.Sign.Encode {
	case consts.EncodeNone: // 不加密
		resp.Data = static.HF_JtoA(v)
	case consts.EncodeAes: // aes + base64
		var bytes []byte
		if reflect.TypeOf(v).Kind() == reflect.String {
			bytes, _ = goEncrypt.AesCTR_Encrypt([]byte(v.(string)), []byte(GetServer().Con.EncodeClientKey))
		} else {
			bytes, _ = goEncrypt.AesCTR_Encrypt([]byte(static.HF_JtoA(v)), []byte(GetServer().Con.EncodeClientKey))
		}
		resp.Data = base64.URLEncoding.EncodeToString(bytes)
	}

	return resp
}

// 消息解密
func getCommonMsg(msg *static.Msg_Header) error {
	switch msg.Sign.Encode {
	case consts.EncodeNone: // 不加密不处理
	case consts.EncodeAes: // aes + base64
		data, err := base64.URLEncoding.DecodeString(msg.Data)
		if err != nil {
			return err
		}

		bytes, err := goEncrypt.AesCTR_Decrypt(data, []byte(GetServer().Con.EncodeClientKey))
		if err != nil {
			return err
		}
		msg.Data = string(bytes)
		xlog.Logger().Infoln("解密后消息:", msg.Data)
	}
	return nil
}

/*
http路由服务器重新读取配置文件接口实现
*/
func reloadConfigs(server int, games []int) error {
	callRPC := func() error {
		msg := &static.Msg_ReloadConfig{
			Games:    games,
			CallGame: server == 0 || server == 3,
			CallHall: server == 0 || server == 2,
		}
		var err error
		if server == 0 || server == 1 {
			if err = GetDBMgr().ReadAllConfig(); err != nil {
				return err
			}
		}
		tips, err := GetServer().CallHall("NewServerMsg", consts.MsgTypeReloadConfig, msg)
		xlog.Logger().Debug("callRPC:", string(tips))
		return err
	}

	xlog.Logger().Debug(">>>>>>>reloadConfigs，wuhan:", server, ";games:", games)
	if err := callRPC(); err != nil {
		return err
	}
	return nil
}

// 用户登录事件
func OnUserLogin(newUser bool, user *models.User, platform int, ip string, machineCode string) error {
	var err error
	// 如果不是新用户 如果用户的最后一次登录时间是在金币场上线之前，则赠送初始金币
	if !newUser && !GetServer().ConServers.NewGoldTime.IsZero() {
		if user.LastLoginAt == nil || (!user.LastLoginAt.IsZero() && user.LastLoginAt.Before(GetServer().ConServers.NewGoldTime)) {
			giveGold := GetServer().ConServers.NewGold - user.Gold
			if giveGold != 0 { // 需要赠送
				xlog.Logger().Infof("old user gold fix: uid(%d), oldGold(%d), newGold(%d), fix(%d), newGoldTime(%s)",
					user.Id,
					user.Gold,
					GetServer().ConServers.NewGold,
					giveGold,
					GetServer().ConServers.NewGoldTime.Format(static.TIMEFORMAT),
				)
				tx := GetDBMgr().GetDBmControl().Begin()
				_, user.Gold, err = wealthtalk.UpdateGold(user.Id, giveGold, models.CostTypeRegister, tx)
				if err != nil {
					tx.Rollback()
					return err
				}
				err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(user.Id, "Gold", user.Gold)
				if err != nil {
					tx.Rollback()
					return err
				}
				tx.Commit()
			}
		}
	}

	addLoginRecord(GetDBMgr().GetDBmControl(), user.Id, platform, ip, machineCode)

	GetServer().UpdateUserImgUrlFromQiNiu(user.Id, user.Imgurl)

	return nil
}

// 新增登录记录
func addLoginRecord(db *gorm.DB, uid int64, platform int, ip string, machineCode string) error {
	var err error
	// 添加登录记录
	var record models.UserLoginRecord
	record.Uid = uid
	record.Platform = platform
	record.Ip = ip
	record.MachineCode = machineCode
	if err = db.Create(&record).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	// 更新用户最近一次登录信息
	updateMap := make(map[string]interface{})
	updateMap["last_login_ip"] = ip
	updateMap["last_login_at"] = time.Now()
	if err = db.Model(models.User{Id: uid}).Updates(updateMap).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	return nil
}
