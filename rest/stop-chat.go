package rest

import (
	"errors"
	"fmt"
	"github.com/XiBao/csc/chat"
	"github.com/XiBao/csc/model"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 结束（客服）会话
// http://csc.xibao100.com/rest/chat/stop?uid=$UID&kefuid=$KEFUID&callback=$CALLBACK
func StopChat(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Query("uid"), 10, 64)
	if uid == 0 {
		json_response(c, nil, errors.New("No uid parameter or it is zero."))
		return
	}

	kefuid, _ := strconv.ParseUint(c.Query("kefuid"), 10, 64)
	if kefuid == 0 {
		json_response(c, nil, errors.New("No kefuid parameter or it is zero."))
		return
	}

	visitor := chatCenter.FindUser(uid)
	if visitor == nil {
		json_response(c, nil, errors.New(chat.Prompts["VISITOR_IS_OFFLINE"]))
		return
	}

	kefu := chatCenter.FindKefu(kefuid)
	if kefu == nil {
		json_response(c, nil, errors.New(chat.Prompts["KEFU_IS_OFFLINE"]))
		return
	}

	if visitor.Kefu != nil && visitor.Kefu.Uid == kefuid && chatCenter.CheckKefuChat(kefuid, uid) {
		visitor.SetKefu(nil, false)
	} else {
		chatCenter.IncreaseChatStat(kefuid, uid, -1)
		chatCenter.NotifyVisitorSts(visitor)
	}

	sysMsg := chat.NewSysMsg(fmt.Sprintf(chat.Prompts["KEFU_STOP_CHAT"], kefu.UName), visitor)
	chat.SaveSysMsgToDb(sysMsg, visitor, kefu, model.KEFU)

	logger.Infof("Call rpc: chat/stop, uid:%d, kefuid:%d, return OK.", uid, kefuid)
	json_response(c, "OK", nil)
}
