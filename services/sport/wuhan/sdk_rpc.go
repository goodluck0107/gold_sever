package wuhan

import (
	"context"
	"encoding/json"
	"github.com/open-source/game/chess.git/pkg/consts"
	network2 "github.com/open-source/game/chess.git/pkg/rpc"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"net"
	"runtime/debug"
	"time"

	"github.com/smallnest/rpcx/client"
)

var rpcCliMap map[string]client.XClient
var isFirstConn bool = true

//InitCliByIP 初始化rpc链接，并保存在rpcxmap中
func InitCliByIP(ip string) {
	if rpcCliMap == nil {
		rpcCliMap = make(map[string]client.XClient)
	}
	rpcCliMap[ip] = initByIP(ip)
	onGameStart()
}

func initByIP(ip string) client.XClient {
	cli := network2.InitClient(ip, "ServerMethod", &ConnectionPlugin{})
	return cli
}

//ConnectionPlugin rpcx 回调插件
type ConnectionPlugin struct {
}

//ClientConnected 客户端连接成功
func (p *ConnectionPlugin) ClientConnected(conn net.Conn) (net.Conn, error) {
	if !isFirstConn {
		onGameStart()
	}
	return conn, nil
}

//ClientConnectionClose 客户端断开连接
func (p *ConnectionPlugin) ClientConnectionClose(conn net.Conn) error {
	isFirstConn = false
	go onGameClose()
	return nil
}
func getHallCli() client.XClient {
	if rpcCliMap == nil {
		rpcCliMap = make(map[string]client.XClient, 1)
		hallCli := initByIP(GetServer().Con.Center)
		rpcCliMap["hall"] = hallCli
		return hallCli
	}
	hallCli, ok := rpcCliMap["hall"]
	if ok {
		return hallCli
	}
	hallCli = initByIP(GetServer().Con.Center)
	rpcCliMap["hall"] = hallCli
	return hallCli
}

//CallHall 调用大厅服
func CallHall(args *static.Msg_MsgBase, reply *[]byte) error {
	buf, err := json.Marshal(args)
	if err != nil {
		xlog.Logger().Errorf("json 解析失败:%v", err)
		return err
	}
	hallCli := getHallCli()
	Done := make(chan struct{}, 1)
	c, cancle := context.WithCancel(context.Background())
	go func(s []byte) {
		select {
		case <-time.After(3 * time.Second):
			cancle()
			xlog.Logger().Errorf("error:time out:%s", s)
		case <-Done:
			break
		}
	}(debug.Stack())
	err = hallCli.Call(c, "ServerMsg", static.Rpc_Args{buf}, reply)
	if err != nil {
		xlog.Logger().Errorf("failed to call: %v", err)
	}
	Done <- struct{}{}
	return err
}

func onGameStart() {
	go func() {
		for {
			time.Sleep(3 * time.Second)
			err := GetServer().RegistGame(consts.ServerStatusOnline)
			if err == nil {
				break
			}
		}
	}()
	return
}

func onGameClose() {
	//GetServer().RegistGame(constant.ServerStatusOffline)
	for {
		time.Sleep(3 * time.Second)
		err := GetServer().RegistGame(consts.ServerStatusOnline)
		if err == nil {
			break
		}
	}
	return
}

//ServerPlugin rpcx服务组件
type ServerPlugin struct {
}

func (p *ServerPlugin) HandleConnAccept(conn net.Conn) (net.Conn, error) {
	xlog.Logger().Infof("client %v connected", conn.RemoteAddr().String())
	return conn, nil
}

func (p *ServerPlugin) HandleConnClose(conn net.Conn) error {
	xlog.Logger().Infof("client %v closed", conn.RemoteAddr().String())
	return nil
}
