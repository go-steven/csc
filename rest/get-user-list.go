package rest

import (
	"errors"
	"github.com/XiBao/csc/model"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

// 查询在线客服(role=2)/客户列表(role=1)
// http://csc.test.xibao100.com/rest/user/list?role=$ROLE&chatsts=&callback=$CALLBACK
//    role=1时chatsts参数可选（取值范围：1，2，3，可以用逗号分隔多个值）
//    取值含义见：CHAT_STS_NO, CHAT_STS_ON, CHAT_STS_WAIT uint = 1, 2, 3
func GetUserList(c *gin.Context) {
	role64, _ := strconv.ParseUint(c.Query("role"), 10, 64)
	role := uint(role64)
	if role != model.VISITOR && role != model.KEFU {
		json_response(c, nil, errors.New("Invalid role parameter."))
		return
	}

	var kefuid uint64
	chatSts := []uint{}
	if role == model.VISITOR {
		kefuid, _ = strconv.ParseUint(c.Query("kefuid"), 10, 64)
		if chatStsStr := c.Query("chatsts"); chatStsStr != "" {
			for _, v := range strings.Split(chatStsStr, ",") {
				if v != "" {
					sts, _ := strconv.ParseUint(v, 10, 64)
					if sts > 0 {
						chatSts = append(chatSts, uint(sts))
					}
				}
			}
		}
	}
	result := []*model.UserCopy{}
	users := chatCenter.GetAllOnlineUsers(role, kefuid, chatSts)
	for _, v := range users {
		result = append(result, v.Copy())
	}
	logger.Infof("Call rpc: user/list, role:%d, kefuid:%d, return records:%d", role, kefuid, len(result))
	json_response(c, result, nil)
}
