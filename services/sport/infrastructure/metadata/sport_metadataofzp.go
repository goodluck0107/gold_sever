package metadata

/* 回放相关结构体
 */
//出牌记录类型 字母代码不要重复
const (
	E_ZP_SendCard    = iota //SA 发牌
	E_ZP_OutCard            //OA 出牌
	E_ZP_Wik_Left           //LA 吃
	E_ZP_Wik_Center         //CA 吃
	E_ZP_Wik_Right          //RA 吃
	E_ZP_Wik_1X2D           //XA 吃，绞
	E_ZP_Wik_2X1D           //YA 吃，绞
	E_ZP_Wik_2710           //ZA 吃，2710
	E_ZP_Peng               //PA碰牌
	E_ZP_TianLong           //TA  天拢
	E_ZP_Gang               //GA 杠牌，明杠
	E_ZP_Hu                 //HA 胡
	E_ZP_HuangZhuang        //NA 荒
	E_ZP_Li_Xian            //LB 离线
	E_ZP_Jie_san            //JA 解散
	E_ZP_Wik_Hua            //HC 滑
	E_ZP_Tuo_Guan           //DA 托管
	E_ZP_Wik_Half           //HB 歪
	E_ZP_Shao               //PB 绍牌，碰牌
	//E_ZP_PIAO						//PB 选漂   荆州花牌 通城个子 冲突了？
	E_ZP_PIAO      //PC 选漂
	E_ZP_StartPIAO //PD 开始选漂
	E_ZP_XiaZhua   //XB 下抓
	E_ZP_TuoCard   //TB 拖牌
	//E_ZP_Wik_Ta                     //TB 踏   荆州花牌   冲突了？
	E_ZP_Wik_Ta   //TC 踏
	E_ZP_Wik_Jian //JB 捡牌
	E_ZP_NoOut    //NB 捏牌

	E_SendCardRight   //AA 发送牌权
	E_HandleCardRight //BA 处理牌权
	E_ReadyStatus     //ZB 处理牌权
	E_GameStart       //BB 开始发牌
	E_Hide            //BC 隐藏牌权
)

//Replay_Ext_Type  字母代码不要重复
const (
	E_ZP_Ext_NUL         = iota //
	E_ZP_Ext_HuXi               //EA 胡息
	E_ZP_Ext_HandHuXi           //EB 手牌胡息，包括组合牌区的胡息
	E_ZP_Ext_Provider           //EC 吃碰杠牌提供者
	E_ZP_Ext_ProvideCard        //ED 吃碰杠牌提供牌,通城个子有花色，需要这个标记
	E_ZP_Ext_PiaoScore          //EE 漂分
	E_ZP_Ext_HouShao            //EF 后绍
	//E_ZP_Ext_Jian		            //EF 捡牌标记   荆州花牌 通城个子 冲突了？
	E_ZP_Ext_Jian       //EQ 捡牌标记
	E_ZP_Ext_XiaZhuaCnt //EG 下抓次数
	E_ZP_Ext_WeaveHas   //EH 是否是操作吃牌区里的牌
	//E_ZP_Ext_Tong					//EH 统数更新    荆州花牌 冲突了？
	E_ZP_Ext_Tong        //EI 统数更新
	E_ZP_Ext_ZiMo        //EJ 自摸(没有弃的按钮)
	E_ZP_Ext_RemoveCards //EK 要删除的牌
	E_ZP_Ext_Bao3Zhang   //EL 报三张
	E_ZP_Ext_BiGua       //EM 必挂
	E_ZP_Ext_AutoAction  //EN 不需要牌权的自动操作
	E_ZP_Ext_TimaAt      //EP 操作时间
)
