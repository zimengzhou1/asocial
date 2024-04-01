//go:build wireinject
// +build wireinject

package wire

import (
	"asocial/pkg/chat"
	"asocial/pkg/common"
	"asocial/pkg/config"

	"github.com/google/wire"
)

func InitializeChatServer(name string) (*common.Server, error) {
	wire.Build(
		config.NewConfig,

		chat.NewMelodyChatConn,
		chat.NewGinServer,

		chat.NewHttpServer,
		wire.Bind(new(common.HttpServer), new(*chat.HttpServer)),
		chat.NewRouter,
		wire.Bind(new(common.Router), new(*chat.Router)),
		common.NewServer,
	)
  return &common.Server{}, nil
}
