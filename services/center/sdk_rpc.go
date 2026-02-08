package center

import (
	"context"
	network2 "github.com/open-source/game/chess.git/pkg/rpc"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"net"
	"runtime/debug"
	"time"

	"github.com/smallnest/rpcx/client"
)

func InitByIP(ip string) *client.XClient {
	cli := network2.InitClient(ip, "ServerMethod", &CliPlugin{})
	return &cli
}

func AddGameServer(ip string) *client.XClient {
	return InitByIP(ip)
}

func CallGame(args *static.Rpc_Args, cli *client.XClient, gameId int) ([]byte, error) {
	reply := []byte{}

	Done := make(chan error, 1)
	c, cancel := context.WithCancel(context.Background())
	go func(s []byte) {
		select {
		case <-time.After(3 * time.Second):
			cancel()
			xlog.Logger().Errorf("error:cli:%+v", cli)
			xlog.Logger().Errorf("error:time out:%s", s)
		case err := <-Done:
			if err != nil {
				reply = []byte("ERR")
				xlog.Logger().Errorf("call rpc done with error: %s", err)
			} else {
				xlog.Logger().Debug("call rpc suc")
			}
			break
		}
	}(debug.Stack())
	err := (*cli).Call(c, "ServerMsg", args, &reply)
	if err != nil {
		xlog.Logger().Errorf("first call rpc error:%s, go to reconnect...", err)
	}
	Done <- err
	return reply, err
}

//CliPlugin rpcx客户端组件
type CliPlugin struct {
}

func (p *CliPlugin) ClientConnected(conn net.Conn) (net.Conn, error) {
	xlog.Logger().Infof("wuhan %v connected", conn.RemoteAddr().String())
	return conn, nil
}

func (p *CliPlugin) ClientConnectionClose(conn net.Conn) error {
	xlog.Logger().Infof("wuhan %v closed", conn.RemoteAddr().String())
	return nil
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
