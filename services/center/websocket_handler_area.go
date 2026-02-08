package center

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sort"
	"strings"
)

// chess新版加入区域
func Proto_AreaEnter(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_AreaIn)
	req.Code = consts.DefaultAreaCode
	// 区域默认有效
	req.IsValid = true

	// 校验区域是否存在的有效性
	if !AreaCodeCheck(req.Code) {
		req.IsValid = false
	}

	// 区域内是否存在游戏包
	if len(GetAreaPackagesByCode(static.AreaPackageKindCard, req.Code, false)) == 0 && len(GetAreaPackagesByCode(static.AreaPackageKindGold, req.Code, false)) == 0 {
		req.IsValid = false
	}

	// 更新mysql
	tx := GetDBMgr().GetDBmControl().Begin()
	if err := tx.Model(&models.User{Id: p.Uid}).Update("area", req.Code).Error; err != nil {
		xlog.Logger().Errorln("AreaEnter.UpdateMysql.Error:", err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	// 更新redis
	if err := GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Area", req.Code); err != nil {
		xlog.Logger().Errorln("AreaEnter.UpdateRedis.Error:", err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	tx.Commit()
	p.Area = req.Code
	return xerrors.SuccessCode, req
}

// chess搜索游戏
func Proto_AreaPackageSeek(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_AreaGameSeek)
	// 去掉关键字的空格和换行符
	req.Keyword = strings.Replace(strings.Replace(req.Keyword, " ", "", -1), "\n", "", -1)

	ack := new(static.AreaPackageSeek)
	ack.Packages = GetAreaPackagesByCode(static.AreaPackageKindCard, req.Code, true)

	// 排序
	if len(req.Code) > 0 && len(ack.Packages) > 1 {
		sort.Sort(&ack.Packages)
	}

	// 分析请求
	switch req.Type {
	case static.AreaSeekTypeRecommend:
		{
			if err, pkg := AreaGameListRecommend(p); err != nil {
				xlog.Logger().Errorln("Proto_AreaPackageSeek.AreaGameListRecommend.Error:", err.Error())
				return xerrors.ResultErrorCode, xerrors.NewXError("获取推荐玩法失败")
			} else {
				ack.Packages = append(ack.Packages, pkg...)
				if len(ack.Packages) >= consts.AREAGAME_RECOMMEND_MAX {
					ack.Packages = ack.Packages[:consts.AREAGAME_RECOMMEND_MAX]
				}
			}
		}
	case static.AreaSeekTypeNearby:
		{
			// TODO
			// if err, pkg := AreGameListNearby(p); err != nil {
			// 	syslog.Logger().Errorln("Proto_AreaPackageSeek.AreGameListNearby.Error:", err.Error())
			// 	return xerrors.ResultErrorCode, xerrors.NewXError("获取推荐玩法失败")
			// } else {
			// 	ack = append(ack, pkg...)
			// 	ack.Limit()
			// }
		}
	default:
		{
			if req.Keyword == "" {
				// 如果为区域内且不按关键字搜索，则追加上通用包
				if AreaCodeCheck(req.Code) {
					universal := GetAreaPackagesUniver(static.AreaPackageKindCard)
					if len(universal) > 1 {
						sort.Sort(&universal)
					}
					for _, pkgUniversal := range universal {
						if pkgUniversal == nil {
							continue
						}
						ack.Packages = append(ack.Packages, pkgUniversal)
					}
				}
			} else {
				ack.Packages = ack.Packages.Search(req.Keyword)
			}
		}
	}
	ack.PackageType = req.PackageType
	ack.Keyword = req.Keyword
	ack.Type = req.Type
	ack.AreaCode = req.Code
	ack.Packages = ack.Packages.SearchByPType(req.PackageType)
	ack.Packages.UnDuplicate()
	return xerrors.SuccessCode, ack
}

// 获取区域内房卡游戏包
func Proto_AreaPackageGameCardListMain(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 小程序执行独立的包管理
	if p.Platform == consts.PlatformWechatApplet {
		return Proto_AreaPackageAppletGameCardListMain(s, p, data)
	}

	var ack []*static.AreaPackageCompiled

	// 展示顺序 最近游戏 -- 区域游戏 -- 通用游戏
	checkHas := func(pkgKey string) bool {
		for _, pkg := range ack {
			if pkg == nil {
				continue
			}
			if pkg.PackageKey == pkgKey {
				return true
			}
		}
		return false
	}

	// 最近游戏的区域包
	if gameHistory, err := GetDBMgr().GetDBrControl().GamePlaysSelectAll(p.Uid); err == nil {
		if len(gameHistory) > 1 {
			sort.Sort(models.UserGameHistoryList(gameHistory))
		}

		for i := 0; i < len(gameHistory); i++ {
			historyAreaPkg := GetAreaPackageByKidAndArea(static.AreaPackageKindCard, gameHistory[i].KindId)
			if historyAreaPkg == nil {
				continue
			}
			if checkHas(GetAreaPackageKeyByKid(gameHistory[i].KindId)) {
				xlog.Logger().Warningln("最近游戏的包 出现一条重复包 pkgKey = ", historyAreaPkg.PackageKey, "游戏包kid = ", gameHistory[i].KindId)
				continue
			}
			// 最近游戏最大展示个数
			if len(ack) >= consts.AREAGAME_HISTORY_MAX {
				xlog.Logger().Warningln("展示最近游戏的区域包已达最大值")
				break
			}
			// 对游戏包排序
			for i := 0; i < len(historyAreaPkg.Games)-1; i++ {
				for j := i + 1; j < len(historyAreaPkg.Games); j++ {
					iSort := 99
					jSort := 99
					// 在 gameHistory 中 先找到的 就是应该排到最前面的
					for k := 0; k < len(gameHistory); k++ {
						if historyAreaPkg.Games[i].KindId == gameHistory[k].KindId {
							iSort = k
						}
						if historyAreaPkg.Games[j].KindId == gameHistory[k].KindId {
							jSort = k
						}
					}
					if iSort > jSort {
						var tmp *static.AreaGameCompiled
						tmp = historyAreaPkg.Games[j]
						historyAreaPkg.Games[j] = historyAreaPkg.Games[i]
						historyAreaPkg.Games[i] = tmp
					}
				}
			}
			ack = append(ack, historyAreaPkg)
		}
		xlog.Logger().Debugln("最近游戏列表：", fmt.Sprintf("%+v", ack))
	}

	// 当前区域包
	areaCode := p.Area
	var areaPkgLists static.AreaPkgCompiledList
	if areaCode == "" {
		// 展示推荐玩法
		//areaPkgLists = GetAreaPackagesRecomd(public.AreaPackageKindCard)
		areaPkgLists = make(static.AreaPkgCompiledList, 0)
	} else {
		// 区域游戏列表
		areaPkgLists = GetAreaPackagesByCode(static.AreaPackageKindCard, areaCode, false)
	}
	// 排序
	if len(areaPkgLists) > 1 {
		sort.Sort(&areaPkgLists)
	}
	xlog.Logger().Debugln("获取玩家当前区域下子包：玩家区域：", areaCode, "游戏列表：", fmt.Sprintf("%+v", areaPkgLists))
	// 加入返回列表中
	for i := 0; i < len(areaPkgLists); i++ {
		if checkHas(areaPkgLists[i].PackageKey) {
			xlog.Logger().Warningln("玩家当前区域包与返回的包重合 出现一条重复包 pkgKey = ", areaPkgLists[i].PackageKey)
			continue
		}
		ack = append(ack, areaPkgLists[i])
	}

	// 通用游戏
	univerPkgLists := GetAreaPackagesUniver(static.AreaPackageKindCard)
	// 排序
	if len(univerPkgLists) > 1 {
		sort.Sort(&univerPkgLists)
	}
	xlog.Logger().Debugln("通用包游戏列表：", fmt.Sprintf("%+v", univerPkgLists))
	// 加入返回列表
	for i := 0; i < len(univerPkgLists); i++ {
		if checkHas(univerPkgLists[i].PackageKey) {
			xlog.Logger().Warningln("通用包与返回的包重合 出现一条重复包 pkgKey =", univerPkgLists[i].PackageKey)
			continue
		}
		ack = append(ack, univerPkgLists[i])
	}

	// 限时免费标签
	freeGameMap := GetServer().GetLimitFreeGameKindIds()
	for _, pkg := range ack {
		for _, game := range pkg.Games {
			if free, ok := freeGameMap[game.KindId]; ok {
				game.TimeLimitFree = free
			} else {
				game.TimeLimitFree = false
			}
		}
	}

	return xerrors.SuccessCode, ack
}

// 获取区域内金币游戏包
func Proto_AreaPackageGameGoldListMain(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 展示顺序 区域游戏 -- 定位区域 -- 包厢所在区域 -- 最近游戏 -- 通用游戏 -- 联运游戏（传奇捕鱼类）

	// 区域包列表
	var ack []*static.AreaPackageCompiled
	// 区域游戏 -- 定位区域 -- 包厢所在区域 -- 最近游戏 重复的不算, 最多展示三个,
	var showAreaCode []string

	// 去重
	checkHas := func(pkgKey string) bool {
		for _, pkg := range ack {
			if pkg == nil {
				continue
			}
			if pkg.PackageKey == pkgKey {
				return true
			}
		}
		return false
	}

	// 判断重复
	isRepeat := func(code string) bool {
		for _, val := range showAreaCode {
			if val == code {
				return true
			}
		}
		return false
	}

	// 选择区域
	areaCode := p.Area
	var areaPkgLists static.AreaPkgCompiledList
	if areaCode == "" {
		// 展示推荐的玩法
		//areaPkgLists = GetAreaPackagesRecomd(public.AreaPackageKindGold)
		areaPkgLists = make(static.AreaPkgCompiledList, 0)
	} else {
		// 区域游戏列表
		areaPkgLists = GetAreaPackagesByCode(static.AreaPackageKindGold, areaCode, false)
		if len(areaPkgLists) > 0 {
			showAreaCode = append(showAreaCode, p.Area)
		}
	}
	// 排序
	if len(areaPkgLists) > 1 {
		sort.Sort(&areaPkgLists)
	}
	xlog.Logger().Debugln("获取玩家选择区域下子包：玩家区域：", areaCode, "游戏列表：", fmt.Sprintf("%+v", areaPkgLists))
	// 加入返回列表中
	for i := 0; i < len(areaPkgLists); i++ {
		if checkHas(areaPkgLists[i].PackageKey) {
			xlog.Logger().Warningln("玩家选择区域包与返回的包重合 出现一条重复包 pkgKey = ", areaPkgLists[i].PackageKey)
			continue
		}
		ack = append(ack, areaPkgLists[i])
	}

	// 定位区域
	if len(p.Area2nd) > 0 && isRepeat(p.Area2nd) == false {
		gpsPkgList := GetAreaPackagesByCode(static.AreaPackageKindGold, p.Area2nd, false)
		if len(gpsPkgList) > 0 {
			showAreaCode = append(showAreaCode, p.Area2nd)
		}
		if len(gpsPkgList) > 1 {
			sort.Sort(&gpsPkgList)
		}
		xlog.Logger().Debugln("获取玩家定位区域下子包：玩家区域：", p.Area2nd, "游戏列表：", fmt.Sprintf("%+v", gpsPkgList))
		for i := 0; i < len(gpsPkgList); i++ {
			if checkHas(gpsPkgList[i].PackageKey) {
				xlog.Logger().Warningln("玩家定位区域包与返回的包重合 出现一条重复包 pkgKey = ", gpsPkgList[i].PackageKey)
				continue
			}
			ack = append(ack, gpsPkgList[i])
		}
	}

	// 包厢所在区域
	if len(p.Area3rd) > 0 {
		if isRepeat(p.Area3rd) == false {
			housePkgList := GetAreaPackagesByCode(static.AreaPackageKindGold, p.Area3rd, false)
			if len(housePkgList) > 0 {
				showAreaCode = append(showAreaCode, p.Area3rd)
			}
			if len(housePkgList) > 1 {
				sort.Sort(&housePkgList)
			}
			xlog.Logger().Debugln("获取玩家包厢区域下子包：玩家区域：", p.Area3rd, "游戏列表：", fmt.Sprintf("%+v", housePkgList))
			for i := 0; i < len(housePkgList); i++ {
				if checkHas(housePkgList[i].PackageKey) {
					xlog.Logger().Warningln("玩家包厢区域与返回的包重合 出现一条重复包 pkgKey = ", housePkgList[i].PackageKey)
					continue
				}
				ack = append(ack, housePkgList[i])
			}
		}
	} else {
		// 查找包厢所在区域
		houseAreaInfo := GetDBMgr().GetHouseAreaList(p.Uid)
		if len(houseAreaInfo) > 0 {
			houseAreaCode := houseAreaInfo[0].Area
			// 更新mysql
			if err := GetDBMgr().GetDBmControl().Model(&models.User{}).Where("id = ?", p.Uid).Update("area3rd", houseAreaCode).Error; err != nil {
				xlog.Logger().Error(err)
			}
			// 更新redis
			if err := GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "Area3rd", houseAreaCode); err != nil {
				xlog.Logger().Error(err)
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
			// 更新缓存信息
			p.Area3rd = houseAreaCode
			if per := GetPlayerMgr().GetPlayer(p.Uid); per != nil {
				per.Info.Area3rd = houseAreaCode
			}
			// 更新返回列表
			if isRepeat(p.Area3rd) == false {
				housePkgList := GetAreaPackagesByCode(static.AreaPackageKindGold, p.Area3rd, false)
				if len(housePkgList) > 0 {
					showAreaCode = append(showAreaCode, p.Area3rd)
				}
				if len(housePkgList) > 1 {
					sort.Sort(&housePkgList)
				}
				xlog.Logger().Debugln("首次获取玩家包厢区域下子包：玩家区域：", p.Area3rd, "游戏列表：", fmt.Sprintf("%+v", housePkgList))
				for i := 0; i < len(housePkgList); i++ {
					if checkHas(housePkgList[i].PackageKey) {
						xlog.Logger().Warningln("首次玩家包厢区域与返回的包重合 出现一条重复包 pkgKey = ", housePkgList[i].PackageKey)
						continue
					}
					ack = append(ack, housePkgList[i])
				}
			}
		}
	}

	// 最近游戏的区域包
	if len(showAreaCode) < 3 {
		// 最近游戏的区域包
		if gameHistory, err := GetDBMgr().GetDBrControl().GamePlaysSelectAll(p.Uid); err == nil {
			if len(gameHistory) > 1 {
				sort.Sort(models.UserGameHistoryList(gameHistory))
			}
			for i := 0; i < len(gameHistory); i++ {
				// 从最近游戏的区域查找（不管是房卡场还是金币场）
				historyAreaPkg := GetAreaPackageByKid(gameHistory[i].KindId)
				if historyAreaPkg == nil {
					continue
				}
				// 最近游戏的区域的金币场
				latestPkgLists := GetAreaPackagesByCode(static.AreaPackageKindGold, historyAreaPkg.Code, false)
				if len(latestPkgLists) == 0 {
					break
				} else {
					showAreaCode = append(showAreaCode, historyAreaPkg.Code)
				}
				if len(latestPkgLists) > 1 {
					sort.Sort(&latestPkgLists)
				}
				// 加入返回列表中
				for i := 0; i < len(latestPkgLists); i++ {
					if checkHas(latestPkgLists[i].PackageKey) {
						xlog.Logger().Warningln("玩家最近一次区域包与返回的包重合 出现一条重复包 pkgKey = ", latestPkgLists[i].PackageKey)
						continue
					}
					ack = append(ack, latestPkgLists[i])
				}
				// 只展示最近的
				break
			}
		}
	}

	// 通用包
	univerPkgLists := GetAreaPackagesUniver(static.AreaPackageKindGold)
	// 排序
	if len(univerPkgLists) > 1 {
		sort.Sort(&univerPkgLists)
	}
	xlog.Logger().Debugln("通用包游戏列表：", fmt.Sprintf("%+v", univerPkgLists))
	// 加入返回列表中
	for i := 0; i < len(univerPkgLists); i++ {
		if checkHas(univerPkgLists[i].PackageKey) {
			xlog.Logger().Warningln("通用包与返回的包重合 出现一条重复包 pkgName =", univerPkgLists[i].PackageKey)
			continue
		}
		ack = append(ack, univerPkgLists[i])
	}

	// 小程序限定展示游戏的个数
	if p.Platform == consts.PlatformWechatApplet {
		for i := 0; i < len(ack); i++ {
			for j := 0; j < len(ack[i].Games); j++ {
				bFind := false
				for _, kid := range AppletGoldShowGame {
					if kid == ack[i].Games[j].KindId {
						bFind = true
						break
					}
				}
				// 踢除未找到的玩法
				if !bFind {
					ack[i].Games = append(ack[i].Games[:j], ack[i].Games[j+1:]...)
				}
			}
		}
	}

	return xerrors.SuccessCode, ack
}

// 获取区域内小程序房卡游戏包
func Proto_AreaPackageAppletGameCardListMain(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	var ack static.AreaGameCompiledList

	// 当前区域包
	areaCode := p.Area
	agcs := new(static.AppletAreaGamesCompiled)
	if areaCode == "" {
		// 展示所有玩法
		for i := 0; i < len(AppletCardShowArea); i++ {
			agcs := GetAppletAreaGamesByCode(AppletCardShowArea[i])
			for j := 0; j < len(agcs.Games); j++ {
				ack = append(ack, agcs.Games[j])
			}
		}
	} else {
		// 区域游戏列表
		agcs = GetAppletAreaGamesByCode(areaCode)
		if len(agcs.Games) == 0 {
			for i := 0; i < len(AppletCardShowArea); i++ {
				agcs = GetAppletAreaGamesByCode(AppletCardShowArea[i])
				for j := 0; j < len(agcs.Games); j++ {
					ack = append(ack, agcs.Games[j])
				}
			}
		} else {
			ack = agcs.Games
		}
	}

	xlog.Logger().Debugln("获取玩家当前区域下子包：玩家区域：", areaCode, "游戏列表：", fmt.Sprintf("%+v", ack))

	return xerrors.SuccessCode, ack
}

// 通过包名得到包下面的子游戏
func Proto_AreaGamesByPackageKey(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_AreaGamesByPkg)
	apc := GetAreaPackageByPKey(req.PackageKey)
	if apc != nil {
		return xerrors.ResultErrorCode, xerrors.NewXError("无效的游戏包")
	}
	return xerrors.SuccessCode, apc.Games
}

// 区域列表
func Proto_AreaList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 还未要求实现
	ack := new(static.Msg_HC_GameRules)
	return xerrors.SuccessCode, ack
}

// 通过桌子号得到桌子所属的包
func Proto_AreaPackageByKId(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_HG_UpdateGameServer)
	pkg := GetAreaPackageByKid(req.KindId)
	if pkg == nil {
		return xerrors.ResultErrorCode, xerrors.NewXError(fmt.Sprintf("区域游戏包不存在:%d", req.KindId)).Msg
	}

	return xerrors.SuccessCode, pkg
}

// 通过桌子号得到桌子所属的包
func Proto_AreaPackageByTId(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_TableDel)
	var kindID int
	if req.Id > 0 {
		table := GetTableMgr().GetTable(req.Id)
		if table == nil {
			// 缓存里面也没有，就报错
			redisTable, err := GetDBMgr().GetDBrControl().GetTableInfoPattern(req.Id)
			if err != nil {
				xlog.Logger().Error(err)
			} else {
				kindID = redisTable.KindId
			}
		} else {
			kindID = table.KindId
		}
	}
	if kindID > 0 {
		pkg := GetAreaPackageByKid(kindID)
		if pkg == nil {
			xer := xerrors.NewXError(fmt.Sprintf("区域游戏包不存在:%d", kindID))
			return xer.Code, xer.Msg
		}
		return xerrors.SuccessCode, pkg
	} else {
		return xerrors.TableIdError.Code, xerrors.TableIdError.Msg
	}
}

func Proto_GetCSWX(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	return xerrors.SuccessCode, GetAreaWeChat(p.Area)
}

// 推荐玩法
func AreaGameListRecommend(p *static.Person) (error, static.AreaPkgCompiledList) {
	ack := make(static.AreaPkgCompiledList, 0)

	// 游戏历史
	if gameHistory, err := GetDBMgr().GetDBrControl().GamePlaysSelectAll(p.Uid); err != nil {
		sort.Sort(models.UserGameHistoryList(gameHistory))
		if len(gameHistory) >= consts.AREAGAME_HISTORY_MAX {
			gameHistory = gameHistory[:consts.AREAGAME_HISTORY_MAX]
		}
		for _, history := range gameHistory {
			if history == nil {
				continue
			}
			if ack.Contains(GetAreaPackageKeyByKid(history.KindId)) {
				continue
			}
			historyGame := GetAreaPackageByKid(history.KindId)
			if historyGame == nil {
				xlog.Logger().Errorln("AreaGameListRecommend.historyGame.Error:筛选出来的历史常玩游戏存在，但是在取区域取游戏信息时为空,kindId:", history.KindId)
				continue
			}
			if historyGame.Engine != p.Engine {
				continue
			}
			ack = append(ack, historyGame)
		}
	}
	// 最后显示官方推荐
	ack = append(ack, GetAreaPkgOfficial()...)
	return nil, ack
}

// 周边玩法
func AreGameListNearby(p *static.Person) (error, static.AreaPkgCompiledList) {
	// TODO
	return nil, nil
}
