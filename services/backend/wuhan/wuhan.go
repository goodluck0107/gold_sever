package wuhan

import (
	"context"
	"encoding/json"
	"fmt"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/rpc"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

type Config struct {
	Host                  string `json:"host"`    //! GM服务器ip
	UserServerAddr        string `json:"user"`    //! login服务器内网ip
	Center                string `json:"center" ` //! hall服务器内网ip
	Gate                  string `json:"gate"`    //! gate服务器ip
	Redis                 string `json:"redis"`   //! redis
	RedisDB               int    `json:"redisdb"`
	RedisAuth             string `json:"redisauth"`       //!redis密码
	DB                    string `json:"db"`              //! 数据库
	Encode                int    `json:"encode"`          // 消息加密方式 0为不加密(使用加密后客户端发的消息也必须加密, 否则忽略)
	EncodeClientKey       string `json:"encodeclientkey"` // 与客户端通信的aes加密秘钥
	EncodePhpKey          string `json:"encodephpkey"`    // 与php通信的aes加密秘钥
	service2.CommonConfig        // 第三方服务配置
}

var serverSingleton *Server = nil

// ! 得到服务器指针
func GetServer() *Server {
	if serverSingleton == nil {
		serverSingleton = new(Server)
		serverSingleton.ShutDown = false
		serverSingleton.Con = new(Config)
		serverSingleton.Wait = new(sync.WaitGroup)
		serverSingleton.BlackIds = make(map[int64]int8)
	}

	return serverSingleton
}

type Server struct {
	Con             *Config         //! 配置
	Wait            *sync.WaitGroup //! 同步阻塞
	ShutDown        bool            //! 是否正在执行关闭
	HallServer      string
	GateServer      string
	BlackIds        map[int64]int8
	DataSynchronism bool //! 数据同步
}

func (s *Server) Host() (string, string) {
	ip := strings.Split(s.Con.Host, ":")
	return ip[0], ip[1]
}

func (s *Server) Start(ctx context.Context) error {
	err := models.InitModel(GetDBMgr().GetDBmControl())
	if err != nil {
		return err
	}
	service2.InitConfig(s.Con.CommonConfig)
	// 服务器启动 更新区域信息
	GetAreaMgr().Update()
	// 服务器启动 更新公告信息
	GetNoticeMgr().Update()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	//! 告诉服务器关闭
	var msg static.Msg_Null
	GetServer().CallHall("NewServerMsg", "adminserverclose", &msg)
	s.ShutDown = true
	s.Wait.Wait()
	GetDBMgr().Close()
	if rpcCliMap != nil {
		for _, v := range rpcCliMap {
			v.Close()
		}
	}
	xlog.Logger().Warn("wuhan shutdown")
	return nil
}

func (s *Server) RegisterRouter(ctx context.Context) {
	http.HandleFunc("/what", SuperOpt)
	http.HandleFunc("/fake", FakerOpt)
	http.HandleFunc("/owner", OwnerOpt)
	http.HandleFunc("/changeow", ChangeOwner)
	http.HandleFunc("/spinrec", SpinRecord)
	http.HandleFunc("/charges", ChargeStat)
	return
}

func (s *Server) RegisterRPC(ctx context.Context) {
	rpc.NewRpcxServer(s.Con.Host, "ServerMethod", new(ServerMethod), nil)
}

func (s *Server) LoadConfig(ctx context.Context) error {
	// 服务器启动 更新区域信息
	GetAreaMgr().Update()
	// 服务器启动 更新公告信息
	GetNoticeMgr().Update()
	return nil
}

func (s *Server) LoadServerConfig(ctx context.Context) error {
	if err := static.LoadServeYaml(s, true); err != nil {
		xlog.Logger().Error("InitConfig_yaml:", err)
		return err
	} else {
		xlog.Logger().Debug("load wuhan yaml etc succeed : ", fmt.Sprintf("%+v", s.Con))
	}
	return nil
}

func (s *Server) SetLoggerLevel() {
	xlog.SetFileLevel("info")
}

func (s *Server) Run(ctx context.Context) {
	// 定时获取在线人数
	go s.saveGameOnline()
	xlog.Logger().Infoln("run,time:", time.Now().Hour())
	ticker := time.NewTicker(time.Second * 3600)
	defer ticker.Stop()
	for {
		<-ticker.C
		if s.ShutDown {
			break
		}
		if time.Now().Hour() == 2 {
			//游戏数据统计
			go s.WriteGameStatisticsData()
		} else if time.Now().Hour() == 3 {
			xlog.Logger().Errorln("ticker run SaveRedisToDBData beg: ", time.Now())
			// redis数据同步写入数据库
			go s.SaveRedisToDBData()
		} else if time.Now().Hour() == 4 {
			// 数据统计
			xlog.Logger().Errorln("ticker run WriteStatisticsData beg: ", time.Now())
			go s.WriteStatisticsData()

			xlog.Logger().Errorln("ticker run WriteGlodUserStatisticsData beg: ", time.Now())
			go s.WriteGlodUserStatisticsData()
		}
	}
}

func (s *Server) Handle(string) error {
	return nil
}

func (s *Server) IsMulti() bool {
	return false
}

func (s *Server) ErrorDeep() int {
	return 1
}

func (s *Server) Name() string {
	return static.ServerNameBackend
}

func (s *Server) Etc() interface{} {
	return s.Con
}

// CallHall 调用大厅服
func (s *Server) CallHall(method string, msghead string, v interface{}) ([]byte, error) {
	args := &static.Msg_MsgBase{}
	buf, err := json.Marshal(v)
	if err != nil {
		xlog.Logger().Errorf("大厅调用失败：%v", err)
		return nil, err
	}
	args.Data = string(buf)
	args.Head = msghead
	buf, err = json.Marshal(args)
	if err != nil {
		xlog.Logger().Errorf("json 解析参数失败：%v", err)
		return nil, err
	}

	rep := CallHall(&static.Rpc_Args{MsgData: buf})
	return rep, nil
}

// CallLogin 调用登录服
func (s *Server) CallLogin(method string, msghead string, v interface{}) ([]byte, error) {
	args := &static.Msg_MsgBase{}
	buf, err := json.Marshal(v)
	if err != nil {
		xlog.Logger().Errorf("大厅调用失败：%v", err)
		return nil, err
	}
	args.Data = string(buf)
	args.Head = msghead
	buf, err = json.Marshal(args)
	if err != nil {
		xlog.Logger().Errorf("json 解析参数失败：%v", err)
		return nil, err
	}
	rep := CallLogin(&static.Rpc_Args{MsgData: buf})
	return rep, nil
}

func (s *Server) SaveDBDataToRedis() error {
	defer func() {
		s.DataSynchronism = false
	}()

	s.DataSynchronism = true

	begTime := time.Now()
	xlog.Logger().Infoln("wuhan run redis to db begin.")

	var err error

	/* 用户数据 */
	var users []*models.User
	if err = GetDBMgr().GetDBmControl().Find(&users).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	// redis
	for _, user := range users {
		err = GetDBMgr().GetDBrControl().AddPerson(user.ConvertModel())
		if err != nil {

			xlog.Logger().Errorln(err)
			return err

		}
	}

	// 包厢数据
	// mysql
	var houses []*models.House
	if err = GetDBMgr().GetDBmControl().Find(&houses).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	// redis
	err = GetDBMgr().GetDBrControl().HouseBaseReSave(houses)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	// 包厢楼层数据
	var floors []*models.HouseFloor
	if err = GetDBMgr().GetDBmControl().Find(&floors).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	// redis
	err = GetDBMgr().GetDBrControl().HouseFloorBaseReSave(floors)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	// 包厢成员数据
	var members []*models.HouseMember
	if err = GetDBMgr().GetDBmControl().Find(&members).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	// redis
	err = GetDBMgr().GetDBrControl().HouseMemberBaseReSave(members)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	//// 房卡统计数据
	//// 7日内数据 时间戳
	//timeStr := time.Now().Format("2006-01-02")
	//t, _ := time.Parse("2006-01-02", timeStr)
	//timeNumber := t.Unix() - (8*3600 + 7*24*3600)
	//
	//sqlstr := "select hid, count(*) playtime, SUM(kacost) kacost, date from " +
	//	"(SELECT hid, kacost, DATE_FORMAT(created_at, '%Y-%m-%d') date from  record_game_cost where created_at > ?) data GROUP BY hid, date ORDER BY date"
	//
	//var rows []model.RecordGameCostMini
	//err = GetDBMgr().GetDBmControl().Raw(sqlstr, timeNumber).Scan(&rows).Error
	//if err != nil {
	//	fmt.Println(err)
	//	return err
	//}
	//var datas []interface{}
	//for _, row := range rows {
	//	date, _ := time.ParseInLocation("2006-01-02", row.Date, time.Local)
	//	data := public.RecordGameCostMini{row.HId, row.PlayTime, row.KaCost, date.Unix()}
	//	datas = append(datas, &data)
	//}
	//err = GetDBMgr().GetDBrControl().HouseRecordCostReSave(datas)
	//if err != nil {
	//	syslog.Logger().Errorln(err)
	//	return err
	//}
	//
	//// 成员统计数据写入到Redis
	//var dayrecords []model.RecordGameDay
	//err = GetDBMgr().GetDBmControl().Where("DATE_SUB(CURDATE(), INTERVAL 7 DAY) <= created_at").Find(&dayrecords).Error
	//if err != nil {
	//	syslog.Logger().Errorln(err)
	//	return err
	//}
	//
	////redis
	//err = GetDBMgr().GetDBrControl().GameDayRecordReSave(&dayrecords)
	//if err != nil {
	//	syslog.Logger().Errorln(err)
	//	return err
	//}
	//
	//// 游戏回放数据写入到Reids
	//var gamereplays []model.RecordGameReplay
	//err = GetDBMgr().GetDBmControl().Where("DATE_SUB(CURDATE(), INTERVAL 2 DAY) <= created_at").Find(&gamereplays).Error
	//if err != nil {
	//	syslog.Logger().Errorln(err)
	//	return err
	//}
	//
	////redis
	//err = GetDBMgr().GetDBrControl().GameRecordReplayReSave(&gamereplays)
	//if err != nil {
	//	syslog.Logger().Errorln(err)
	//	return err
	//}
	//
	//// 战绩详情写入到Redis
	//var recordrounds []model.RecordGameRound
	//err = GetDBMgr().GetDBmControl().Where("DATE_SUB(CURDATE(), INTERVAL 2 DAY) <= created_at").Find(&recordrounds).Error
	//if err != nil {
	//	syslog.Logger().Errorln(err)
	//	return err
	//}
	////redis
	//err = GetDBMgr().GetDBrControl().GameRecordRoundReSave(&recordrounds)
	//if err != nil {
	//	syslog.Logger().Errorln(err)
	//	return err
	//}
	//
	//// 战绩数据写到Redis
	//var recordtotals []model.RecordGameTotal
	//err = GetDBMgr().GetDBmControl().Where("DATE_SUB(CURDATE(), INTERVAL 2 DAY) <= created_at").Find(&recordtotals).Error
	//if err != nil {
	//	syslog.Logger().Errorln(err)
	//	return err
	//}
	////redis
	//err = GetDBMgr().GetDBrControl().GameRecordTotalReSave(&recordtotals)
	//if err != nil {
	//	syslog.Logger().Errorln(err)
	//}

	// 任务数据
	var utasks []*models.UserTask
	if err = GetDBMgr().GetDBmControl().Find(&utasks).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	err = GetDBMgr().GetDBrControl().UserTaskReSave(utasks)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	// 包厢活动
	var hacts []*models.HouseActivity
	if err = GetDBMgr().GetDBmControl().Find(&hacts).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	err = GetDBMgr().GetDBrControl().HouseActivityReSave(hacts)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	err = GetDBMgr().GetDBrControl().HouseActivityReSave(hacts)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	// 包厢活动数据
	var hactrecs []*models.HouseActivityRecord
	if err = GetDBMgr().GetDBmControl().Find(&hactrecs).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	err = GetDBMgr().GetDBrControl().HouseActivityRecordReSave(hactrecs)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	err = GetDBMgr().GetDBrControl().HouseActivityRecordReSave(hactrecs)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	// 加盟商数据
	// 加盟商基础信息
	var leagueInfo []*models.League
	if err := GetDBMgr().GetDBmControl().Where("freeze=0").Find(&leagueInfo).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	err = GetDBMgr().GetDBrControl().LeagueBaseReSave(leagueInfo)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	//加盟商用户信息
	leagueUser := []*models.LeagueUser{}
	err = GetDBMgr().GetDBmControl().Where("freeze = 0 ").Find(&leagueUser).Group("league_id").Error
	if err != nil {
		return err
	}
	err = GetDBMgr().GetDBrControl().LeagueUserReSave(leagueUser)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	houseMixFloorTableInfo := []*models.HousemixfloorTable{}
	err = GetDBMgr().GetDBmControl().Find(&houseMixFloorTableInfo).Group("hid").Error
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	err = GetDBMgr().GetDBrControl().SaveHftDBToRedis(houseMixFloorTableInfo)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}

	fmt.Println("SaveDBDataToRedis Done costs:", time.Now().Unix()-begTime.Unix())
	return nil
}

// SaveRedisToDBData 从redis导入到db
func (s *Server) SaveRedisToDBData() error {
	defer func() {
		endTime := time.Now()
		s.DataSynchronism = false
		xlog.Logger().Warnf("done sync redis to mysql at %s", endTime.Format(static.TIMEFORMAT))
	}()

	s.DataSynchronism = true
	begTime := time.Now()
	xlog.Logger().Warnf("start sync redis to mysql at %s", begTime.Format(static.TIMEFORMAT))

	xlog.Logger().Infof("wuhan run redis to db begin.")

	// 用户数据
	persons, err := GetDBMgr().GetDBrControl().UserBaseReLoad()
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	for _, person := range persons {
		if person == nil || person.Uid <= 0 {
			xlog.Logger().Errorf("error person data:%+v", person)
			continue
		}
		updateMap := make(map[string]interface{})
		updateMap["describe_info"] = person.DescribeInfo
		updateMap["nickname"] = person.Nickname
		updateMap["sex"] = person.Sex
		updateMap["win_count"] = person.WinCount
		updateMap["lost_count"] = person.LostCount
		updateMap["draw_count"] = person.DrawCount
		updateMap["flee_count"] = person.FleeCount
		updateMap["total_count"] = person.TotalCount
		updateMap["idcard"] = person.Idcard
		updateMap["rename"] = person.ReName
		updateMap["games"] = person.Games
		updateMap["area"] = person.Area
		if person.LastOffLineTime > 0 {
			updateMap["last_offline_at"] = time.Unix(person.LastOffLineTime, 0)
		}
		err := GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", person.Uid).Updates(updateMap).Error
		if err != nil {
			xlog.Logger().Errorln(err)
			return err
		}

		var userLocation models.UserLocation
		userLocation.Id = person.Uid
		userLocation.Ip = person.Ip
		userLocation.Longitude = person.Longitude
		userLocation.Latitude = person.Latitude
		userLocation.Address = person.Address
		err = GetDBMgr().GetDBmControl().Save(&userLocation).Error
		if err != nil {
			xlog.Logger().Errorln(err)
			return err
		}
	}
	xlog.Logger().Errorln("person merge done.")

	// 包厢数据
	datas, err := GetDBMgr().GetDBrControl().HouseBaseReLoad()
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	for _, house := range datas {
		updateMap := make(map[string]interface{})
		updateMap["hid"] = house.HId
		updateMap["uid"] = house.UId
		updateMap["area"] = house.Area
		updateMap["name"] = house.Name
		updateMap["notify"] = house.Notify
		updateMap["table_sum"] = house.TableSum
		updateMap["is_checked"] = house.IsChecked
		updateMap["is_frozen"] = house.IsFrozen
		updateMap["is_member_hide"] = house.IsMemHide
		updateMap["is_hid_hide"] = house.IsHidHide
		updateMap["is_vitamin_hide"] = house.IsVitaminHide
		updateMap["is_vitamin_modi"] = house.IsVitaminModi
		updateMap["is_partner_hide"] = house.IsPartnerHide
		updateMap["is_partner_modi"] = house.IsPartnerModi
		updateMap["is_game_pause"] = house.IsGamePause
		updateMap["is_member_send"] = house.IsMemberSend
		updateMap["is_vitamin"] = house.IsVitamin
		updateMap["ai_check"] = house.AICheck
		updateMap["ai_total_score_limit"] = house.AITotalScoreLimit
		updateMap["dialog"] = house.Dialog
		updateMap["dialog_active"] = house.DialogActive
		updateMap["auto_pay_partner"] = house.AutoPayPartnrt
		updateMap["is_member_exit"] = house.IsMemExit
		updateMap["is_partner_apply"] = house.IsPartnerApply
		updateMap["onlyquick"] = house.OnlyQuickJoin
		updateMap["table_join_type"] = house.TableJoinType
		updateMap["mix_table_num"] = house.MixTableNum
		updateMap["mix_active"] = house.MixActive
		updateMap["merge_hid"] = house.MergeHId
		updateMap["table_show_count"] = house.TableShowCount
		updateMap["min_table_num"] = house.MinTableNum
		updateMap["max_table_num"] = house.MaxTableNum
		updateMap["empty_table_back"] = house.EmptyTableBack
		updateMap["table_sort_type"] = house.TableSortType
		updateMap["empty_table_max"] = house.EmptyTableMax
		updateMap["is_head_hide"] = house.IsHeadHide
		updateMap["is_mem_uid_hide"] = house.IsMemUidHide
		updateMap["dis_vitamin_junior"] = house.DisVitaminJunior
		updateMap["partner_kick"] = house.PartnerKick
		updateMap["reward_balanced"] = house.RewardBalanced
		updateMap["reward_balanced_type"] = house.RewardBalancedType
		updateMap["fangka_tips_min_num"] = house.FangKaTipsMinNum
		updateMap["record_time_interval"] = house.RecordTimeInterval
		updateMap["new_table_sort_type"] = house.NewTableSortType
		updateMap["create_table_type"] = house.CreateTableType
		updateMap["RankRound"] = house.RankRound
		updateMap["RankWiner"] = house.RankWiner
		updateMap["RankRecord"] = house.RankRecord
		updateMap["RankOpen"] = house.RankOpen
		updateMap["is_not_effect_2P"] = house.IsNotEft2PTale
		updateMap["no_skip_vitamin_set"] = house.NoSkipVitaminSet
		err := GetDBMgr().GetDBmControl().Model(models.House{}).Where("id = ?", house.Id).Updates(updateMap).Error
		if err != nil {
			xlog.Logger().Errorln(err)
			return err
		}
	}
	xlog.Logger().Errorln("house merge done.")

	// 包厢楼层数据
	floors, err := GetDBMgr().GetDBrControl().HouseFloorBaseReLoad()
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	for _, floor := range floors {
		updateMap := make(map[string]interface{})
		updateMap["hid"] = floor.DHId
		updateMap["rule"] = floor.Rule
		updateMap["ai_super_sum"] = floor.AiSuperNum
		updateMap["name"] = floor.Name
		updateMap["is_mix"] = floor.IsMix
		updateMap["ai_super_sum"] = floor.AiSuperNum
		updateMap["is_vitamin"] = floor.IsVitamin
		updateMap["is_game_pause"] = floor.IsGamePause
		updateMap["vitamin_lowerlimit"] = floor.VitaminLowLimit
		updateMap["vitamin_highlimit"] = floor.VitaminHighLimit
		updateMap["vitamin_lowerlimit_pause"] = floor.VitaminLowLimitPause

		err := GetDBMgr().GetDBmControl().Model(models.HouseFloor{}).Where("id = ?", floor.Id).Updates(updateMap).Error
		if err != nil {
			xlog.Logger().Errorln(err)
			return err
		}
	}
	xlog.Logger().Errorln("housefloor merge done.")

	// 包厢成员数据
	mems, err := GetDBMgr().GetDBrControl().HouseMemberBaseReLoad()
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	for _, data := range mems {
		member := data.(*static.HouseMember)
		updateMap := make(map[string]interface{})

		updateMap["hid"] = member.DHId
		updateMap["fid"] = member.FId
		updateMap["ref"] = member.Ref
		updateMap["uid"] = member.UId
		updateMap["urole"] = member.URole
		updateMap["uremark"] = member.URemark
		updateMap["p_remark"] = member.PRemark

		if member.ApplyTime > 0 {
			updateMap["apply_time"] = time.Unix(member.ApplyTime, 0)
		}
		if member.AgreeTime > 0 {
			updateMap["agree_time"] = time.Unix(member.AgreeTime, 0)
		}

		updateMap["bw_times"] = member.BwTimes
		updateMap["play_times"] = member.PlayTimes
		updateMap["forbid"] = member.Forbid
		updateMap["partner"] = member.Partner
		updateMap["superior"] = member.Superior
		updateMap["agent"] = member.Agent
		updateMap["uvitamin"] = member.UVitamin
		updateMap["vitamin_admin"] = member.VitaminAdmin
		updateMap["vice_partner"] = member.VicePartner
		updateMap["no_floors"] = member.NoFloors
		updateMap["ref"] = member.Ref

		err := GetDBMgr().GetDBmControl().Model(models.HouseMember{}).Where("id = ?", member.Id).Updates(updateMap).Error
		if err != nil {
			xlog.Logger().Errorln(err)
			continue
		}
	}
	xlog.Logger().Errorln("housemember merge done.")

	// 包厢活动
	hadatas, err := GetDBMgr().GetDBrControl().HouseActivityReLoad()
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	for _, data := range hadatas {
		hact := data.(*static.HouseActivity)

		updateMap := make(map[string]interface{})
		updateMap["hid"] = hact.DHId
		updateMap["fid"] = hact.FId
		updateMap["kind"] = hact.Kind
		updateMap["name"] = hact.Name
		updateMap["status"] = hact.Status
		updateMap["begtime"] = time.Unix(hact.BegTime, 0)
		updateMap["endtime"] = time.Unix(hact.EndTime, 0)
		err := GetDBMgr().GetDBmControl().Model(models.HouseActivity{}).Where("id = ?", hact.Id).Updates(updateMap).Error
		if err != nil {
			xlog.Logger().Errorln(err)
			return err
		}
	}
	xlog.Logger().Info("houseactivity merge done.")

	// 包厢活动数据
	harecdatas, err := GetDBMgr().GetDBrControl().HouseActivityRecordReLoad()
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	for _, data := range harecdatas {
		hactrec := data.(*static.ActRecordItem)
		updateMap := make(map[string]interface{})
		updateMap["actid"] = hactrec.ActId
		updateMap["uid"] = hactrec.UId
		updateMap["ranksore"] = hactrec.Score
		if err = GetDBMgr().GetDBmControl().Model(models.HouseActivityRecord{}).Where("actid = ? and uid = ?", hactrec.ActId, hactrec.UId).Updates(updateMap).Error; err != nil {
			xlog.Logger().Errorln(err)
			return err
		}
	}
	xlog.Logger().Info("houseactivitydatas merge done.")

	fmt.Println("SaveRedisToDBData Done costs:", time.Now().Unix()-begTime.Unix())
	return nil
}

//func (self *Server) RunTimeSeconds() {
//	syslog.Logger().Infoln("run,time:", time.Now().Hour())
//	ticker := time.NewTicker(time.Second)
//	for {
//		<-ticker.C
//		if self.ShutDown {
//			break
//		}
//
//		if time.Now().Hour() == 0 && time.Now().Minute() == 0 && time.Now().Second() == 0 {
//			syslog.Logger().Errorln("ticker run HouseMemVitaminLeftStatistic beg: ", time.Now())
//			self.HouseMemVitaminLeftStatistic()
//			syslog.Logger().Errorln("ticker run HouseMemVitaminLeftStatistic end: ", time.Now())
//		}
//
//		time.Sleep(2 * time.Second)
//	}
//
//	ticker.Stop()
//}

// 每2分钟获取在线人数
func (s *Server) saveGameOnline() {
	now := int(time.Now().Unix())
	time.Sleep(time.Duration((120 - now%120) * 1000000000))
	ticker := time.NewTicker(time.Minute * 2)
	for {
		if s.ShutDown {
			break
		}

		<-ticker.C

		value, err := GetServer().CallHall("NewServerMsg", consts.MsgTypeGetOnlineNumber, nil)
		if err != nil {
			xlog.Logger().Errorln("getHouseVitaminPoolSum online number from hall failed:", err.Error())
			continue
		}

		numMap := make(map[int]int)
		err = json.Unmarshal(value, &numMap)
		if err != nil {
			xlog.Logger().Errorln("get online number from hall failed:", err.Error())
			continue
		}

		// 写入数据库
		tx := GetDBMgr().db_M.Begin()
		for k, v := range numMap {
			var obj models.GameOnline
			obj.KindId = k
			obj.Value = v
			if err = tx.Create(&obj).Error; err != nil {
				xlog.Logger().Errorln("save online number failed:", err.Error())
				break
			}
		}
		tx.Commit()
	}
}

// 写入统计数据
func (s *Server) WriteStatisticsData() {
	taskStartTime := time.Now()
	xlog.Logger().Errorln("统计数据计算开始")

	statisticsData := time.Now().AddDate(0, 0, -1)
	selectDayStr := fmt.Sprintf("%d-%02d-%02d", statisticsData.Year(), statisticsData.Month(), statisticsData.Day())

	db := GetDBMgr().GetDBmControl()
	var err error

	/* 用户统计 */
	userData := new(models.StatisticsUser)
	userData.Date = statisticsData

	// 查询房卡场时间节点总人数
	userData.TotalCount, err = models.FangKaUserGetCountByTime(db, statisticsData)
	if err != nil {
		xlog.Logger().Errorf("[房卡场统计信息][查询玩家总数失败]:%v", err)
		return
	}

	// 查询房卡场当日新增
	userData.NewCount, err = models.FangKaNewUserGetCountByTime(db, statisticsData)
	if err != nil {
		xlog.Logger().Errorf("[房卡场统计信息][查询玩家总数失败]:%v", err)
		return
	}

	userData.ActiveCount, err = models.GetFangKaUserActiveCount(db, statisticsData)
	if err != nil {
		xlog.Logger().Errorf("[房卡场统计信息][查询活跃总数失败]:%v", err)
		return
	}

	// 金币场、房卡场新增人数
	userData.GoldNewUser, err = models.GoldNewUserNewStataGetCountByTime(db, statisticsData)
	if err != nil {
		xlog.Logger().Errorf("[查询新增金币用户失败]:%v", err)
		return
	} else {
		userData.FangKaNewUser = userData.NewCount - userData.GoldNewUser
	}

	// 查询安卓、ios 新增人数
	userData.AndroidNewUser, userData.IOSNewUser, err = models.PlatformNewUserGetCountByTime(db, statisticsData)
	if err != nil {
		xlog.Logger().Errorf("[查询anroid、ios新增用户失败]:%v", err)
		return
	}

	if err = GetDBMgr().db_M.Create(userData).Error; err != nil {
		xlog.Logger().Errorln(err)
		return
	}

	// 更新七日留存统计
	for i := 0; i <= 7; i++ {
		updateDay := statisticsData.AddDate(0, 0, -i)
		UserRetention, UserRetention1, UserRetention3, UserRetention7 := models.GetFangKaUserRetentionStatistics(db, updateDay)
		GameRetention, GameRetention1, GameRetention3, GameRetention7 := models.GetFangKaGameRetentionStatistics(db, updateDay)

		err := models.FangKaUserStatisticsUpdateRetention(db, updateDay, UserRetention, UserRetention1, UserRetention3, UserRetention7,
			GameRetention, GameRetention1, GameRetention3, GameRetention7)
		if err != nil {
			xlog.Logger().Errorf("[房卡场统计信息][更新玩家留存失败]:%v", err)
			continue
		}
	}

	type queryResult struct {
		Count int `json:"count"`
	}

	tx := GetDBMgr().db_M.Begin()
	/* 财富统计 */
	getArr := []int8{models.CostTypeGm, models.CostTypeCheckin, models.CostTypeAllowances, models.CostTypeRegister} // 新增
	costArr := []int8{models.CostTypeRevenue, models.CostTypeSticker}                                               //消耗
	for _, item := range getArr {
		res := new(queryResult)
		if err = GetDBMgr().db_M.Model(models.UserWealthCost{}).Select("sum(cost) as count").Where(`date_format(created_at, "%Y-%m-%d") = ? and wealth_type = ? and cost_type = ?`, selectDayStr, consts.WealthTypeGold, item).Scan(&res).Error; err != nil {
			tx.Rollback()
			xlog.Logger().Errorln(err)
			return
		} else {
			wealthData := new(models.StatisticsWealth)
			wealthData.Date = statisticsData
			wealthData.WealthType = consts.WealthTypeGold
			wealthData.CostType = item
			wealthData.Num = res.Count
			if err = tx.Create(&wealthData).Error; err != nil {
				tx.Rollback()
				xlog.Logger().Errorln(err)
				return
			}
		}
	}
	for _, item := range costArr {
		res := new(queryResult)
		if err = GetDBMgr().db_M.Model(models.UserWealthCost{}).Select("sum(cost) as count").Where(`date_format(created_at, "%Y-%m-%d") = ? and wealth_type = ? and cost_type = ?`, selectDayStr, consts.WealthTypeGold, item).Scan(&res).Error; err != nil {
			tx.Rollback()
			xlog.Logger().Errorln(err)
			return
		} else {
			wealthData := new(models.StatisticsWealth)
			wealthData.Date = statisticsData
			wealthData.WealthType = consts.WealthTypeGold
			wealthData.CostType = item
			wealthData.Num = res.Count * -1
			if err = tx.Create(&wealthData).Error; err != nil {
				tx.Rollback()
				xlog.Logger().Errorln(err)
				return
			}
		}
	}
	tx.Commit()

	xlog.Logger().Errorln(fmt.Sprintf("统计数据计算结束, 耗时：%v", time.Now().Sub(taskStartTime)))
}

// 写入统计数据
func (s *Server) WriteGlodUserStatisticsData() {
	/* 用户统计 */
	statisticsData := time.Now().AddDate(0, 0, -1)
	userData := new(models.GlodUserStatistics)
	userData.Date = statisticsData

	db := GetDBMgr().GetDBmControl()
	var err error
	userData.TotalCount, err = models.GoldUserGetCountByTime(db, statisticsData)
	if err != nil {
		xlog.Logger().Errorf("[统计信息][查询玩家总数失败]:%v", err)
		return
	}

	userData.NewUserCount, err = models.GoldNewUserGetCountByTime(db, statisticsData)
	if err != nil {
		xlog.Logger().Errorf("[统计信息][查询新增玩家总数失败]:%v", err)
		return
	}

	userData.DayGameCount, err = models.GoldGameUserCountSelectByTime(db, statisticsData)
	if err != nil {
		xlog.Logger().Errorf("[统计信息][查询当日游戏玩家总数失败]:%v", err)
		return
	}
	userData.NewUserGameCount, err = models.GoldGameNewUserCountSelectByTime(db, statisticsData)
	if err != nil {
		xlog.Logger().Errorf("[统计信息][查询新增用户当日游戏玩家总数失败]:%v", err)
		return
	}

	err = db.Create(&userData).Error
	if err != nil {
		xlog.Logger().Errorf("[统计信息][保存统计信息失败]:%v", err)
		return
	}

	// 更新七日留存统计
	for i := 0; i <= 7; i++ {
		updateDay := statisticsData.AddDate(0, 0, -i)
		UserRetention, UserRetention1, UserRetention3, UserRetention7 := models.GetUserRetentionStatistics(db, updateDay)
		GameRetention, GameRetention1, GameRetention3, GameRetention7 := models.GetGameRetentionStatistics(db, updateDay)

		err := models.GlodUserStatisticsUpdateRetention(db, updateDay, UserRetention, UserRetention1, UserRetention3, UserRetention7,
			GameRetention, GameRetention1, GameRetention3, GameRetention7)
		if err != nil {
			xlog.Logger().Errorf("[统计信息][更新玩家留存失败]:%v", err)
			continue
		}
	}
}

func (s *Server) HouseVitaminPoolSum() {
	_, err := GetServer().CallHall("NewServerMsg", consts.MsgHouseVitaminPoolSum, nil)
	if err != nil {
		xlog.Logger().Errorln(" HouseVitaminPoolSum failed:", err.Error())
	}
}

func (s *Server) HouseMemVitaminLogSum() {
	_, err := GetServer().CallHall("NewServerMsg", consts.MsgHouseMemVitaminSum, nil)
	if err != nil {
		xlog.Logger().Errorln(" HouseMemVitaminLogSum failed:", err.Error())
	}
}

func (s *Server) HouseMemVitaminLeftStatistic() {
	_, err := GetServer().CallHall("NewServerMsg", consts.MsgHouseMemLeftStatistic, nil)
	if err != nil {
		xlog.Logger().Errorln(" HouseMemVitaminLeftStatistic failed:", err.Error())
	}
}

func (s *Server) HouseAutoPayPartner() {
	//_, err := GetServer().CallHall("NewServerMsg", constant.MsgHouseAutoPayPartnerMsg, nil)
	//if err != nil {
	//	syslog.Logger().Errorln(" HouseMemVitaminLogSum failed:", err.Error())
	//}
}

func (s *Server) WriteGameStatisticsData() {
	taskStartTime := time.Now()
	xlog.Logger().Info("统计数据计算开始")

	var err error
	var config_rows []models.ConfigGameTypes
	if err = GetDBMgr().GetDBmControl().Model(models.ConfigGameTypes{}).Find(&config_rows).Error; err != nil {
		xlog.Logger().Error(err)
		return
	}

	type queryResult struct {
		Count int `json:"count"`
	}

	tx := GetDBMgr().db_M.Begin()
	for _, config := range config_rows {
		var gamedata models.StatisticsGame

		gamedata.KindId = config.KindId
		gamedata.SiteType = config.SiteType
		gamedata.Date, _ = time.ParseInLocation("2006-01-02", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), time.Local)

		dayActiveUserCount := 0 //日活跃玩家(有对局的玩家)
		dayActiveQuery := new(queryResult)
		if err = GetDBMgr().db_M.Model(models.GameResultDetail{}).Select("Count(distinct(uid)) as count").Where("kind_id = ? AND site_type = ? AND DATE_FORMAT(created_at,'%Y-%m-%d') = DATE_SUB(CURDATE(), INTERVAL 1 DAY) AND uid NOT IN(SELECT mid FROM robot)", config.KindId, config.SiteType).Scan(&dayActiveQuery).Error; err != nil {
			xlog.Logger().Error(err)
			return
		}
		dayActiveUserCount = dayActiveQuery.Count
		//每日玩牌人数
		gamedata.DayPlayUser = dayActiveUserCount

		//玩家平均玩牌局数
		playCount := 0
		if err = GetDBMgr().db_M.Model(models.GameResultDetail{}).Where("kind_id = ? AND site_type = ? AND DATE_FORMAT(created_at,'%Y-%m-%d') = DATE_SUB(CURDATE(), INTERVAL 1 DAY)  AND uid NOT IN(SELECT mid FROM robot)", config.KindId, config.SiteType).Count(&playCount).Error; err != nil {
			xlog.Logger().Error(err)
			return
		}
		if dayActiveUserCount > 0 {
			gamedata.AvgPlayCount = playCount / dayActiveUserCount
		}

		dayActiveRobotCount := 0 //日活跃机器人(有对局的玩家)
		dayActiveRobotQuery := new(queryResult)
		if err = GetDBMgr().db_M.Model(models.GameResultDetail{}).Select("Count(distinct(uid)) as count").Where("kind_id = ? AND site_type = ? AND DATE_FORMAT(created_at,'%Y-%m-%d') = DATE_SUB(CURDATE(), INTERVAL 1 DAY) AND uid IN(SELECT mid FROM robot)", config.KindId, config.SiteType).Scan(&dayActiveRobotQuery).Error; err != nil {
			xlog.Logger().Error(err)
			return
		}
		dayActiveRobotCount = dayActiveRobotQuery.Count

		//机器人平均玩牌局数
		robotplaycount := 0
		if err = GetDBMgr().db_M.Model(models.GameResultDetail{}).Where("kind_id = ? AND site_type = ? AND DATE_FORMAT(created_at,'%Y-%m-%d') = DATE_SUB(CURDATE(), INTERVAL 1 DAY)  AND uid IN(SELECT mid FROM robot)", config.KindId, config.SiteType).Count(&robotplaycount).Error; err != nil {
			xlog.Logger().Error(err)
			return
		}
		if dayActiveRobotCount > 0 {
			gamedata.RobotAvgPlayCount = robotplaycount / dayActiveRobotCount
		}

		//玩家平均玩牌时长
		playtime := 0  //玩牌时长(单位秒)
		WinScore := 0  //赢分
		LoseScore := 0 //输分
		winCount := 0
		loseCount := 0
		var result []models.GameResultDetail
		if err = GetDBMgr().GetDBmControl().Model(models.GameResultDetail{}).Where("kind_id = ? AND site_type = ? AND DATE_FORMAT(created_at,'%Y-%m-%d') = DATE_SUB(CURDATE(), INTERVAL 1 DAY) AND uid NOT IN(SELECT mid FROM robot)", config.KindId, config.SiteType).Find(&result).Error; err != nil {
			xlog.Logger().Error(err)
			return
		}
		for _, row := range result {
			playtime += int(row.EndTime.Sub(row.BeginTime).Seconds())
			if row.Score > 0 {
				WinScore += int(row.AfterScore - row.BeforeScore)
				winCount++
			} else {
				LoseScore += int(row.AfterScore - row.BeforeScore)
				loseCount++
			}
		}
		if dayActiveUserCount > 0 {
			gamedata.AvgPlayTime = playtime / dayActiveUserCount
		}
		//平均单局结算赢分
		if winCount > 0 {
			gamedata.AvgWinScore = int(WinScore / winCount)
		}
		//平均单局结算输分
		if loseCount > 0 {
			gamedata.AvgLoseScore = int(LoseScore / loseCount)
		}

		//dayActiveRobotCount := 0 //日活跃机器人(有对局的机器人)
		//dayActiveRobotQuery := new(queryResult)
		//if err = GetDBMgr().db_M.Model(model.GameResultDetail{}).Select("Count(distinct(uid)) as count").Where("kind_id = ? AND site_type = ? AND DATE_FORMAT(created_at,'%Y-%m-%d') = DATE_SUB(CURDATE(), INTERVAL 1 DAY)  AND uid IN(SELECT mid FROM robot)", config.KindId, config.SiteType).Scan(&dayActiveRobotQuery).Error; err != nil {
		//	syslog.Logger().Error(err)
		//	return
		//}
		//dayActiveRobotCount = dayActiveRobotQuery.Count

		robotWinScore := 0  //赢分
		robotLoseScore := 0 //输分
		robotwinCount := 0
		robotloseCount := 0
		RobotAvgWinScore := 0
		RobotAvgLoseScore := 0
		var rows []models.GameResultDetail
		if err = GetDBMgr().GetDBmControl().Model(models.GameResultDetail{}).Where("kind_id = ? AND site_type = ? AND DATE_FORMAT(created_at,'%Y-%m-%d') = DATE_SUB(CURDATE(), INTERVAL 1 DAY) AND uid IN(SELECT mid FROM robot)", config.KindId, config.SiteType).Find(&rows).Error; err != nil {
			xlog.Logger().Error(err)
			return
		}
		for _, row := range rows {
			if row.Score > 0 {
				robotWinScore += int(row.AfterScore - row.BeforeScore)
				robotwinCount++
			} else {
				robotLoseScore += int(row.AfterScore - row.BeforeScore)
				robotloseCount++
			}
		}
		//平均单局结算赢分
		if robotwinCount > 0 {
			RobotAvgWinScore = int(robotWinScore / robotwinCount)
		}
		//平均单局结算输分
		if robotloseCount > 0 {
			RobotAvgLoseScore = int(robotLoseScore / robotloseCount)
		}
		gamedata.RobotAvgScore = RobotAvgWinScore + RobotAvgLoseScore

		//每日机器人具体输赢金币总数
		var robotScore []float64
		sumCostSQL := "SELECT ifnull(SUM(SumCost),0) FROM ((select  ifnull(after_score - before_score,0) as SumCost  from game_result_detail WHERE " + fmt.Sprintf("kind_id = %d AND site_type = %d", config.KindId, config.SiteType) + " AND DATE_FORMAT(game_result_detail.created_at,'%Y-%m-%d') = DATE_SUB(CURDATE(), INTERVAL 1 DAY) AND uid IN(SELECT mid FROM robot)) AS TmpTable)"
		if err = GetDBMgr().GetDBmControl().Model(models.GameResultDetail{}).Raw(sumCostSQL).Pluck("ifnull(SUM(SumCost),0)", &robotScore).Error; err != nil {
			xlog.Logger().Error(err)
			return
		} else {
			gamedata.RobotDayCost = int(robotScore[0])
		}

		if err = tx.Create(&gamedata).Error; err != nil {
			tx.Rollback()
			xlog.Logger().Error(err)
			return
		}

	}

	tx.Commit()
	xlog.Logger().Info(fmt.Sprintf("统计数据计算结束, 耗时：%v", time.Now().Sub(taskStartTime)))
}
