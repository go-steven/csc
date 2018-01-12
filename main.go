package main

import (
	"flag"
	"fmt"
	"github.com/XiBao/csc/chat"
	"github.com/XiBao/csc/model"
	"github.com/XiBao/csc/rest"
	ws "github.com/XiBao/csc/websocket"
	"github.com/bububa/goconfig/config"
	"github.com/bububa/mymysql/autorc"
	_ "github.com/bububa/mymysql/thrsafe"
	"github.com/gin-gonic/gin"
	"math/rand"
	"runtime"
	"time"
)

const (
	_CONFIG_FILE = "/var/code/go/config.cfg"
)

var (
	logFlag  = flag.String("log", "", "set log path")
	portFlag = flag.Int("port", 10080, "set port")

	MasterDB   *autorc.Conn
	ChatCenter *chat.ChatCenter
)

func init() {
	cfg, _ := config.ReadDefault(_CONFIG_FILE)

	hostMaster, _ := cfg.String("masterdb", "host")
	userMaster, _ := cfg.String("masterdb", "user")
	passwdMaster, _ := cfg.String("masterdb", "passwd")
	dbnameMaster, _ := cfg.String("masterdb", "dbname")

	MasterDB = autorc.New("tcp", "", hostMaster, userMaster, passwdMaster, dbnameMaster)
	MasterDB.Register("set names utf8")
	MasterDB.Reconnect()

	ChatCenter = chat.NewChatCenter()

	rand.Seed(time.Now().UnixNano())
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	logger = setGlobalLogger(*logFlag)

	chat.SetLogger(logger)
	ws.SetLogger(logger)
	rest.SetLogger(logger)
	model.SetLogger(logger)

	chat.SetDb(MasterDB)
	ws.SetDb(MasterDB)
	chat.SetChatCenter(ChatCenter)
	ws.SetChatCenter(ChatCenter)
	rest.SetChatCenter(ChatCenter)

	gin.SetMode(gin.ReleaseMode)
	router := NewRouter()
	router.Use(gin.Recovery())

	logger.Infof("Go CSC Server started at:0.0.0.0:%d", *portFlag)
	defer func() {
		logger.Infof("Go CSC Server exit from:0.0.0.0:%d", *portFlag)
		if MasterDB != nil && MasterDB.Raw.IsConnected() {
			MasterDB.Reconnect()
			MasterDB.Raw.Close()
		}
	}()
	// listen and server on 0.0.0.0:8180
	logger.Fatal(router.Run(fmt.Sprintf(":%d", *portFlag)))
}
