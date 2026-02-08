package card_mgr

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
)

//考虑通用2个赖子的情况
//这个结构中的num统计是要根据规则来的
type GodCardInfo struct {
	ID  byte //牌值
	Num byte //这个值是手牌中真正god牌的数据 因为别人打的赖子牌要还原
}

/*
目前我知道的只有赖子能万用，别的还不知道，但是会根据来源区别是不是能万用，所以设定一个mask
判断的牌几种来源：
抢杠 不知道赖子能不能杠，但是可能出现在杠位上
杠后 （别人杠后 打的热铳，和自己杠起的杠上花）
别人的
自己的
设定万用，必然为1
目前崇阳放放 11101
真麻烦，修改word
修改为int ，上8是别人的，下8是自己的
*/
//20190119 这个结构可能要大改一下
type CheckGodOrg struct {
	GodCardInfo []*GodCardInfo
	GodNum      byte //未处理的时候鬼牌计数
	//判胡后要设置这个数的，因为判胡的时候只算补牌数，对于成对和成克的赖子并没统计，追加参考 废弃，要算番，就要有选择
	NeedGodNum byte //3n+2 状态下，需要的godnum  各番型会独立算，因为这样的情况  软清一色 硬将将胡
}

//可能有的游戏就没有赖子，那么这个地方mask就可能为0
func NewGodCardOrg(mask uint) *GodCardOrg {
	return &GodCardOrg{
		checkMask: mask,
		guicards:  make([]byte, 0),
		// GodCardInfo: make([]*GodCardInfo, 0),
		// GodNum:      0,
		// NeedGodNum:  0,
	}
}
func NewCheckGodOrg() *CheckGodOrg {
	return &CheckGodOrg{
		GodCardInfo: make([]*GodCardInfo, 0),
		GodNum:      0,
		NeedGodNum:  0,
	}
}
func NewGodCardInfo(id byte, num byte) (*GodCardInfo, error) {
	if !mahlib2.IsMahjongCard(id) || num > 4 {
		return nil, errors.New(fmt.Sprintf("创建鬼牌信息失败：（%x）越界（37）或者牌数越界（%d）", id, num))
	}
	return &GodCardInfo{
		ID:  id,
		Num: num,
	}, nil
}

func (this *GodCardOrg) Setguicard(guicard byte) error {
	if this.guicards != nil {
		//设定上限
		if len(this.guicards) >= MAX_GUI_NUM /*|| guindex == 0*/ {
			return errors.New(fmt.Sprintf("赖子最多2个,目前赖子（%v）", this.guicards))
		}
		//重置
		if guicard == static.INVALID_BYTE {
			this.guicards = []byte{}
			return nil
		}
		if !mahlib2.IsMahjongCard(guicard) {
			return errors.New(fmt.Sprintf("无效赖子牌(%x)", guicard))
		}
		this.guicards = append(this.guicards, guicard)
	}
	return nil
}

//得到癞子牌
func (this *GodCardOrg) GetGuiInfo() []byte {
	return this.guicards
}

//坑爹的地方没有word，这个要测一下，不然就要加个单元专门处理
func (this *GodCardOrg) CanBeGod(origin byte) bool {
	//20190117目前来源就4个，分为别人（桌面，倒牌），自己（手牌，发牌），如果是别人的，移动一下
	if origin&(ORIGIN_NOM|ORIGIN_HAND) > 0 {
		//只会有一种情况
		if this.checkMask&uint(origin&ORIGIN_NOM) > 0 {
			// 在发牌区
			return true
		}
		if this.checkMask&uint(origin&ORIGIN_HAND) > 0 {
			// 在手牌区
			return true
		}
	} else {
		//这个是别人的
		// if this.checkMask&uint((origin&ORIGIN_WARN)<<8) > 0 {
		if this.checkMask&uint(origin&ORIGIN_WARN) > 0 {
			//倒牌 （抢杠胡。。。我感觉多余的，不排除可以蓄杠赖子。。。）
			return true
		}
		// if this.checkMask&uint((origin&ORIGIN_TABILE)<<8) > 0 {
		if this.checkMask&uint(origin&ORIGIN_TABILE) > 0 {
			//桌面
			return true
		}
	}
	return false
}
func GetGuiNumAndEyeClorer(hu_struct *[4][2]byte) (guinum byte, eyecolor []byte) {
	if hu_struct == nil {
		return static.INVALID_BYTE, nil
	}
	for i := 0; i < 4; i++ {
		guinum = guinum + hu_struct[i][0]
		if hu_struct[i][1] > 0 {
			eyecolor = append(eyecolor, byte(i))
		}
	}
	return
}

//20190306 设置needGuiNum 这个是只是3n+2 设置的
func (this *CheckGodOrg) SetNeedGuiNum(chiHuKind uint16, huStruct *[4][2]byte) {
	if chiHuKind&static.CHK_PING_HU_MAGIC > 0 && huStruct != nil {
		_, this.NeedGodNum = RestoreNeedGui(huStruct)
	}
}
