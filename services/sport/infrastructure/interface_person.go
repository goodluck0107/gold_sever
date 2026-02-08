package infrastructure

import (
	"github.com/open-source/game/chess.git/pkg/static"
)

type PersonBase interface {
	GetInfo() static.Person
	GetIp() string
}
