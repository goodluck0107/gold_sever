// 区域相关的辅助函数
package center

import (
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

// 区域码效验
func AreaCodeCheck(code string) bool {
	return GetDBMgr().GetDBrControl().CheckAreaCodeExist(static.AreaPackageKindCard, code)
}

// 通过游戏id得到包key
func GetAreaPackageKeyByKid(kid int) string {
	k, _ := GetDBMgr().GetDBrControl().RedisV2.HGet(static.AreaKindIdRedisKey(static.AreaPackageKindCard), static.HF_Itoa(kid)).Result()
	if len(k) == 0 {
		return GetDBMgr().GetDBrControl().RedisV2.HGet(static.AreaKindIdRedisKey(static.AreaPackageKindGold), static.HF_Itoa(kid)).Val()
	} else {
		return k
	}
}

// 通过游戏包key得到包
func GetAreaPackageByPKey(key string) *static.AreaPackageCompiled {
	apc := new(static.AreaPackageCompiled)

	cmd := GetDBMgr().GetDBrControl().RedisV2.HGet(static.AreaPackageRedisKey(static.AreaPackageKindCard), key)
	if cmd.Err() != nil {
		cmd = GetDBMgr().GetDBrControl().RedisV2.HGet(static.AreaPackageRedisKey(static.AreaPackageKindGold), key)
	}

	err := cmd.Scan(apc)
	if err != nil {
		return nil
	}

	return apc
}

// 得到区域微信信息
func GetAreaWeChat(code string) *static.AreaGmWeChatInfo {
	return GetDBMgr().GetDBrControl().SelectAreaWxByCode(code)
}

// 得到所有包
func GetAreaPackages() static.AreaPkgCompiledList {
	return GetDBMgr().GetDBrControl().SelectAllAreaPkgs()
}

// 得到所有包
func GetAreaPackagesByKind(kind static.AreaPackageKind) static.AreaPkgCompiledList {
	apcs, err := GetDBMgr().GetDBrControl().SelectAllAreaPkgsByKind(kind)
	if err != nil {
		xlog.Logger().Errorln("SelectAllAreaPkgsByKind error:", err)
	}
	return apcs
}

// 根据游戏id得到区域包
func GetAreaPackageByKid(kid int) *static.AreaPackageCompiled {
	apc, err := GetDBMgr().GetDBrControl().SelectAreaPkgByKindId(static.AreaPackageKindCard, kid)
	if err != nil {
		apc, err = GetDBMgr().GetDBrControl().SelectAreaPkgByKindId(static.AreaPackageKindGold, kid)
		if err != nil {
			xlog.Logger().Error(err.Error())
			return nil
		} else {
			return apc
		}
	} else {
		return apc
	}
}

// 根据游戏id得到区域包
func GetAreaPackageByKidAndArea(areaKind static.AreaPackageKind, kid int) *static.AreaPackageCompiled {
	apc, err := GetDBMgr().GetDBrControl().SelectAreaPkgByKindId(areaKind, kid)
	if err != nil {
		xlog.Logger().Error(err)
		return nil
	} else {
		return apc
	}
}

// 根据游戏id得到游戏包
func GetAreaGameByKid(kid int) *static.AreaGameCompiled {
	apc := GetAreaPackageByKid(kid)
	if apc == nil {
		return nil
	}
	for _, game := range apc.Games {
		if game == nil {
			continue
		}
		if game.KindId == kid {
			return game
		}
	}
	return nil
}

//func GetAreaGamesByKids(kids ...interface{}) static.AreaGameCompiledList {
//	agcs := make(static.AreaGameCompiledList, 0)
//	apcs, err := GetDBMgr().GetDBrControl().SelectAreaPkgsByKindIds(kids...)
//
//	if err == nil {
//		for _, apc := range apcs {
//			if apc == nil {
//				continue
//			}
//			agcs = append(agcs, apc.Games...)
//		}
//	} else {
//		syslog.Logger().Errorln("SelectAreaPkgsByKindIds", err)
//	}
//
//	agcs.UnDuplicate()
//	return agcs
//}

// 得到区域下所有包
func GetAreaPackagesByCode(kind static.AreaPackageKind, code string, nullIsAll bool) static.AreaPkgCompiledList {
	res := make(static.AreaPkgCompiledList, 0)
	if code == "" {
		if nullIsAll {
			return GetAreaPackagesByKind(kind)
		} else {
			return res
		}
	}
	var err error
	res, err = GetDBMgr().GetDBrControl().SelectAreaPkgsByCode(kind, code)
	if err != nil {
		xlog.Logger().Errorln("GetAreaPackagesByCode.SelectAreaPkgByCode.error", code, err)
	}
	return res
}

// 得到小程序中区域内的所有游戏包
func GetAppletAreaGamesByCode(code string) *static.AppletAreaGamesCompiled {
	res, err := GetDBMgr().GetDBrControl().SelectAppletAreaGamesByCode(code)
	if err != nil {
		xlog.Logger().Errorln("GetAreaPackagesByCode.SelectAreaPkgByCode.error", code, err)
	}
	return res
}

// 得到通用包
func GetAreaPackagesUniver(kind static.AreaPackageKind) static.AreaPkgCompiledList {
	apcs, err := GetDBMgr().GetDBrControl().SelectAreaPkgsUniversal(kind)
	if err != nil {
		xlog.Logger().Errorln("SelectAreaPkgsUniversal error:", err)
	}
	return apcs
}

// 得到推荐包
func GetAreaPackagesRecomd(kind static.AreaPackageKind) static.AreaPkgCompiledList {
	apcs, err := GetDBMgr().GetDBrControl().SelectAreaPkgsRecommend(kind)
	if err != nil {
		xlog.Logger().Errorln("SelectAreaPkgsRecommend error:", err)
	}
	return apcs
}

// 是否为官方自营区域
func IsOfficialWorkArea(code string) bool {
	if !AreaCodeCheck(code) {
		return false
	}

	agentAreas, err := GetAllianceMgr().GetAgentLeagueArea()
	if err != nil {
		xlog.Logger().Errorln("GetAgentLeagueArea error:", err)
		return false
	}

	in := func(lcode string) bool {
		for _, ac := range agentAreas {
			if ac == lcode {
				return true
			}
		}
		return false
	}

	return !in(code)
}

// 得到官方自营包
func GetAreaPkgOfficial() static.AreaPkgCompiledList {
	result := make(static.AreaPkgCompiledList, 0)
	agentAreas, err := GetAllianceMgr().GetAgentLeagueArea()
	if err != nil {
		xlog.Logger().Errorln("GetAgentLeagueArea error:", err)
		return result
	}
	codes := GetDBMgr().GetDBrControl().SelectAllAreaCodesByKind(static.AreaPackageKindCard).WeedOut(agentAreas...).UnDuplicate()

	apcs, err := GetDBMgr().GetDBrControl().SelectAreaPkgsByCode(static.AreaPackageKindCard, codes...)

	if err != nil {
		xlog.Logger().Errorln("GetAreaPkgOfficial.SelectAreaPkgsByCode error.", err)
	}

	return apcs
}

// 获取区域下的游戏列表
func getGameCollections() []*static.Msg_S2C_GameCollections {
	result := make([]*static.Msg_S2C_GameCollections, 0)

	// 先做一个假数据，只有跑的快

	var gameCollection static.Msg_S2C_GameCollections
	gameCollection.ID = 2
	gameCollection.Name = "扑克"
	gameCollection.GameList = make([]*static.Msg_S2C_GameList, 0)
	var game static.Msg_S2C_GameList
	game.Name = "跑得快(金币场)"
	game.KindId = 387
	game.PackageName = "跑得快(金币场)"
	game.PackageKey = "pdkgd"
	game.PackageVersion = "1.0.2.9"
	game.Icon = ""
	game.Online = 100
	game.MatchFlag = 0
	// game.OrderId = 8
	game.SiteType = make([]*static.Msg_S2C_SiteType, 0)
	var site static.Msg_S2C_SiteType
	site.Name = "初级场"
	site.Type = 1
	site.Online = 0
	site.MinScore = 500
	site.MaxScore = 0
	site.Difen = 500

	game.SiteType = append(game.SiteType, &site)
	gameCollection.GameList = append(gameCollection.GameList, &game)
	result = append(result, &gameCollection)
	return result
	// collections := GetAreaMgr().GetGamesCollections()
	// result := make([]*public.Msg_S2C_GameCollections, 0)
	// for _, c := range collections {
	// 	collection := new(public.Msg_S2C_GameCollections)
	// 	collection.Name = c.Name
	// 	collection.ID = c.ID
	// 	for _, game := range c.Games {
	// 		obj := new(public.Msg_S2C_GameList)
	// 		obj.PackageVersion = game.PackageVersion
	// 		obj.PackageName = game.PackageName
	// 		obj.PackageKey = game.PackageKey
	// 		obj.Name = game.Name
	// 		obj.KindId = game.KindId
	// 		obj.Icon = game.Icon
	// 		obj.OrderId = game.OrderId
	// 		//暂时屏蔽掉obj.Appid = game.Appid
	// 		siteTypes := GetServer().GetSiteTypeByKindId(public.GAME_TYPE_GOLD, game.KindId)
	// 		totalOnline := 0
	// 		for _, item := range siteTypes {
	// 			v := &public.Msg_S2C_SiteType{
	// 				Type: item,
	// 			}
	// 			// 获取房间配置
	// 			c := GetServer().GetRoomConfig(obj.KindId, item)
	// 			if c != nil {
	// 				v.Name = c.Name
	// 				v.MaxScore = c.MaxScore
	// 				v.MinScore = c.MinScore
	// 				v.Difen = int(c.Config["difen"].(float64))
	// 				// 获取该游戏该场次的在线人数(加上基础人数)
	// 				v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type) + c.BaseNum
	// 				v.SitMode = c.SitMode
	// 				v.MatchFlag = 0
	// 				mat := GetServer().GetMatchConfig(game.KindId, v.Type) //是否开启了排位赛
	// 				if mat != nil {
	// 					for _, matitem := range mat {
	// 						if matitem.Flag > 0 {
	// 							obj.MatchFlag = matitem.Flag
	// 							if matitem.State == 1 {
	// 								v.MatchFlag = matitem.Flag
	// 								//有任何一个排位赛开启了，就认为开启了
	// 								break
	// 							}
	// 						}
	// 					}
	// 				}
	// 			} else {
	// 				v.SitMode = 0
	// 				v.MaxScore = 0
	// 				v.MinScore = 0
	// 				// 获取该游戏该场次的在线人数(加上基础人数)
	// 				v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type)
	// 			}
	// 			totalOnline = totalOnline + v.Online
	// 			obj.SiteType = append(obj.SiteType, v)
	// 		}
	//
	// 		obj.Online = totalOnline
	// 		collection.GameList = append(collection.GameList, obj)
	// 	}
	// 	result = append(result, collection)
	// }
	// return result
}

// 获取区域下的游戏列表
func getGameList(area string) []*static.Msg_S2C_GameList {
	// 得到所有金币场的包
	goldPkgs := GetAreaPackagesByKind(static.AreaPackageKindGold)
	result := make([]*static.Msg_S2C_GameList, 0)

	for _, pkg := range goldPkgs {
		if pkg == nil {
			continue
		}
		for _, game := range pkg.Games {
			obj := new(static.Msg_S2C_GameList)
			obj.PackageVersion = game.PackageVersion
			obj.PackageName = game.PackageName
			obj.PackageKey = game.PackageKey
			obj.Name = game.Name
			obj.KindId = game.KindId
			obj.Icon = game.Icon
			// obj.OrderId = game.OrderId
			siteTypes := GetServer().GetSiteTypeByKindId(static.GAME_TYPE_GOLD, game.KindId)
			totalOnline := 0
			for _, item := range siteTypes {
				v := &static.Msg_S2C_SiteType{
					Type: item,
				}
				// 获取房间配置
				c := GetServer().GetRoomConfig(obj.KindId, item)
				if c != nil {
					v.Name = c.Name
					v.MaxScore = c.MaxScore
					v.MinScore = c.MinScore
					v.Difen = int(c.Config["difen"].(float64))
					// 获取该游戏该场次的在线人数(加上基础人数)
					v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type) + c.BaseNum
					v.SitMode = c.SitMode
					v.MatchFlag = 0
					// mat := GetServer().GetMatchConfig(game.KindId, v.Type) // 是否开启了排位赛
					// if mat != nil {
					// 	for _, matitem := range mat {
					// 		if matitem.Flag > 0 {
					// 			obj.MatchFlag = matitem.Flag
					// 			if matitem.State == 1 {
					// 				v.MatchFlag = matitem.Flag
					// 				//有任何一个排位赛开启了，就认为开启了
					// 				break
					// 			}
					// 		}
					// 	}
					// }
				} else {
					v.SitMode = 0
					v.MaxScore = 0
					v.MinScore = 0
					// 获取该游戏该场次的在线人数(加上基础人数)
					v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type)
				}
				totalOnline = totalOnline + v.Online
				obj.SiteType = append(obj.SiteType, v)
			}
			obj.Online = totalOnline
			result = append(result, obj)
		}
	}
	return result
}

// 游戏的场次信息
func getGameSiteList(kindId int) *static.Msg_S2C_SiteList {
	game := GetAreaGameByKid(kindId)
	if game == nil {
		return nil
	}

	var ack static.Msg_S2C_SiteList
	siteTypes := GetServer().GetSiteTypeByKindId(static.GAME_TYPE_GOLD, game.KindId)
	for _, typeId := range siteTypes {
		v := &static.Msg_S2C_SiteType{
			Type: typeId,
		}
		// 获取房间配置
		c := GetServer().GetRoomConfig(game.KindId, typeId)
		if c != nil {
			v.Name = c.Name
			v.MaxScore = c.MaxScore
			v.MinScore = c.MinScore
			v.Difen = int(c.Config["difen"].(float64))
			// 获取该游戏该场次的在线人数(加上基础人数)
			v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type) + c.BaseNum
			v.SitMode = c.SitMode
			v.MatchFlag = 0
		} else {
			v.SitMode = 0
			v.MaxScore = 0
			v.MinScore = 0
			// 获取该游戏该场次的在线人数(加上基础人数)
			v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type)
		}
		// 服务器状态
		if v.Online <= consts.SERVER_FREE_PLAYER_NUM {
			v.Sta = consts.SERVER_STA_FREE
		} else if v.Online <= consts.SERVER_NORMAL_PLAYER_NUM {
			v.Sta = consts.SERVER_STA_NORMAL
		} else if v.Online <= consts.SERVER_BUSY_PLAYER_NUM {
			v.Sta = consts.SERVER_STA_BUSY
		} else {
			v.Sta = consts.SERVER_STA_HOT
		}
		ack.SiteList = append(ack.SiteList, v)
	}

	return &ack
}
