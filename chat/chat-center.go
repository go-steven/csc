package chat

import (
	"fmt"
	"github.com/XiBao/csc/model"
	"time"
)

type ChatCenter struct {
	visitors *ChatCenter_Visitors
	kefus    *ChatCenter_Kefus
	start_t  time.Time
}

func NewChatCenter() *ChatCenter {
	return &ChatCenter{
		visitors: NewChatCenterVisitors(),
		kefus:    NewChatCenterKefus(),
		start_t:  time.Now(),
	}
}

func (this *ChatCenter) RegistorUser(user *ChatUser) error {
	this.visitors.Save(user)
	return nil
}

func (this *ChatCenter) FindUser(uid uint64) *ChatUser {
	return this.visitors.Find(uid)
}

func (this *ChatCenter) RegistorKefu(kefu *ChatUser) error {
	this.kefus.Save(kefu)
	return nil
}

func (this *ChatCenter) FindKefu(kefuid uint64) *ChatUser {
	return this.kefus.Find(kefuid)
}

func (this *ChatCenter) AssignOnlineKefu() *ChatUser {
	return this.kefus.AssignOnlineKefu()
}

func (this *ChatCenter) CheckKefuChat(kefuid, visitorId uint64) bool {
	return this.kefus.CheckKefuChat(kefuid, visitorId)
}

func (this *ChatCenter) IncreaseChatStat(kefuid, visitorId uint64, num int) {
	this.kefus.IncreaseChatStat(kefuid, visitorId, num)
}

func (this *ChatCenter) DispatchNoKefuVisitors() {
	for visitorId, _ := range this.visitors.NoKefuVisitors() {
		visitor := this.visitors.Find(visitorId)
		if visitor == nil || !visitor.IsOnline() {
			this.visitors.DelNoKefu(visitorId)
			logger.Info("visitor is not online.")
			continue
		}

		kefu := this.AssignOnlineKefu()
		if kefu == nil || !kefu.IsOnline() {
			logger.Info("kefu is not online.")
			continue
		}

		logger.Info("set kefu.")
		this.visitors.DelNoKefu(visitorId)

		visitor.SetKefu(kefu, false)
	}
}

func (this *ChatCenter) GetAllOnlineUsers(role uint, kefuid uint64, chatSts []uint) []*model.User {
	if role == model.VISITOR {
		return this.visitors.OnlineUsers(kefuid, chatSts)
	} else {
		return this.kefus.OnlineUsers()
	}
}

func (this *ChatCenter) AddNoKefu(uid uint64) {
	this.visitors.AddNoKefu(uid)
}

func (this *ChatCenter) DelNoKefu(uid uint64) {
	this.visitors.DelNoKefu(uid)
}

func (this *ChatCenter) GetKefuVisitors(kefuid uint64) map[uint64]bool {
	return this.kefus.GetAllVisitors(kefuid)
}

func (this *ChatCenter) NotifyVisitorSts(visitor *ChatUser) {
	for _, v := range this.kefus.GetAllKefus() {
		v.NotifyVisitorSts(this.visitors.OnlineUser(visitor.Uid))
	}
}

func (this *ChatCenter) NotifyVisitorOnline(visitor *ChatUser) {
	user := this.visitors.OnlineUser(visitor.Uid)
	if user.GetUint64Field(model.REAL_KEFU_ID) == 0 && user.GetUint64Field(model.KEFU_ID) > 0 {
		user.SetUint64Field(model.REAL_KEFU_ID, user.GetUint64Field(model.KEFU_ID))
		user.SetStrField(model.REAL_KEFU_NAME, user.GetStrField(model.KEFU_NAME))
	}

	for _, v := range this.kefus.GetAllKefus() {
		v.NotifyVisitorSts(user)
	}
}

func (this *ChatCenter) GetOnlineUser(role uint, uid uint64) *model.User {
	if role == model.KEFU {
		return nil
	}

	return this.visitors.OnlineUser(uid)
}

func (this *ChatCenter) CleanUpUser(role uint, uid, relatedKefuid uint64) {
	if role == model.KEFU {
		logger.Infof("Cleanup kefu: %v", uid)
		this.kefus.CleanUpKefu(uid)
	} else {
		logger.Infof("Cleanup visitor: %v", uid)
		if relatedKefuid > 0 {
			this.kefus.CleanUpVisitor(relatedKefuid, uid)
		}
		this.visitors.CleanUpVisitor(uid)
	}
}

func (this *ChatCenter) Report() (ret []string) {
	ret = append(ret, fmt.Sprintf("========================== Server Running Stats ==============="))
	ret = append(ret, fmt.Sprintf("start time : %v", this.start_t.Format("2006-01-02 3.04.05 PM")))
	ret = append(ret, fmt.Sprintf("total duration : %v", time.Now().Sub(this.start_t)))

	ret = append(ret, fmt.Sprintf("========================== Kefus Stats ==============="))
	ret = append(ret, this.kefus.Report()...)

	ret = append(ret, fmt.Sprintf("========================== Visitors Stats ==============="))
	ret = append(ret, this.visitors.Report()...)
	return
}
