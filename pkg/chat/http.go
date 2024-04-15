package chat

import (
	"asocial/pkg/config"
	"context"
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
	msgSvc        MessageService
	msgSubscriber *MessageSubscriber
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

func NewHttpServer(name string, config *config.Config, svr *gin.Engine, mc MelodyChatConn, msgSvc MessageService, msgSubscriber *MessageSubscriber) *HttpServer {
	return &HttpServer{
		name:          name,
		svr:           svr,
		mc:            mc,
		httpPort:      config.Chat.Http.Server.Port,
		msgSvc:        msgSvc,
		msgSubscriber: msgSubscriber,
	}
}

func (r *HttpServer) RegisterRoutes() {
	r.msgSubscriber.RegisterHandler()

	r.svr.GET("/api/chat", r.StartChat)

	r.mc.HandleMessage(r.HandleChatOnMessage)
	r.mc.HandleConnect(r.HandleChatOnConnect)
	r.mc.HandleClose(r.HandleChatOnClose)
}

func (r *HttpServer) Run() {
	go func() {
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
	}()
	go func() {
		err := r.msgSubscriber.Run()
		if err != nil {
			fmt.Printf("message subscriber error: %v", err)
			os.Exit(1)
		}
	}()
}

func (r *HttpServer) GracefulStop(ctx context.Context) error {
	err := MelodyChat.Close()
	if err != nil {
		return err
	}
	err = r.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}
	err = r.msgSubscriber.GracefulStop()
	if err != nil {
		return err
	}
	return nil
}