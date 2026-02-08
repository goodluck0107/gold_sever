package static

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

var defaultAddr = "127.0.0.1:6379"

type RedisError string

func (err RedisError) Error() string {
	return "Redis Error: " + string(err)
}

type RedisCli struct {
	pool *redis.Pool
}

func InitRedis(ip string, db int, auth string) *RedisCli {

	pRedisPool := &redis.Pool{
		MaxIdle:     50,
		MaxActive:   12000,
		IdleTimeout: 0,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ip)
			if err != nil {
				return nil, err
			}
			if auth != "" {
				//redis校验
				if _, err := c.Do("AUTH", auth); err != nil {
					c.Close()
					return nil, err
				}
			}

			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}

			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	xlog.Logger().Infoln("redis connect!", db)

	return &RedisCli{
		pool: pRedisPool,
	}
}

func (client *RedisCli) GlobalLock(key string, redisLockTimeout int, f func()) {
	//从连接池中娶一个con链接，pool可以自己定义。
	con := client.pool.Get()
	defer con.Close()
	for {
		if _, err := redis.String(con.Do("set", key, 1, "ex", redisLockTimeout, "nx")); err != nil {
			//间隔半秒去抢占锁
			time.Sleep(500 * time.Millisecond)
			continue
		}
		break
	}
	//业务逻辑
	NO_ERROR(f)
	//释放锁
	_, _ = con.Do("del", key)
}

func (client *RedisCli) Close() {
	if client != nil {
		client.pool.Close()
		client.pool = nil
	}
}

// 数据库切换
func (client *RedisCli) SelectDB(db uint8) error {
	cn := client.pool.Get()
	defer cn.Close()
	_, err := cn.Do("SELECT", db)
	return err
}

// key 相关方法
func (client *RedisCli) Keys(pattern string) ([][]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("KEYS", pattern)

	if err != nil {
		return nil, err
	}

	var vdatas [][]byte
	vres := res.([]interface{})
	for _, datas := range vres {
		var vdata []byte
		for _, vbyte := range datas.([]byte) {
			vdata = append(vdata, vbyte)
		}
		vdatas = append(vdatas, vdata)
	}

	//ret := make([]string, len(keydata))
	//for i, k := range keydata {
	//	ret[i] = string(k)
	//}

	return vdatas, nil
}

func (client *RedisCli) Del(key string) (bool, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("DEL", key)
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, nil
}

// string 相关方法
func (client *RedisCli) Set(key string, val []byte) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("SET", key, string(val))
	if err != nil {
		return err
	}
	return nil
}

func (client *RedisCli) Get(key string) ([]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("GET", key)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, RedisError("Key `" + key + "` does not exist")
	}

	data := res.([]byte)
	return data, nil
}

// string 相关方法
func (client *RedisCli) Exists(key string) bool {
	cn := client.pool.Get()
	defer cn.Close()

	v, err := cn.Do("EXISTS", key)
	if err != nil {
		return false
	}
	return v.(int64) > 0
}

// list 相关方法
func (client *RedisCli) Llen(key string) (int, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("LLEN", key)
	if err != nil {
		return -1, err
	}
	return int(res.(int64)), nil
}

func (client *RedisCli) Lrange(key string, start int, end int) ([][]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("LRANGE", key, strconv.Itoa(start), strconv.Itoa(end))
	if err != nil {
		return nil, err
	}

	var vdatas [][]byte
	vres := res.([]interface{})
	for _, datas := range vres {
		var vdata []byte
		for _, vbyte := range datas.([]byte) {
			vdata = append(vdata, vbyte)
		}
		vdatas = append(vdatas, vdata)
	}

	return vdatas, nil
}

func (client *RedisCli) Lindex(key string, index int) ([]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("LINDEX", key, strconv.Itoa(index))
	if err != nil {
		return nil, err
	}
	return res.([]byte), nil
}

func (client *RedisCli) Lset(key string, index int, value []byte) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("LSET", key, strconv.Itoa(index), string(value))
	if err != nil {
		return err
	}
	return nil
}

func (client *RedisCli) Lpush(key string, val []byte) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("LPUSH", key, string(val))
	if err != nil {
		return err
	}
	return nil
}

func (client *RedisCli) BLpop(key string) ([]interface{}, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("blpop", key, 100)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	return res.([]interface{}), nil
}

func (client *RedisCli) Lpop(key string) ([]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("LPOP", key)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	return res.([]byte), nil
}

func (client *RedisCli) Rpush(key string, val []byte) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("RPUSH", key, string(val))

	if err != nil {
		return err
	}
	return nil
}

func (client *RedisCli) Rpop(key string) ([]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("RPOP", key)
	if err != nil {
		return nil, err
	}
	return res.([]byte), nil
}

func (client *RedisCli) Lrem(key string, value []byte) (int, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("LREM", key, 0, string(value))
	if err != nil {
		return -1, err
	}
	return int(res.(int64)), nil
}

//! sorted set 相关
func (client *RedisCli) Zadd(key string, value []byte, score float64) (bool, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("ZADD", key, strconv.FormatFloat(score, 'f', -1, 64), string(value))
	if err != nil {
		return false, err
	}

	return res.(int64) == 1, nil
}

func (client *RedisCli) Zrem(key string, value []byte) (bool, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("ZREM", key, string(value))
	if err != nil {
		return false, err
	}

	return res.(int64) == 1, nil
}

func (client *RedisCli) Zincrby(key string, value []byte, score float64) (float64, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("ZINCRBY", key, strconv.FormatFloat(score, 'f', -1, 64), string(value))
	if err != nil {
		return 0, err
	}

	data := string(res.([]byte))
	f, _ := strconv.ParseFloat(data, 64)
	return f, nil
}

func (client *RedisCli) Zrange(key string, start int, end int) ([][]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("ZRANGE", key, strconv.Itoa(start), strconv.Itoa(end))
	if err != nil {
		return nil, err
	}

	var vdatas [][]byte
	vres := res.([]interface{})
	for _, datas := range vres {
		if datas != nil {
			var vdata []byte
			for _, vbyte := range datas.([]byte) {
				vdata = append(vdata, vbyte)
			}
			vdatas = append(vdatas, vdata)
		}
	}

	return vdatas, nil
}

func (client *RedisCli) Zrevrange(key string, start int, end int) ([][]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("ZREVRANGE", key, strconv.Itoa(start), strconv.Itoa(end))
	if err != nil {
		return nil, err
	}

	var vdatas [][]byte
	vres := res.([]interface{})
	for _, datas := range vres {
		if datas != nil {
			var vdata []byte
			for _, vbyte := range datas.([]byte) {
				vdata = append(vdata, vbyte)
			}
			vdatas = append(vdatas, vdata)
		}
	}

	return vdatas, nil
}

func (client *RedisCli) ZrevrangeWithScore(key string, start int, end int) ([][]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("ZREVRANGE", key, strconv.Itoa(start), strconv.Itoa(end), "WITHSCORES")
	if err != nil {
		return nil, err
	}

	var vdatas [][]byte
	vres := res.([]interface{})
	for _, datas := range vres {
		if datas != nil {
			var vdata []byte
			for _, vbyte := range datas.([]byte) {
				vdata = append(vdata, vbyte)
			}
			vdatas = append(vdatas, vdata)
		}
	}

	return vdatas, nil
}

//! hash 相关

func (client *RedisCli) HSet(key string, field string, value []byte) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("HSET", key, field, string(value))
	if err != nil {
		return err
	}

	return nil
}

func (client *RedisCli) HGet(key string, field string) ([]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("HGET", key, field)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, RedisError("Key `" + key + "` does not exist")
	}
	return res.([]byte), nil
}

func (client *RedisCli) HExists(key string, field string) bool {
	cn := client.pool.Get()
	defer cn.Close()

	v, err := cn.Do("HEXISTS", key, field)
	if err != nil {
		return false
	}
	return v.(int64) > 0
}

func (client *RedisCli) HMsetall(key string, obj interface{}) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("HMSET", redis.Args{}.Add(key).AddFlat(obj)...)
	if err != nil {
		return err
	}

	return nil
}

func (client *RedisCli) HMset(key string, args ...interface{}) error {
	cn := client.pool.Get()
	defer cn.Close()

	args = append([]interface{}{key}, args...)
	_, err := cn.Do("HMSET", args...)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (client *RedisCli) HMget(key string, field string) ([][]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("HMGET", key, field)
	if err != nil {
		return nil, err
	}

	var vdatas [][]byte
	vres := res.([]interface{})
	for _, datas := range vres {
		if datas != nil {
			var vdata []byte
			for _, vbyte := range datas.([]byte) {
				vdata = append(vdata, vbyte)
			}
			vdatas = append(vdatas, vdata)
		}
	}

	return vdatas, nil
}

func (client *RedisCli) Hgetall(key string, obj interface{}) error {
	cn := client.pool.Get()
	defer cn.Close()

	//startAt := time.Now()
	v, err := redis.Values(cn.Do("HGETAll", key))
	if err != nil {
		return err
	}
	//xlog.Logger().Warn("Hgetall spend", time.Since(startAt).String())
	if len(v) == 0 {
		return RedisError(fmt.Sprintf("[key: %s] is not exists", key))
	}

	if err := redis.ScanStruct(v, obj); err != nil {
		return err
	}
	//xlog.Logger().Warn("scanStruct spend", time.Since(startAt).String())
	return nil
}

func (client *RedisCli) Hvals(key string) ([][]byte, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("HVALS", key)
	if err != nil {
		return nil, err
	}

	var vdatas [][]byte
	vres := res.([]interface{})
	for _, datas := range vres {
		var vdata []byte
		for _, vbyte := range datas.([]byte) {
			vdata = append(vdata, vbyte)
		}
		vdatas = append(vdatas, vdata)
	}

	return vdatas, nil
}

func (client *RedisCli) HDel(key string, field string) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("HDEL", key, field)

	return err
}

func (client *RedisCli) Hlen(key string) (int, error) {
	cn := client.pool.Get()
	defer cn.Close()

	res, err := cn.Do("HLEN", key)
	if err != nil {
		return -1, err
	}
	return int(res.(int64)), nil
}

func (client *RedisCli) HIncrBy(key, field string, count int64) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("HINCRBY", key, field, count)
	return err
}

//! 设置数据过期时间
func (client *RedisCli) Expire(key string, second int64) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("EXPIRE", key, second)

	return err
}

/* Stream */
// 添加
type XaddMsg struct {
	Key   string // 表名
	Value string // 数据结构json
}

// 写入数据到队列
func (client *RedisCli) Xadd(key string, object *XaddMsg) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("XADD", key, "*", object.Key, object.Value)
	if err != nil {
		return err
	}

	return nil
}

// 创建消费者组
func (client *RedisCli) Xgroup(key string, name string, id string) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("XGROUP", "CREATE", key, name, id)
	if err != nil {
		return err
	}

	return nil
}

// 从消费组队列读一条最新且未被消费的数据(消息到目前为止从未传递给其他消费者, 但是不会取到旧的历史消息)
func (client *RedisCli) Xreadgroup(groupName, customerName, key, id string) (*XaddMsg, string, error) {
	cn := client.pool.Get()
	defer cn.Close()

	datas, err := redis.Values(cn.Do("XREADGROUP", "GROUP", groupName, customerName, "block", 0, "streams", key, ">"))
	if err != nil {
		return nil, "", err
	}

	// 只取一条
	if len(datas) > 0 {
		var keyInfo = datas[0].([]interface{})
		var idList = keyInfo[1].([]interface{})

		if len(idList) > 0 {
			var idInfo = idList[0].([]interface{})
			var id = string(idInfo[0].([]byte))
			var fieldList = idInfo[1].([]interface{})
			var field = string(fieldList[0].([]byte))
			var value = string(fieldList[1].([]byte))

			object := &XaddMsg{
				Key:   field,
				Value: value,
			}
			return object, id, nil
		}
	}

	err = errors.New("can't find stream")
	return nil, "", err
}

// 从指定stream队列根据id顺序获取一条消息(阻塞)
func (client *RedisCli) Xread(key, id string) (*XaddMsg, string, error) {
	cn := client.pool.Get()
	defer cn.Close()

	datas, err := redis.Values(cn.Do("XREAD", "block", 0, "count", 1, "streams", key, id))
	if err != nil {
		return nil, "", err
	}

	// 只取一条
	if len(datas) > 0 {
		var keyInfo = datas[0].([]interface{})
		var idList = keyInfo[1].([]interface{})

		if len(idList) > 0 {
			var idInfo = idList[0].([]interface{})
			var id = string(idInfo[0].([]byte))
			var fieldList = idInfo[1].([]interface{})
			var field = string(fieldList[0].([]byte))
			var value = string(fieldList[1].([]byte))

			object := &XaddMsg{
				Key:   field,
				Value: value,
			}
			return object, id, nil
		}
	}

	err = errors.New("can't find stream")
	return nil, "", err
}

// 从consumer的已处理消息列表中删除
func (client *RedisCli) Xack(key, groupName, id string) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("XACK", key, groupName, id)
	if err != nil {
		return err
	}

	return nil
}

// 从队列删除数据
func (client *RedisCli) Xdel(key, id string) error {
	cn := client.pool.Get()
	defer cn.Close()

	_, err := cn.Do("XDEL", key, id)
	if err != nil {
		return err
	}

	return nil
}

// 集合相关操作
func (client *RedisCli) Sadd(key, val interface{}) error {
	cn := client.pool.Get()
	defer cn.Close()
	_, err := cn.Do("Sadd", key, val)
	return err
}

func (client *RedisCli) SADDS(key interface{}, val []int64) error {
	cn := client.pool.Get()
	defer cn.Close()
	interfaceStruct := []interface{}{key}
	for _, item := range val {
		interfaceStruct = append(interfaceStruct, item)
	}
	_, err := cn.Do("Sadd", interfaceStruct...)
	return err
}
func (client *RedisCli) SRem(key, val interface{}) error {
	cn := client.pool.Get()
	defer cn.Close()
	_, err := cn.Do("SREM", key, val)
	return err
}

func (client *RedisCli) SisMember(key, val interface{}) bool {
	cn := client.pool.Get()
	defer cn.Close()
	res, err := cn.Do("SISMEMBER", key, val)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return res.(int64) == 1
}

func (client *RedisCli) SMembers(key string) ([]interface{}, error) {
	cn := client.pool.Get()
	defer cn.Close()
	res, err := cn.Do("SMembers", key)
	if err != nil {
		fmt.Println(err)
	}
	return res.([]interface{}), nil

}
