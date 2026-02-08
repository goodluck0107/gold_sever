package infrastructure

//游戏入口
//kindid
const (
	KIND_ID_CY         = 897 //崇阳麻将
	KIND_ID_CYFF       = 896 //崇阳放放
	KIND_ID_JM         = 981 //荆门晃晃
	KIND_ID_Test       = 123 //崇阳测试专供
	KIND_ID_Test1      = 124 //崇阳放放测试专供
	KIND_ID_JMSK       = 963 //chess麻将荆门双开
	KIND_ID_HZLZG      = 984 //chess麻将红中赖子杠
	KIND_ID_JZJL       = 599 //新江陵玩法
	KIND_ID_XTMJ_LH    = 598 //仙桃麻将赖晃
	KIND_ID_XTMJ_YLDD  = 597 //仙桃麻将一赖到底
	KIND_ID_XTMJ_YLKZC = 596 //仙桃麻将一赖可捉统
	KIND_ID_JZMJ_YJLY  = 595 //荆州麻将一脚赖油
	KIND_ID_HHHJ_YH    = 594 //晃晃合集硬晃
	KIND_ID_HHHJ_LH    = 593 //晃晃合计赖晃
	KIND_ID_HHHJ_YLKZC = 592 //晃晃合集一赖可捉统
	KIND_ID_CBMJ_DD    = 591 //赤壁麻将剁刀
	KIND_ID_CBMJ_YH    = 590 //赤壁麻将硬晃
	KIND_ID_CBMJ_HZ    = 589 //赤壁麻将红中玩法
	KIND_ID_TMMJ_YLKZC = 588 //天门麻将一赖可捉统
	KIND_ID_TMMJ_YH    = 584 //天门麻将硬晃
	KIND_ID_TMMJ_YLDD  = 583 //天门麻将一赖到底
	KIND_ID_TMMJ_PLZ   = 582 //天门麻将痞癞子
	KIND_ID_QJMJ_QJHH  = 578 //潜江麻将潜江晃晃
	KIND_ID_QJMJ_JQDT  = 577 //潜江麻将经典单挑
	KIND_ID_QJMJ_QJHZ  = 576 //潜江麻将潜江红中
	KIND_ID_HM_WHMJ    = 889 //斑马汉麻武汉麻将
	KIND_ID_HM_WHHH    = 587 //斑马汉麻武汉晃晃
	KIND_ID_JY_HZLZG   = 581 //嘉鱼红中癞子杠
	KIND_ID_JY_HH      = 580 //嘉鱼晃晃
	KIND_ID_JY_YQ      = 579 //嘉鱼硬巧
	KIND_ID_TC_HZLZG   = 572 //通城红中赖子杠
	KIND_ID_TS_TSMJ    = 574 //通山麻将
	KIND_ID_TS_TSHH    = 575 //通山晃晃
	KIND_ID_TS_TSTDH   = 476 //通山推倒胡
	KIND_ID_EZ_510K    = 486 //鄂州510K
	KIND_ID_WX_GF      = 448 //武穴隔反
	//20190403 洪湖3款麻将 苏大强
	KIND_HH_YLDD          = 570
	KIND_HH_LH            = 569
	KIND_HH_YH            = 568
	KIND_JZ_TDH           = 564 //焦作推倒胡
	KIND_JZ_YBDD          = 565 //焦作硬报到底
	KIND_ID_XX_MJ         = 567
	KIND_ID_XX_HJ         = 566
	KIND_ID_JM_ZXTM       = 562 //荆门钟祥推磨
	KIND_ID_CZ_TDH        = 561 //安徽滁州推倒胡
	KIND_ID_CZ_57F        = 560 //滁州57番
	KIND_ID_CZ_3F         = 559 //滁州老3番
	KIND_ID_DG_TS_HY      = 558 //通山打拱4人好友
	KIND_ID_DG_CB_HY      = 481 //赤壁打拱4人好友
	KIND_ID_XN_XNHH       = 480 //咸宁晃晃
	KIND_ID_XN_HZLZG      = 477 //咸宁红中赖子杠
	KIND_ID_DG_XTQF_HY    = 479 //仙桃千分4人好友
	KIND_ID_HS_HSHH       = 467 //黄石晃晃
	KIND_ID_DY_DYHH       = 466 //大冶晃晃
	KIND_ID_DAYEKAIKOUFAN = 471 //大冶开口番麻将
	KIND_ID_DG_XN_HY      = 478 //咸宁打拱4人好友
	KIND_ID_LT_HZLZG      = 473 //罗田红中赖子杠
	KIND_ID_DG_CB3_HY     = 475 //赤壁打拱3人好友
	KIND_ID_DG_CB2_HY     = 474 //赤壁打拱2人好友
	KIND_ID_DG_DY2_HY     = 470 //大冶打拱2人好友
	KIND_ID_DG_DY3_HY     = 469 //大冶打拱3人好友
	KIND_ID_DG_DY_HY      = 468 //大冶打拱4人好友
	KIND_ID_JL_JLMJ       = 472 //监利麻将
	KIND_ID_HS_YXMJ       = 463 //阳新麻将

	KIND_ID_QC_QCDG_HY  = 465 //蕲春打拱好友房
	KIND_ID_JL_JLKJ     = 464 //监利开机
	KIND_ID_ZP_DYZP     = 462 //大冶字牌
	KIND_ID_NS_NSMJ     = 460 //恩施麻将
	KIND_ID_DG_CY_HY    = 461 //崇阳打滚
	KIND_ID_MJ_TUANFENG = 458 //团风麻将

	KIND_ID_MJ_FY_YZ_GF = 457 //阜阳颍州杠番
	KIND_ID_MJ_WXMJ     = 453 //武穴麻将
	KIND_ID_PK_510K_WX  = 398 //武穴510K

	KIND_ID_MJ_AH_CZ_LA_L3  = 456  //安徽滁州来安老三番
	KIND_ID_DG_YX           = 454  //阳新打拱好友房
	KIND_ID_MJ_JSMJ         = 451  //京山麻将
	KIND_ID_DG_PDK          = 452  //跑得快3人好友房
	KIND_ID_DG_YX3          = 450  //阳新打拱3人好友房
	KIND_ID_HC_CXZ          = 449  //汉川麻将搓虾子
	KIND_ID_HC_LAIHUANG     = 434  //汉川赖晃
	KIND_ID_MJ_YCKWX        = 445  //应城卡五星
	KIND_ID_HG_HZMJ         = 446  //黄冈红中麻将
	KIND_ID_ZP_HCSJ         = 447  //汉川数斤3人好友房
	KIND_ID_SJ_HQJ          = 443  //崇阳画圈脚
	KIND_ID_YC_HH           = 444  //应城晃晃
	KIND_ID_JZ_CXZ          = 442  //荆州搓虾子
	KIND_ID_HZ_HH           = 441  //黄州晃晃
	KIND_ID_ZP_YXZP         = 440  //阳新字牌
	KIND_ID_DG_PDK15        = 439  //跑得快15张
	KIND_ID_ZP_TCGZ         = 459  //通城个子
	KIND_ID_ZP_YCSDR        = 437  //应城上大人
	KIND_ID_ZJK_PDK         = 433  //张家口跑得快
	KIND_ID_CZ_PDK          = 432  //滁州跑得快
	KIND_ID_XT_PDK          = 431  //仙桃跑得快
	KIND_ID_MJ_HC_WHMJ      = 429  //汉川武汉麻将
	KIND_ID_DY_510K         = 428  //大治510k
	KIND_ID_JH_PDK          = 423  //江汉跑得快
	KIND_ID_DG_JY           = 418  //嘉鱼打滚
	KIND_ID_XS_MJ           = 438  //浠水麻将
	KIND_ID_ZP_JZHP         = 436  //荆州花牌
	KIND_ID_MJ_ZJKFHE       = 435  //张家口番混儿
	KIND_ID_MJ_PHDL         = 430  //张家口平胡多癞
	KIND_ID_MJ_XGKWX        = 427  //孝感卡五星
	KIND_ID_MJ_XLCH         = 424  //血流成河
	KIND_ID_MJ_HMMJ         = 421  //黄梅麻将
	KIND_ID_MJ_ZJKTDH       = 422  //张家口推倒胡
	KIND_ID_MJ_JZHZG        = 419  //荆州红中杠
	KIND_ID_MJ_GDTDH        = 426  //广东推倒胡
	KIND_ID_MJ_JLHZLZG      = 420  //监利红中癞子杠
	KIND_ID_ZP_ESSH         = 425  //恩施绍胡
	KIND_ID_CR_MJ           = 409  //铳儿麻将
	KIND_ID_NEW_WHKAIKOUFAN = 411  //新武汉麻将开口番
	KIND_ID_NEW_WHKOUKOUFAN = 410  //新武汉麻将口口番
	KIND_ID_MJ_ZZ           = 413  //转转麻将
	KIND_ID_MJ_ESHFMLZ      = 412  //恩施鹤峰焖癞子
	KIND_ID_MJ_ESYING       = 417  //恩施硬麻将
	KIND_ID_ZP_HFBH         = 408  //鹤峰百胡
	KIND_ID_MJ_SHAH         = 400  //石首捱晃
	KIND_ID_DG_CY2_HY       = 406  //崇阳打滚2人玩
	KIND_ID_DG_CY3_HY       = 405  //崇阳打滚3人玩
	KIND_ID_HONGHU_510K     = 399  //洪湖510K
	KIND_ID_DG_DDZ          = 407  //癞子斗地主
	KIND_ID_XTQFBD          = 394  //仙桃千分必打
	KIND_ID_SS_510K         = 395  //石首510K
	KIND_ID_TONGCHENGDAGUN  = 391  //通城打滚
	KIND_ID_SJ_TCBG         = 393  //通城巴锅
	KIND_ID_ES_PDK          = 390  //恩施跑得快
	KIND_ID_MC_MJ           = 381  //麻城麻将
	KIND_ID_HS_510k         = 385  //黄石510k
	KIND_ID_TY_GDY          = 380  //通用干瞪眼
	KIND_ID_DG_YX2          = 379  //阳新二人拱
	KIND_ID_QCMJ            = 374  //蕲春麻将
	KIND_ID_QJ_SHZ          = 376  //潜江说胡子
	KIND_ID_QJ_510KBD       = 375  //潜江510K必打
	KIND_ID_ZP_LCSDR        = 377  //利川上大人
	KIND_ID_LC_DDZ          = 373  //利川斗地主
	NEW_KIND_ID_MJ_XGKWX    = 368  //新孝感卡五星
	KIND_ID_MJ_XYKWX        = 371  //襄阳卡五星
	KIND_ID_MJ_SYKWX        = 370  //十堰卡五星
	KIND_ID_MJ_SZKWX        = 369  //随州卡五星
	KIND_ID_JZ_PDK          = 367  //决战跑得快
	KIND_ID_MJ_QBMJ         = 366  //蕲北麻将
	KIND_ID_XS_DQ           = 365  //浠水打七
	KIND_ID_MJ_YC_XLCH      = 362  //宜昌血流成河
	KIND_ID_SDR_YC          = 363  //宜昌上大人
	KIND_ID_ZP_YCHP         = 1000 //宜昌花牌
	KIND_ID_XN_PDK          = 1021 //咸宁跑得快
	KIND_ID_XN_PDK15        = 1022 //咸宁跑得快15张
)

//纸牌游戏的kindid

const (
	KIND_ID_DG_QC4 = 996 //蕲春打拱4人金币
	KIND_ID_P3     = 585 //拼三（大吉大利）
	KIND_ID_DG_QC3 = 586 //蕲春打拱3人金币
	//----------麻将金币场游戏-----
	KIND_ID_GOLD_XTMJ_LH    = 499 //仙桃麻将赖晃    金币
	KIND_ID_GOLD_XTMJ_YLDD  = 498 //仙桃麻将一赖到底  金币
	KIND_ID_GOLD_XTMJ_YLKZC = 497 //仙桃麻将一赖可捉统 金币
	KIND_ID_GOLD_JZMJ_YJLY  = 496 //荆州麻将一脚赖油  金币
	KIND_ID_GOLD_HHHJ_YH    = 495 //晃晃合集硬晃    金币
	KIND_ID_GOLD_HHHJ_LH    = 494 //晃晃合计赖晃    金币
	KIND_ID_GOLD_HHHJ_YLKZC = 493 //晃晃合集一赖可捉统 金币
	KIND_ID_GOLD_CBMJ_DD    = 492 //赤壁麻将剁刀    金币
	KIND_ID_GOLD_CBMJ_YH    = 491 //赤壁麻将硬晃    金币
	KIND_ID_GOLD_CBMJ_HZ    = 480 //赤壁麻将红中玩法  金币
	KIND_ID_GOLD_TMMJ_YLKZC = 489 //天门麻将一赖可捉统 金币
	//----------麻将金币场游戏------
	KIND_ID_DG_DY4  = 488 //大冶打拱4人金币
	KIND_ID_DG_DY3  = 487 //大冶打拱3人金币
	KIND_ID_TC_MJ   = 571 //通城麻将
	KIND_ID_EZ_HH   = 573 //鄂州晃晃
	KIND_ID_DY_RAR  = 415 //大冶肉挨肉
	KIND_ID_ES_GDY  = 397 //恩施干瞪眼
	KIND_ID_TM_510K = 392 //天门510K
	KIND_ID_JS_CH   = 386 //建始楚胡
	KIND_ID_HG_DBD  = 378 //黄冈斗板凳
)

func HF_IsHHMG(kindId int) bool {
	/*洪湖
	  一赖到底 麻将 570
	  洪湖赖晃 麻将 569
	  洪湖硬晃 麻将 568
	*/
	return (kindId >= KIND_HH_YH && kindId <= KIND_HH_YLDD)
}

//赤壁晃晃
func HF_IsHHCHiBi(kindId int) bool {
	return kindId == KIND_ID_CBMJ_DD || kindId == KIND_ID_CBMJ_YH || kindId == KIND_ID_CBMJ_HZ
}

//仙桃晃晃
func HF_IsHHXianTao(kindId int) bool {
	return kindId == KIND_ID_XTMJ_LH || kindId == KIND_ID_XTMJ_YLDD || kindId == KIND_ID_XTMJ_YLKZC
}

// 判断是不是晃晃游戏
func HF_IsHHGame(kindId int) bool {
	return (kindId >= KIND_ID_TMMJ_YLKZC && kindId <= KIND_ID_XTMJ_LH) || (kindId >= KIND_ID_TMMJ_PLZ && kindId <= KIND_ID_TMMJ_YH) || (kindId >= KIND_ID_QJMJ_QJHZ && kindId <= KIND_ID_QJMJ_QJHH)
}

// 判断是不是晃晃金币游戏
func HF_IsGoldHHGame(kindId int) bool {
	return kindId >= KIND_ID_GOLD_TMMJ_YLKZC && kindId <= KIND_ID_GOLD_XTMJ_LH
}
