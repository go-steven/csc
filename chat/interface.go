package chat

import (
	"encoding/json"
	"fmt"
	"github.com/XiBao/csc/model"
	"github.com/bububa/mymysql/autorc"
	"github.com/bububa/mymysql/mysql"
	_ "github.com/bububa/mymysql/thrsafe"
	"github.com/igm/sockjs-go/sockjs"
	log "github.com/kdar/factorlog"
	"strings"
	"time"
)

var (
	logger *log.FactorLog

	DB         *autorc.Conn
	chatCenter *ChatCenter
)

// Set to use a specific logger
func SetLogger(alogger *log.FactorLog) {
	logger = alogger
}

func SetChatCenter(center *ChatCenter) {
	chatCenter = center
}

func SetDb(db *autorc.Conn) {
	DB = db
}

func NowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Json(entry interface{}) string {
	b, err := json.Marshal(entry)
	if err != nil {
		return ""
	}

	return string(b)
}

func logStr(uid, kefuid uint64, conn sockjs.Session) string {
	if conn == nil {
		return fmt.Sprintf("[%d-%d-]", uid, kefuid)
	} else {
		return fmt.Sprintf("[%d-%d-%s]", uid, kefuid, conn.ID())
	}
}

func GetOfflineMsgs(visitorId, kefuid uint64, initiator uint) ([]string, error) {
	var query string
	var err error
	var rows []mysql.Row

	if kefuid == 0 {
		query = `SELECT m.id, m.create_time, v.name AS visitor_name, k.name as kefu_name, k.nick as kefu_nick, m.content, v.Affiliate FROM csc_msgs m
     					LEFT JOIN csc_users v ON (v.uid = m.visitor_id AND v.role = 1)
     					LEFT JOIN csc_users k ON (k.uid = m.kefu_id AND k.role = 2)
   	 			  WHERE m.visitor_id = %d AND m.initiator = %d AND m.msg_type = 1 AND m.status = 0 ORDER BY m.create_time ASC`
		rows, _, err = DB.Query(query, visitorId, initiator)
	} else {
		query = `SELECT m.id, m.create_time, v.name AS visitor_name, k.name as kefu_name, k.nick as kefu_nick, m.content, v.Affiliate FROM csc_msgs m
     		      		LEFT JOIN csc_users v ON (v.uid = m.visitor_id AND v.role = 1)
     					LEFT JOIN csc_users k ON (k.uid = m.kefu_id AND k.role = 2)
    			 WHERE m.visitor_id = %d AND (m.kefu_id = 0 OR m.kefu_id = %d) AND m.msg_type = 1 AND m.initiator = %d AND m.status = 0 ORDER BY m.create_time ASC`
		rows, _, err = DB.Query(query, visitorId, kefuid, initiator)
	}

	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	ids := []string{}
	msgs := []string{}
	for _, row := range rows {
		ids = append(ids, fmt.Sprintf("'%s'", DB.Escape(row.Str(0))))

		visitorUser := model.NewUserModel()
		visitorName := row.Str(2)
		visitorNick := row.Str(2)
		visitorUser.SetUint64Field(model.UID, visitorId)
		visitorUser.SetStrField(model.NAME, visitorName)
		visitorUser.SetStrField(model.NICK, visitorNick)
		visitorUser.SetUintField(model.ROLE, model.VISITOR)
		visitorUser.SetStrField(model.AFFILIATE, row.Str(6))

		kefuUser := model.NewUserModel()
		kefuName := row.Str(3)
		kefuNick := row.Str(4)
		kefuUser.SetUint64Field(model.UID, kefuid)
		kefuUser.SetStrField(model.NAME, kefuName)
		kefuUser.SetStrField(model.NICK, kefuNick)
		kefuUser.SetUintField(model.ROLE, model.KEFU)

		chatMsg := NewChatMsg(row.Localtime(1).Format("2006-01-02 15:04:05"), &ChatUser{Uid: visitorId, UName: visitorName, UNick: visitorNick, User: visitorUser}, &ChatUser{Uid: kefuid, UName: kefuName, UNick: kefuNick, User: kefuUser}, row.Str(5), initiator)
		msgs = append(msgs, chatMsg.Json())
	}

	if len(ids) > 0 {
		updateQuery := `update csc_msgs set status = 1, send_time = now() WHERE visitor_id = %d AND msg_type = 1 AND initiator = %d AND status = 0 AND id IN(%s)`
		_, _, err = DB.Query(updateQuery, visitorId, initiator, strings.Join(ids, ","))
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
	}

	return msgs, nil
}

func SaveSysMsgToDb(msg *ChatMsg, visitor, kefu *ChatUser, initiator uint) error {
	if msg.MsgType != MSG_TYPE_SYS {
		return nil
	}

	message := &model.Msg{
		VisitorId:  0,
		KefuId:     0,
		MsgType:    msg.MsgType,
		Initiator:  initiator,
		Status:     1,
		CreateTime: msg.MsgTime,
		SendTime:   msg.MsgTime,
		Content:    msg.Msg,
		DB:         DB,
	}
	if visitor != nil && visitor.User.GetUintField(model.ROLE) == model.VISITOR {
		message.VisitorId = visitor.Uid
	}
	if kefu != nil && kefu.User.GetUintField(model.ROLE) == model.KEFU {
		message.KefuId = kefu.Uid
	}

	err := message.SaveToDb()
	if err != nil {
		logger.Error(err.Error())
	}

	return err
}
