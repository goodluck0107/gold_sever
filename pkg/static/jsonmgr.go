package static

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"io/ioutil"
	"os"
	"strings"
)

type JsonMgr struct {
}

var jsonmgrsingleton *JsonMgr = nil

//! public
func GetJsonMgr() *JsonMgr {
	if jsonmgrsingleton == nil {
		jsonmgrsingleton = new(JsonMgr)
	}
	return jsonmgrsingleton
}

//获取指定目录下的所有文件,包含子目录下的文件
func GetAllFiles(dirPth string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() {
			// 目录, 递归遍历
			dirs = append(dirs, dirPth+PthSep+fi.Name())
			GetAllFiles(dirPth + PthSep + fi.Name())
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), ".json")
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}
	// 读取子目录下文件
	for _, table := range dirs {
		temp, _ := GetAllFiles(table)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}
	return files, nil
}
func (self *JsonMgr) ReadData(path string, name string, v interface{}) bool {
	fullpath := fmt.Sprintf("%s/%s.json", path, name)
	config, err := ioutil.ReadFile(fullpath)
	if err != nil {
		//log.Fatal(fmt.Sprintf("JsonMgr ReadData ReadFile err : %s.json", name))
		xlog.Logger().Warningln(fmt.Sprintf("JsonMgr ReadData ReadFile err1 : %s.json", name), err)
		return false
	}
	err = json.Unmarshal(config, v)
	if err != nil {
		//log.Fatal(fmt.Sprintf("JsonMgr ReadData Unmarshal err : %s.json", name, err))
		xlog.Logger().Warningln(fmt.Sprintf("JsonMgr ReadData Unmarshal err2 : %s.json:", name), err)
		return false
	}

	return true
}
