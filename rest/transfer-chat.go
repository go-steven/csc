package rest

import (
	"errors"
	"fmt"
	"github.com/XiBao/csc/chat"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 客服（会话）转移
// http://csc.xibao100.com/rest/chat/transfer?uid=$UID&fromkefu=$FROMKEFUID&tokefu=$TOKEFUID&callback=$CALLBACK
func TransferChat(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Query("uid"), 10, 64)
	if uid == 0 {
		json_response(c, nil, errors.New("No uid parameter or it is zero."))
		return
	}
	visitor := chatCenter.FindUser(uid)
	if visitor == nil {
		json_response(c, nil, errors.New("No online visitor found."))
		return
	}

	fromKefuId, _ := strconv.ParseUint(c.Query("fromkefu"), 10, 64)
	if fromKefuId == 0 {
		json_response(c, nil, errors.New("No fromkefu parameter or it is zero."))
		return
	}
	fromKefu := chatCenter.FindKefu(fromKefuId)
	if fromKefu == nil || !fromKefu.IsOnline() {
		json_response(c, nil, errors.New("No online from kefu found."))
		return
	}

	toKefuId, _ := strconv.ParseUint(c.Query("tokefu"), 10, 64)
	if toKefuId == 0 {
		json_response(c, nil, errors.New("No tokefu parameter or it is zero."))
		return
	}

	toKefu := chatCenter.FindKefu(toKefuId)
	if toKefu == nil || !toKefu.IsOnline() {
		json_response(c, nil, errors.New("No online to kefu found."))
		return
	}

	visitor.SetKefu(toKefu, true)
	toKefu.SysMsg(fmt.Sprintf(chat.Prompts["KEFU_TRANSFERED_TO"], fromKefu.UName, visitor.UName), visitor)

	logger.Infof("Call rpc: chat/transfer, uid:%d, fromKefuId:%d, toKefuId:%d, return OK.", uid, fromKefuId, toKefuId)
	json_response(c, "OK", nil)
}
