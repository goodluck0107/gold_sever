package rpc

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"net"
	"strings"
	"time"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
)

func NewPeer2PeerDiscovery(server, metadata string) client.ServiceDiscovery {
	ret, _ := client.NewPeer2PeerDiscovery(server, metadata)
	return ret
}

func InitClient(ipstr, serPath string, inPlug interface{}) client.XClient {
	ip := strings.Split(ipstr, ":")
	if len(ip) < 2 {
		return nil
	}
	opt := client.DefaultOption
	opt.Heartbeat = false
	opt.HeartbeatInterval = 3 * time.Second
	opt.Retries = 1
	cli := client.NewXClient(serPath, client.Failtry, client.RandomSelect,
		NewPeer2PeerDiscovery(fmt.Sprintf("tcp@%s:1%s", ip[0], ip[1]), ""), opt)
	plug := client.NewPluginContainer()
	if inPlug == nil {
		inPlug = &CliPlugin{}
	}
	plug.Add(inPlug)
	cli.SetPlugins(plug)
	return cli
}

func NewRpcxServer(ipstr, serPath string, rvcr, inPlug interface{}) {
	s := server.NewServer()
	if inPlug == nil {
		inPlug = &CliPlugin{}
	}
	s.Plugins.Add(inPlug)
	ip := strings.Split(ipstr, ":")
	s.RegisterName(serPath, rvcr, "")
	err := s.Serve("tcp", fmt.Sprintf("0.0.0.0:1%s", ip[1]))
	if err != nil {
		if server.ErrServerClosed != err {
			panic(err)
		}
		xlog.Logger().Error("closed:", err)
	}
}

//CliPlugin rpcx客户端组件
type CliPlugin struct {
}

func (p *CliPlugin) ClientConnected(conn net.Conn) (net.Conn, bool) {
	xlog.Logger().Infof("wuhan %v connected", conn.RemoteAddr().String())
	return conn, true
}

func (p *CliPlugin) ClientConnectionClose(conn net.Conn) bool {
	xlog.Logger().Infof("wuhan %v closed", conn.RemoteAddr().String())
	return true
}

type ServerPlugin struct {
}

func (p *ServerPlugin) HandleConnAccept(conn net.Conn) (net.Conn, bool) {
	xlog.Logger().Infof("client %v connected", conn.RemoteAddr().String())
	return conn, true
}

func (p *ServerPlugin) HandleConnClose(conn net.Conn) bool {
	xlog.Logger().Infof("client %v closed", conn.RemoteAddr().String())
	return true
}
