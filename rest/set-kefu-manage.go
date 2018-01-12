package rest

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 客服设置自动分配标志
// http://csc.test.xibao100.com/rest/user/manage?kefuid=$KEFUID&auto=$AUTO&callback=$CALLBAC
//    $AUTO: 1-自动分配，0：非自动分配
func SetKefuManage(c *gin.Context) {
	kefuid, _ := strconv.ParseUint(c.Query("kefuid"), 10, 64)
	if kefuid == 0 {
		json_response(c, nil, errors.New("No kefuid parameter or it is zero."))
		return
	}

	auto, _ := strconv.ParseUint(c.Query("auto"), 10, 64)
	var manager uint
	if auto > 0 {
		manager = 1
	}

	kefu := chatCenter.FindKefu(kefuid)
	if kefu == nil {
		json_response(c, nil, errors.New("Not found online kefu."))
		return
	}

	kefu.SetManagerSts(manager)

	if kefu.IsOnline() && manager == 1 {
		chatCenter.DispatchNoKefuVisitors()
	}

	logger.Infof("Call rpc: user/manage, kefuid:%d, auto:%d, return OK.", kefuid, auto)
	json_response(c, "OK", nil)
}
