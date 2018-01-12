package main

import (
	"github.com/XiBao/csc/rest"
	ws "github.com/XiBao/csc/websocket"
	//"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/igm/sockjs-go/sockjs"
	jsonp "github.com/jim3ma/gin-jsonp"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(jsonp.Handler())
	// 设置websocket路由
	sockjs_router(r)
	// 设rest路由
	rest_router(r)
	// 设置client路由
	client_router(r)

	status_router(r)

	return r
}

func sockjs_router(r *gin.Engine) {
	r_socket := r.Group("/chat")
	{
		r_socket.GET("/visitor/*action", func(c *gin.Context) {
			h := sockjs.NewHandler(`/chat/visitor/[0-9]+`, sockjs.DefaultOptions, ws.VisitorChatHandler)
			h.ServeHTTP(c.Writer, c.Request)
		})
		r_socket.GET("/kefu/*action", func(c *gin.Context) {
			h := sockjs.NewHandler(`/chat/kefu/[0-9]+`, sockjs.DefaultOptions, ws.KefuChatHandler)
			h.ServeHTTP(c.Writer, c.Request)
		})
	}
}

func client_router(r *gin.Engine) {
	//r.Use(static.Serve("/client/", static.LocalFile("./client", true)))
}

// 设置rest路由
func rest_router(r *gin.Engine) {
	r_rest_user := r.Group("/rest/user")
	{
		r_rest_user.GET("/list", rest.GetUserList)
		r_rest_user.GET("/info", rest.GetKefuInfo)
		r_rest_user.GET("/kefu", rest.SetKefuOnline)
		r_rest_user.GET("/manage", rest.SetKefuManage)
		r_rest_user.GET("/call", rest.KefuCall)
	}
	r_rest_chat := r.Group("/rest/chat")
	{
		r_rest_chat.GET("/transfer", rest.TransferChat)
		r_rest_chat.GET("/stop", rest.StopChat)
	}
}

func status_router(r *gin.Engine) {
	r.LoadHTMLFiles("/var/code/go/src/github.com/XiBao/csc/templates/status.html")
	r.GET("/status", func(c *gin.Context) {
		c.HTML(200, "status.html", gin.H{
			"Title": "CSC status report",
			"Items": ChatCenter.Report(),
		})
	})
}
