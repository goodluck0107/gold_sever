package engine

import "context"

type IServer interface {
	Name() string

	Host() (string, string)

	Start(ctx context.Context) error

	Stop(ctx context.Context) error

	RegisterRouter(ctx context.Context)

	RegisterRPC(ctx context.Context)

	LoadConfig(ctx context.Context) error

	LoadServerConfig(ctx context.Context) error

	SetLoggerLevel()

	Run(ctx context.Context)
}
