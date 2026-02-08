package eve

//申请解散房间状态
const (
	AGREE           = 1 //同意
	DISAGREE        = 2 //不同意
	DISMISS_CREATOR = 3 //解散房间发起者
	WATING          = 4 //等等中
	DISSMISS_MAX    = 5
)

//出牌记录类型
const (
	E_SendCard               = iota //发牌
	E_OutCard                       //出牌
	E_OutCard_TG                    //托管出牌
	E_Wik_Left                      //吃
	E_Wik_Center                    //吃
	E_Wik_Right                     //吃
	E_Peng                          //碰牌
	E_Gang                          //杠牌，明杠
	E_Gang_PiziGand                 //杠牌，皮子杠
	E_Gang_HongZhongGand            //杠牌，红中杠
	E_Gang_FaCaiGand                //杠牌，发财杠
	E_Gang_LaiziGand                //杠牌，癞子杠
	E_Gang_XuGand                   //蓄杠，即回头杠
	E_Gang_AnGang                   //暗杠
	E_Gang_SmallChaoTianGand        //小朝天杠（手上两个皮子 其他人打了一个皮子算小朝天杠）
	E_Gang_ChaoTianGand             //朝天杠
	E_Qiang                         //抢暗杠
	E_Hu                            //胡
	E_HuangZhuang                   //荒
	E_Bird                          //抓鸟
	E_Li_Xian                       //离线
	E_Jie_san                       //解散
	E_Pao                           //跑
	E_SendCardRight                 //发送牌权
	E_HandleCardRight               //处理牌权
	E_Baoqing                       //记录报清
	E_Baojiang                      //记录报将
	E_Baofeng                       //记录报风
	E_Baojing                       //记录报警
	E_Baoqi                         //记录报弃
	E_BaoTing                       //记录报听
	E_GameScore                     //玩家剩余分
	E_BuHua                         //补花
	E_BuHua_TG                      //托管补花
	E_Liang                         //亮牌
	E_Change_Pizhi                  //换痞子
	E_OutCard_TG_PiZi               //托管出皮子
	E_OutCard_TG_Magic              //托管出赖子

	E_ExchageThreeCard //换三张
	E_ExchageThreeEnd  //换三张结束

	E_K5x_LiangPai    //卡五星亮牌
	E_LastCard        //牌堆最后一张牌
	E_HaiDiLao        //开始海底捞
	E_HuanPai_Type    //换牌方式
	E_HardHu          //硬胡软胡
	E_FanLaiZi        //翻癞子
	E_TuoGuan         //托管
	E_JZSendCardRight //发送牌权(荆州搓虾子)
	E_ReadyStatus     //准备
	E_GameStart       //游戏小局开始
)

//承包类型
const (
	ENChengBaoNull = iota //没有承包
	ENQuanQiuRen          //全求人承包
	ENQingYiSe            //清一色承包
	ENQiangGang           //抢杠承包
	ENXiaYu               //下雨承包
	ENFengYiSe            //风一色承包
	ENJiangYiSe           //将一色承包
)

const (
	GONG_ALL    = 0xff      //全杠支持
	GONG_MING   = 1 << iota //明杠
	GONG_XU                 //蓄杠
	GONG_HUITOU             //回头杠
	GONG_AN                 //暗杠
	GONG_PIZI
	GONG_LAIZI
	GONG_HONGZHONG
	GONG_FACAI
)

const (
	Gang_Type_HongZhongGang = 1 //1红中杠勾选,0红中杠未勾选
)

const (
	Gang_Type_HongZhongFaCai = 1
	Gang_Type_QiPiSiLai      = 2
	Gang_Type_ShiYiPiSiLai   = 3
)

const (
	KaiKouFan_FengDingDianShu = 400
	KouKouFan_FengDingDianShu = 800
)
const (
	JingDing_Score      = 10
	YangGuangDing_Score = 20
	SanYangDing_Score   = 24
)

const (
	XiaYu_XiaoYu = 1
	XiaYu_DaYu   = 2
)
