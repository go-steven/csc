package rest

import (
	"github.com/XiBao/csc/chat"
	"github.com/gin-gonic/gin"
	log "github.com/kdar/factorlog"
)

var (
	logger     *log.FactorLog
	chatCenter *chat.ChatCenter
)

// Set to use a specific logger
func SetLogger(alogger *log.FactorLog) {
	logger = alogger
}

func SetChatCenter(center *chat.ChatCenter) {
	chatCenter = center
}

func json_response(c *gin.Context, data interface{}, err error) {
	if err != nil {
		logger.Error(err.Error())
		c.JSON(200, gin.H{
			"data":  nil,
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"data":  data,
		"error": nil,
	})
}
