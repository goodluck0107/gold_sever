package user

import (
	"context"
	network2 "github.com/open-source/game/chess.git/pkg/rpc"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/smallnest/rpcx/client"
)

var rpcCliMap map[string]client.XClient

func InitByIP(ip string) {
	if rpcCliMap == nil {
		rpcCliMap = make(map[string]client.XClient, 1)
	}
	rpcCliMap[ip] = initByIP(ip)
}

func initByIP(ip string) client.XClient {
	return network2.InitClient(ip, "ServerMethod", nil)
}
func getCliByIP(ip string) client.XClient {
	if rpcCliMap == nil {
		rpcCliMap = make(map[string]client.XClient, 1)
	}
	cli, ok := rpcCliMap[ip]
	if ok {
		return cli
	}
	cli = initByIP(ip)
	rpcCliMap[ip] = initByIP(ip)
	return cli
}

// CallHall 调用大厅服
func CallHall(args *static.Rpc_Args) []byte {
	reply := []byte{}
	hallCli := getCliByIP(GetServer().Con.Center)
	err := hallCli.Call(context.Background(), "ServerMsg", args, &reply)
	if err != nil {
		xlog.Logger().Errorf("failed to call: %v\n", err)
	}

	return reply
}
