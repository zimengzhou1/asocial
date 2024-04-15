package chat

import (
	"asocial/pkg/common"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

var (
	sessCidKey = "sesscid"
)

func (r *HttpServer) StartChat(c *gin.Context) {
	fmt.Println("start socket")
	if err := r.mc.HandleRequest(c.Writer, c.Request); err != nil {
		fmt.Println("upgrade websocket error: ", err.Error())
		response(c, http.StatusInternalServerError, common.ErrServer)
		return
	}
}

func (r *HttpServer) HandleChatOnMessage(sess *melody.Session, data []byte) {
	fmt.Printf("received message: %s from session: %s\n", data, sess.Request.RemoteAddr)

	msg, err := DecodeToMessage(data)
	if err != nil {
		fmt.Println("decode message error: ", err.Error())
		return
	}

	if err := r.msgSvc.BroadcastTextMessage(context.Background(), msg); err != nil {
		fmt.Println("broadcast message error: ", err.Error())
		return
	}

	//r.mc.BroadcastOthers(data, sess)


}

func (r *HttpServer) HandleChatOnConnect(sess *melody.Session) {
	fmt.Println("connected, current session: ", sess.Request.RemoteAddr)
	userID := sess.Request.URL.Query().Get("uid")

	// temporary hard code
	channelID := "default"
	err := r.initializeChatSession(sess, channelID, userID)
	if err != nil {
		fmt.Println("initialize chat session error: ", err.Error())
		return
	}
}

func (r *HttpServer) initializeChatSession(sess *melody.Session, channelID, userID string) error {
	sess.Set(sessCidKey, channelID)
	sess.Set("user", userID)
	return nil
}

func (r *HttpServer) HandleChatOnClose(sess *melody.Session, i int, s string) error {
	fmt.Println("session closed: ", sess.Request.RemoteAddr)
	return nil
}

func response(c *gin.Context, httpCode int, err error) {
	message := err.Error()
	c.JSON(httpCode, common.ErrResponse{
		Message: message,
	})
}