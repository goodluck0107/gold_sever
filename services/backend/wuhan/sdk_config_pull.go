// 从php拉取配置 写入redis
package wuhan

// // 从PHP拉取一些初始化配置
// func PullConfig(host string) {
// 	var err error
// 	err = pullCfgAllowanceGift(host)
// 	if err != nil {
// 		syslog.Logger().Error(err)
// 		{ // 测试代码
// 			syslog.AnyErrors(GetDBMgr().GetDBrControl().RedisV2.Set(constant.RedisKeyAllowanceGift, &public.AllowanceGift{
// 				ExchangeGoldNum: 38888,
// 				NeedDiamondNum:  30,
// 			}, 0).Err(),
// 				"插入默认的低保礼包数据出错")
// 		}
// 	} else {
// 		syslog.Logger().Info("pull allowance gift config succeed...")
// 	}
// }
//
// // 拉取低保礼包配置
// func pullCfgAllowanceGift(host string) error {
// 	data, err := util.HttpGet(host, nil)
// 	if err != nil {
// 		syslog.Logger().Error("http get explain error:", err)
// 		return err
// 	}
// 	res := new(public.MsgAllowanceGift)
// 	err = json.Unmarshal(data, res)
// 	if err != nil {
// 		return err
// 	}
// 	if res.Code != 0 {
// 		return fmt.Errorf("php explain response error: %s", res.Msg)
// 	}
// 	return GetDBMgr().GetDBrControl().RedisV2.Set(constant.RedisKeyAllowanceGift, res.Data, 0).Err()
// }
