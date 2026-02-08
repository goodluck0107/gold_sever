// 一个yaml解析工具，目前的作用是代替目前所有的服务器配置json文件
package static

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

const (
	ServerNameGate    = "gate"
	ServerNameBackend = "backend"
	ServerNameUser    = "user"
	ServerNameCenter  = "center"
	ServerNameSport   = "sport"
	ServerNameApi     = "api"
)

type Server_Yamler interface {
	Name() string
	Etc() interface{}
}

func LoadServeYaml(server Server_Yamler, need_service bool) error {
	return new(ServerYaml).yaml2JsonStruct(server.Name(), need_service, server.Etc())
}

/*
 * @Description //把yaml配置数据转换为二进制数据在json解析到服务器配置结构体
 * @Date 17:22 2019/2/14
 * @Param courseName 进程名
 * @Param b_service  是否需要第三方服务配置
 * @Param structdst  结构体指针
 * @return
 **/
func (yaml *ServerYaml) yaml2JsonStruct(courseName string, b_service bool, structdst interface{}) error {
	if err := yaml.load(); err != nil {
		return fmt.Errorf("Yaml2JsonStruct:%s", err)
	}
	data, err := yaml.configByName(courseName)
	if err != nil {
		return fmt.Errorf("Yaml2JsonStruct:%s", err)
	}
	if b_service {
		if err := json.Unmarshal(yaml.serviceConfig(), structdst); err != nil {
			return fmt.Errorf("Yaml2JsonStruct:%s", err)
		}
	}
	return json.Unmarshal(data, structdst)
}

/*
 * @Description //对应 wuhan.yaml 配置文件的结构体
 * @Date 17:03 2019/2/14
 **/
type ServerYaml struct {
	// 第三方服务配置
	Service struct {
		Wxappid     string `yaml:"wxappid" json:"wxappid"`
		Wxappsecert string `yaml:"wxappsecert" json:"wxappsecert"`
		Smshost     string `yaml:"smshost" json:"smshost"`
		Adminhost   string `yaml:"adminhost" json:"adminhost"`
		DrobotToken string `yaml:"drobot" json:"drobot"`
	} `yaml:"service"`
	// 各服务器配置 进程名> 映射 >配置结构
	ServerConfig map[string]*struct {
		ID              int    `yaml:"id" json:"id"`
		Port            int    `yaml:"port" json:"port"`
		Safehost        string `yaml:"safehost" json:"safehost"`
		Usesafeip       int    `yaml:"usesafeip" json:"usesafeip"`
		Host            string `yaml:"host" json:"host"`
		Inhost          string `yaml:"inhost" json:"inhost"`
		Gate            string `yaml:"gate" json:"gate"`
		User            string `yaml:"user" json:"user"`
		SafeCenter      string `yaml:"safecenter" json:"safecenter"`
		Center          string `yaml:"center" json:"center"`
		Api             string `yaml:"api" json:"api"`
		Incenter        string `yaml:"incenter" json:"incenter"`
		EnabledYk       bool   `yaml:"enabledyk" json:"enabledyk"`
		Encode          int    `yaml:"encode" json:"encode"`
		Encodeclientkey string `yaml:"encodeclientkey" json:"encodeclientkey"`
		Encodephpkey    string `yaml:"encodephpkey" json:"encodephpkey"`
		Redis           string `yaml:"redis" json:"redis"`
		Redisdb         int    `yaml:"redisdb" json:"redisdb"`
		Redisauth       string `yaml:"redisauth" json:"redisauth"`
		PubRedis        string `yaml:"pubredis" json:"pubredis"`
		PubRedisdb      int    `yaml:"pubredisdb" json:"pubredisdb"`
		PubRedisauth    string `yaml:"pubredisauth" json:"pubredisauth"`
		StoreRedis      string `yaml:"storeredis" json:"storeredis"`
		StoreRedisdb    int    `yaml:"storeredisdb" json:"storeredisdb"`
		StoreRedisauth  string `yaml:"storeredisauth" json:"storeredisauth"`
		Rcredis         string `yaml:"rcredis" json:"rcredis"`
		Rcredisdb       int    `yaml:"rcredisdb" json:"rcredisdb"`
		DB              string `yaml:"db" json:"db"`
		CppDB           string `yaml:"cppdb" json:"cppdb"`
		DB_mysql        struct {
			User      string `yaml:"user"`
			PassWord  string `yaml:"password"`
			Port      int    `yaml:"port"`
			Addr      string `yaml:"addr"`
			DBname    string `yaml:"dbname"`
			TimeOut   string `yaml:"timeout"`
			ParseTime bool   `yaml:"parsetime"`
			Loc       string `yaml:"loc"`
			Charset   string `yaml:"charset"`
		} `yaml:"db_mysql"`

		DB_sqlserver struct {
			User     string `yaml:"user"`
			PassWord string `yaml:"password"`
			Port     int    `yaml:"port"`
			Addr     string `yaml:"addr"`
			DBname   string `yaml:"dbname"`
			Encrypt  string `yaml:"encrypt"`
			AppName  string `yaml:"appname"`
		} `yaml:"db_sqlserver"`

		DB_redis struct {
			Auth     string `yaml:"auth"`
			SelectDB int    `yaml:"db"`
			Addr     string `yaml:"addr"`
			Port     int    `yaml:"port"`
		} `yaml:"db_redis"`
		DB_Pubredis struct {
			Auth     string `yaml:"auth"`
			SelectDB int    `yaml:"db"`
			Addr     string `yaml:"addr"`
			Port     int    `yaml:"port"`
		} `yaml:"db_pubredis"`
		DB_Storeredis struct {
			Auth     string `yaml:"auth"`
			SelectDB int    `yaml:"db"`
			Addr     string `yaml:"addr"`
			Port     int    `yaml:"port"`
		} `yaml:"db_storeredis"`
	} `yaml:"server_config"`
}

/*
 * @Description //根据进程名得到其配置的二进制数据
 * @Date 17:06 2019/2/14
 * @Param courseName 进程名
 * @return
 **/
func (self *ServerYaml) configByName(courseName string) ([]byte, error) {
	svr, ok := self.ServerConfig[courseName]
	if !ok {
		return nil, fmt.Errorf("configByName:No config named: %s", courseName)
	}
	return HF_JtoB(svr), nil
}

/*
 * @Description //得到配置的第三方服务数据
 * @Date 17:06 2019/2/14
 * @return
 **/
func (self *ServerYaml) serviceConfig() []byte {
	return HF_JtoB(&self.Service)
}

/*
 * @Description //加载配置
 * @Date 17:06 2019/2/14
 * @Param
 * @return
 **/
func (self *ServerYaml) load() error {
	fp, err := filepath.Abs("../../configs/etc/")
	if err != nil {
		return err
	}
	if err := HF_ReadYaml(fp, "server.yaml", self); err != nil {
		return fmt.Errorf("load error:%s", err)
	}
	return self.convert()
}

/*
 * @Description //转换配置 使兼容当前需要配置
 * @Date 17:06 2019/2/14
 * @Param
 * @return
 **/
func (self *ServerYaml) convert() error {
	var centerhost, centerInhost, userhost, gatehost, apihost string
	for k, v := range self.ServerConfig {
		// 服务器主机地址
		v.Host = fmt.Sprintf("%s:%d", v.Host, v.Port)
		// 服务器内网地址
		v.Inhost = fmt.Sprintf("%s:%d", v.Inhost, v.Port)
		// 服务器安全地址
		v.Safehost = fmt.Sprintf("%s:%d", v.Safehost, v.Port)
		// redis 地址
		v.Redis = fmt.Sprintf("%s:%d", v.DB_redis.Addr, v.DB_redis.Port)
		// redis 库
		v.Redisdb = v.DB_redis.SelectDB
		// redis auth
		v.Redisauth = v.DB_redis.Auth

		// redis 地址
		v.PubRedis = fmt.Sprintf("%s:%d", v.DB_Pubredis.Addr, v.DB_Pubredis.Port)
		// redis 库
		v.PubRedisdb = v.DB_Pubredis.SelectDB
		// redis auth
		v.PubRedisauth = v.DB_Pubredis.Auth

		// redis 地址
		v.StoreRedis = fmt.Sprintf("%s:%d", v.DB_Storeredis.Addr, v.DB_Storeredis.Port)
		// redis 库
		v.StoreRedisdb = v.DB_Storeredis.SelectDB
		// redis auth
		v.StoreRedisauth = v.DB_Storeredis.Auth

		// 数据库
		v.DB = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&timeout=%s&parseTime=%t&loc=%s",
			v.DB_mysql.User,
			v.DB_mysql.PassWord,
			v.DB_mysql.Addr,
			v.DB_mysql.Port,
			v.DB_mysql.DBname,
			v.DB_mysql.Charset,
			v.DB_mysql.TimeOut,
			v.DB_mysql.ParseTime,
			v.DB_mysql.Loc,
		)
		v.CppDB = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&&encrypt=%s&app+name=%s",
			v.DB_sqlserver.User,
			v.DB_sqlserver.PassWord,
			v.DB_sqlserver.Addr,
			v.DB_sqlserver.Port,
			v.DB_sqlserver.DBname,
			v.DB_sqlserver.Encrypt,
			v.DB_sqlserver.AppName,
		)
		// v.DB = "root:root@tcp(192.168.1.169:3306)/kygame?charset=utf8mb4&timeout=10s&parseTime=true&loc=Local"
		if k == ServerNameCenter {
			centerhost = v.Host
			centerInhost = v.Inhost
		}

		if k == ServerNameUser {
			userhost = v.Host
		}

		if k == ServerNameGate {
			gatehost = v.Host
		}

		if k == ServerNameApi {
			apihost = v.Host
		}
	}

	for _, v := range self.ServerConfig {
		// 大厅地址
		v.Center = centerhost
		// 内网大厅地址
		v.Incenter = centerInhost
		// 登录服地址
		v.User = userhost
		// 大厅服安全地址
		v.SafeCenter = centerhost
		// 网关地址
		v.Gate = gatehost
		// api 服务器地址
		v.Api = apihost
		//fmt.Println(fmt.Sprintf("%+v", v))
	}

	return nil
}
