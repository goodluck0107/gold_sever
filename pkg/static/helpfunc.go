package static

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static/goEncrypt"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	goRedis "github.com/go-redis/redis"
)

const TIMEFORMAT = "2006-01-02 15:04:05" // 时间格式化

// user信息加密key
const UserEncodeKey = "facai20190110$#"

//! 解消息
func HF_DecodeMsg(msg []byte) (string, *Msg_Sign, int16, []byte, bool, int64, string) {
	var msgbase Msg_MsgBase
	err := json.Unmarshal(msg, &msgbase)
	if err != nil {
		xlog.Logger().Errorln(err)
		return "", nil, 0, []byte(""), false, 0, ""
	}

	return msgbase.Head, &msgbase.Sign, msgbase.ErrCode, []byte(msgbase.Data), true, msgbase.Uid, msgbase.IP
}

//! 加密消息
func HF_EncodeMsg(msghead string, errCode int16, v interface{}, encode int, encodeKey string, uid int64) []byte {
	var sign Msg_Sign
	sign.Time = time.Now().Unix()
	sign.Encode = encode

	var msgbase Msg_MsgBase
	msgbase.ErrCode = errCode
	msgbase.Head = msghead
	msgbase.Sign = sign
	msgbase.Uid = uid

	//if !zip {
	if v != nil {
		kind := reflect.TypeOf(v).Kind()
		switch kind {
		case reflect.String:
			msgbase.Data = v.(string)
		default:
			//syslog.Logger().Errorln("msg type error: ", kind)
			msgbase.Data = HF_JtoA(v)
		}

		if msghead != "ping" && msghead != consts.MsgTypeHouseMemberOnline {
			// if msgbase.ErrCode == 0 {
			xlog.Logger().WithFields(logrus.Fields{
				"_head":   msghead,
				"data":    msgbase.Data,
				"uid":     msgbase.Uid,
				"encode:": encode,
				"errcode": msgbase.ErrCode,
			}).Infoln("【SEND MESSAGE】")
		}

		// 需要做加密处理
		if encode == consts.EncodeAes {
			datas, err := goEncrypt.AesCTR_Encrypt([]byte(msgbase.Data), []byte(encodeKey))
			if err != nil {
				xlog.Logger().Errorln("encode msg filed:", err)
			} else {
				msgbase.Data = base64.URLEncoding.EncodeToString(datas)
			}
		}
	} else {
		if msghead != "ping" && msghead != consts.MsgTypeHouseMemberOnline {
			xlog.Logger().WithFields(logrus.Fields{
				"msghead:": msghead,
				"data":     msgbase.Data,
				"uid":      msgbase.Uid,
				"encode:":  msgbase.ErrCode,
			}).Infoln("【SEND NULL MESSAGE】")
		}
	}

	return HF_JtoB(&msgbase)
}

//! 字符串解码
func HF_DecodeStr(srcStr []byte, key string) (string, error) {
	decodeStr := ""
	base64Buf, err := base64.URLEncoding.DecodeString(string(srcStr))
	if err != nil {
		return decodeStr, err
	}
	decodeBuf, err := goEncrypt.AesCTR_Decrypt(base64Buf, []byte(key))
	if err != nil {
		return decodeStr, err
	}
	decodeStr = HF_Bytestoa(decodeBuf)
	return decodeStr, err
}

//! 字符串加密
func HF_EncodeStr(srcStr []byte, key string) (string, error) {
	encodeStr := ""
	aesBuf, err := goEncrypt.AesCTR_Encrypt(srcStr, []byte(key))
	if err != nil {
		xlog.Logger().Errorln("encode msg fail", err)
	} else {
		encodeStr = base64.URLEncoding.EncodeToString(aesBuf)
	}
	return encodeStr, err
}

//! int取最小
func HF_MinInt(a int, b int) int {
	if a < b {
		return a
	}

	return b
}

//! int取最大
func HF_MaxInt(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func HF_MaxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

func HF_Abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}

//! 数字 转 []byte

func HF_Itobytes(i int) []byte {
	return []byte(strconv.Itoa(i))
}
func HF_I64tobytes(i int64) []byte {
	return HF_Atobytes(HF_I64toa(i))
}

//! []byte 转 数字
func HF_Bytestoi(b []byte) int {
	return HF_Atoi(HF_Bytestoa(b))
}

func HF_Bytestoi64(b []byte) int64 {
	return HF_Atoi64(HF_Bytestoa(b))
}

func HF_Bytestof64(b []byte) float64 {
	return HF_Atof64(HF_Bytestoa(b))
}

//! 字符串转数字
func HF_Atoi(s string) int {
	num, _ := strconv.Atoi(s)
	return num
}

func HF_Itoa(i int) string {
	return strconv.Itoa(i)
}

func HF_I64stoas(i64s ...int64) []string {
	res := make([]string, len(i64s))
	for i, i64 := range i64s {
		res[i] = strconv.FormatInt(i64, 10)
	}
	return res
}

func HF_Atoi64(s string) int64 {
	num, _ := strconv.ParseInt(s, 10, 64)
	return num
}

func HF_F64toa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

//! 数字转string
func HF_I64toa(i int64) string {
	return strconv.FormatInt(i, 10)
}

//! string 转 []byte
func HF_Atobytes(s string) []byte {
	return []byte(s)
}

//! []byte 转 string
func HF_Bytestoa(b []byte) string {
	return string(b[:])
}

//! 结构转json串
func HF_JtoA(v interface{}) string {
	s, _ := json.Marshal(v)
	return string(s)
}

//! 结构转json串
func HF_JtoB(v interface{}) []byte {
	s, _ := json.Marshal(v)
	return s
}

//! 字符串转float32
func HF_Atof(s string) float32 {
	num, _ := strconv.ParseFloat(s, 32)
	return float32(num)
}

//! 字符串转float64
func HF_Atof64(s string) float64 {
	num, _ := strconv.ParseFloat(s, 64)
	return num
}

//! int数组转换为字符串
func HF_IntsToStr(ints []int) string {
	vStr := ""
	for i := 0; i < len(ints); i++ {
		vStr += fmt.Sprintf("%d", ints[i])
		if i != len(ints)-1 {
			vStr += ","
		}
	}
	return vStr
}

//! 字符串转换为int数组
func HF_StrToInts(intstr string) []int {
	ints := []int{}
	if intstr == "" {
		return []int{}
	}
	splitStr := strings.Split(intstr, ",")
	for i := 0; i < len(splitStr); i++ {
		ints = append(ints, HF_Atoi(splitStr[i]))
	}
	return ints
}

//! 得到一个随机数
func HF_GetRandom(num int) int {
	return rand.New(rand.NewSource(time.Now().UnixNano() + rand.Int63n(1000))).Intn(num)
}

//! 得到一个随机字符串（例如：邀请码）
func HF_GetRandomString(max int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < max; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

//! 得到一个随机数字符串
func HF_GetRandomNumberString(max int) string {
	str := "0123456789"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < max; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

//! 克隆对象 dst为指针
func HF_DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

//! 是否合法
func HF_IsLicitName(name []byte) bool {
	for i := 0; i < len(name); i++ {
		if name[i] == '\r' {
			return false
		} else if name[i] == '\'' {
			return false
		} else if name[i] == '\n' {
			return false
		} else if name[i] == ' ' {
			return false
		} else if name[i] == '	' {
			return false
		}
	}

	return true
}

//! 得到ip
func HF_GetHttpIP(req *http.Request) string {
	ip := req.Header.Get("Remote_addr")
	if ip == "" {
		ip = req.RemoteAddr
	}
	return strings.Split(ip, ":")[0]
}

func getReturn(header string, info string, code int) []byte {
	resultHead := new(Msg_Header)
	resultHead.Header = header
	resultHead.Data = info
	resultHead.Sign.Encode = code
	resultHead.Sign.Time = time.Now().UnixNano() / 1000000
	return HF_JtoB(resultHead)
}

// 通用得判错函数
func HF_CheckErr(e error) {
	// defer func() {
	// 	x := recover()
	// 	if x != nil {
	// 		syslog.Logger().Errorln(x, string(sysdebug.Stack()))
	// 	}
	// }()
	if e != nil {
		xlog.Logger().Errorln(e)
	}
	return
}

/*
	往切片中插入一个元素，需提供：切片，插入坐标，插入元素
	这个函数不会检查类型是否一致，使用前确保切片和插入元素类型一致
*/
func HF_InsertOne(slice interface{}, pos int, value interface{}) interface{} {
	v := reflect.ValueOf(slice)
	v = reflect.Append(v, reflect.ValueOf(value))
	reflect.Copy(v.Slice(pos+1, v.Len()), v.Slice(pos, v.Len()))
	v.Index(pos).Set(reflect.ValueOf(value))
	return v.Interface()
}

/*
	根据字段名字设置字段的值
	struc_dst	结构体地址
	column   	字段名
	set      	设置的目标值
*/
func Setter(structDst interface{}, column string, set interface{}) error {
	struc := reflect.ValueOf(structDst).Elem()
	if !(struc.Kind() == reflect.Struct) {
		return errors.New("the args indexed 0 is not a struct")
	}
	field := struc.FieldByName(column)
	_set := reflect.ValueOf(set)
	if !field.IsValid() {
		return errors.New(fmt.Sprintf("Setter: the column named:%s not exist", column))
	}
	if !field.CanSet() {
		return errors.New("setter: can't set")
	}
	if _set.Kind() == reflect.Invalid {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}
	if !(field.Kind() == _set.Kind()) {
		return errors.New("setter: type mismatch")
	}
	field.Set(_set)
	return nil
}

/*
	字符串decode给结构体
	dst：结构体指针
*/
func HF_Atostru(data string, dst interface{}) {
	HF_CheckErr(json.Unmarshal(HF_Atobytes(data), dst))
}

/*
	判断文件或文件夹是否存在
*/
func HF_PathExists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

/*
	元素去重
*/
func HF_RemoveRep(slc []byte) []byte {
	if len(slc) < 1024 {
		// 切片长度小于1024的时候，循环来过滤
		return removeRepByLoop(slc)
	} else {
		// 大于的时候，通过map来过滤
		return removeRepByMap(slc)
	}
}

// 通过两重循环过滤重复元素
func removeRepByLoop(slc []byte) []byte {
	result := make([]byte, 0) // 存放结果
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i] == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}

// 通过map主键唯一的特性过滤重复元素
func removeRepByMap(slc []byte) []byte {
	result := make([]byte, 0)
	tempMap := map[byte]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// slice int 转 byte
func HF_IntsToBytes(slc []int) []byte {
	byteCard := make([]byte, 0)
	for _, value := range slc {
		byteCard = append(byteCard, byte(value))
	}
	return byteCard
}

// slice byte 转 []int
func HF_BytesToInts(slc []byte) []int {
	intCard := make([]int, 0)
	for _, value := range slc {
		intCard = append(intCard, int(value))
	}
	return intCard
}

// int64 转 []byte 2进制流
func HF_Int64ToBytes(num int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(num))
	return buf
}

// []byte 转 int64
func HF_BytesToInt64(slc []byte) int64 {
	return int64(binary.BigEndian.Uint64(slc))
}

// int 转 []byte
func HF_IntToBytes(num int) []byte {
	return []byte(strconv.Itoa(num))
}

// 通过方法名调用对象的方法
func HF_CallMethod(structdst interface{} /*结构体指针*/, callMethodname string /*调用的方法名字*/, args ...interface{} /*方法所需要的参数*/) ([]reflect.Value, error) {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	if v := reflect.ValueOf(structdst).MethodByName(callMethodname); v.String() == "<invalid Value>" {
		//syslog.Logger().Errorln("Call method named: [" + callMethodname + "] #FAILED#!")
		return nil, errors.New("The method invoked does not exist")
	} else {
		//syslog.Logger().Errorln("Call method named: [" + callMethodname + "] #SUCCEED#!")
		return v.Call(inputs), nil
	}
}

// 得到函数执行时间
// 使用方法：在函数第一行调用：
// func ep(){
//   defer static.HF_FuncElapsedTime()()
// }
func HF_FuncElapsedTime() func() {
	starttime := time.Now()
	if pc, f, l, ok := runtime.Caller(1); ok {
		return func() {
			xlog.Logger().Debug(fmt.Sprintf("%s:%d %s() elapsed time: %d(ms).", f, l, runtime.FuncForPC(pc).Name(), time.Since(starttime).Nanoseconds()/1e6))
		}
	}
	return func() {
		xlog.Logger().Debug(fmt.Sprintf("elapsed time: %d(ms).", time.Since(starttime).Nanoseconds()/1e6))
	}
}

// 得到函数调用连信息
// 调试用，平时关掉，有性能开支
func HF_FuncCallerInfo() {
	pc := make([]uintptr, 10) // at least 1 entry needed
	n := runtime.Callers(0, pc)
	if n <= 2 { // Two levels below depth does not print
		return
	}
	str := "\n<Function caller link eve, Not panic>\n"
	frames := runtime.CallersFrames(pc[2 : n-2])
	for {
		frame, more := frames.Next()
		str += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	xlog.Logger().Debug(str)
}

// 得到执行路径
func HF_GetExecPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	p, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	return p, nil
}

// 得到进程名字
func HF_GetCourseName() string {
	courseName := filepath.Base(os.Args[0])
	if runtime.GOOS == "windows" {
		courseName = strings.Replace(courseName, ".exe", "", -1)
	}
	return courseName
}

// 读取yaml文件
func HF_ReadYaml(path string, name string, v interface{}) error {
	cf := filepath.Join(path, name)
	data, err := ioutil.ReadFile(cf)
	if err != nil {
		return fmt.Errorf("HF_ReadYaml error:%s", err)
	}
	if err = yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("HF_ReadYaml:%s", err)
	}
	//fmt.Println(fmt.Sprintf("ReadYaml succeed : %+v ",v))
	return nil
}

// 合并二进制流
func HF_BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

// 对方端口是否打开
func HF_PortIsOpen(raddr string, timeout int) error {
	xlog.Logger().Debug(fmt.Sprintf("Try to ping %s", raddr))

	conn, err := net.DialTimeout("tcp", raddr, time.Duration(timeout)*time.Second)
	if nil != err {
		return err
	}

	defer conn.Close()
	return nil
}

// 得到过去的目标时间距离当前时间的天数
func HF_GetApartDays(t int64) int {
	apartHours := int(time.Now().Sub(time.Unix(t, 0)).Hours()) - time.Now().Hour()
	if apartHours <= 0 {
		return 0
	}
	days := apartHours / 24
	if apartHours%24 > 0 {
		days++
	}
	return days
}

// 得到今天剩余的秒数
func HF_GetTodayRemainSecond() int64 {
	todayLast := time.Now().Format("2006-01-02") + " 23:59:59"
	todayLastTime, _ := time.ParseInLocation("2006-01-02 15:04:05", todayLast, time.Local)
	return todayLastTime.Unix() - time.Now().Local().Unix()
}

// 机器码效验
func HF_IsValidMachineCode(machineCode string) bool {
	return !(machineCode == "" || machineCode == consts.UnknownMachineCode)
}

// 游戏引擎效验
func HF_IsValidGameEngine(engine int) bool {
	return engine == consts.PHPCocosCreator || engine == consts.PHPCocosJs
}

// 得到游戏引擎标志位
// 0 和 1 为cocos creator   其他为Js
func HF_GetGameEngine(engine int) int {
	if engine > 1 {
		return consts.EngineCocosJs
	}
	return consts.EngineCocosCreator
}

// 转换游戏引擎
func HF_CvtGameEngine(phpEngine int) int {
	if phpEngine == consts.PHPCocosJs {
		return consts.EngineCocosJs
	}
	return consts.EngineCocosCreator
}

// 浮点数除法取整
func HF_DecimalDivide(dividend, divisor float64, n uint8) float64 {
	if divisor == 0 {
		return dividend
	}
	f := fmt.Sprintf("%%0.%df", n)
	// fmt.Println(f)
	v, _ := strconv.ParseFloat(fmt.Sprintf(f, dividend/divisor), 64)
	return v
}

//GetTimeLastDayStartAndEnd 返回给定时间的昨天的开始和结束时间
func GetTimeLastDayStartAndEnd(t time.Time) (time.Time, time.Time) {
	y, m, d := t.Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	end := start.Add(-24 * time.Hour)
	return end, start
}

//GetTimeLastSixHourStartAndEnd 返回给定时间的六小时整点的开始和结束时间
func GetTimeLastSixHourStartAndEnd(t time.Time) (time.Time, time.Time) {
	y, m, d := t.Date()
	hour := t.Hour()
	hour = hour / 6
	hour = hour * 6
	start := time.Date(y, m, d, hour, 0, 0, 0, time.Local)
	end := start.Add(-6 * time.Hour)
	return end, start
}

// 浮点数加法取整
func HF_DecimalAddition(a, b float64, n uint8) float64 {
	f := fmt.Sprintf("%%0.%df", n)
	// fmt.Println(f)
	v, _ := strconv.ParseFloat(fmt.Sprintf(f, a+b), 64)
	return v
}

func SwitchIntVitamin(vitamin int) int64 {
	return int64(vitamin) * consts.VitaminExchangeRate
}

func SwitchVitaminInt(vitamin int64) int {
	return int(vitamin / consts.VitaminExchangeRate)
}

func SwitchVitaminInt64(vitamin int64) int64 {
	return vitamin * consts.VitaminExchangeRate
}

func SwitchVitaminToF64(vitamin int64) float64 {
	if vitamin == consts.VitaminInvalidValueSrv {
		return consts.VitaminInvalidValueCli
	}
	return HF_DecimalDivide(float64(vitamin), consts.VitaminExchangeRate, 2)
}

func SwitchF64ToVitamin(vitamin float64) int64 {
	if vitamin == consts.VitaminInvalidValueCli {
		return consts.VitaminInvalidValueSrv
	}
	if vitamin > 0 {
		return int64((vitamin + 0.005) * consts.VitaminExchangeRate)
	} else if vitamin < 0 {
		return int64((vitamin - 0.005) * consts.VitaminExchangeRate)
	}
	return int64((vitamin /*+0.005*/) * consts.VitaminExchangeRate)
}

// 两数相除 保留指定的小数点位数 并向下取整（满足疲劳值平均扣除需求）
func DecimalDivideFloor(dividend, divisor float64, n float64) float64 {
	if divisor == 0 {
		return dividend
	}
	return math.Floor(dividend/divisor*math.Pow(10, float64(n))) / math.Pow(10, float64(n))
}

func GetGoid() int64 {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		return 0
	}
	return int64(id)
}

func GetZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func GetLastTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 0, d.Location())
}

func RemoveFromSlice(dest []int64, val int64) []int64 {
	if dest == nil {
		return nil
	}
	length := len(dest)
	for i, v := range dest {
		if v == val {
			if i == length-1 {
				dest = dest[0:i]
			} else {
				dest = append(dest[0:i], dest[i+1:]...)
			}
		}
	}
	return dest
}

func ErrorsCheck(errs ...error) error {
	for i := 0; i < len(errs); i++ {
		if err := errs[i]; err != nil {
			return err
		}
	}
	return nil
}

func TxCommit(tx *gorm.DB, err error) bool {
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return false
	} else {
		if err = tx.Commit().Error; err != nil {
			xlog.Logger().Error(err)
			tx.Rollback()
			return false
		}
		return true
	}
}

func I64sToA(slice []int64) string {
	l := len(slice)
	if l > 0 {
		ss := make([]string, l)
		for i, i64 := range slice {
			ss[i] = fmt.Sprint(i64)
		}
		return strings.Join(ss, ",")
	}
	return ""
}

func AtoI64s(s string) []int64 {
	ss := strings.Split(s, ",")
	l := len(ss)
	slice := make([]int64, l)
	for i, si := range ss {
		i64, err := strconv.ParseInt(si, 10, 64)
		if err != nil {
			xlog.Logger().Error(err)
		} else {
			slice[i] = i64
		}
	}
	return slice
}

func In64(slice []int64, elem int64) bool {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == elem {
			return true
		}
	}
	return false
}

func In(slice []int, elem int) bool {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == elem {
			return true
		}
	}
	return false
}

// 是否包含敏感词
func IsContainSensitiveWord(redis *goRedis.Client, text string) bool {
	if redis == nil {
		return false
	}

	// 转UT8处理（中文） 减少遍历次数
	rText := []rune(text)

	// 字符串是否包含敏感词
	for i := 0; i < len(rText); i++ {
		for j := i + 1; j <= len(rText); j++ {
			content := string(rText[i:j])
			if redis.HExists(consts.REDIS_KEY_SENSITIVE_WORDS, content).Val() {
				return true
			}
		}
	}

	return false
}

// 检查敏感词并使用*号替换
func CheckSensitiveWord(redis *goRedis.Client, text string) string {
	if redis == nil {
		return text
	}

	// 转UT8处理（中文） 减少遍历次数
	rText := []rune(text)
	newText := text

	// 字符串是否包含敏感词
	for i := 0; i < len(rText); i++ {
		for j := i + 1; j <= len(rText); j++ {
			content := string(rText[i:j])
			if redis.HExists(consts.REDIS_KEY_SENSITIVE_WORDS, content).Val() {
				var str string
				for idx := 0; idx < len([]rune(content)); idx++ {
					str += "*"
				}
				newText = strings.Replace(newText, content, str, 1)
			}
		}
	}

	return newText
}

// 根据身份证号码获取年龄
func HF_GetAgeFromIdcard(idcard string) int {
	// 身份证解密
	idcard, _ = HF_DecodeStr(HF_Atobytes(idcard), UserEncodeKey)
	if len(idcard) <= 0 || len(idcard) > 18 {
		return -1
	}
	// 获取一下当前时间
	timeObj := time.Now()
	year := timeObj.Year()
	month := int(timeObj.Month())
	day := timeObj.Day()
	// 合法的身份证号 获取一下出生年月日
	birthYear, err1 := strconv.Atoi(idcard[6:10])
	birthMonth, err2 := strconv.Atoi(idcard[10:12])
	birthDay, err3 := strconv.Atoi(idcard[12:14])
	if err1 != nil || err2 != nil || err3 != nil {
		return -1
	} else {
		if year > birthYear {
			if month > birthMonth || (month == birthMonth && day >= birthDay) {
				return year - birthYear
			} else {
				return year - birthYear - 1
			}
		} else if year == birthYear {
			return 0
		} else {
			return -1
		}
	}
}
