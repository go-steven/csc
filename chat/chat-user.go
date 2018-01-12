package chat

import (
	"fmt"
	"github.com/XiBao/csc/model"
	"github.com/XiBao/csc/util"
	"github.com/igm/sockjs-go/sockjs"
	"sync"
	"time"
)

type ChatUser struct {
	m *sync.RWMutex

	Uid      uint64
	UName    string
	UNick    string
	UManager uint
	Role     uint
	User     *model.User
	Chat     *util.Publisher
	Conns    *util.MapData

	Status uint
	Kefu   *ChatUser

	LastManualStatus uint
	LastConnClosed   int64 // time stamp
}

func NewChatUser(uid uint64, role uint) *ChatUser {
	user := model.NewUser(uid, role, DB)
	if user == nil {
		return nil
	}
	//var status uint = OFFLINE
	//if role == model.VISITOR {
	//	status = model.ONLINE
	//}
	if role == model.VISITOR {
		user.SetUintField(model.STATUS, model.ONLINE)
	}

	userCopy := user.Copy()
	return &ChatUser{
		m:        new(sync.RWMutex),
		Uid:      userCopy.Uid,
		UName:    userCopy.Name,
		UNick:    userCopy.Nick,
		Role:     userCopy.Role,
		UManager: userCopy.Manager,

		User:   user,
		Chat:   util.NewPublisher(),
		Conns:  util.NewMapData(),
		Status: userCopy.Status,

		LastManualStatus: model.OFFLINE,
		LastConnClosed:   0,
	}
}

func NewOffLineVisitor(uid uint64) *ChatUser {
	visitor := NewChatUser(uid, model.VISITOR)
	if visitor == nil {
		return nil
	}
	visitor.Status = model.OFFLINE
	visitor.User.SetUintField(model.STATUS, model.OFFLINE)
	chatCenter.RegistorUser(visitor)

	return visitor
}

func (this *ChatUser) SetKefu(kefu *ChatUser, noKefuNotify bool) {
	if this.Role != model.VISITOR || (kefu != nil && kefu.Role != model.KEFU) {
		return
	}

	if this.Kefu != nil && (kefu == nil || this.Kefu.Uid != kefu.Uid) && chatCenter.CheckKefuChat(this.Kefu.Uid, this.Uid) {
		chatCenter.IncreaseChatStat(this.Kefu.Uid, this.Uid, -1)
	}

	this.m.Lock()
	this.Kefu = kefu
	this.m.Unlock()

	if kefu == nil {
		this.User.SetUint64Field(model.KEFU_ID, 0)
		this.User.SetStrField(model.KEFU_NAME, "")
	} else {
		this.User.SetUint64Field(model.KEFU_ID, kefu.Uid)
		this.User.SetStrField(model.KEFU_NAME, kefu.UName)
	}

	if err := this.User.SaveToDb(); err != nil {
		logger.Errorf("Failed to save user to database, error:%s.", err.Error())
	}

	if kefu == nil {
		chatCenter.AddNoKefu(this.Uid)
	} else {
		chatCenter.IncreaseChatStat(kefu.Uid, this.Uid, 1)
		this.SysMsg(fmt.Sprintf(Prompts["KEFU_ONLINE"], kefu.UNick), kefu)
		if !noKefuNotify {
			kefu.SysMsg(fmt.Sprintf(Prompts["MSG_USER_WAITING"], this.UName), this)
		}
		if msgs, err := GetOfflineMsgs(this.Uid, kefu.Uid, model.VISITOR); err == nil && len(msgs) > 0 {
			for _, v := range msgs {
				for _, val := range kefu.Conns.Values() {
					if conn, ok := val.(sockjs.Session); ok {
						if conn.GetSessionState() == sockjs.SessionActive {
							conn.Send(v)
						}
					}
				}
			}
		}
	}
	chatCenter.NotifyVisitorSts(this)
}

func (this *ChatUser) AddConn(conn sockjs.Session) (int, error) {
	this.checkConns()

	if conn.GetSessionState() == sockjs.SessionActive {
		this.Conns.Set(conn.ID(), conn)
	}

	this.User.SetStrField(model.SIGNIN_AT, NowStr())
	this.User.SetStrField(model.SESSION, conn.ID())

	if err := this.User.SaveToDb(); err != nil {
		logger.Errorf("Failed to save user to database, error:%s.", err.Error())
		return 0, err
	}
	if err := this.User.UpdateInfoFromDb(); err != nil {
		logger.Errorf("Failed to get user ext info from database, error:%s.", err.Error())
		return 0, err
	}

	return this.Conns.Len(), nil
}

func (this *ChatUser) checkConns() {
	for k, v := range this.Conns.Copy() {
		if conn, ok := v.(sockjs.Session); ok {
			if conn.GetSessionState() != sockjs.SessionActive {
				this.Conns.Del(k)
			}
		}
	}
}

func (this *ChatUser) IsOnline() bool {
	this.m.RLock()
	defer this.m.RUnlock()

	return this.Status == model.ONLINE
}

func (this *ChatUser) SetStatus(sts uint) {
	this.m.Lock()
	this.Status = sts
	this.m.Unlock()
	this.User.SetUintField(model.STATUS, sts)

	if sts == model.OFFLINE {
		this.User.SetStrField(model.SIGNUP_AT, NowStr())
	} else {
		this.User.SetStrField(model.SIGNIN_AT, NowStr())
	}

	if err := this.User.SaveToDb(); err != nil {
		logger.Errorf("Failed to save user to database, error:%s.", err.Error())
	}

	if this.Role == model.KEFU {
		// notify all related visitors for this kefu's status changed
		visitors := chatCenter.GetKefuVisitors(this.Uid)
		for k, _ := range visitors {
			v := chatCenter.FindUser(k)
			/*
				if v != nil && v.IsOnline() && v.Kefu != nil && v.Kefu.Uid == this.Uid {
					if this.Status == OFFLINE {
						v.SysMsg(fmt.Sprintf(Prompts["KEFU_OFFLINE"], this.UNick), this)
					} else {
						v.SysMsg(fmt.Sprintf(Prompts["KEFU_ONLINE"], this.UNick), this)
					}
				}
			*/
			chatCenter.NotifyVisitorSts(v)
		}

		if this.IsOnline() {
			chatCenter.DispatchNoKefuVisitors()
		}
	}
}

func (this *ChatUser) SetManagerSts(manager uint) {
	if this.Role != model.KEFU {
		return
	}

	this.m.Lock()
	this.UManager = manager
	this.m.Unlock()
	this.User.SetUintField(model.MANAGER, manager)

	if err := this.User.SaveToDb(); err != nil {
		logger.Errorf("Failed to save user to database, error:%s.", err.Error())
	}
}

func (this *ChatUser) CloseConn(conn sockjs.Session) {
	this.m.Lock()
	this.LastConnClosed = time.Now().Unix()
	this.m.Unlock()

	conn.Send(NewNotifyMsg(Prompts["SESSION_CLOSED"]))
	conn.Close(0, "")

	this.Conns.Del(conn.ID())
	this.checkConns()
	if this.Conns.Len() == 0 {
		this.m.Lock()
		this.Status = model.OFFLINE
		this.m.Unlock()

		this.User.SetUintField(model.STATUS, model.OFFLINE)
		this.User.SetStrField(model.SIGNUP_AT, NowStr())

		if err := this.User.SaveToDb(); err != nil {
			logger.Errorf("Failed to save user to database, error:%s.", err.Error())
		}

		if this.Role == model.KEFU {
			// notify all related visitors for this kefu's status changed
			visitors := chatCenter.GetKefuVisitors(this.Uid)
			for k, _ := range visitors {
				v := chatCenter.FindUser(k)
				/*
					if v != nil && v.IsOnline() && v.Kefu != nil && v.Kefu.Uid == this.Uid {
						v.SysMsg(fmt.Sprintf(Prompts["KEFU_OFFLINE"], this.UNick), this)
					}
				*/
				chatCenter.NotifyVisitorSts(v)
			}
		}

		// cleanup chat-center data
		var relatedKefuid uint64
		if this.Role == model.VISITOR {
			this.m.RLock()
			if this.Kefu != nil {
				relatedKefuid = this.Kefu.Uid
			}
			this.m.RUnlock()
		}

		chatCenter.CleanUpUser(this.Role, this.Uid, relatedKefuid)
	}
}

func (this *ChatUser) SysMsg(msg string, toUser *ChatUser) {
	chatMsg := NewSysMsg(msg, toUser)

	for _, v := range this.Conns.Values() {
		if conn, ok := v.(sockjs.Session); ok {
			if conn.GetSessionState() == sockjs.SessionActive {
				conn.Send(chatMsg.Json())
			}
		}
	}

	var visitor, kefu *ChatUser
	if toUser != nil {
		if toUser.Role == model.VISITOR {
			visitor = toUser
		} else {
			kefu = toUser
		}
	}

	if this.Role == model.VISITOR {
		visitor = this
	} else {
		kefu = this
	}

	if err := SaveSysMsgToDb(chatMsg, visitor, kefu, this.Role); err != nil {
		logger.Error(err.Error())
	}
}

func (this *ChatUser) NotifyVisitorSts(user *model.User) {
	if user == nil {
		return
	}
	userCopy := user.Copy()
	if this.Role != model.KEFU || userCopy.Role != model.VISITOR {
		return
	}

	msg := &VisitorStsMsg{
		MsgTime:      NowStr(),
		MsgType:      MSG_VISITOR_STS,
		Uid:          userCopy.Uid,
		Name:         userCopy.Name,
		Affiliate:    userCopy.Affiliate,
		Status:       userCopy.Status,
		RealKefuId:   userCopy.RealKefuId,
		RealKefuName: userCopy.RealKefuName,
		ChatStatus:   userCopy.ChatStatus,
		ScLevel:      userCopy.ScLevel,
		DepositCnt:   userCopy.DepositCnt,
	}
	json := Json(msg)

	for _, v := range this.Conns.Values() {
		if conn, ok := v.(sockjs.Session); ok {
			if conn.GetSessionState() == sockjs.SessionActive {
				conn.Send(json)
			}
		}
	}
}

func (this *ChatUser) CheckKefuRefresh() {
	if this.Role != model.KEFU {
		return
	}

	this.m.RLock()
	lastConnClosed := this.LastConnClosed
	this.m.RUnlock()

	// If the connection is re-connected in 10s, we think it as refresh action
	if lastConnClosed > 0 && time.Now().Unix()-lastConnClosed <= 10 {
		this.SetStatus(this.LastManualStatus)
	}
}
