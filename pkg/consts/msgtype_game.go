package consts

const (
	MsgCommonToGameContinue = "table2common2game_continue" // 游戏Common层通知子游戏继续游戏（现只限用于及时结算暂停恢复后发牌）
)

const (
	//游戏房间消息
	MsgTypeGameGVoiceMember       = "gvoicemember"       //即时语音
	MsgTypeGameTimeMsg            = "gameTimeMsg"        //计时器事件
	MsgTypeGameOperateNotify      = "gameOperateNotify"  //操作提示
	MsgTypeGameStart              = "gameStart"          //开始游戏
	MsgTypeGamePaoSetting         = "paoSetting"         //下跑设置
	MsgTypeGameXiapao             = "xiapao"             //下跑
	MsgTypeGamePiaoSetting        = "piaoSetting"        //选漂设置
	MsgTypeGameXuanpiao           = "xuanpiao"           //选漂
	MsgTypeGameUserStatus         = "userStatus"         //玩家状态发送
	MsgTypeGameBalanceGame        = "balanceGame"        //总结算
	MsgTypeGameSendCard           = "sendcard"           //发牌
	MsgTypeGameEnd                = "gameend"            //小结算
	MsgTypeGameOperateResult      = "operateresult"      //操作结果
	MsgTypeGameOperateBuHuaNotify = "operatebuhuanotify" //补花通知（监利麻将）
	MsgTypeGameGPSReq             = "gps"                //GPS数据
	MsgTypeGameStatusFree         = "statusFree"         //游戏空闲状态
	MsgTypeGameStatusPlay         = "statusplay"         //游戏状态
	MsgTypeGameStatusHiDi         = "StatusHiDi"         //游戏状态海底捞
	MsgTypeGameTableStatus        = "tablestatus"        //桌子状态
	MsgTypeGameReadyState         = "readystate"         //准备状态
	MsgTypeGameDissMissRoom       = "dismissRoom"        //解散
	MsgTypeGameDismissFriendRep   = "dissmissfriendrep"  //申请解散房间返回
	MsgTypeGameMessage            = "gamemessage"        //游戏消息
	MsgTypeGameUserOfflineTime    = "OfflineTime"        //离线剩余时间
	MsgTypeGameUserVitaminLowTime = "VitaminLowTime"     //比赛分过低游戏暂停剩余时间
	MsgTypeGameSitFailed          = "sitfailed"          //坐下失败
	MsgTypeGameUserInfoHead       = "userinfohead"       //玩家数据下发
	MsgTypeGameSendHDV            = "sendhdv"            //发送海底捞
	MsgTypeGameForceExit          = "forceexit"          //强退
	MsgTypeGameForceTableDel      = "forcetabledel"      //服务器强踢
	MsgTypeGameFewerInfo          = "fewerinfo"          //少人开局信息
	MsgTypeGameFewerStartClose    = "fewerclose"         //申请少人开局面板关闭
	MsgTypeGameFewerStartShow     = "fewershow"          //申请少人开局面板关闭
	MsgTypeGameFewerStartHide     = "fewerhide"          //申请少人开局面板关闭
	MsgTypeGameTerminate          = "gameterminate"      //游戏达到最小需要保留的牌堆数，还未胡牌
	MsgTypeGamePause              = "gamepause"          //游戏暂停
	MsgTypeGameContinue           = "gamecontinue"       //游戏继续
	MsgTypeGameDiPaiCard          = "gamedipaicard"      //底牌信息
	MsgTypeGameExchangeSetting    = "exchangeSetting"    //换三张设置
	MsgTypeGameExchangeBoradcast  = "exchangeboradcast"  //广播换三张的状态(哪些人选好了，哪些人没有选好)
	MsgTypeGameExchangeNewCard    = "exchangenewcard"    //换三张后的新牌
	MsgTypeGameOperateScore       = "operatescore"       //操作分
	MsgTypeGameAlarmNotify        = "gameAlarmNotify"    //报警提示
	MsgTypeGameStartPunishNotify  = "startpunishnotify"  //开始罚时提醒
	MsgTypeGameHuResult           = "gamehuresult"       //血流成河胡牌结果
	MsgTypeGameFanLaiZi           = "gamefanlaizi"       //翻癞子
	MsgTypeGameBombCheck          = "gamebombcheck"      //炸弹检测
	MsgTypeGameUpdateBombNum      = "updatebombnum"      //更新桌面炸弹数量
	MsgTypeGameUpdateHandCards    = "updatehandcards"    //更新玩家手牌数据

	//c2s
	MsgTypeGameReady               = "gameready"               //准备
	MsgTypeGameCancelReady         = "gamecancelready"         //取消准备
	MsgTypeGameinfo                = "gameinfo"                //游戏信息
	MsgTypeGameBalanceGameReq      = "balanceGameReq"          //玩家请求总结算数据
	MsgTypeGameOutCard             = "gameOutCard"             //出牌命令
	MsgTypeGameBaoTingStatus       = "gamebaotingstatus"       //报听状态
	MsgTypeGameOutMagic            = "gameOutMagic"            //打出癞子命令
	MsgTypeGameOperateCard         = "gameOperateCard"         //操作命令
	MsgTypeGameTrustee             = "gameTrustee"             //用户托管
	MsgTypeGameHD                  = "gameHD"                  ////海底操作
	MsgTypeGameGoOnNextGame        = "gameGoOnNextGame"        // 一局完成后点击再来一局
	MsgTypeGameGangHouBuPai        = "gameganghoubupai"        //杠后补牌消息
	MsgTypeGameDismissFriendReq    = "dissmissfriend"          //申请解散房间
	MsgTypeGameDismissFriendResult = "dissmissresult"          //申请解散玩家选择
	MsgTypeGameFewerStartReq       = "fewerfriend"             //申请少人开局
	MsgTypeGameFewerStartResult    = "fewerresult"             //申请少人开局选择
	MsgTypeGameUserChat            = "userchat"                //聊天信息
	MsgTypeGameUserYYInfo          = "yyinfo"                  //语音聊天
	MsgTypeGameNotificationMessage = "gamenotificationmessage" //游戏通知消息
	MsgTypeGameOpStatusMessage     = "gameopstatusmessage"     //游戏通知消息
	MsgTypeGameBaoQing             = "gamebaoqing"             //报1清
	MsgTypeGameChooseDiPai         = "gamechoosedipai"         //选择底牌
	MsgTypeGameExchangeThreeCard   = "gameexchangethreecard"   //换三张 操作
	MsgTypeGameLiangPai            = "gameliangpai"            //亮牌
	MsgTypeGameGuoHu               = "gameGuoHuFanFan"         //过胡翻番
	MsgTypeGameLiangNo13           = "gameLiangNo13"           //不够13张不能亮牌
	MsgTypeGameSupperMessage       = "gamesuppermessage"       //超级防作弊
	MsgTypeGameTimeOutAutoHu       = "gametimeoutautohu"       //超时自动胡
	MsgTypeGamePLOutCard           = "gamePLOutCard"           //皮癞杠后不补牌
	MsgTypeGameExchangeCard        = "gameexchangecard"        //换牌 操作
	MsgTypeGameChatLogs            = "gamechatlogs"            //聊天记录
	MsgTypeGameIsOutCard           = "gameIsOutCard"           //通知客户端能否出牌
	MsgTypeGameChangeSeat          = "gamechangeseat"          //换座
	MsgTypeGameChangeSeatBegin     = "gamechangeseatbegin"     //换桌开始
	MsgTypeGameOutMagicCard        = "gameoutmagiccard"        //打出赖子个数推送
	MsgTypeGameTotalTimer          = "gametotaltimer"          //累计时间 恩施用
	MsgTypeGameAIOutCard           = "gameAIOutCard"           //机器人出牌命令
	MsgTypeGameAIOperateCard       = "gameAIOperateCard"       //机器人操作命令
	MsgTypeGameLeftCards           = "gameleftcards"           //游戏剩余牌
	MsgTypeGameWantCard            = "gamewantcard"            //玩家想要的牌
	MsgTypeGamePeepCard            = "gamepeepcard"            //玩家偷看牌
	MsgTypeGameWantGood            = "gamewantgood"            //玩家想要好牌
)

const (
	MsgContentGameGiveUpGang       = "弃杠后不能再杠"
	MsgContentGamePassDealer       = "过张不能胡"
	MsgContentGameTingAll          = "见字胡不能胡"
	MsgContentGameTingChiHu        = "见字胡不能炮胡"
	MsgContentGameTingMagic        = "甩牌不能胡"
	MsgContentHGameInviteSucceed   = "邀请已发出，请耐心等待玩家加入。"
	MsgContentHGameInviteOverFlow  = "超出邀请上限，请稍后再试。"
	MsgContentHGameInviteUserNil   = "用户不存在。"
	MsgContentHGameInviteFailed    = "邀请失败。"
	MsgContentHGameInviteOnBegin   = "游戏已开始，暂时不能邀请。"
	MsgContentHGameInviteOnFull    = "牌桌已坐满，暂时不能邀请。"
	MsgContentGamePausePlaying     = "您的比赛分已低于下限，请联系盟主处理后再进行游戏。"
	MsgContentGamePauseOnNext      = "您的比赛分已低于下限，请联系盟主处理后再进行下一局游戏。"
	MsgContentGamePauseOtherPlayer = "玩家%s已进入防沉迷，暂无法进行游戏，请耐心等待解除。"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////////
//打拱游戏专用
/////////////////////////////////////////////////////////////////////////////////////////////////////////////
const (
	//游戏房间消息
	MsgTypeSendPaiCount     = "sendpaicount"     //玩家手牌数目
	MsgTypeSendPower        = "sendpower"        //发送牌权
	MsgTypeRoar             = "roar"             //吼牌消息
	MsgTypeEndRoar          = "endroar"          //结束吼牌消息
	MsgTypeEndOut           = "endout"           //结束一轮
	MsgTypePlaySound        = "playsound"        //发送声音
	MsgTypePlayScore        = "playscore"        //用户分数
	MsgTypeTurnScore        = "turnscore"        //本轮分数
	MsgTypeSendTurn         = "sendturn"         //玩家几游
	MsgTypeTeamerPai        = "teamerpai"        //队友的牌
	MsgTypeGameRule         = "gamerule"         //基础规则
	MsgTypeReLinkTip        = "relinktip"        //有人断线重连回来的提示
	MsgTypeShow             = "showji"           //有人明鸡
	MsgLongCount            = "longCount"        //发送拱笼的笼数
	RandTaskID              = "randtaskid"       //系统选中的随机任务
	FinishedTaskID          = "finishedtaskid"   //游戏任务完成状态
	MsgTypeNeedSplitCard    = "needsplitcard"    //需要切牌
	MsgTypeSplitCardStart   = "splitcardstart"   //切牌开始
	MsgTypeSplitCard        = "splitcard"        //切牌消息
	MsgType4KingScore       = "do4kingscore"     //4王换分
	MsgTypeAntiBrand        = "antiBrand"        //反牌消息
	MsgTypeEndAntiBrand     = "endAntiBrand"     //结束反牌消息
	MsgTypePlayJiFen        = "playjifenscore"   //用户积分
	MsgTypeChangeTableInRet = "changetableinret" //进入新桌
	MsgTypeRobSpring        = "qiangchun"        //抢春
	MsgTypeEndRobSpring     = "endqiangchun"     //结束抢春
	MsgTypeToolList         = "toollist"         //用户可以兑换的记牌器列表
	MsgTypeToolExchange     = "toolexchange"     //用户请求兑换记牌器，客户端直接请求到房间的
	MsgTypeToolExchangeHall = "toolexchangehall" //用户请求兑换记牌器，大厅转发的
	MsgTypeSendCardRecorder = "sendcardrecorder" //发送用户的记牌器数据
)

//升级游戏，其他和打拱共用
const (
	//游戏房间消息
	MsgTypeHaveExtCallTime = "extcalltime"   //额外的叫牌等待时间
	MsgTypeLastTurn        = "lastturn"      //上一轮牌数据请求
	MsgTypeConcealCard     = "concealcard"   //底牌消息
	MsgTypeCallCard        = "callcard"      //叫主消息
	MsgTypeTurnBalance     = "turnbalance"   //一轮结算
	MsgTypeThrowResult     = "throwresult"   //甩牌结果
	MsgTypeGamePlay        = "gameplay"      //开始出牌
	MsgTypeSendOrder       = "sendorder"     //发送玩家级数。
	MsgTypeCallScore       = "callscore"     //叫分消息。
	MsgTypeGiveUp          = "giveup"        //用户投降
	MsgTypeShowMainCards   = "showmaincards" //展示主牌和副牌数量
	MsgTypeMaiCard         = "maicard"       //买主消息
	MsgTypeMingQi          = "mingqi"        //明七消息
	MsgTypeAntiMain        = "antimain"      //反主消息
)

//拼三(大吉大利)
const (
	MsgTypeXiaZhu          = "xiazu"     //下注
	MsgTypeQiPai           = "qipai"     //弃牌
	MsgTypeKanPai          = "kanpai"    //看牌
	MsgTypeLiangPai        = "liangpai"  //亮牌
	MsgTypeBiPai           = "bipai"     //比牌
	MsgTypeLunShu          = "lunshu"    //轮数更新
	MsgTypeOpenCard        = "opencard"  //
	MsgTypeShowCard        = "showcard"  //系统亮牌
	MsgTypeGuZhuYiZi       = "guzhuyizi" //孤注一掷
	MsgTypeAuto            = "autop3"    //自动跟注
	MsgTypeEndAuto         = "endautop3"
	MsgTypeUserScoreUpdate = "userscore"      //更新玩家分数
	MsgTypeUserGoldUpdate  = "usergoldupdate" //更新玩家分数
)

//字牌,部分和打拱共用
const (
	MsgTypeTIANLONG   = "tianlong"  //天拢
	MsgTypeTIANHU     = "tianhu"    //天胡
	MsgTypeKAICHAOOUT = "kaicaoout" //开朝后出牌
	MsgTypeHUXI       = "huxi"      //发送胡息
	MsgTypeHandHUXI   = "handhuxi"  //发送手牌胡息
	MsgTypeNoOut      = "noout"     //不出牌消息
	MsgTypeHide       = "hide"      //隐藏牌权按钮
	MsgTypeZhuaNum    = "zhuanum"   //发送抓数
	MsgTypeTuoCard    = "tuocard"   //拖牌
	MsgTypeTingInfo   = "tinginfo"  //听牌
	MsgTypeTingTag    = "tingtag"   //听牌标签
)
