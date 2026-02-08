package backboard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

func init() {
	//_ = GetMjDataTableInstance()
}

//初始化本地表
type (
	mjDataTable struct {
		mTbl          [8]MjLoadTable
		mEyeTbl       [8]MjLoadTable
		mEye258Tbl    [8]MjLoadTable
		mFlowerTbl    [8]MjLoadTable
		mFlowerEyeTbl [8]MjLoadTable
	}
	MjLoadTable struct {
		tbl map[int]int
	}
)

var (
	__mjDataTableIns__ *mjDataTable = nil
	mjDataOnce         sync.Once
)

func GetMjDataTableInstance() *mjDataTable {
	if __mjDataTableIns__ == nil {
		mjDataOnce.Do(func() {
			__mjDataTableIns__ = &mjDataTable{}
			__mjDataTableIns__.initLoad()
		})
	}
	return __mjDataTableIns__
}

func (_this *mjDataTable) initLoad() {
	//_this.LoadTable()
	//__mjDataTableIns__,_ = modules.GetHuTable()

}

func (_this *mjDataTable) GetTable(guiNum int, eye byte, chi bool) *MjLoadTable {
	var tbl *MjLoadTable = nil
	if chi {
		switch eye {
		case 1:
			//混将
			tbl = &_this.mEyeTbl[guiNum]
		case 2:
			//258
			tbl = &_this.mEye258Tbl[guiNum]
		default:
			tbl = &_this.mTbl[guiNum]
		}
	} else {
		if eye > 0 {
			tbl = &_this.mFlowerEyeTbl[guiNum]
		} else {
			tbl = &_this.mFlowerTbl[guiNum]
		}
	}
	return tbl
}

func (_this *mjDataTable) Check(key int, guiNum int, eye byte, chi bool) bool {

	tbl := _this.GetTable(guiNum, eye, chi)
	if tbl == nil {
		return false
	}
	return tbl.Check(key)
}

func (_this *mjDataTable) LoadTable() {
	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl_ini/TBL_258/eye258_table_%d.ini", i)
		_this.mEye258Tbl[i].load(name)
	}

	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl_ini/TBL_ALL/table_%d.ini", i)
		_this.mTbl[i].load(name)

	}

	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl_ini/TBL_ALL/eye_table_%d.ini", i)
		_this.mEyeTbl[i].load(name)
	}

	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl_ini/TBL_ALL/feng_table_%d.ini", i)
		_this.mFlowerTbl[i].load(name)
	}

	for i := 0; i <= 8; i++ {
		name := fmt.Sprintf("./tbl_ini/TBL_ALL/feng_eye_table_%d.ini", i)
		_this.mFlowerEyeTbl[i].load(name)
	}
}

func (_this *MjLoadTable) Check(key int) bool {
	_, ok := _this.tbl[key]
	return ok
}

func (_this *MjLoadTable) load(name string) error {
	_this.tbl = make(map[int]int, 10000)
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		buf, _, err := reader.ReadLine()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return err
		}
		str := string(buf)
		k, kerr := strconv.Atoi(str)
		if kerr == nil {
			_this.tbl[k] = 1
		}
	}
	return nil
}

/*
使用扩展版本表结构的时候  就要使用以下方法 以下方法 还需要优化 这版本不管
*/

//func (_this *mjDataTable) LoadTable() {
//	for i := 0; i <= 8; i++ {
//		name := fmt.Sprintf("./tbl_Json/TBL_258/eye_table_%d.json", i)
//		_this.mEye258Tbl[i].load(name)
//	}
//
//	for i := 0; i <= 8; i++ {
//		name := fmt.Sprintf("./tbl_Json/TBL_ALL/table_%d.json", i)
//		_this.mTbl[i].load(name)
//	}
//
//	for i := 0; i <= 8; i++ {
//		name := fmt.Sprintf("./tbl_Json/TBL_ALL/eye_table_%d.json", i)
//		_this.mEyeTbl[i].load(name)
//	}
//
//	for i := 0; i <= 8; i++ {
//		name := fmt.Sprintf("./tbl_Json/TBL_ALL/feng_table_%d.json", i)
//		_this.mFlowerTbl[i].load(name)
//	}
//
//	for i := 0; i <= 8; i++ {
//		name := fmt.Sprintf("./tbl_Json/TBL_ALL/feng_eye_table_%d.json", i)
//		_this.mFlowerEyeTbl[i].load(name)
//	}
//}
//
//func (_this *MjLoadTable) load(name string) error {
//	datas, err := ioutil.ReadFile(name)
//	if err != nil {
//		fmt.Print("读取文件错误")
//		return err
//	}
//	//老的 strat
//	//m := make([] string, 0 , 3000)
//	m := make(map[string]int)
//	err = json.Unmarshal(datas, &m)
//	if err != nil {
//		fmt.Print("表数据解析错误")
//		return err
//	}
//	for i, _ := range m {
//		if i == "" {
//			_this.tbl.Store(0, 1)
//		} else {
//			k, kerr := strconv.Atoi(i)
//			if kerr == nil {
//				_this.tbl.Store(k, 1)
//			}
//		}
//	}
//	//老的 end
//	return nil
//}
