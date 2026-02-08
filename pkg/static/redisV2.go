package static

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

var (
	redisClient map[string]*redis.Client
	// Nil redis返回空
	Nil = redis.Nil
)

// InitRedis redis init 可以进行多db初始化
func InitRedisV2(ip string, db int, auth string) *redis.Client {
	xlog.Logger().Warnf("Redis %s[%d] ...%s", ip, db, auth)
	redisClient = make(map[string]*redis.Client)
	redisCli := initClient(ip, db, auth)
	redisClient["zero"] = redisCli
	return redisCli
}

func initClient(ip string, db int, auth string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:        ip,
		Password:    auth,
		DB:          db,
		IdleTimeout: time.Second * 60,
	})
	s := client.Ping()
	if err := s.Err(); err != nil {
		panic(err)
	}
	return client
}

// GetRedisConn 获取客户端链接
func GetRedisConn(dbName string) *redis.Client {
	return redisClient[dbName]
}

// GetSessionConn 获取session缓存库
func GetDBZero() *redis.Client {
	return redisClient["zero"]
}
func CoverMapToStruct(data map[string]string, s interface{}) error {
	v := reflect.ValueOf(s).Elem()
	if !v.CanAddr() {
		return fmt.Errorf("must be a pointer")
	}
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i)
		tag := fieldInfo.Tag.Get("json")
		if tag == "" {
			tag = strings.ToLower(fieldInfo.Name)
		}
		if tag == "-" {
			continue
		}
		if value, ok := data[tag]; ok {
			kind := v.FieldByName(fieldInfo.Name).Kind()
			switch {
			case kind == reflect.Int64, kind == reflect.Int, kind == reflect.Int8,
				kind == reflect.Int16, kind == reflect.Int32:
				{
					val, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						panic(err)
					}
					fi := v.FieldByName(fieldInfo.Name)
					if !fi.CanSet() {
						return fmt.Errorf("can not set value of:%s", fieldInfo.Name)
					}
					fi.SetInt(val)
				}
			case kind == reflect.String:
				{
					fi := v.FieldByName(fieldInfo.Name)
					if !fi.CanSet() {
						return fmt.Errorf("can not set value of:%s", fieldInfo.Name)
					}
					fi.SetString(value)
				}
			case kind == reflect.Bool:
				{

					val, err := strconv.ParseBool(value)
					if err != nil {
						panic(err)
					}
					fi := v.FieldByName(fieldInfo.Name)
					if !fi.CanSet() {
						return fmt.Errorf("can not set value of:%s", fieldInfo.Name)
					}
					fi.SetBool(val)

				}
			case kind == reflect.Float32, kind == reflect.Float64:
				{
					val, err := strconv.ParseFloat(value, 64)
					if err != nil {
						panic(err)
					}
					fi := v.FieldByName(fieldInfo.Name)
					if !fi.CanSet() {
						return fmt.Errorf("can not set value of:%s", fieldInfo.Name)
					}
					fi.SetFloat(val)
				}
			default:
				if len(fieldInfo.Name) > 4 && fieldInfo.Name[len(fieldInfo.Name)-4:] == "Time" {
					t, err := time.Parse("2006-01-02T15:04:05Z", value) //redis 时间格式
					if err != nil {
						panic(err)
					}
					fi := v.FieldByName(fieldInfo.Name)
					if !fi.CanSet() {
						return fmt.Errorf("can not set value of:%s", fieldInfo.Name)
					}
					j := reflect.ValueOf(t)
					fi.Set(j)
				}
				continue
			}
		}

	}
	return nil
}
