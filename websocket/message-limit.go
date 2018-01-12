package websocket

import (
	"github.com/XiBao/csc/chat"
	"github.com/igm/sockjs-go/sockjs"
	"time"
)

const (
	ALLOWED_MSG_MAX_LEN  = 500
	ALLOWED_MSG_MAX_RATE = 5 // per second
)

type MessageLimit struct {
	Second    int
	SecondCnt uint
}

func NewMessageLimit() *MessageLimit {
	return &MessageLimit{}
}

func (this *MessageLimit) NotAllowedMsgLen(msgLen int, conn sockjs.Session, toUser *chat.ChatUser) bool {
	if msgLen <= 0 {
		if conn.GetSessionState() == sockjs.SessionActive {
			conn.Send(chat.NewSysMsg(chat.Prompts["MSG_LIMIT_TOO_SHORT"], toUser).Json())
		}
		return true
	}

	if msgLen >= ALLOWED_MSG_MAX_LEN {
		if conn.GetSessionState() == sockjs.SessionActive {
			conn.Send(chat.NewSysMsg(chat.Prompts["MSG_LIMIT_TOO_LONG"], toUser).Json())
		}
		return true
	}

	return false
}

func (this *MessageLimit) NotAllowedMsgRate(conn sockjs.Session, toUser *chat.ChatUser) bool {
	second := time.Now().Second()
	if this.Second != second {
		this.Second = second
		this.SecondCnt = 0
	}

	this.SecondCnt = this.SecondCnt + 1

	if this.SecondCnt >= ALLOWED_MSG_MAX_RATE {
		if conn.GetSessionState() == sockjs.SessionActive {
			conn.Send(chat.NewSysMsg(chat.Prompts["MSG_LIMIT_TOO_FAST"], toUser).Json())
		}
		time.Sleep(5 * time.Second)
		return true
	}

	return false
}
