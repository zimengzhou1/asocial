package chat

import (
	"asocial/pkg/config"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

var (
	MelodyChat MelodyChatConn
)

type MelodyChatConn struct {
	*melody.Melody
}

type HttpServer struct {
	name          string
	svr           *gin.Engine
	mc            MelodyChatConn
	httpPort      string
	httpServer    *http.Server
}

func NewMelodyChatConn(config *config.Config) MelodyChatConn {
	m := melody.New()
	m.Config.MaxMessageSize = config.Chat.Message.MaxSizeByte
	MelodyChat = MelodyChatConn{
		m,
	}
	return MelodyChat
}

func NewGinServer() *gin.Engine {
	router := gin.Default()
	return router
}

func NewHttpServer(name string, config *config.Config, svr *gin.Engine, mc MelodyChatConn) *HttpServer {
	return &HttpServer{
		name:          name,
		svr:           svr,
		mc:            mc,
		httpPort:      config.Chat.Http.Server.Port,
	}
}

func (r *HttpServer) RegisterRoutes() {
	r.svr.GET("/api/chat", r.StartChat)
	r.svr.GET("/", homePage)

	r.mc.HandleMessage(r.HandleChatOnMessage)
	r.mc.HandleConnect(r.HandleChatOnConnect)
}

func (r *HttpServer) Run() {
	addr := ":" + r.httpPort
	r.httpServer = &http.Server{
		Addr:    addr,
		Handler: r.svr, // done to add obervability later
	}
	fmt.Printf("http server listening addr %v", addr)
	err := r.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("http server error: %v", err)
		os.Exit(1)
	}
}