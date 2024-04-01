package chat

import (
	"asocial/pkg/common"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
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
	r.mc.BroadcastOthers(data, sess)
}

func (r *HttpServer) HandleChatOnConnect(sess *melody.Session) {
	fmt.Println("connected, current session: ", sess.Request.RemoteAddr)
}

func homePage(c *gin.Context) {
	c.String(http.StatusOK, "This is my home page")
}

func response(c *gin.Context, httpCode int, err error) {
	message := err.Error()
	c.JSON(httpCode, common.ErrResponse{
		Message: message,
	})
}