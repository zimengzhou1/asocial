package chat

import (
	"asocial/pkg/common"
)

type Router struct {
	httpServer common.HttpServer
}


func NewRouter(httpServer common.HttpServer) *Router {
	return &Router{httpServer}
}

func (r *Router) Run() {
	r.httpServer.RegisterRoutes()
	r.httpServer.Run()
}
