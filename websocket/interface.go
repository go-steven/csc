package websocket

import (
	"fmt"
	"github.com/XiBao/csc/chat"
	"github.com/XiBao/csc/model"
	"github.com/bububa/mymysql/autorc"
	_ "github.com/bububa/mymysql/thrsafe"
	"github.com/igm/sockjs-go/sockjs"
	log "github.com/kdar/factorlog"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

var (
	logger *log.FactorLog

	ChatCenter *chat.ChatCenter
	DB         *autorc.Conn
)

// Set to use a specific logger
func SetLogger(alogger *log.FactorLog) {
	logger = alogger
}

func SetChatCenter(center *chat.ChatCenter) {
	ChatCenter = center
}

func SetDb(db *autorc.Conn) {
	DB = db
}

func GetUidFromURI(uri string) (uid uint64, err error) {
	reg := regexp.MustCompile(`(\d+)`)

	matches := reg.FindAllString(uri, -1)
	if len(matches) >= 1 {
		uid, _ = strconv.ParseUint(matches[0], 10, 64)
	}

	return
}

func NowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func logStr(uid, kefuid uint64, conn sockjs.Session) string {
	return fmt.Sprintf("[%d-%d-%s]", uid, kefuid, conn.ID())
}

func handler_recover(conn sockjs.Session) {
	defer func(c sockjs.Session) {
		if c != nil && c.GetSessionState() == sockjs.SessionActive {
			c.Send(chat.NewNotifyMsg(chat.Prompts["SESSION_CLOSED"]))
			c.Close(0, "")
		}
	}(conn)

	if r := recover(); r != nil {
		switch rval := r.(type) {
		case nil:
			return
		case model.CscErr:
			if conn != nil && conn.GetSessionState() == sockjs.SessionActive {
				conn.Send(chat.NewNotifyMsg(rval.Error()))
			}
		default:
			var st = func(all bool) string {
				// Reserve 1K buffer at first
				buf := make([]byte, 512)

				for {
					size := runtime.Stack(buf, all)
					// The size of the buffer may be not enough to hold the stacktrace,
					// so double the buffer size
					if size == len(buf) {
						buf = make([]byte, len(buf)<<1)
						continue
					}
					break
				}

				return string(buf)
			}
			logger.Errorf("panic:%v\nstack:%v", r, st(false))
		}
	}
}

func SaveMsgToDb(fromUser, toUser *chat.ChatUser, msg *chat.ChatMsg, msgStatus uint) error {
	initiator := fromUser.User.GetUintField(model.ROLE)

	message := &model.Msg{
		VisitorId:  0,
		KefuId:     0,
		Initiator:  initiator,
		MsgType:    msg.MsgType,
		Status:     msgStatus,
		CreateTime: msg.MsgTime,
		Content:    msg.Msg,
		DB:         DB,
	}
	if initiator == model.VISITOR {
		message.VisitorId = fromUser.Uid

		if toUser != nil {
			message.KefuId = toUser.Uid
		}
	}
	if initiator == model.KEFU {
		message.KefuId = fromUser.Uid
		if toUser != nil {
			message.VisitorId = toUser.Uid
		}
	}

	if msgStatus == 1 {
		message.SendTime = msg.MsgTime
	}

	return message.SaveToDb()
}
