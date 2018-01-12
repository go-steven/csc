package rest

import (
	"errors"
	"github.com/XiBao/csc/model"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 客服上线/离线
// http://csc.xibao100.com/rest/user/kefu?kefuid=$KEFUID&online=$ONLINE&callback=$CALLBACK
//   $ONLINE: 1-上线，0：下线
func SetKefuOnline(c *gin.Context) {
	kefuid, _ := strconv.ParseUint(c.Query("kefuid"), 10, 64)
	if kefuid == 0 {
		json_response(c, nil, errors.New("No kefuid parameter or it is zero."))
		return
	}

	online, _ := strconv.ParseUint(c.Query("online"), 10, 64)
	var status uint = model.OFFLINE
	if online > 0 {
		status = model.ONLINE
	}

	kefu := chatCenter.FindKefu(kefuid)
	if kefu == nil {
		json_response(c, nil, errors.New("Not found online kefu."))
		return
	}

	kefu.LastManualStatus = status
	if kefu.Status != status {
		kefu.SetStatus(status)
	}
	logger.Infof("Call rpc: user/kefu, kefuid:%d, online:%d, return OK.", kefuid, online)

	json_response(c, "OK", nil)
}
