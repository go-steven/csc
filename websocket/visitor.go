package websocket

import (
	"fmt"
	"github.com/XiBao/csc/chat"
	"github.com/XiBao/csc/model"
	"github.com/igm/sockjs-go/sockjs"
	"strings"
)

func VisitorChatHandler(conn sockjs.Session) {
	logger.Info("new sockjs session established")
	defer handler_recover(conn)

	uid, _ := GetUidFromURI(conn.Request().RequestURI)
	if uid == 0 {
		logger.Errorf("%s Invalid uid, url:%s.", logStr(uid, 0, conn), conn.Request().RequestURI)
		panic(model.NewCscErr(chat.Prompts["INVALID_SESSION_PARAMS"]))
	}

	visitor := ChatCenter.FindUser(uid)
	if visitor == nil {
		logger.Info("Not found visitor, created new one.")
		visitor = chat.NewChatUser(uid, model.VISITOR)
		if visitor == nil {
			panic(model.NewCscErr("Not found in users table."))
		}
		ChatCenter.RegistorUser(visitor)
	}
	visitor.SetStatus(model.ONLINE)
	visitor.User.UpdateExtInfo()

	if msgs, err := chat.GetOfflineMsgs(uid, 0, model.KEFU); err == nil && len(msgs) > 0 {
		for _, msg := range msgs {
			conn.Send(msg)
		}
	}

	conn.Send(chat.NewNotifyMsg(FAQStr()))

	if visitor.Kefu == nil {
		if visitor_kefuId := visitor.User.GetUint64Field(model.KEFU_ID); visitor_kefuId > 0 {
			kefu := ChatCenter.FindKefu(visitor_kefuId)
			if kefu != nil && kefu.IsOnline() {
				visitor.SetKefu(kefu, true)
			}
		}
	}

	ChatCenter.NotifyVisitorOnline(visitor)

	var closedSession = make(chan struct{})
	go func() {
		defer handler_recover(conn)

		reader, _ := visitor.Chat.SubChannel(nil)
		for {
			select {
			case <-closedSession:
				return
			case message := <-reader:
				if msg, ok := message.(*chat.ChatMsg); ok {
					if err := conn.Send(msg.Json()); err != nil {
						logger.Error(err.Error())
						return
					}
				}
			}

		}
	}()

	msgLimit := NewMessageLimit()
	for {
		if msg, err := conn.Recv(); err == nil {
			logger.Info("Recv msg: " + msg)
			msg = strings.TrimSpace(msg)
			msgLen := len([]rune(msg))
			if msgLimit.NotAllowedMsgRate(conn, visitor) {
				// ignore the message
				continue
			}

			if msgLimit.NotAllowedMsgLen(msgLen, conn, visitor) {
				// ignore the message
				continue
			}
			if isFAQ, answer := FAQCheck(msg); isFAQ {
				if conn.GetSessionState() == sockjs.SessionActive {
					chatMsg := chat.NewChatMsg(NowStr(), visitor, nil, fmt.Sprintf("%v", msg), model.VISITOR)
					conn.Send(chatMsg.Json())
					SaveMsgToDb(visitor, nil, chatMsg, 1)

					conn.Send(chat.NewNotifyMsg(answer))
				}

				continue
			}
			if visitor.Kefu == nil || !visitor.Kefu.IsOnline() || !ChatCenter.CheckKefuChat(visitor.Kefu.Uid, visitor.Uid) {
				kefu := ChatCenter.AssignOnlineKefu()
				if kefu != nil && kefu.IsOnline() {
					visitor.SetKefu(kefu, false)
				} else {
					visitor.SetKefu(nil, false)
				}
			}

			chatMsg := chat.NewChatMsg(NowStr(), visitor, visitor.Kefu, fmt.Sprintf("%v", msg), model.VISITOR)
			visitor.Chat.Publish(chatMsg)

			var msgStatus uint
			if visitor.Kefu != nil && visitor.Kefu.IsOnline() {
				visitor.Kefu.Chat.Publish(chatMsg)
				msgStatus = 1
			} else {
				sysMsg := chat.NewSysMsg(chat.Prompts["KEFU_NOT_FOUND"], visitor)
				visitor.Chat.Publish(sysMsg)
				chat.SaveSysMsgToDb(sysMsg, visitor, nil, model.VISITOR)
			}

			if err := SaveMsgToDb(visitor, visitor.Kefu, chatMsg, msgStatus); err != nil {
				//logger.Errorf("%s Failed to save msg[%s] to db, error:%s", this.logStr(), msg, err.Error())
				//return err
			}

			continue
		}
		break
	}
	close(closedSession)

	visitor.CloseConn(conn)
	ChatCenter.NotifyVisitorSts(visitor)

	logger.Info("sockjs session closed")
}
