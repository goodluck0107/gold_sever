package lib_win

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sync"
)

var (
	__mjDataTableIns__ *TableMgr = nil
	mjDataOnce         sync.Once
)

func GetOldDataTableInstance() (*TableMgr, error) {
	if __mjDataTableIns__ == nil {
		mjDataOnce.Do(func() {
			xlog.Logger().Debug("我就看看我出现了几次")
			__mjDataTableIns__ = &TableMgr{}
			__mjDataTableIns__.Init()
			err := __mjDataTableIns__.LoadTable()
			if err != nil {
				xlog.Logger().Errorln("[hu table]load error，please check At:", err)
			}
			err = __mjDataTableIns__.LoadFengTable()
			if err != nil {
				xlog.Logger().Errorln("[hu feng table]load error，please check At:", err)
			}
		})
	}
	return __mjDataTableIns__, nil
}

//这里只是胡牌
type TableMgr struct {
	m_tbl          [9]*Table
	m_eye_tbl      [9]*Table //乱将表
	m_eye258_tbl   [9]*Table //258将表
	m_feng_tbl     [9]*Table
	m_feng_eye_tbl [9]*Table
}

//func (this *TableMgr) GetB() {
//	syslog.Logger().Println(this.m_tbl[8])
//	syslog.Logger().Errorln("bbbbbssds:", this.m_tbl[5])
//}

func (this *TableMgr) Init() {
	// 不带将的表
	for i := 0; i < 9; i++ {
		this.m_tbl[i] = &Table{}
		this.m_tbl[i].init()
	}
	//带将表
	for i := 0; i < 9; i++ {
		this.m_eye_tbl[i] = &Table{}
		this.m_eye_tbl[i].init()
	}
	//20190107 苏大强 258将表
	for i := 0; i < 9; i++ {
		this.m_eye258_tbl[i] = &Table{}
		this.m_eye258_tbl[i].init()
	}
	//字牌不带将表
	for i := 0; i < 9; i++ {
		this.m_feng_tbl[i] = &Table{}
		this.m_feng_tbl[i].init()
	}
	//字牌带将表
	for i := 0; i < 9; i++ {
		this.m_feng_eye_tbl[i] = &Table{}
		this.m_feng_eye_tbl[i].init()
	}
}

//查表
func (this *TableMgr) getTable(gui_num int, eye byte, chi bool) *Table {
	//return common.GetMjDataTableInstance().GetTable(gui_num, eye, chi)
	var tbl *Table
	if chi {
		switch eye {
		case 1:
			//混将
			tbl = this.m_eye_tbl[gui_num]
		case 2:
			//258
			tbl = this.m_eye258_tbl[gui_num]
		default:
			tbl = this.m_tbl[gui_num]
		}
	} else {
		if eye > 0 {
			tbl = this.m_feng_eye_tbl[gui_num]
		} else {
			tbl = this.m_feng_tbl[gui_num]
		}
	}
	return tbl
}

//查表
func (this *TableMgr) GetNewTable(gui_num int, eye byte, chi bool) *Table {
	//return common.GetMjDataTableInstance().GetTable(gui_num, eye, chi)
	var tbl *Table
	if chi {
		switch eye {
		case 1:
			//混将
			tbl = this.m_eye_tbl[gui_num]
		case 2:
			//258
			tbl = this.m_eye258_tbl[gui_num]
		default:
			tbl = this.m_tbl[gui_num]
		}
	} else {
		if eye > 0 {
			tbl = this.m_feng_eye_tbl[gui_num]
		} else {
			tbl = this.m_feng_tbl[gui_num]
		}
	}
	return tbl
}

//感觉这两个没啥用处 有用，这个回去看一下
//func (this *TableMgr) Add(key int, gui_num int, eye byte, chi bool) {
//	tbl := this.getTable(gui_num, eye, chi)
//	tbl.add(key)
//}

//应承晃晃或独立算胡包的游戏 查表的接口
func (this *TableMgr) CheckTBl(key int, gui_num int, eye bool, eyemask byte, chi bool) bool {
	//拿表
	var tbl *Table = nil
	if eye {
		tbl = this.getTable(gui_num, eyemask, chi)
	} else {
		tbl = this.getTable(gui_num, 0, chi)
	}
	return tbl.check(key)
}

//这个是检查 eye 代表要查有眼的情况
func (this *TableMgr) check(key int, gui_num int, eye bool, eyemask byte, chi bool) bool {
	//拿表
	var tbl *Table = nil
	if eye {
		tbl = this.getTable(gui_num, eyemask, chi)
	} else {
		tbl = this.getTable(gui_num, 0, chi)
	}
	return tbl.check(key)
}

//加载胡牌规则
//20190107 追加258表
func (this *TableMgr) LoadTable() error {
	for i := 0; i < 9; i++ {
		// name := fmt.Sprintf("tbl/table_%d.tbl", i)
		name := fmt.Sprintf(g_tartablefile, i)
		err := this.m_tbl[i].load(name)
		if err != nil {
			return err
		}
	}
	//带将 混将
	for i := 0; i < 9; i++ {
		// name := fmt.Sprintf("tbl/eye_table_%d.tbl", i)
		name := fmt.Sprintf(g_tareye_table, i)
		err := this.m_eye_tbl[i].load(name)
		if err != nil {
			return err
		}
	}
	//258
	for i := 0; i < 9; i++ {
		// name := fmt.Sprintf("tbl/eye_table_%d.tbl", i)
		name := fmt.Sprintf(g_tareye258_table, i)
		err := this.m_eye258_tbl[i].load(name)
		if err != nil {
			return err
		}
	}
	return nil
}

//写回去？
func (this *TableMgr) DumpTable() {
	for i := 0; i < 9; i++ {
		// name := fmt.Sprintf("tbl/table_%d.tbl", i)
		name := fmt.Sprintf(g_tartablefile, i)
		this.m_tbl[i].dump(name)
	}

	for i := 0; i < 9; i++ {
		// name := fmt.Sprintf("tbl/eye_table_%d.tbl", i)
		name := fmt.Sprintf(g_tareye_table, i)
		this.m_eye_tbl[i].dump(name)
	}

}

func (this *TableMgr) LoadFengTable() error {
	//赖子个数将表
	for i := 0; i < 9; i++ {
		name := fmt.Sprintf(g_tarfeng_table, i)
		// name := fmt.Sprintf("tbl/feng_table_%d.tbl", i)
		err := this.m_feng_tbl[i].load(name)
		if err != nil {
			return err
		}
	}
	//字牌带将表
	for i := 0; i < 9; i++ {
		// name := fmt.Sprintf("tbl/feng_eye_table_%d.tbl", i)
		name := fmt.Sprintf(g_tarfeng_eye_table, i)
		err := this.m_feng_eye_tbl[i].load(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *TableMgr) DumpFengTable() {
	for i := 0; i < 9; i++ {
		name := fmt.Sprintf(g_tarfeng_table, i)
		// name := fmt.Sprintf("tbl/feng_table_%d.tbl", i)
		this.m_feng_tbl[i].dump(name)
	}

	for i := 0; i < 9; i++ {
		// name := fmt.Sprintf("tbl/feng_eye_table_%d.tbl", i)
		name := fmt.Sprintf(g_tarfeng_eye_table, i)
		this.m_feng_eye_tbl[i].dump(name)
	}
}

// 应城晃晃 执行这个 load
func (_this *TableMgr) LoadAllTable() {
	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl/eye258_table_%d.tbl", i)
		_this.m_eye258_tbl[i].load(name)
	}

	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl/table_%d.tbl", i)
		_this.m_tbl[i].load(name)

	}

	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl/eye_table_%d.tbl", i)
		_this.m_eye_tbl[i].load(name)
	}

	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl/feng_table_%d.tbl", i)
		_this.m_feng_tbl[i].load(name)
	}

	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl/feng_eye_table_%d.tbl", i)
		_this.m_feng_eye_tbl[i].load(name)
	}
}
