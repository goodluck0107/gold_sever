package static

type ChuZouHuType struct {
	Type_yingque    byte //硬缺 3
	Type_ruangque   byte //软缺 2
	Type_duiduihu   byte //对对胡 3
	Type_duan19     byte //断19 3
	Type_gangkai    byte //杠开 3
	Type_haidi      byte //海底捞 3
	Type_yitiaolong byte //一条龙 3
	Type_jiemei     byte //姊妹铺 3
	Type_jiemei2    byte //双姊妹铺 30
	Type_sankan     byte //三坎 3
	Type_sikan      byte //四坎 30
	Type_qiyise     byte //清一色 30
	Type_hunyise    byte //混一色 3
	Type_ziyise     byte //子一色 60
	Type_quan19     byte //全幺九 30
	Type_mengqin    byte //门清 1
	Type_siyunzi    byte //四云子 +1
	Type_kanjiang   byte //坎将
	Type_sicha      byte //四叉
	Type_quanqiuren byte //z全球人
	Type_sandajiang byte //三大将
	Type_sijifeng   byte //四季风
	Type_pinghu     byte //平胡
	Type_qianggang  byte //抢杠
}

//游戏可胡牌类型设置
func (self *ChuZouHuType) Init() {
	self.Type_yingque = 0    //硬缺 3
	self.Type_ruangque = 0   //软缺 2
	self.Type_duiduihu = 0   //对对胡 3
	self.Type_duan19 = 0     //断19 3
	self.Type_gangkai = 0    //杠开 3
	self.Type_haidi = 0      //海底捞 3
	self.Type_yitiaolong = 0 //一条龙 3
	self.Type_jiemei = 0     //姊妹铺 3
	self.Type_jiemei2 = 0    //双姊妹铺 30
	self.Type_sankan = 0     //三坎 3
	self.Type_sikan = 0      //四坎 30
	self.Type_qiyise = 0     //清一色 30
	self.Type_hunyise = 0    //混一色 3
	self.Type_ziyise = 0     //子一色 60
	self.Type_quan19 = 0     //全幺九 30
	self.Type_mengqin = 0    //门清 1
	self.Type_siyunzi = 0    //四云子 +1
	self.Type_kanjiang = 0   //坎将
	self.Type_sicha = 0      //四叉
	self.Type_quanqiuren = 0 //z全球人
	self.Type_sandajiang = 0 //三大将
	self.Type_sijifeng = 0   //四季风
	self.Type_pinghu = 0     //平胡
	self.Type_qianggang = 0  //抢杠
}

func (self *ChuZouHuType) GetFan() int {
	return int(self.Type_yingque + self.Type_ruangque + self.Type_duiduihu + self.Type_duan19 +
		self.Type_gangkai + self.Type_haidi + self.Type_yitiaolong + self.Type_jiemei + self.Type_jiemei2 +
		self.Type_sankan + self.Type_sikan + self.Type_qiyise + self.Type_hunyise + self.Type_ziyise + self.Type_quan19 +
		self.Type_mengqin + self.Type_siyunzi + self.Type_kanjiang + self.Type_sicha + self.Type_quanqiuren + self.Type_sandajiang +
		self.Type_sijifeng + self.Type_pinghu)
}

func (self *ChuZouHuType) Copy(_type *ChuZouHuType) {
	self.Type_yingque = _type.Type_yingque       //硬缺 3
	self.Type_ruangque = _type.Type_ruangque     //软缺 2
	self.Type_duiduihu = _type.Type_duiduihu     //对对胡 3
	self.Type_duan19 = _type.Type_duan19         //断19 3
	self.Type_gangkai = _type.Type_gangkai       //杠开 3
	self.Type_haidi = _type.Type_haidi           //海底捞 3
	self.Type_yitiaolong = _type.Type_yitiaolong //一条龙 3
	self.Type_jiemei = _type.Type_jiemei         //姊妹铺 3
	self.Type_jiemei2 = _type.Type_jiemei2       //双姊妹铺 30
	self.Type_sankan = _type.Type_sankan         //三坎 3
	self.Type_sikan = _type.Type_sikan           //四坎 30
	self.Type_qiyise = _type.Type_qiyise         //清一色 30
	self.Type_hunyise = _type.Type_hunyise       //混一色 3
	self.Type_ziyise = _type.Type_ziyise         //子一色 60
	self.Type_quan19 = _type.Type_quan19         //全幺九 30
	self.Type_mengqin = _type.Type_mengqin       //门清 1
	self.Type_siyunzi = _type.Type_siyunzi       //四云子 +1
	self.Type_kanjiang = _type.Type_kanjiang     //坎将
	self.Type_sicha = _type.Type_sicha           //四叉
	self.Type_quanqiuren = _type.Type_quanqiuren //z全球人
	self.Type_sandajiang = _type.Type_sandajiang //三大将
	self.Type_sijifeng = _type.Type_sijifeng     //四季风
	self.Type_pinghu = _type.Type_pinghu         //平胡
	self.Type_qianggang = _type.Type_qianggang   //抢杠
}
