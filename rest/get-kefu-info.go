package rest

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 查询在线客服信息
// http://csc.xibao100.com/rest/user/info?kefuid=$KEFUID&callback=$CALLBACK
func GetKefuInfo(c *gin.Context) {
	kefuid, _ := strconv.ParseUint(c.Query("kefuid"), 10, 64)
	if kefuid == 0 {
		json_response(c, nil, errors.New("No kefuid parameter or it is zero."))
		return
	}

	kefu := chatCenter.FindKefu(kefuid)
	if kefu == nil {
		json_response(c, nil, errors.New("Not found online kefu."))
		return
	}

	logger.Infof("Call rpc: user/info, kefuid:%d, return OK.", kefuid)
	json_response(c, kefu.User.Copy(), nil)
}
