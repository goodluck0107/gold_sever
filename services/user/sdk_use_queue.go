package user

import (
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static/chanqueue"
	"github.com/open-source/game/chess.git/pkg/static/chanqueue/proto"
)

type QueueHandler struct {
	pool map[string]*chanqueue.Worker
}

// 初始化HTTP服务
func (qh *QueueHandler) OnInit() {
	qh.pool = make(map[string]*chanqueue.Worker)
	qh.Register(proto.MsgHeadQiNiuUserImg, &proto.QiNiuUserImgMsg{}, QueueFetchUserImgToQiNiu)
}

func (qh *QueueHandler) Register(header string, v interface{}, handlerFunc chanqueue.HandlerFunc) {
	qh.pool[header] = chanqueue.NewWorker(v, handlerFunc)
}

func (qh *QueueHandler) WorkerByProto(proto string) (*chanqueue.Worker, bool) {
	worker, ok := qh.pool[proto]
	return worker, ok
}

func QueueFetchUserImgToQiNiu(v interface{}) error {
	if !GetServer().ConQiNiu.Able {
		return nil
	}

	// 这里可以保证断言成功，chan queue包做了proto的类型校验
	msg := v.(*proto.QiNiuUserImgMsg)

	// 生成七牛图片下载地址
	newURL, err := service2.GetQiNiuClient(GetServer().ConQiNiu).FetchUserImg(msg.Uid, msg.ImgUrl)
	if err != nil {
		return err
	}

	// 检测图片是否违规
	if !service2.GetQiNiuClient(GetServer().ConQiNiu).CheckUserImg(msg.Uid, newURL) {
		err = service2.GetQiNiuClient(GetServer().ConQiNiu).DeleteUserImg(msg.Uid)
		if err != nil {
			return nil
		}
		newURL = ""
	}

	// 更新用户头像
	err = GetDBMgr().GetDBmControl().Model(&models.User{}).Where("id = ?", msg.Uid).Update("imgurl", newURL).Error
	if err != nil {
		return err
	}

	// 更新用户头像
	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(msg.Uid, "Imgurl", newURL)
	if err != nil {
		return err
	}

	return nil
}
