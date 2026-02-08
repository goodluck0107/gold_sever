package components

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	scoringlib2 "github.com/open-source/game/chess.git/services/sport/backboard/lib_scoring"
	fanlib2 "github.com/open-source/game/chess.git/services/sport/backboard/lib_scoring/fanlib"
	hulib2 "github.com/open-source/game/chess.git/services/sport/backboard/lib_win"
	"io/ioutil"
)

//20181209 胡表生成一个单例生成一个单例
var huTable *hulib2.TableMgr = nil

//20190106 查番表生成
var G_fanTable *fanlib2.PublicFanManager = nil

func GetHuTable() (*hulib2.TableMgr, error) {
	//return nil, nil
	if huTable == nil {
		//huTable = &lib_win.TableMgr{}
		//		//huTable.Init()
		//err := huTable.LoadTable()
		//if err != nil {
		//	syslog.Logger().Errorln("[hu table]load error，please check At:", err)
		//	//return nil, err
		//}
		//err = huTable.LoadFengTable()
		//if err != nil {
		//	syslog.Logger().Errorln("[hu feng table]load error，please check At:", err)
		//	//return nil, err
		//}

		huTable, _ = hulib2.GetOldDataTableInstance()
	}
	return huTable, nil
}

//20190106
func GetFanTable() (newfan *fanlib2.PublicFanManager, err error) {
	if G_fanTable == nil {
		newfan, err = fanlib2.NewPublicScoringManager(nil)
		if err != nil {
			xlog.Logger().Errorln("GetFanTable err :", err.Error())
			return nil, err
		}
	}
	return
}

func init() {
	//读取番表文件夹来干活
	var err error
	if G_fanTable == nil {
		G_fanTable, err = GetFanTable()
	}
	if err != nil {
		xlog.Logger().Errorln("fanRuleFiles err :", err.Error())
		return
	}
	var fanRuleFiles []string
	fanRuleFiles, err = static.GetAllFiles("./fanRule")
	if err != nil {
		xlog.Logger().Errorln("fanRuleFiles err :", err.Error())
		return
	}
	xlog.Logger().Debug(fmt.Sprintf("%v", fanRuleFiles))
	for _, filename := range fanRuleFiles {
		config, err := ioutil.ReadFile(filename)
		if err != nil {
			xlog.Logger().Errorln(fmt.Sprintf("读取查番规则文件（%s） err : %v", filename, err))
			panic(fmt.Sprintf("读取查番规则文件（%s） err : %v", filename, err))
			continue
		}
		check := &scoringlib2.GameKindRule{}
		err = json.Unmarshal(config, check)
		if err != nil {
			xlog.Logger().Debug(fmt.Sprintf("番规则文件（%s）格式有误: %v", filename, err))
			panic(fmt.Sprintf("番规则文件（%s）格式有误: %v", filename, err))
			continue
		}
		//syslog.Logger().Debug(fmt.Sprintf("番规则文件（%s）: %v", filename, check))
		//if len(check.ScoringMask) != 0 {
		//	for _, v := range check.ScoringMask {
		//		syslog.Logger().Debug(fmt.Sprintf("番号（%d）番名（%s）番数（%d）mask（%b）", v.FanInfo.FanID, v.FanInfo.Name, v.FanInfo.FanShu, v.HuMask))
		//		if v.ModifyInfo != nil {
		//			syslog.Logger().Debug(fmt.Sprintf("modify 番号（%d）", v.ModifyInfo.ModifyID))
		//		}
		//	}
		//}
		NewBalanceRule, err := fanlib2.NewBalanceRule(check, fanlib2.G_ScoringManager)
		if err != nil {
			fmt.Println(err)
			return
		}
		//加入到服务器列表中去
		G_fanTable.SetNewRule(NewBalanceRule)
	}
}
