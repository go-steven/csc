package chat

import (
	"fmt"
	"github.com/XiBao/csc/model"
	"github.com/XiBao/csc/util"
	"sync"
)

type ChatCenter_Kefus struct {
	m_s *sync.RWMutex

	Users          *util.MapData // map[uint64]*ChatUser
	kefuchat_stats map[uint64]map[uint64]bool
}

func NewChatCenterKefus() *ChatCenter_Kefus {
	return &ChatCenter_Kefus{
		Users:          util.NewMapData(),
		kefuchat_stats: make(map[uint64]map[uint64]bool),
		m_s:            new(sync.RWMutex),
	}
}

func (this *ChatCenter_Kefus) Save(user *ChatUser) {
	this.Users.Set(user.Uid, user)
}

func (this *ChatCenter_Kefus) Find(kefuid uint64) *ChatUser {
	v := this.Users.Get(kefuid)
	if v != nil {
		if u, ok := v.(*ChatUser); ok {
			return u
		}
	}

	return nil
}

func (this *ChatCenter_Kefus) CheckKefuChat(kefuid, visitorId uint64) bool {
	this.m_s.RLock()
	defer this.m_s.RUnlock()

	if _, ok := this.kefuchat_stats[kefuid]; !ok {
		return false
	}

	if _, ok := this.kefuchat_stats[kefuid][visitorId]; !ok {
		return false
	}

	return true
}

func (this *ChatCenter_Kefus) IncreaseChatStat(kefuid, visitorId uint64, num int) {
	this.m_s.Lock()
	defer this.m_s.Unlock()

	if _, ok := this.kefuchat_stats[kefuid]; !ok {
		this.kefuchat_stats[kefuid] = make(map[uint64]bool)
	}

	if num > 0 {
		this.kefuchat_stats[kefuid][visitorId] = true
	}

	if num < 0 {
		if _, ok := this.kefuchat_stats[kefuid][visitorId]; ok {
			delete(this.kefuchat_stats[kefuid], visitorId)
		}
	}
}

const ONE_KEFU_MAX_VISITORS = 10

func (this *ChatCenter_Kefus) AssignOnlineKefu() *ChatUser {
	this.m_s.RLock()
	defer this.m_s.RUnlock()

	var (
		min, n uint
		minK   uint64
		ok     bool
	)

	for key, val := range this.Users.Copy() {
		k, ok1 := key.(uint64)
		v, ok2 := val.(*ChatUser)
		if !ok1 || !ok2 {
			continue
		}

		if v.IsOnline() && v.UManager == 1 {
			if _, ok = this.kefuchat_stats[k]; !ok {
				return v
			}

			n = uint(len(this.kefuchat_stats[k]))
			if n == 0 {
				return v
			}
			//if n > ONE_KEFU_MAX_VISITORS {
			//	continue
			//}

			if min == 0 || n < min {
				min = n
				minK = k
			}
		}
	}

	if minK > 0 {
		return this.Find(minK)
	}

	return nil
}

func (this *ChatCenter_Kefus) OnlineUsers() []*model.User {
	this.m_s.RLock()
	defer this.m_s.RUnlock()

	users := []*model.User{}

	for _, val := range this.Users.Values() {
		v, ok := val.(*ChatUser)
		if !ok {
			continue
		}

		if v.IsOnline() {
			if _, ok := this.kefuchat_stats[v.Uid]; ok {
				v.User.SetUintField(model.CHAT_CNT, uint(len(this.kefuchat_stats[v.Uid])))
			}
			users = append(users, v.User)
		}
	}

	return users
}

func (this *ChatCenter_Kefus) GetAllVisitors(kefuid uint64) map[uint64]bool {
	this.m_s.RLock()
	defer this.m_s.RUnlock()

	if _, ok := this.kefuchat_stats[kefuid]; !ok {
		this.kefuchat_stats[kefuid] = make(map[uint64]bool)
	}

	return this.kefuchat_stats[kefuid]
}

func (this *ChatCenter_Kefus) GetAllKefus() []*ChatUser {
	result := []*ChatUser{}
	for _, val := range this.Users.Copy() {
		if v, ok := val.(*ChatUser); ok {
			result = append(result, v)
		}
	}
	return result
}

func (this *ChatCenter_Kefus) CleanUpVisitor(kefuid, visitorId uint64) {
	this.m_s.Lock()
	defer this.m_s.Unlock()

	if _, ok := this.kefuchat_stats[kefuid]; !ok {
		return
	}

	if _, ok := this.kefuchat_stats[kefuid][visitorId]; !ok {
		return
	}

	delete(this.kefuchat_stats[kefuid], visitorId)
}

func (this *ChatCenter_Kefus) CleanUpKefu(kefuid uint64) {
	this.m_s.Lock()
	if _, ok := this.kefuchat_stats[kefuid]; ok {
		delete(this.kefuchat_stats, kefuid)
	}
	this.m_s.Unlock()

	this.Users.Del(kefuid)
}

func (this *ChatCenter_Kefus) Report() (ret []string) {
	ret = append(ret, fmt.Sprintf("Total online admin users : %v", this.Users.Len()))
	this.m_s.RLock()
	ret = append(ret, fmt.Sprintf("Total online kefus : %v", len(this.kefuchat_stats)))
	var chatCnt int
	for kefuid, visitors := range this.kefuchat_stats {
		chatCnt += len(visitors)

		ret = append(ret, fmt.Sprintf("    kefu[id=%v] : %v visitors", kefuid, len(visitors)))
	}
	this.m_s.RUnlock()
	ret = append(ret, fmt.Sprintf("Total chats : %v", chatCnt))
	return
}
