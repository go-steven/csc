package rest

import (
	"errors"
	"fmt"
	"github.com/XiBao/csc/chat"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 客服主动呼叫
// http://csc.test.xibao100.com/rest/user/call?kefuid=$KEFUID&uid=$UID&callback=$CALLBAC
func KefuCall(c *gin.Context) {
	kefuid, _ := strconv.ParseUint(c.Query("kefuid"), 10, 64)
	if kefuid == 0 {
		json_response(c, nil, errors.New("No kefuid parameter or it is zero."))
		return
	}

	uid, _ := strconv.ParseUint(c.Query("uid"), 10, 64)
	if uid == 0 {
		json_response(c, nil, errors.New("No uid parameter or it is zero."))
		return
	}

	visitor := chatCenter.FindUser(uid)
	if visitor == nil {
		json_response(c, nil, errors.New(chat.Prompts["VISITOR_IS_OFFLINE"]))
		return
	}

	if visitor.Kefu != nil && visitor.Kefu.Uid != kefuid && chatCenter.CheckKefuChat(visitor.Kefu.Uid, uid) {
		json_response(c, nil, errors.New(fmt.Sprintf(chat.Prompts["OTHER_KEFU_PROCESSED"], visitor.Kefu.UName)))
		return
	}

	kefu := chatCenter.FindKefu(kefuid)
	if kefu == nil || !kefu.IsOnline() {
		json_response(c, nil, errors.New(chat.Prompts["KEFU_IS_OFFLINE"]))
		return
	}

	chatCenter.DelNoKefu(uid)
	visitor.SetKefu(kefu, false)

	logger.Infof("Call rpc: user/call, kefuid:%d, uid:%d, return OK.", kefuid, uid)
	json_response(c, "OK", nil)
}
