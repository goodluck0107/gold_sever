package static

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const NoticeRedisKeyPrefix = "{noticedata}"

const (
	ShowTypeDaily     = 1 // 每天一次
	ShowTypeLogin     = 2 // 登录一次
	ShowTypeUniversal = 3 // 通用
)

type NoticePositionType int

const (
	NoticePositionTypeAll      NoticePositionType = 0
	NoticePositionTypeDialog   NoticePositionType = 1 // 大厅公告(弹窗)
	NoticePositionTypeMaintain NoticePositionType = 2 // 维护公告
	NoticePositionTypeMarquee  NoticePositionType = 3 // 跑马灯
	NoticePositionTypeOption   NoticePositionType = 4 // 功能
)

type NoticeMaintainServerType int

const (
	NoticeMaintainServerAllServer = iota - 1 // 所有服务器
	NoticeMaintainServerAllGame              // 所有游戏服
)

func (npt NoticePositionType) String() string {
	if b, err := npt.MarshalText(); err == nil {
		return string(b)
	} else {
		return "unknown"
	}
}

func ParsePosType(pt string) (NoticePositionType, error) {
	switch strings.ToLower(pt) {
	case "all":
		return NoticePositionTypeAll, nil
	case "dialog":
		return NoticePositionTypeDialog, nil
	case "maintain":
		return NoticePositionTypeMaintain, nil
	case "marquee":
		return NoticePositionTypeMarquee, nil
	case "option":
		return NoticePositionTypeOption, nil
	}
	var npt NoticePositionType
	return npt, fmt.Errorf("not a valid notice position type: %q", pt)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (npt *NoticePositionType) UnmarshalText(text []byte) error {
	l, err := ParsePosType(string(text))
	if err != nil {
		return err
	}

	*npt = NoticePositionType(l)

	return nil
}

func (npt NoticePositionType) MarshalText() ([]byte, error) {
	switch npt {
	case NoticePositionTypeAll:
		return []byte("all"), nil
	case NoticePositionTypeDialog:
		return []byte("dialog"), nil
	case NoticePositionTypeMaintain:
		return []byte("maintain"), nil
	case NoticePositionTypeMarquee:
		return []byte("marquee"), nil
	case NoticePositionTypeOption:
		return []byte("option"), nil
	}

	return nil, fmt.Errorf("not a valid logrus level %d", npt)
}

type NoticeList []*Notice

func (ns *NoticeList) MarshalBinary() (data []byte, err error) {
	return json.Marshal(ns)
}

func (ns *NoticeList) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, ns)
}

func (ns *NoticeList) RmExpired() {
	// ns.GameTimer()
	for i := 0; i < len(*ns); {
		n := (*ns)[i]
		if n == nil {
			i++
			continue
		}
		if !n.Start.IsZero() && time.Now().After(n.Start) && (n.End.IsZero() || n.End.After(time.Now())) {
			i++
		} else {
			copy((*ns)[i:], (*ns)[i+1:])
			*ns = (*ns)[:len(*ns)-1]
		}
	}
}

func (ns *NoticeList) WaitTime() time.Duration {
	var t time.Duration
	for _, n := range *ns {
		if !n.Flag {
			if tt := n.Start.Sub(time.Now()); tt > 0 {
				if t == 0 {
					t = tt
				} else if tt < t {
					t = tt
				}
			}
		}
	}
	return t
}

func (ns *NoticeList) Timer() {
	for _, n := range *ns {
		if n == nil {
			continue
		}
		n.Timer()
	}
}

func (ns NoticeList) ToMap() map[NoticePositionType]NoticeList {
	result := make(map[NoticePositionType]NoticeList)
	for _, n := range ns {
		if nl, ok := result[NoticePositionType(n.PositionType)]; ok {
			nl = append(nl, n)
			result[NoticePositionType(n.PositionType)] = nl
		} else {
			result[NoticePositionType(n.PositionType)] = NoticeList{n}
		}
	}
	return result
}

func (ns NoticeList) ToObjects() []interface{} {
	result := make([]interface{}, 0)
	for _, n := range ns {
		result = append(result, n)
	}
	return result
}

// 公告
type Notice struct {
	Id           int                      `json:"id"`             // id
	ContentType  int                      `json:"content_type"`   // 内容类型(文字 or 图片)
	Image        string                   `json:"image"`          // 图片内容
	Content      string                   `json:"content"`        // 文字内容
	PositionType int                      `json:"position_type"`  // 公告类型
	ShowType     int                      `json:"show_type"`      // 展示类型(每天一次 or 登录一次)
	StartAt      string                   `json:"start_at"`       // 开始时间
	EndAt        string                   `json:"end_at"`         // 结束时间
	KindId       int                      `json:"kind_id"`        // 关联子游戏
	GameServerId NoticeMaintainServerType `json:"game_server_id"` // 游戏服务器id
	Tag          string                   `json:"tag"`
	Start        time.Time                `json:"-"`
	End          time.Time                `json:"-"`
	Flag         bool                     `json:"flag"`       // 是否广播标志位
	UpdateAt     string                   `json:"updated_at"` // 更新时间
}

func (n *Notice) SetWorked() {
	n.Flag = true
}

func (n *Notice) IsWorked() bool {
	return n.Flag
}

func (n *Notice) MarshalBinary() (data []byte, err error) {
	return json.Marshal(n)
}

func (n *Notice) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, n)
}

func (n *Notice) Timer() {
	n.Start, _ = time.ParseInLocation("2006-01-02 15:04:05", n.StartAt, time.Local)
	n.End, _ = time.ParseInLocation("2006-01-02 15:04:05", n.EndAt, time.Local)
	// n.End, _ = time.ParseInLocation("2006-01-02 15:04:05", n.EndAt, time.Local)
}

func NoticeDataRedisKey() string {
	return fmt.Sprintf("%s*", NoticeRedisKeyPrefix)
}

func NoticePTypeRedisKey(positionType NoticePositionType) string {
	return fmt.Sprintf("%s:%s", NoticeRedisKeyPrefix, positionType)
}
