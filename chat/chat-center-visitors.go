package chat

import (
	"fmt"
	"github.com/XiBao/csc/model"
	"github.com/XiBao/csc/util"
)

const (
	CHAT_STS_NO, CHAT_STS_ON, CHAT_STS_WAIT uint = 1, 2, 3
)

type ChatCenter_Visitors struct {
	Users           *util.MapData //map[uint64]*ChatUser
	nokefu_visitors *util.MapData // map[uint64]bool
}

func NewChatCenterVisitors() *ChatCenter_Visitors {
	return &ChatCenter_Visitors{
		Users:           util.NewMapData(),
		nokefu_visitors: util.NewMapData(),
	}
}

func (this *ChatCenter_Visitors) Save(user *ChatUser) {
	this.Users.Set(user.Uid, user)
}

func (this *ChatCenter_Visitors) Find(uid uint64) *ChatUser {
	v := this.Users.Get(uid)
	if v != nil {
		if u, ok := v.(*ChatUser); ok {
			return u
		}
	}

	return nil
}

func (this *ChatCenter_Visitors) AddNoKefu(uid uint64) {
	v := this.Find(uid)
	if v != nil && v.IsOnline() {
		this.nokefu_visitors.Set(uid, true)
	}
}

func (this *ChatCenter_Visitors) DelNoKefu(uid uint64) {
	this.nokefu_visitors.Del(uid)
}

func (this *ChatCenter_Visitors) NoKefuVisitors() map[uint64]bool {
	data := this.nokefu_visitors.Copy()
	result := make(map[uint64]bool)
	for k, v := range data {
		uid, ok1 := k.(uint64)
		val, ok2 := v.(bool)
		if ok1 && ok2 {
			result[uid] = val
		}
	}
	return result
}

func (this *ChatCenter_Visitors) OnlineUsers(kefuid uint64, chatSts []uint) []*model.User {
	users := []*model.User{}

	for _, val := range this.Users.Values() {
		v, ok := val.(*ChatUser)
		if !ok {
			continue
		}

		if kefuid > 0 && v.Kefu != nil && v.Kefu.Uid != kefuid {
			continue
		}

		if v.IsOnline() {
			this.updateUserRealInfo(v)

			if chatSts != nil && len(chatSts) > 0 {
				var matched bool
				for _, sts := range chatSts {
					if sts == v.User.GetUintField(model.CHAT_STATUS) {
						matched = true
						break
					}
				}

				if !matched {
					continue
				}
			}

			users = append(users, v.User)
		}
	}

	return users
}

func (this *ChatCenter_Visitors) OnlineUser(uid uint64) *model.User {
	v := this.Find(uid)
	if v == nil {
		return nil
	}

	this.updateUserRealInfo(v)
	return v.User
}

func (this *ChatCenter_Visitors) updateUserRealInfo(v *ChatUser) {
	v.User.SetUint64Field(model.REAL_KEFU_ID, 0)
	v.User.SetStrField(model.REAL_KEFU_NAME, "")

	if v.Kefu != nil {
		kefu := chatCenter.FindKefu(v.Kefu.Uid)
		if kefu != nil && kefu.IsOnline() && chatCenter.CheckKefuChat(v.Kefu.Uid, v.Uid) {
			v.User.SetUint64Field(model.REAL_KEFU_ID, kefu.Uid)
			v.User.SetStrField(model.REAL_KEFU_NAME, kefu.UName)
		}
	}

	if v.User.GetUint64Field(model.REAL_KEFU_ID) > 0 {
		v.User.SetUintField(model.CHAT_STATUS, CHAT_STS_ON)
	} else {
		val := this.nokefu_visitors.Get(v.Uid)
		if val != nil {
			v.User.SetUintField(model.CHAT_STATUS, CHAT_STS_WAIT)
		} else {
			v.User.SetUintField(model.CHAT_STATUS, CHAT_STS_NO)
		}
	}
}

func (this *ChatCenter_Visitors) CleanUpVisitor(uid uint64) {
	this.Users.Del(uid)
	this.nokefu_visitors.Del(uid)
}

func (this *ChatCenter_Visitors) Report() (ret []string) {
	ret = append(ret, fmt.Sprintf("Total online visitors : %v", this.Users.Len()))
	ret = append(ret, fmt.Sprintf("Total no kefu visitors : %v", this.nokefu_visitors.Len()))
	return
}
