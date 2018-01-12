package websocket

import (
	"fmt"
	"github.com/XiBao/csc/chat"
	"github.com/XiBao/csc/model"
	"github.com/igm/sockjs-go/sockjs"
	"strconv"
	"strings"
)

func KefuChatHandler(conn sockjs.Session) {
	logger.Info("new sockjs session established")
	defer handler_recover(conn)

	kefuid, _ := GetUidFromURI(conn.Request().RequestURI)
	if kefuid == 0 {
		logger.Errorf("%s Invalid kefuid, url:%s.", logStr(0, kefuid, conn), conn.Request().RequestURI)
		panic(model.NewCscErr(chat.Prompts["INVALID_SESSION_PARAMS"]))
	}

	kefu := ChatCenter.FindKefu(kefuid)
	if kefu == nil {
		logger.Info("Not found kefu, created new one.")
		kefu = chat.NewChatUser(kefuid, model.KEFU)
		if kefu == nil {
			panic(model.NewCscErr("Not found in users table."))
		}

		ChatCenter.RegistorKefu(kefu)
	}
	kefu.AddConn(conn)
	kefu.CheckKefuRefresh()

	var closedSession = make(chan struct{})

	go func() {
		defer handler_recover(conn)

		reader, _ := kefu.Chat.SubChannel(nil)
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

	// initiate kefu UI: send other visitors list
	visitors := ChatCenter.GetAllOnlineUsers(model.VISITOR, 0, nil)
	for _, user := range visitors {
		if user.GetUint64Field(model.REAL_KEFU_ID) == 0 && user.GetUint64Field(model.KEFU_ID) > 0 {
			user.SetUint64Field(model.REAL_KEFU_ID, user.GetUint64Field(model.KEFU_ID))
			user.SetStrField(model.REAL_KEFU_NAME, user.GetStrField(model.KEFU_NAME))
		}

		kefu.NotifyVisitorSts(user)
	}

	msgLimit := NewMessageLimit()
	for {
		if input, err := conn.Recv(); err == nil {
			logger.Info("Recv input: " + input)
			input = strings.TrimSpace(input)
			uid, msg := parseKefuInput(input)
			if uid == 0 {
				continue
			}
			msgLen := len([]rune(msg))
			visitor := ChatCenter.FindUser(uid)
			if visitor == nil {
				visitor = chat.NewOffLineVisitor(uid)
				if visitor == nil {
					continue
				}
			}

			if msgLimit.NotAllowedMsgRate(conn, visitor) {
				// ignore the message
				continue
			}

			if msgLimit.NotAllowedMsgLen(msgLen, conn, visitor) {
				// ignore the message
				continue
			}

			if msgLen == 2 {
				if isFAQ, answer := FAQKefuCheck(msg); isFAQ > 0 {
					if isFAQ == 2 {
						// show FAQ list on kefu side
						if conn.GetSessionState() == sockjs.SessionActive {
							conn.Send(chat.NewChatMsg(NowStr(), visitor, kefu, fmt.Sprintf("%v", msg), model.KEFU).Json())
							conn.Send(chat.NewSysMsg(answer, visitor).Json())
						}
						continue
					}

					// send the FAQ answer to visitor directly
					msg = answer
				}
			}

			if !kefu.IsOnline() {
				conn.Send(chat.NewSysMsg(chat.Prompts["KEFU_IS_OFFLINE"], visitor).Json())
				continue
			}

			var msgStatus uint
			chatMsg := chat.NewChatMsg(NowStr(), visitor, kefu, fmt.Sprintf("%v", msg), model.KEFU)
			if !visitor.IsOnline() {
				conn.Send(chat.NewSysMsg(chat.Prompts["VISITOR_IS_OFFLINE"], visitor).Json())
			} else {
				if visitor.Kefu != nil && visitor.Kefu.IsOnline() && visitor.Kefu.Uid != kefuid && ChatCenter.CheckKefuChat(visitor.Kefu.Uid, uid) {
					conn.Send(chat.NewSysMsg(fmt.Sprintf(chat.Prompts["OTHER_KEFU_PROCESSED"], visitor.Kefu.UName), visitor).Json())
					continue
				}
				if visitor.Kefu == nil || !visitor.Kefu.IsOnline() || !ChatCenter.CheckKefuChat(visitor.Kefu.Uid, visitor.Uid) {
					visitor.SetKefu(kefu, false)
				}
				visitor.Chat.Publish(chatMsg)
				msgStatus = 1
			}
			if err := SaveMsgToDb(kefu, visitor, chatMsg, msgStatus); err != nil {
				logger.Error(err.Error())
				//return err
			}
			kefu.Chat.Publish(chatMsg)
			continue
		}
		break
	}
	close(closedSession)

	kefu.CloseConn(conn)

	logger.Info("sockjs session closed")
}

const KEFU_INPUT_SEPERATOR = ";;;;"

func parseKefuInput(input string) (uid uint64, msg string) {
	pos := strings.Index(input, KEFU_INPUT_SEPERATOR)
	if pos < 0 {
		return
	}

	uid, _ = strconv.ParseUint(input[0:pos], 10, 64)
	msg = input[pos+len(KEFU_INPUT_SEPERATOR):]

	return
}
