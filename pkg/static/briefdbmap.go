package static

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

type BriefDb struct {
	*sql.DB
	m_dbName string //! 库名
}

type IBriefDBMap interface {
	// 查询所有
	SqlSelectAll(tx *sql.Tx, db *sql.DB) ([]interface{}, error)
	// 条件查询所有
	SqlSelectAllCondition(tx *sql.Tx, db *sql.DB) ([]interface{}, error)
	// 主键查询
	SQLSelectByPriKey(tx *sql.Tx, db *sql.DB) (interface{}, error)
	//
	SQLSelectByFieldNameInDB(tx *sql.Tx, db *sql.DB, nameInDB string) ([]interface{}, error)
	// 插入
	SQLInsert(tx *sql.Tx, db *sql.DB) (int64, error)
	// 主键更新
	SQLUpdateByPriKey(tx *sql.Tx, db *sql.DB) error
	//
	SQLUpdateByPriKeyAssign(tx *sql.Tx, db *sql.DB, assignTags []string) error
	// 指定键值更新
	SQLUpdateByKeyAssign(tx *sql.Tx, db *sql.DB, updateKey string, assignTags []string) error
	// 指定忽略键值更新
	SQLUpdateByPriKeyIgnore(tx *sql.Tx, db *sql.DB, ignoreTags []string) error
	// 主键删除
	SQLDelByPriKey(tx *sql.Tx, db *sql.DB) error
}

func NewBriefDb(driverName, dataSourceName string) (*BriefDb, error) {

	//! 解析库名
	name := dataSourceName
	begin := strings.Index(name, "/")
	end := strings.Index(name, "?")
	dbname := name[begin+1 : end]

	fmt.Println("mysql:", dataSourceName)
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		log.Panic("db open ping fail!  err:", err, " dns:", dbname)
	}

	db.SetMaxIdleConns(500)

	briefDB := &BriefDb{db, dbname}
	return briefDB, nil
}

type BriefDBField struct {
	Name       string
	Tag        string
	Type       string
	Addr       interface{}
	IntSave    sql.NullInt64
	StringSave sql.NullString
	FloatSave  sql.NullFloat64
	BoolSave   sql.NullBool
}

type BriefDBMap struct {
	objPtr  interface{}
	refType reflect.Type
	fields  []BriefDBField
	table   string

	IBriefDBMap
}

func NewBriefDBMap(table string, objPtr interface{}) (*BriefDBMap, error) {
	elem := reflect.ValueOf(objPtr).Elem()
	refType := elem.Type()

	var fields []BriefDBField
	for i, fnums := 0, refType.NumField(); i < fnums; i++ {
		var field BriefDBField
		field.Type = refType.Field(i).Type.String()
		if field.Type != "int64" && field.Type != "float64" &&
			field.Type != "bool" && field.Type != "string" &&
			field.Type != "int" && field.Type != "time.Time" {
			continue
			//！这里跳过不支持的数据类型
			//return nil, errors.New("Unknow Type : " + field.Type)
		}
		field.Name = refType.Field(i).Name
		field.Tag = refType.Field(i).Tag.Get("json")
		field.Addr = elem.Field(i).Addr().Interface()
		fields = append(fields, field)
	}

	return &BriefDBMap{
		objPtr:  objPtr,
		refType: refType,
		fields:  fields,
		table:   table,
	}, nil
}

func (self *BriefDBMap) SQlParamStr() string {

	// `objA`, `objB`, `objC`
	var paramStr string
	for i, fnum := 0, len(self.fields); i < fnum; i++ {
		if len(paramStr) > 0 {
			paramStr += ", "
		}
		paramStr += "`"
		//paramStr += self.fields[i].Name
		paramStr += self.fields[i].Tag
		paramStr += "`"
	}

	if len(paramStr) > 0 {
		paramStr += " "
		paramStr = " " + paramStr
	}

	return paramStr
}

func (self *BriefDBMap) SQlParamStrForSet() string {

	// `objA`=?, `objB`=?, `objC`=?
	var tagsStr string
	for i, flen := 0, len(self.fields); i < flen; i++ {
		if len(tagsStr) > 0 {
			tagsStr += ", "
		}
		tagsStr += "`"
		tagsStr += self.fields[i].Tag
		tagsStr += "`"
		tagsStr += " = ?"
	}
	if len(tagsStr) > 0 {
		tagsStr += " "
		tagsStr = " " + tagsStr
	}

	return tagsStr
}

func (self *BriefDBMap) SQlParamStrCustomAssign(assignTags []string) (string, []interface{}, error) {

	//验证tags正确性
	for _, tag := range assignTags {
		if self.GetFieldIndex(tag) < 0 {
			return "", nil, errors.New(fmt.Sprintf("dbmap error tag not exist. [%s]", tag))
		}
	}

	// `objA`=?, `objB`=?, `objC`=?
	var paramStr string
	var paramInf []interface{}

	for _, v := range assignTags {
		if len(paramStr) > 0 {
			paramStr += ", "
		}

		paramStr += "`"
		paramStr += v
		paramStr += "`"
		paramStr += "=?"

		paramInf = append(paramInf, self.GetFieldValue(self.GetFieldIndex(v)))
	}

	return paramStr, paramInf, nil
}

func (self *BriefDBMap) SQlParamStrCustomIgnore(ignoreTags []string) (string, []interface{}, error) {

	//验证tags正确性
	for _, tag := range ignoreTags {
		if self.GetFieldIndex(tag) < 0 {
			return "", nil, errors.New(fmt.Sprintf("dbmap error tag not exist. [%s]", tag))
		}
	}

	// `objC`=? -- ignore key objA, ignore objB
	var paramStr string
	var paramInf []interface{}
	for i, fnum := 0, len(self.fields); i < fnum; i++ {

		bIgnore := false
		for _, v := range ignoreTags {
			if self.fields[i].Tag == v {
				bIgnore = true
			}
		}
		if bIgnore {
			continue
		}
		if len(paramStr) > 0 {
			paramStr += ", "
		}
		paramStr += "`"
		paramStr += self.fields[i].Tag
		paramStr += "`"
		paramStr += "=?"

		paramInf = append(paramInf, self.GetFieldValue(i))
	}

	if len(paramStr) > 0 {
		paramStr += " "
		paramStr = " " + paramStr
	}

	return paramStr, paramInf, nil
}

func (self *BriefDBMap) SQlConditionStrCustom(tags []string) (string, []interface{}, error) {

	//验证tags正确性
	for _, tag := range tags {
		if self.GetFieldIndex(tag) < 0 {
			return "", nil, errors.New(fmt.Sprintf("dbmap error tag not exist. [%s]", tag))
		}
	}

	// `objA`=? AND `objB`=? AND`objC`=?
	var strCondition string
	var paramInf []interface{}
	for i, fnum := 0, len(tags); i < fnum; i++ {
		if len(strCondition) > 0 {
			strCondition += " AMD "
		}
		strCondition += "`"
		strCondition += tags[i]
		strCondition += "`"
		strCondition += "=?"

		paramInf = append(paramInf, self.GetFieldValue(i))
	}

	if len(strCondition) > 0 {
		strCondition += " "
		strCondition = " " + strCondition
	}

	return strCondition, paramInf, nil
}

func (self *BriefDBMap) GetFieldIndex(tag string) int {

	for i, v := range self.fields {
		if v.Tag == tag {
			return i
		}
	}

	return -1
}

func (self *BriefDBMap) GetFields() []BriefDBField {
	return self.fields
}

func (self *BriefDBMap) GetFieldValue(idx int) interface{} {
	switch self.fields[idx].Type {
	case "int":
		return self.fields[idx].Addr.(*int)
	case "int64":
		return self.fields[idx].Addr.(*int64)
	case "float64":
		return self.fields[idx].Addr.(*float64)
	case "bool":
		return self.fields[idx].Addr.(*bool)
	case "string":
		return self.fields[idx].Addr.(*string)
	case "time.Time":
		return self.fields[idx].Addr.(*time.Time)
	default:
	}
	return nil
}

func (self *BriefDBMap) GetFieldValues() []interface{} {
	var values []interface{}
	for i, flen := 0, len(self.fields); i < flen; i++ {
		values = append(values, self.GetFieldValue(i))
	}
	return values
}

func (self *BriefDBMap) GetFieldValAddr(idx int) interface{} {
	switch self.fields[idx].Type {
	case "int":
		return &self.fields[idx].IntSave
	case "int64":
		return &self.fields[idx].IntSave
	case "float64":
		return &self.fields[idx].FloatSave
	case "bool":
		return &self.fields[idx].BoolSave
	case "string":
		return &self.fields[idx].StringSave
	case "time.Time":
		return &self.fields[idx].StringSave
	default:
	}
	return nil
}

func (self *BriefDBMap) GetFieldValAddrs() []interface{} {

	var addrs []interface{}
	for i, fcnt := 0, len(self.fields); i < fcnt; i++ {
		addrs = append(addrs, self.GetFieldValAddr(i))
	}

	return addrs
}

func (self *BriefDBMap) BackToObject() interface{} {

	for i, flen := 0, len(self.fields); i < flen; i++ {
		switch self.fields[i].Type {
		case "int":
			if self.fields[i].IntSave.Valid {
				*self.fields[i].Addr.(*int) = int(self.fields[i].IntSave.Int64)
			}
			break
		case "int64":
			if self.fields[i].IntSave.Valid {
				*self.fields[i].Addr.(*int64) = self.fields[i].IntSave.Int64
			}
			break
		case "string":
			if self.fields[i].StringSave.Valid {
				*self.fields[i].Addr.(*string) = self.fields[i].StringSave.String
			}
			break
		case "time.Time":
			if self.fields[i].StringSave.Valid {

				datetime := self.fields[i].StringSave.String
				datetime = datetime[:strings.Index(datetime, "+")]
				idx := strings.Index(datetime, "T")
				datetime = datetime[:idx] + " " + datetime[idx+1:]

				timeLayout := "2006-01-02 15:04:05"
				loc, _ := time.LoadLocation("Local")
				theTime, _ := time.ParseInLocation(timeLayout, datetime, loc)
				*self.fields[i].Addr.(*time.Time) = theTime
			}
			break
		case "float64":
			if self.fields[i].FloatSave.Valid {
				*self.fields[i].Addr.(*float64) = self.fields[i].FloatSave.Float64
			}
			break
		case "bool":
			if self.fields[i].BoolSave.Valid {
				*self.fields[i].Addr.(*bool) = self.fields[i].BoolSave.Bool
			}
			break
		default:
		}
	}
	return self.objPtr
}

func (self *BriefDBMap) PrepareStmt(tx *sql.Tx, db *sql.DB,
	sqlstr string) (*sql.Stmt, error) {

	if tx != nil {
		return tx.Prepare(sqlstr)
	}
	if db != nil {
		return db.Prepare(sqlstr)
	}
	return nil, errors.New("tx & db both nil")
}

// 获取所有数据
func (self *BriefDBMap) SqlSelectAll(tx *sql.Tx, db *sql.DB) ([]interface{}, error) {
	sqlstr := fmt.Sprintf("select * from %s", self.table)

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rs, err := stmt.Query()
	if rs == nil {
		return nil, errors.New("row is nil")
	}

	var objs []interface{}
	for rs.Next() {
		obj := reflect.New(self.refType).Interface()
		fieldsMap, err := NewBriefDBMap(self.table, obj)
		if err != nil {
			return nil, err
		}

		err = rs.Scan(fieldsMap.GetFieldValAddrs()...)
		if err != nil {
			return nil, err
		}
		fieldsMap.BackToObject()
		objs = append(objs, obj)
	}

	return objs, nil
}

//！获取单条数据
func (self *BriefDBMap) SqlSelectOne(tx *sql.Tx, db *sql.DB) (interface{}, error) {
	sqlstr := fmt.Sprintf("select * from %s", self.table)

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rs, err := stmt.Query()
	if rs == nil {
		return nil, errors.New("row is nil")
	}
	obj := reflect.New(self.refType).Interface()
	for rs.Next() {
		fieldsMap, err := NewBriefDBMap(self.table, obj)
		if err != nil {
			return nil, err
		}

		err = rs.Scan(fieldsMap.GetFieldValAddrs()...)
		if err != nil {
			return nil, err
		}
		fieldsMap.BackToObject()
		break
	}

	return obj, nil
}

func (self *BriefDBMap) SqlSelectAllCondition(tx *sql.Tx, db *sql.DB, strcon string, params ...interface{}) ([]interface{}, error) {
	sqlstr := fmt.Sprintf("select * from %s %s", self.table, strcon)

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rs, err := stmt.Query(params)
	if rs == nil {
		return nil, errors.New("row is nil")
	}

	var objs []interface{}
	for rs.Next() {
		obj := reflect.New(self.refType).Interface()
		fieldsMap, err := NewBriefDBMap(self.table, obj)
		if err != nil {
			return nil, err
		}

		err = rs.Scan(fieldsMap.GetFieldValAddrs()...)
		if err != nil {
			return nil, err
		}
		fieldsMap.BackToObject()
		objs = append(objs, obj)
	}

	return objs, nil
}

func (self *BriefDBMap) SQLSelectByPriKey(tx *sql.Tx, db *sql.DB) (interface{}, error) {

	sqlstr := "SELECT " + self.SQlParamStr() +
		" FROM `" + self.table + "` " + " where `" + self.fields[0].Tag + "` = ? "

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(self.GetFieldValue(0))
	if row == nil {
		return nil, errors.New("row is nil")
	}

	err = row.Scan(self.GetFieldValAddrs()...)
	if err != nil {
		return nil, err
	}

	return self.BackToObject(), nil
}

func (self *BriefDBMap) SQLSelectByFieldNameInDB(tx *sql.Tx, db *sql.DB, nameInDB string) ([]interface{}, error) {

	idx := -1
	for i, flen := 0, len(self.fields); i < flen; i++ {
		if self.fields[i].Tag == nameInDB {
			idx = i
			break
		}
	}

	if idx < 0 {
		return nil, errors.New("no field match `sql` tag:" + nameInDB)
	}

	sqlstr := "SELECT " + self.SQlParamStr() +
		" FROM `" + self.table + "` " + " where `" + self.fields[idx].Tag + "` = ? "

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rs, err := stmt.Query(self.GetFieldValue(idx))
	if err != nil {
		return nil, err
	}

	var objs []interface{}
	for rs.Next() {
		obj := reflect.New(self.refType).Interface()
		fieldsMap, err := NewBriefDBMap(self.table, obj)
		if err != nil {
			return nil, err
		}

		err = rs.Scan(fieldsMap.GetFieldValAddrs()...)
		if err != nil {
			return nil, err
		}
		fieldsMap.BackToObject()
		objs = append(objs, obj)
	}

	return objs, nil
}

func (self *BriefDBMap) SQLInsert(tx *sql.Tx, db *sql.DB) (int64, error) {

	var vs string
	for i, fnum := 0, len(self.fields); i < fnum; i++ {
		if len(vs) > 0 {
			vs += ", "
		}
		vs += "?"
	}

	sqlstr := "INSERT INTO `" + self.table + "` (" + self.SQlParamStr() + ") " +
		"VALUES (" + vs + ")"

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(self.GetFieldValues()...)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (self *BriefDBMap) SQLUpdateByPriKey(tx *sql.Tx, db *sql.DB) error {

	sqlstr := "UPDATE `" + self.table + "` SET " + self.SQlParamStrForSet() +
		" WHERE `" + self.fields[0].Tag + "` = ?"

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	values := self.GetFieldValues()
	values = append(values, self.GetFieldValue(0))
	_, err = stmt.Exec(values...)
	if err != nil {
		return err
	}
	return nil
}

// assignTags must not include keyTags
func (self *BriefDBMap) SQLUpdateByPriKeyAssign(tx *sql.Tx, db *sql.DB, assignTags []string) error {

	//包含判定
	for _, vIgnore := range assignTags {
		if self.fields[0].Tag == vIgnore {
			return errors.New("assignTags must not include keyTags")
		}
	}

	strcust, infParam, err := self.SQlParamStrCustomAssign(assignTags)
	if err != nil {
		return err
	}

	sqlstr := "UPDATE `" + self.table + "` SET " + strcust + " WHERE `" + self.fields[0].Tag + "` = ?"

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	//params
	var params []interface{}
	params = append(params, infParam...)
	params = append(params, self.GetFieldValue(0))
	_, err = stmt.Exec(params...)
	if err != nil {
		return err
	}
	return nil
}

func (self *BriefDBMap) SQLUpdateByKeyAssign(tx *sql.Tx, db *sql.DB, updateKey string, assignTags []string) error {
	//包含判定
	for _, vIgnore := range assignTags {
		if self.fields[0].Tag == vIgnore {
			return errors.New("assignTags must not include keyTags")
		}
	}

	var value interface{}
	for i, flen := 0, len(self.fields); i < flen; i++ {
		if self.fields[i].Tag == updateKey {
			value = self.GetFieldValue(i)
		}
	}

	strcust, infParam, err := self.SQlParamStrCustomAssign(assignTags)
	if err != nil {
		return err
	}

	sqlstr := "UPDATE `" + self.table + "` SET " + strcust + " WHERE `" + updateKey + "` = ?"

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	//params
	var params []interface{}
	params = append(params, infParam...)
	params = append(params, value)
	_, err = stmt.Exec(params...)
	if err != nil {
		return err
	}
	return nil
}

// ignoreTags Auto include keyTags
func (self *BriefDBMap) SQLUpdateByPriKeyIgnore(tx *sql.Tx, db *sql.DB, ignoreTags []string) error {

	//去重
	for _, vIgnore := range ignoreTags {
		if self.fields[0].Tag == vIgnore {
			return errors.New("不允许包含主键")
		}
	}

	strcust, infParam, err := self.SQlParamStrCustomIgnore(ignoreTags)
	if err != nil {
		return err
	}

	sqlstr := "UPDATE `" + self.table + "` SET " + strcust + " WHERE `" + self.fields[0].Tag + "` = ?"

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	//params
	var params []interface{}
	params = append(params, infParam...)
	params = append(params, self.GetFieldValue(0))
	_, err = stmt.Exec(params...)
	if err != nil {
		return err
	}
	return nil
}

func (self *BriefDBMap) SQLDelByPriKey(tx *sql.Tx, db *sql.DB) error {

	sqlstr := "DELETE FROM `" + self.table + "` " + " where `" + self.fields[0].Tag + "` = ? "

	stmt, err := self.PrepareStmt(tx, db, sqlstr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(self.GetFieldValue(0))
	if err != nil {
		return err
	}
	return nil
}
