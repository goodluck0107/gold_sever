package consts

const (
	MsgCloseConnection      = "closesession" // 关闭连接
	MsgForceCloseConnection = "forceclose"   // 通知客户端关闭 不重连

	// 客户端http服务消息
	MsgTypeCheckYK        = "checkYK"        // 检测是否是游客账号
	MsgTypeLoginYK        = "loginYK"        // 游客登录
	MsgTypeLoginMobile    = "loginMobile"    // 手机登录(手机号+验证码)
	MsgTypeLoginMobileV2  = "loginMobile_v2" // 手机登录v2(手机号+密码)
	MsgTypeMobileRegister = "mobileRegister" // 手机注册
	MsgTypeResetPassword  = "resetpassword"  // 重置密码
	MsgTypeLoginWechat    = "loginWechat"    // 微信登录
	MsgTypeLoginWechatV2  = "loginWechat_v2" // 小程序微信登录
	MsgTypeLoginApple     = "loginApple"     // 苹果登录
	MsgTypeLoginHW        = "loginHW"        // 华为登录
	MsgTypeLoginToken     = "loginToken"     // 令牌直接登录
	MsgTypeSmscode        = "smscode"        // 短信验证码
	MsgTypeCheckUpdate    = "checkUpdate"    // 检查更新
	MsgTypeGetReplay      = "getreplay"      // 获取回放记录

	// 游戏后台服务消息
	MsgTypeDeliverProduct         = "deliverProduct"         // 下发产品
	MsgTypeUpdWealth              = "updwealth"              // 财富更新
	MsgTypeNoticeUpdate           = "noticeupdate"           // 公告更新
	MsgTypeExplainUpdate          = "explainupdate"          // 游戏说明更新
	MsgTypeAreaUpdate             = "areaupdate"             // 区域游戏更新
	MsgTypeReloadConfig           = "reloadconfig"           //! 一条http路由让所有服务器重读配置文件
	MsgTypeGetValidKindid         = "getvalidid"             //! php后台得到服务器有效的kindid接口
	MsgTypeHouseConUpdate         = "updatehousecon"         //! 应要求，得到一个字段，对应修改config_house表，成功后通知大厅重读
	MsgTypeClearUserInfo          = "clearuserinfo"          //! 应要求，得到一个字段，在user表和redis同时重置这个字段，如手机号，身份证等
	MsgTypeCompulsoryDiss         = "tablecompuldissmiss"    //! 应要求，得到一个玩家id，如果他处的这个房间开始了，就强制解散他，如果没开始就踢出去
	MsgTypeSaveToRedis            = "savetoredis"            //! 数据库数据写入redis
	MsgTypeSaveToDatabase         = "savetodb"               //! redis数据写入数据库
	MsgTypeSetBlack               = "setblack"               //! 设置用户黑名单
	MsgTypeGetOnlineNumber        = "getonline"              //! 获取当前在线人数
	MsgTypeGetUserInfo            = "getuserinfo"            //! 获取用户详情
	MsgTypeDialogNotice           = "dialognotice"           // 弹窗公告
	MsgTypeDialogNoticeList       = "dialognoticelist"       // 公告列表
	MsgTypeMarqueenNotice         = "marqueenotice"          // 跑马灯公告
	MsgTypeMaintainNotice         = "maintainnotice"         // 维护公告
	MsgTypeExecStatisticsTask     = "execstatisticstask"     // 执行统计任务
	MsgTypeExecGameStatisticsTask = "execgamestatisticstask" // 执行游戏统计任务
	MsgTypeHosueIDChange          = "houseidchange"          // 修改包厢id
	MsgTypeHouseTableInviteRecv   = "htinvite_ntf"           // 收到包厢牌桌邀请
	MsgTypeHouseTableInviteSend   = "htinvite_send"          // 发送包厢牌桌邀请
	MsgTypeHouseTableInviteResp   = "htinvite_ack"           // 受邀人应答
	MsgTypeHouseJoinInviteRecv    = "housejoininvitentf"     // 收到包厢牌桌邀请
	MsgTypeHouseJoinInviteSend    = "housejoininvitesend"    // 发送包厢牌桌邀请
	MsgTypeHouseJoinInviteResp    = "housejoininviteack"     // 受邀人应答
	MsgTypeDeliveryInfoUpd        = "deliveryinfoupdate"     // 个人名片更新
	MsgTypeUpdateGameServer       = "updategameserver"       // 更细游戏服务器
	MsgTypeOnFewerStart           = "onfewerstart"           // 少人开局
	MstTypeWatcherList            = "tablewatcherlist"       // 观战列表
	MstTypeWatcherQuit            = "tablewatcherquit"       // 退出观战
	MstTypeWatcherSwitch          = "tablewatcherswitch"     // 切换观战的位置

	// 大厅消息
	MsgTypeHallLogin               = "setuid"                  // 登录
	MsgTypeHallCertification       = "certification"           // 实名认证
	MsgTypeHallBindWechat          = "bindwechat"              // 绑定微信
	MsgTypeHallAuthWechat          = "authwechat"              // 授权微信用户信息
	MsgTypeHallBindMobile          = "bindmobile"              // 绑定手机
	MsgTypeHallBindMobileV2        = "bindmobilev2"            // 绑定手机(金币场)
	MsgTypeHallUnbindMobile        = "unbindmobile"            // 解绑手机
	MsgTypeHallSaveGames           = "savegames"               // 保存用户游戏列表
	MsgTypeHallGameRecord          = "gamerecord"              // 大厅历史战绩
	MsgTypeGameRecordInfo          = "gamerecordinfo"          // 战绩详情
	MsgTypeRelogin                 = "relogin"                 // 重复登录
	MsgTypeUserOnline              = "useronline"              // 用户上线
	MsgTypeUserOffline             = "useronoffline"           // 用户离线
	MsgTypeCheckReplayId           = "checkreplayid"           // 检查回放码
	MsgTypeGetAreas                = "getareas"                // 获取区域列表
	MsgTypeAreaIn                  = "areain"                  // 加入区域
	MsgTypeSaveInsureGold          = "saveinsuregold"          // 保存金币至保险箱
	MsgTypeUpdateUserInfo          = "updateuserinfo"          // 更新用户信息
	MsgTypeUpdateDescribeInfo      = "updatedescribeinfo"      // 更新用户签名
	MsgTypeUserBroadcast           = "userbroadcast"           // 用户广播消息
	MsgTypeGameList_Ntf            = "gamelist_ntf"            // 游戏列表
	MsgTypeGameCollections_Ntf     = "gameCollections_ntf"     // 游戏合集
	MsgTypeGameOnline              = "gameonline_ntf"          // 在线人数
	MsgTypeDisposeAllowances       = "disposeallowances"       // 发放低保
	MsgTypeGetAllowancesDouble     = "getallowancesdouble"     // 领取双倍低保
	MsgTypeFakeTableIn             = "faketablein"             // 进入假桌子
	MsgTypeUserAreaBroadcast       = "userareabroadcast"       // 用户区域广播
	MsgTypeWelcomeGift             = "welcomegift"             // 老用户见面礼
	MsgTypeGetWelcomeGift          = "getwelcomegift"          // 获取见面礼
	MsgTypeCheckin                 = "checkin"                 // 每日签到
	MsgTypeTaskCheckin             = "task_checkin"            // 每日签到任务详情
	MsgTypeGetGoldRecord           = "getgoldrecord"           // 获取礼卷奖励列表
	MsgTypePhoneBind               = "shopphonebind"           // 绑定兑换手机号
	MsgTypeShopExchange            = "shopexchange"            // 商品兑换
	MsgTypeGetPhoneBind            = "getphonebind"            // 获取绑定手机
	MsgTypeGetShopLists            = "getshoplist"             // 获取兑换商品
	MsgTypeGetShopRecord           = "getshoprecord"           // 获取兑换记录
	MsgTypeAreaEnter               = "areaenter"               // chess新版大厅进入区域消息
	MsgTypeAreaGameCardListMain    = "areagamecardmain"        // 获取区域内房卡游戏包
	MsgTypeAreaGameGoldListMain    = "areagamegoldmain"        // 获取区域内金币游戏包
	MsgTypeAreaGameSeek            = "areagameseek"            // chess新版大厅游戏搜索
	MsgTypeAreaPkgGames            = "areapkggames"            // 通过包名得到包下面的子游戏
	MsgTypeAreaGameRules           = "areagamerules"           // 通过kindids得到对应的gamerules
	MsgTypeAreaGameExplain         = "areagameexplain"         // 通过kindid得到对应的玩法说明
	MsgTypeAreaPkgByTId            = "areapkgbytid"            // 通过桌子号得到桌子所在的区域包游戏信息
	MsgTypeAreaPkgByKId            = "areapkgbykid"            // 通过桌子号得到桌子所在的区域包游戏信息
	MsgTypeAreaWX                  = "areacswx"                // 区域客服微信
	MsgTypeHouseDialogEdit         = "housedialogedit"         // 包厢弹窗编辑
	MsgTypeHousePartnerGetCode     = "housepartnergetcode"     // 包厢队长得到邀请码
	MsgTypeHouseJoinByPCode        = "housejoinbypcode"        // 包厢队长得到邀请码
	MsgTypeHouseSetTblShowCount    = "housesettblshowcount"    // 包厢队长得到邀请码
	MsgTypeHouseFloorWaitNumSet    = "housefloorwaitnumset"    // 包厢楼层修改等待人数
	MsgTypeHouseFloorWaitNumGet    = "housefloorwaitnumget"    // 包厢楼层修改等待人数
	MsgTypeHouseVitaminAdminSet    = "housevitadminset"        // 包厢比赛分管理员设置
	MsgTypeHouseVicePartnerSet     = "housevicepartnerset"     // 包厢比赛分管理员设置
	MsgTypeHouseVitAdminSet_Ntf    = "housevitadminset_ntf"    // 包厢比赛分管理员设置
	MsgTypeHouseVicePartnerSet_Ntf = "housevicepartnerset_ntf" // 包厢比赛分管理员设置

	//用户分组
	MsgTypeHouseGroupAdd         = "housememgroupadd"
	MsgTypeHouseGroupDel         = "housememgroupdel"
	MsgTypeHouseGroupInfo        = "housememgroupinfo"
	MsgTypeHouseGroupUserList    = "housememgroupuserlist"
	MsgTypeHouseGroupUserAddList = "housememgroupaddlist"
	MsgTypeGroupUserAdd          = "housememgroupuseradd"
	MsgTypeGroupUserRemove       = "housememgroupuserdel"

	// 牌桌消息
	MsgTypeTableCreate     = "tablecreate"     // 创建牌桌
	MsgTypeTableCreate_Req = "tablecreate_req" // 创建牌桌
	MsgTypeTableCreate_Ack = "tablecreate_ack" // 创建牌桌
	MsgTypeTableExit_Ntf   = "tableexit_ntf"   // 离开牌桌
	MsgTypeTableDel_Ntf    = "tabledel_ntf"    // 解散牌桌
	MsgTypeTableRes_Ntf    = "tableres_ntf"    // 牌桌大结算信息通知

	MsgTypeTableIn             = "tablein"             // 加入牌桌
	MsgTypeTableIn_Ack         = "tablein_ack"         // 加入牌桌
	MsgTypeTableIn_Ntf         = "htablein_ntf"        // 加入牌桌
	MsgTypeTableExit           = "tableexit"           // 离开牌桌
	MsgTypeTableDel            = "tabledel"            // 解散牌桌
	MsgTypeTableTimeOutDel     = "tabletimeoutdel"     // 超时解散牌桌
	MsgTypeTableCheckConn      = "tablecheckconn"      // 检测用户连接状态
	MsgTypeTableInfo           = "tableinfo"           // 牌桌详情
	MsgTypeHouseTableInfoByUid = "housetableinfobyuid" // 根据uid查询牌桌详情
	MsgTypeTableNumChange      = "tablenumchange"      // 桌子数量变化
	MsgTtypeTableSetBegin      = "setbegin"            // 通知牌桌局数变化
	MsgTtypeTableSetBegin_NTF  = "housetablestep_ntf"  // 通知牌桌局数变化
	MsgTypeLeagueCardAdd       = "leaguecardadd"       // 加盟商卡池变化
	MsgTypeFloorTableInfo      = "housefloorinfo"      // 获取当前楼层牌桌信息
	MsgTypeTableInLastGame     = "tableinlast"         // 加入牌桌时返回最后一局小结算 功能优化，拆分原来的大消息为几个小消息

	MsgTypeTableWatchIn_Ntf  = "htablewatchin_ntf"  // 加入牌桌观战
	MsgTypeTableWatchOut_Ntf = "htablewatchout_ntf" // 加入牌桌观战

	// 包厢动作消息
	HFChOptHallTableIn     = "chopt_halltablein"     //! 玩家进入牌桌
	HFChOptFloorRuleModify = "chopt_floorrulemodify" //! 修改包厢玩法
	HFChOptMemIn           = "chopt_memin"           //! 玩家进入楼层
	HFChOptMemInV2         = "chopt_meminV2"         //! 混排大厅版本玩家进入楼层
	HFChOptMemOut          = "chopt_memout"          //! 玩家离开楼层
	HFChOptFloorDel        = "chopt_floordel"        //! 玩家删除楼层
	HFChoptFloorActive     = "chopt_flooractive"     //! 玩家激活楼层
	HFChOptTableIn         = "chopt_tablein"         //! 玩家进入牌桌
	HFChOptTableInByID     = "chopt_tableinbyid"     //! 玩家进入牌桌
	HFChOptTableCreate_Ack = "chopt_tablecreate_ack" //! 玩家创建牌桌
	HFChOptTableIn_Ack     = "chopt_tablein_ack"     //! 玩家进入牌桌
	HFChOptTableIn_Ntf     = "chopt_tablein_ntf"     //! 玩家进入牌桌
	HFChOptTableOut_Ntf    = "chopt_tableout_ntf"    //! 玩家离开牌桌
	HFChOptTableDel_Req    = "chopt_tabledel_req"    //! 游戏服解散牌桌
	HFChOptTableDel_Ack    = "chopt_tabledel_ack"    //! 游戏服解散牌桌
	HFChOptTableDel_Ntf    = "chopt_tabledel_ntf"    //! 游戏服解散牌桌
	HFChOptFloorDestory    = "chopt_floordestory"    //! 游戏服解散牌桌
	HFCHOptTableUpdate     = "chopt_tableupdate"     //! 桌子信息更新
	HFCHOptTableStepUpdate = "chopt_tablestepupdate" //! 桌子信息更新

	MsgTypeHTableCreate_Req = "htablecreate_req" // 创建牌桌
	MsgTypeHTableCreate_Ack = "htablecreate_ack" // 创建牌桌
	MsgTypeHTableIn_Req     = "htablein_req"     // 加入牌桌
	MsgTypeHTableIn_Ack     = "htablein_ack"     // 加入牌桌
	MsgTypeHTableDel_Req    = "htabledel_req"    // 解散牌桌
	MsgTypeHTableDel_Ack    = "htabledel_ack"    // 解散牌桌
	MsgTypeHFInfo           = "hfloorinfo"       //! 包厢楼层消息
	MsgTypeSetLogFileLevel  = "setlogfilelevel"  // 设置日志文件打印等级

	//包厢消息
	MsgTypeHouseOptIsMemCheck            = "houseoptionismembercheck"  //! 包厢设置审核
	MsgTypeHouseOptIsMemExit             = "houseoptionismemberexit"   //! 包厢设置退圈开关
	MsgTypeHouseOptIsParnterMemCheck     = "houseoptionisparntercheck" //! 包厢队长审核
	MsgTypeHouseOptIsMemCheck_Ntf        = "houseoptionismembercheck_ntf"
	MsgTypeHouseOptIsMemExitCheck_Ntf    = "houseoptionismemberexitcheck_ntf"
	MsgTypeHouseOptIsParnterMemCheck_Ntf = "houseoptionisparntercheck_ntf"
	MsgTypeHouseOptIsFrozen              = "houseoptionisfrozen"             //! 包厢设置冻结
	MsgTypeHouseOptIsFrozen_Ntf          = "houseoptionisfrozen_ntf"         //! 包厢设置冻结
	MsgTypeHouseOptIsMemHide             = "houseoptionismemhide"            //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsHidHide             = "houseoptionishidhide"            //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsHeadHide            = "houseoptionisheadhide"           //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsMemUidHide          = "houseoptionismemuidhide"         //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsOnlineHide          = "houseoptionisonlinehide"         //! 包厢设置成员列表隐藏
	MsgTypeHouseOptPartnerKick           = "houseoptionpartnerkick"          //! 包厢设置成员列表隐藏
	MsgTypeHouseOptRewardBalanced        = "houseoptionrewardbalanced"       //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsMemHide_Ntf         = "houseoptionismemberhide_ntf"     //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsHidHide_Ntf         = "houseoptionishidhide_ntf"        //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsHeadHide_Ntf        = "houseoptionisheadhide_ntf"       //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsMemUidHide_Ntf      = "houseoptionismemuidhide_ntf"     //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsOnlineHide_Ntf      = "houseoptionisonlinehide_ntf"     //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsPartnerKick_Ntf     = "houseoptionispartnerkick_ntf"    //! 包厢设置成员列表隐藏
	MsgTypeHouseOptIsRewardBalanced_Ntf  = "houseoptionisrewardbalanced_ntf" //! 包厢设置成员列表隐藏
	MsgTypeHousePartner                  = "housepartner"                    //! 包厢队长权限
	MsgTypeHousePartnerList              = "housepartnerlist"                //! 包厢队长列表
	MsgTypeHousePartnerMemCustom         = "housepartnermemcustom"           //! 包厢队长未绑定队长玩家
	MsgTypeHousePartnerBindUser          = "housepartnerbinduser"            //! 包厢队长名下玩家

	MsgTypeHousePartnerGreate       = "housepartnercreate"       //! 包厢队长权限创建
	MsgTypeHousePartnerGreate_Ntf   = "housepartnercreate_ntf"   //! 包厢队长权限创建通知
	MsgTypeHousePartnerJunior_Ntf   = "housepartnerjunior_ntf"   //! 包厢队长权限创建通知
	MsgTypeHousePartnerDelete       = "housepartnerdelete"       //! 包厢队长权限删除
	MsgTypeHousePartnerDelete_Ntf   = "housepartnerdelete_ntf"   //! 包厢队长权限删除通知
	MsgTypeHousePartnerGen          = "housepartnergen"          //! 包厢队长调配
	MsgTypeHousePartnerGen_Ntf      = "housepartnergen_ntf"      //! 包厢队长调配
	MsgTypeHouseCreate              = "housecreate"              //! 包厢创建
	MsgTypeHouseCreate_Ntf          = "housecreate_ntf"          //! 包厢创建通知
	MsgTypeHouseDelete              = "housedelete"              //! 包厢删除
	MsgTypeHouseDelete_Ntf          = "housedelete_ntf"          //! 包厢删除通知
	MsgTypeHouseBaseInfo            = "housebaseinfo"            //! 包厢基本信息
	MsgTypeHouseMemberOnline        = "housememberonline"        //! 包厢在线玩家数
	MsgTypeHousePlayOften           = "houseplayoften"           //! 包厢常玩
	MsgTypeHouseBaseNNModify        = "housebasennmodify"        //! 包厢名称公告修改
	MsgTypeHouseBaseNNModify_Ntf    = "housebasennmodify_ntf"    //! 包厢修改名称公告通
	MsgTypeHouseFloorCreate         = "housefloorcreate"         //! 包厢楼层创建
	MsgTypeHouseFloorCreate_Ntf     = "housefloorcreate_ntf"     //! 包厢楼层创建通知
	MsgTypeHouseFloorDelete         = "housefloordelete"         //! 包厢楼层删除
	MsgTypeHouseFloorDelete_Ntf     = "housefloordelete_ntf"     //! 包厢楼层删除通知
	MsgTypeHouseFloorList           = "housefloorlist"           //! 包厢楼层列表
	MsgTypeHouseFloorRuleModify     = "housefloorrulemodify"     //! 包厢楼层玩法修改
	MsgTypeHouseFloorRuleModify_Ntf = "housefloorrulemodify_ntf" //! 包厢楼层玩法修改
	MsgTypeHouseTableBaseInfo       = "housetablebaseinfo"       //! 包厢牌桌详细数据
	MsgTypeHouseTableIn             = "housetablein"             //! 包厢加入牌桌
	MsgTypeHouseTableInByID         = "housetableinbyid"         //! 包厢根据桌子id加入牌桌
	// MsgTypeHouseTableIn_Ntf         = "housetablein_ntf"         //! 包厢加入牌桌
	// MsgTypeHouseTableOut_Ntf        = "housetableout_ntf"        //! 包厢离开牌桌
	// MsgTypeHouseTableCreate_Ntf     = "housetablecreate_ntf"     //! 包厢创建牌桌
	// MsgTypeHouseTableDissovle_Ntf   = "housetabledissovle_ntf"   //! 包厢解散牌桌
	MsgTypeHouseTableUpdate_Ntf  = "housetableupdate_ntf"  //! 包厢某个牌桌发生变化
	MsgTypeHouseMemberList       = "housememberlist"       //! 包厢玩家列表
	MsgTypeHouseApplyInfo        = "houseapplyinfo"        //! 包厢玩家列表
	MsgTypeHouseMemberTrackPoint = "housemembertrackpoint" //! 包厢成员小红点
	MsgTypeHousepartnerAddList   = "housepartneraddlist"   //! 包厢队长添加列表
	MsgTypeHouseMemberJoin       = "housememberjoin"       //! 包厢玩家加入
	MsgTypeHouseMemberApply_Ntf  = "housememberapply_ntf"  //! 包厢玩家加入通知
	MsgTypeHouseFloorIn_NTF      = "housefloorin_ntf"      // 玩家上线
	MsgTypeHouseMemOnline_NTF    = "housememonline_ntf"    // 玩家上线
	MsgTypeHouseMemOffline_NTF   = "housememoffline_ntf"   // 玩家下线
	MsgTypeHouseMemInGame_NTF    = "housememingame_ntf"    // 玩家进入游戏
	MsgTypeHouseMemOutGame_NTF   = "housememoutgame_ntf"   // 玩家进入游戏

	MsgTypeHouseMemberAgree           = "housememberagree"           //! 包厢玩家过审
	MsgTypeHouseMemberAgree_NTF       = "housememberagree_ntf"       //! 包厢玩家过审通知
	MsgTypeHouseMemberRefused         = "housememberrefused"         //! 包厢玩家拒审
	MsgTypeHouseMemberRefused_NTF     = "housememberrefused_ntf"     //! 包厢玩家拒审通知
	MsgTypeHouseMemberExit            = "housememberexit"            //! 包厢玩家退出
	MsgTypeHouseMemberKick            = "housememberkick"            //! 包厢玩家剔出
	MsgTypeHouseMemberKick_Ntf        = "housememberkick_ntf"        //! 包厢玩家剔出通知
	MsgTypeHouseMemberblacklistInsert = "housememberblacklistinsert" //! 包厢玩家黑名单加入
	MsgTypeHouseMemberblacklistDelete = "housememberblacklistdelete" //! 包厢玩家黑名单删除
	MsgTypeHouseMemberRemark          = "housememberremark"          //! 包厢玩家备注
	MsgTypeHouseMemberRoleGen         = "housememberrolegen"         //! 包厢玩家角色授权
	MsgTypeHouseMemberRoleGen_Ntf     = "housememberrolegen_ntf"     //! 包厢玩家角色变更
	MsgTypeHouseMemberIn              = "housememberin"              //! 包厢玩家进入楼层
	MsgTypeHouseMemberOut             = "housememberout"             //! 包厢玩家离开楼层
	MsgTypeMemberHouseList            = "memberhouselist"            //! 入驻包厢列表
	MsgTypeMemberStatistics           = "memberstatistics"           //! 成员统计查询
	MsgTypeMemberStatisticsTotal      = "memberstatisticstotal"      //! 成员统计查询总计
	MsgTypeHouseGameRecord            = "housegamerecord"            //! 包厢战绩查询
	MsgTypeHouseOperationalStatus     = "houseoperationalstatus"     //! 包厢经营状况
	MsgTypeHouseValidRoundScoreGet    = "housevalidroundscoreget"    //! 包厢有效对局积分查询
	MsgTypeHouseValidRoundScoreSet    = "housevalidroundscoreset"    //! 包厢有效对局积分设置
	MsgTypeHouseMyRecord              = "housemyrecord"              //! 包厢我的战绩查询
	MsgTypeHouseRecord                = "houserecord"                //! 包厢战绩查询
	MsgTypeHouseRecordStatus          = "houserecordstatus"          //! 包厢大赢家对局统计
	MsgTypeHouseRecordHeart           = "houserecordheart"           //! 包厢战绩点赞功能
	MsgTypeHouseRecordStatusClean     = "houserecordstatusclean"     //! 包厢大赢家对局统计清除
	MsgTypeHouseRecordStatusCleanAll  = "houserecordstatuscleanall"  //! 包厢大赢家对局统计清除所有
	MsgTypeHouseRecordKaCost          = "houserecordkacost"          //! 包厢sortType
	MsgTypeHouseActivityCreate        = "houseactivitycreate"        //! 包厢楼层活动创建
	MsgTypeHouseActivityDelete        = "houseactivitydelete"        //! 包厢楼层活动删除
	MsgTypeHouseActivityDeleteNTF     = "houseactivitydelete_ntf"    //! 包厢楼层活动删除

	MsgTypeHouseActivityList = "houseactivitylist"      //! 包厢楼层活动列表
	MsgTypeHouseActivityInfo = "houseactivityinfo"      //! 包厢楼层活动详情
	MsgTypeHouseAreaGameList = "housegamelist"          //! 包厢楼层活动详情
	MsgTypeHouseIDChange     = "houseidchange"          //! 包厢hid修改
	MsgTypeHouseIDChange_NTF = "houseidchange_ntf"      //! 包厢hid修改
	MsgTypeLimitUserPlay     = "houseuserlimitgame"     //禁止用户游戏
	MsgTypeLimitUserPlay_NTF = "houseuserlimitgame_ntf" //禁止用户游戏通知

	MsgTypeHouseVitaminStatus_ntf     = "housevitaminstatus_ntf"  //! 包厢防沉迷开关
	MsgTypeHouseVitaminValues         = "housevitaminvalues"      //! 包厢疲劳值数值
	MsgTypeHouseVitaminInfo           = "housevitamininfo"        //! 包厢疲劳值开关信息
	MsgTypeHouseVitaminOptHide_Ntf    = "housevitaminopthide_ntf" //! 包厢管理员可见
	MsgTypeHouseVitaminSet            = "housevitaminset"         //! 包厢疲劳值修改
	MsgTypeHouseVitaminSet_Ntf        = "housevitaminset_ntf"     //! 包厢疲劳值修改通知
	MsgTypeHouseVitaminChange_Ntf     = "housevitaminchange_ntf"  //! 包厢疲劳值修改通知
	MsgTypeHouseVitaminSetRecords     = "housevitaminsetrecords"  //! 包厢疲劳值修改记录
	MsgTypeHouseVitaminClear          = "housevitaminclear"       //! 包厢疲劳值清空
	MsgTypeHouseFloorVitaminDeductGet = "hfvitamindeductget"      //! 包厢楼层疲劳值扣除相关获取
	MsgTypeHouseFloorVitaminDeductSet = "hfvitamindeductset"      //! 包厢楼层疲劳值扣除相关获取
	MsgTypeHouseFloorVitaminEffectGet = "hfvitamineffectget"      //! 包厢楼层疲劳值扣除相关获取
	MsgTypeHouseFloorVitaminEffectSet = "hfvitamineffectset"      //! 包厢楼层疲劳值扣除相关获取

	MsgTypeHousePartnerVitaminStatistic = "housepartnervitaminstatistics" //! 包厢队长疲劳值统计
	MsgTypeHouseVitaminStatistic        = "housevitaminstatistics"        //! 包厢疲劳值统计
	MsgTypeHouseVitaminStatisticClear   = "housevitaminstatisticsclear"   //! 包厢疲劳值统计清零
	MsgTypeHouseVitaminMgr              = "housevitaminmgrlist"           //! 包厢疲劳值管理列表

	MsgTypeLeagueCardPoolChange = "leaguecardpoolchange_ntf" //! 加盟商包厢卡池权限变更
	MsgTypeHouseCardPoolChange  = "housecardpoolchange_ntf"  //! 包厢卡池权限变更
	MsgTypeHouseMsg             = "housemsg"                 //! 包厢消息

	// 混排大厅
	MsgTypeHouseFloorRename            = "housefloorrename"             // 修改楼层名称
	MsgTypeHouseFloorRenameNTF         = "housefloorrename_ntf"         // 包厢名称变更通知
	MsgTypeHouseMixFloor               = "housemixfloor"                // 增加混排大厅规则
	MsgTypeHouseMixFloorNTF            = "housemixfloor_ntf"            // 混排大厅通知
	MsgTypeHouseMixFloorEditor         = "housemixflooreditor"          // 编辑混排大厅规则
	MsgTypeHouseMixFloorEdirotNTF      = "housemixflooreditor_ntf"      // 编辑混排大厅通知
	MsgTYpeHouseMixFloorInfo           = "housemixfloorinfo"            // 混排大厅信息
	MsgTYpeHouseMixFloorTableCreate    = "housemixfloortablecreate"     // 创建混排大厅楼层桌子
	MsgTypeHouseMixFloorTableChangeNTF = "housemixfloortablechange_ntf" // 编辑混排大厅通知
	MsgTYpeHouseMixFloorTableDelete    = "housemixfloortabledelete"     // 销毁混排大厅楼层桌子
	MsgTYpeHouseMixFloorTableChange    = "housemixfloortablechange"     // 批量修改混排大厅桌子

	// 任务
	MsgTypeTaskList         = "tasklist"          //! 任务列表
	MsgTypeTaskCommit       = "taskcommit"        //! 任务数据提交
	MsgTypeTaskReward       = "taskreward"        //! 任务奖励领取
	MsgTypeTaskCompletedNtf = "taskcompleted_ntf" //! 任务完成通知

	//游戏内任务
	MsgTypeGameTaskProCommit      = "gametaskprocommit"      //! 游戏内任务数据提交
	MsgTypeGameTaskList           = "gametasklist"           //! 游戏内任务列表
	MsgTypeGameTaskReward         = "gametaskreward"         //! 游戏内任务奖励领取
	MsgTypeGameTaskCompletedCount = "gametaskcompletedcount" //! 游戏内任务已完成&未领取任务数量
	MsgTypeGameTaskCompletedNtf   = "gametaskcompleted_ntf"  //! 游戏内任务奖励通知可领取
	MsgTypeGameTaskDataUpdate     = "gametaskdataupdate"     //! 游戏内任务数据更新

	//场次相关
	MsgTypeSiteIn        = "sitein"        // 加入场次
	MsgTypeSiteIn_Ack    = "sitein_ack"    // 加入场次
	MsgTypeSiteIn_Ntf    = "sitein_ntf"    // 加入场次
	MsgTypeSiteInfo      = "siteinfo"      // 场次详情
	MsgTypeSiteExit      = "siteexit"      // 离开场次
	MsgTypeSiteExit_Ntf  = "siteexit_ntf"  // 离开场次
	MsgTypeSiteTableIn   = "sitetablein"   // 进入牌桌
	MsgTypeChangeTableIn = "changetablein" // 换桌
	MsgTypeSiteListIn    = "sitelistin"    // 加入场次列表
	MsgTypeSiteListOut   = "sitelistout"   // 退出场次列表
	MsgTypeSiteTableList = "sitetablelist" // 场次牌桌列表
	MsgTypeSiteTable_Ntf = "sitetable_ntf" // 牌桌信息变化通知

	//任务相关
	MsgGameTaskComplete = "gametaskcomplete" // 通知牌桌局数变化

	MsgTypeMatchList        = "matchlist"        //! 排位赛列表
	MsgTypeMatchRankingList = "matchrankinglist" //! 排位赛排位列表
	MsgTypeMatchAwardList   = "matchawardlist"   //! 排位赛奖励列表
	// 加盟商相关接口
	MsgTypeAddLeague          = "addleague"          // 添加加盟商
	MsgTypeAddLeagueUser      = "addleagueuser"      // 添加加盟商用户
	MsgTypeLeagueFreeze       = "leaguefreeze"       // 冻结加盟商
	MsgTypeLeagueUnFreeze     = "leagueunfreeze"     // 解冻加盟商
	MsgTypeLeagueUserFreeze   = "leagueuserfreeze"   // 冻结加盟商用户
	MsgTypeLeagueUserUnFreeze = "leagueuserunfreeze" // 解冻加盟商用户
	// 禁止同圈相关功能
	MsgTypeHouseMemberTableLimit      = "housemembertablelimitlist"  // 包厢编辑禁止同桌时候列表
	MsgTypeHouseTableLimitGroupAdd    = "housetablelimitgroupadd"    // 添加包厢禁止同桌分组
	MsgTypeHouseTableLimitGroupRemove = "housetablelimitgroupremove" // 移除包厢禁止同桌分组
	MsgTypeHouseTableLimitUserAdd     = "housetablelimituseradd"     // 添加包厢禁止同桌用户
	MsgTypeHouseTableLimitUserRemove  = "housetablelimituserremove"  // 移除包厢禁止同桌用户
	MsgTypeHouseTableLimitInfo        = "housetablelimitinfo"        // 包厢禁止同桌信息
	MsgHouse2PTableLimitNotEffect     = "house2ptablelimitnoteffect" // 2人桌子禁止同桌不生效 是否勾选

	//新疲劳值
	MsgVitaminSend             = "housevitaminsend"        //玩家间赠送疲劳值
	MsgVitaminSend_NTF         = "housevitaminsend_ntf"    //玩家间赠送疲劳值
	MsgHouseMemberGetById      = "housememgetbyid"         //根据用户id精确搜索用户
	MsgHouseVitaminPoolAdd     = "housevitaminpooladd"     // 疲劳值仓库
	MsgHouseVitaminPoolAdd_NTF = "housevitaminpooladd_ntf" //疲劳值仓库操作记录
	MsgHouseVitaminPoolLog     = "housevitaminpoollog"     //仓库操作日志

	MsgHouseVitaminPoolSum   = "housevitaminpoolsum"          //通知大厅统计疲劳值仓库
	MsgHouseMemVitaminSum    = "housememvitaminsum"           //通知大厅统计用户疲劳值
	MsgHouseMemLeftStatistic = "housememvitaminleftstatistic" //通知大厅统计用户疲劳值

	MsgHouseFloorTableSync = "housefloortablesync" //通知大厅统计用户疲劳值
	MsgHouseFloorTableSort = "housefloortablesort" //通知大厅统计用户疲劳值

	// 合并包厢相关
	MsgTypeHouseMergeIntention    = "housemergeintention"     // 合并包厢意向
	MsgTypeHouseMergeCheck        = "housemergecheck"         // 合并包厢检查
	MsgTypeHouseMergeRequest      = "housemergereq"           // 发起合并包厢请求
	MsgTypeHouseMergeResponse     = "housemergersp"           // 响应合并包厢请求
	MsgTypeHouseMergeReqRevoke    = "housemergereqrevoke"     // 撤销合并包厢请求
	MsgTypeHouseMergeReqRevokeNtf = "housemergereqrevoke_ntf" // 撤销合并包厢请求通知
	MsgTypeHouseRevokeRequest     = "houserevokereq"          // 发起撤销合并包厢请求
	MsgTypeHouseRevokeResponse    = "houserevokersp"          // 响应撤销合并包厢请求
	MsgTypeHouseMergeRecord       = "housemergerecord"        // 合并包厢记录
	MsgTypeHouseMergeOk           = "housemergeok"            // 通知包厢被合并
	MsgTypeHouseRevokeOk          = "houserevokeok"           // 通知包厢被撤销合并包厢
	MsgTypeHouseMerge_NTF         = "housemerge_ntf"          // 合并包厢请求推送
	MsgTypeHouseRevoke_NTF        = "houserevoke_ntf"         // 撤销合并包厢请求推送

	MsgHouseJoinTableSet          = "housejointableset"           //快速入桌开关
	MsgHouseJoinTableChangeNtf    = "housejointablechange_ntf"    //快速入桌变更通知
	MsgHouseTblShowCountChangeNtf = "housetblshowcountchange_ntf" //快速入桌变更通知

	MsgTypeHallNotifyPush = "notify" // 通知
	MsgUserPhoneChange    = "userphonechange"
	MsgHouseRevokeGm      = "houserevoke"
	MsgAddHotVersion      = "addhotversion"

	MsgHouseOwnerChange = "houseownerchange" // 修改盟主id

	MsgHouseChangeJoinType = "housetablejointypechange"
	// 队长分层相关
	MsgHouseParnterRoyaltySet            = "houseparnterroyaltyset"            // 设置队长分层
	MsgHouseParnterRoyaltyGet            = "houseparnterroyaltyget"            // 获取队长分层
	MsgHouseOwnerRoyaltySet              = "houseownerroyaltyset"              // 设置队长分层
	MsgHouseOwnerRoyaltyGet              = "houseownerroyaltyget"              // 获取队长分层
	MsgHouseParnterSuperiorList          = "houseparntersuperiorlist"          // 绑定上级队长
	MsgHouseParnterBindSuperior          = "houseparnterbindsuperior"          // 绑定上级队长
	MsgHouseParnterBindSuperiorNtf       = "houseparnterbindsuperior_ntf"      // 绑定上级队长
	MsgHouseParnterBindJunior            = "houseparnterbindjunior"            // 绑定下级队长
	MsgHouseParnterFloorStatistics       = "houseparnterfloorstatistics"       // 队长分层统计
	MsgHouseParnterFloorJuniorStatistics = "houseparnterfloorjuniorstatistics" // 队长下级分层统计
	MsgHouseParnterFloorMemStatistics    = "houseparnterfloormemstatistics"    // 队长名下分层统计

	MsgHousePartnerHistoryFloorStatistics       = "housepartnerfloorhistorystatistics"       // 队长分层删除楼层历史数据
	MsgHousePartnerHistoryFloorDetailStatistics = "housepartnerfloorhistorydtrailstatistics" // 队长分成删除楼层历史数据详情

	MsgHouseAutoPayPartner          = "houseautopaypartner" //队长自动划扣开关
	MsgHouseAutoPayPartnerMsg       = "houseautopaypartnermsg"
	MsgHouseAutoPayPartnerNtf       = "houseautopaypartner_ntf"
	MsgHouseSingleAutoPayPartnerMsg = "housesingleautopaypartnermsg"

	MsgHouseFloorMappingNumUpdate = "housefloormappingupdate"

	MsgHouseMemKick = "msgadminkickhousemem"

	MsgPartnerRoyaltyForMe   = "houseparnterroyaltyforme"
	MsgPartnerRoyaltyHistory = "houseparnterroyaltyhistory"

	MsgNoLeagueStatistics       = "housenoleaguestatistics"
	MsgNoLeagueDetailStatistics = "housenoleaguedetailstatistics"

	MsgGameSwitch    = "housegameswitch"
	MsgGameSwitchNtf = "housegameswitch_ntf"

	MsgUserAgentUpdate     = "useragentupdate"
	MsgHouseAgentUpdateNtf = "houseagentupdate_ntf"

	MsgHousePrizeInfo   = "houseprizeinfo"
	MsgHousePrizeSet    = "houseprizeset"
	MsgHousePrizeSetNtf = "houseprizeset_ntf"

	MsgHouseFloorGearPayGet = "housefloorgearpayget"
	MsgHouseFloorGearPaySet = "housefloorgearpayset"

	MsgHouseLuckSet                    = "houseluckconfig"
	MsgHouseLuckInfo                   = "houseluckinfo"
	MsgHouseMemberCanLuck              = "housemembercanluck"
	MsgHouseMemberGetLuck              = "housemembergetluck"
	MsgHouseActDetail                  = "houseactdetail"
	MsgHouseMemberLuckRecord           = "housememberluckrecord"
	MsgTypeHouseSetVipFloorNtf         = "housevipfloorset_ntf"        // vip楼层变更通知
	MsgTypeHouseFloorVipUserSetNtf     = "housefloorvipuserset_ntf"    // vip楼层vip玩家变更通知
	MsgTypeHouseVipFloorGet            = "housevipfloorget"            // 得到所有楼层的vip信息
	MsgTypeHouseVipFloorSet            = "housevipfloorset"            // 设置所有楼层的vip信息
	MsgTypeHouseFloorVipUsersGet       = "housefloorvipusersget"       // 得到楼层的vip玩家
	MsgTypeHouseFloorEverymanGet       = "houseflooreverymanget"       // 得到楼层的普通玩家
	MsgTypeHouseFloorSetVipUser        = "housefloorsetvipuser"        // 设置楼层的vip玩家
	MsgTypeHouseFloorSetAllVipUser     = "housefloorsetallvipuser"     // 一键设置楼层的vip玩家
	MsgTypeHouseFloorSetAllVipUser_Ntf = "housefloorsetallvipuser_ntf" // 一键设置楼层的vip玩家

	MsgHouseTableDistanceLimitGet    = "housetabledistancelimitget"
	MsgHouseTableDistanceLimitSet    = "housetabledistancelimitset"
	MsgHouseTableDistanceLimitSetNtf = "housetabledistancelimitsetntf"

	MsgHouseApplySwitch = "houseapplyswitch"

	MsgHouseFloorHideImg    = "housefloorhideimg"
	MsgHouseFloorHideImgNTF = "housefloorhideimg_ntf"
	MsgHouseFloorFakeTable  = "housefloorfaketable"

	MsgHouseRecordGameLike = "houserecordgamelike"
	MsgHouseRecordUserLike = "houserecorduserlike"

	MsgHousePrivateGPSGet     = "houseprivategpsget"
	MsgHousePrivateGPSSet     = "houseprivategpsset"
	MsgHousePrivateGPSSet_ntf = "houseprivategpsset_ntf"
	MsgRecommendGoldGameGet   = "recommendgoldgameget"

	MsgChangeBlankUser     = "changeblankuser"
	MsgChangeBlankHouse    = "changeblankhouse"
	MsgBlankUserChangeNTF  = "userblankchange_ntf"
	MsgBlankHouseChangeNTF = "houseblankchange_ntf"

	MsgTypeHouseFloorColorSet    = "housefloorcolorset"
	MsgTypeHouseFloorColorSetNtf = "housefloorcolorset_ntf"

	MsgTypeTableUserKick = "tableuserkick"
	MsgTypePartnerRemark = "partnerremark"

	MsgTypeDailyReward    = "dailyreward"    // 每日礼包
	MsgTypeDailyRewardReq = "getdailyreward" // 领取每日礼包

	MsgTypeGameSiteList = "gamesitelist"  // 游戏场次信息
	MsgTypeGoldGameList = "goldgamelist"  // 金币游戏列表
	MsgEnterGoldGame    = "entergoldgame" // 进入金币场
	MsgTypeShareCfg     = "sharecfg"      // 获取分享配置信息
	MsgTypeShareSuc     = "sharesuc"      // 分享成功

	MsgTypeCheckinInfo = "checkininfo" // 金币游戏列表

	MsgTypeGetAllowanceInfo       = "getallowanceinfo"
	MsgTypeCheckBuyBankruptcyGift = "checkbuybgift"

	MsgTypeOneStepBuyGift = "bankruptcygift" // 破产礼包

	MsgTypeSetFangKaTipsMinNumReq = "setfangkatipsminnumreq" // 设置房卡低于xx时提示盟主 请求
	MsgTypeSetFangKaTipsMinNumRsp = "setfangkatipsminnumrsp" // 设置房卡低于xx时提示盟主 返回

	MsgTypeLimitCaptainPlay     = "housecaptainlimitgame"     //禁止队长或队长和成员 游戏
	MsgTypeLimitCaptainPlay_NTF = "housecaptainlimitgame_ntf" //禁止队长或队长和成员 游戏
	MsgHallSearchUser           = "searchuser"                //搜索玩家
	MsgHouseMtAddUser           = "housemtadduser"            //手动添加玩家
	MsgUserRefuseInvite         = "setuserrefuseinvite"       //设置玩家 拒绝或者接受 包厢邀请

	MsgTypeHmLookUserRight       = "hmlookuserright"       // 查询当前用户权限
	MsgTypeHmUpdateUserRight     = "hmupdateuserright"     // 修改当前用户权限
	MsgTypeHmUpdateUserRight_NTF = "hmupdateuserright_ntf" // 修改当前用户权限_通知

	MsgTypeHmSetSwitch     = "hmsetswitch"     //修改包厢功能开关
	MsgTypeHmSetSwitch_ntf = "hmsetswitch_ntf" //修改包厢功能开关

	MsgSetRecordTimeIntervalReq = "setrecordtimeintervalreq" // 设置战绩筛选时段 请求
	MsgSetRecordTimeIntervalRsp = "setrecordtimeintervalrsp" // 设置战绩筛选时段 返回

	MsgHotUpdate_Ntf       = "hotupdate_ntf"   //热更新推送
	MsgTypeTableStatus_Ntf = "tablestatus_ntf" // 牌桌变化通知

	MsgTypeHouseRankSet               = "houserankset"                  // 排行榜设置
	MsgTypeHouseRankSet_ntf           = "houserankset_ntf"              // 修改包厢功能开关
	MsgTypeHouseRankGet               = "houserankget"                  // 排行榜获取设置
	MsgTypeHouseRankInfoGet           = "houserankinfoget"              // 获取排行榜数据
	MsgTypeOffWork                    = "offwork"                       // 盟主小队或者队长小队 打烊了
	MsgTypeOffWork_ntf                = "offwork_ntf"                   // 盟主小队或者队长小队 打烊了  通知p
	MsgUserSetSex                     = "usersetsex"                    // 修改性别
	MsgTHDismissRoomDet               = "thdismissroomdet"              // 获取包厢中途解散的房间的解散详情
	MsgGpsToSvr                       = "gpstosvr"                      // 将获取到的gps信息发给服务器
	MsgTypeLimitPlayTime              = "limitplaytime"                 // 游戏时长已达上限
	MsgTypeHouseTeamBan               = "houseteamban"                  // 整队禁止
	MsgTypeHouseTeamBan_ntf           = "houseteamban_ntf"              // 整队禁止
	MsgTypePartnerAlarmValueSet_ntf   = "housepartneralarmvalueset_ntf" // 整队禁止
	MsgTypePartnerAlarmValueSet       = "housepartneralarmvalueset"     // 设置队长的警戒值
	MsgTypePartnerAlarmValueGet       = "housepartneralarmvalueget"     // 获取队长的警戒值
	MsgTypePartnerRewardSet_ntf       = "housepartnerrewardset_ntf"     // 设置队长的警戒值
	MsgTypePartnerRewardSet           = "housepartnerrewardset"         // 设置队长的警戒值
	MsgTypePartnerRewardGet           = "housepartnerrewardget"         // 获取队长的警戒值
	MsgTypePartnerAACostSet           = "houseparteraaset"              // 设置队长的AAk扣卡
	MsgTypePartnerAACostSet_ntf       = "houseparteraaset_ntf"          // 设置队长的AAk扣卡
	MsgTypeHouseTeamKick              = "houseteamkick"                 // 整队踢出
	MsgTypeHouseTeamKick_ntf          = "houseteamkick_ntf"             // 整队踢出
	MsgTypeHouseCardStatistics        = "housecardstatistics"           // 包厢房卡统计
	MsgTypeHouseRewardStatistics      = "houserewardstatistics"         // 包厢赛点奖励统计
	MsgTypeHouseClearMyReward         = "houseclearmyreward"            // 包厢清零我自己的奖励
	MsgTypeHouseMemberNoFloorsGet     = "housemembernofloorsget"        // 包厢限制获取
	MsgTypeHouseMemberNoFloorsSet     = "housemembernofloorsset"        // 包厢限制set
	MsgTypeHouseMemberNoFloorsSet_ntf = "housemembernofloorsset_ntf"    // 包厢限制set
	MsgTypeUserBattleLevel            = "userbattlelevel"               // 用户对战等级
	MsgTypeActivitySpinInfo           = "activityspininfo"
	MsgTypeActivitySpinDo             = "activityspindo"
	MsgTypeActivitySpinRecord         = "activityspinrecord"
	MsgTypeActivityCheckinInfo        = "activitycheckininfo"
	MsgTypeActivityCheckinDo          = "activitycheckindo"
	MsgTypeActivityCheckinRecord      = "activitycheckinrecord"
	MsgTypeGetBattleRank              = "getbattlerank"

	MsgTypePaymentCreateOrderId = "paymentcreateorderid" // 创建订单号
	MsgTypePaymentCreateOrder   = "paymentcreateorder"   // 创建订单
	MsgTypePaymentResultNtf     = "paymentresult_ntf"    // 支付结果推送
)

// 大厅HTTP协议
const (
	MsgTypeHouseAnotherGame = "houseanothergame" // 包厢房间再来一局
	MsgTypeHouseChangeTable = "housechangetable" // 游戏服换桌
	MsgHouseTableReset      = "housetablereset"  // 将该大厅重新加载
	MsgResetHmUserRight     = "resethmuserright" // 重置用户指定包厢权限
	MsgForceHotter          = "forcehotter"      // 热更新推送
	MsgWriteOffUser         = "writeoffuser"     // 注销用户
	MsgTypePlayTime         = "userplaytime"     // 未实名、已实名未成年的游戏时长
)
