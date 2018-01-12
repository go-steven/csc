package chat

const (
	MSG_TYPE_CHAT, MSG_TYPE_SYS uint = 1, 9
	MSG_VISITOR_STS             uint = 21
)

type ChatMsg struct {
	MsgTime   string `json:"msg_time"`
	MsgType   uint   `json:"msg_type"`
	UserId    uint64 `json:"user_id"`
	UserName  string `json:"user_name"`
	Affiliate string `json:"affiliate"`
	KefuId    uint64 `json:"kefu_id"`
	KefuName  string `json:"kefu_name"`
	KefuNick  string `json:"kefu_nick"`
	Initiator uint   `json:"initiator"`
	Msg       string `json:"msg"`

	FromUser *ChatUser `json:"-"`
}

func NewChatMsg(msgTime string, visitor, kefu *ChatUser, msg string, initiator uint) *ChatMsg {
	chatMsg := &ChatMsg{
		MsgTime:   msgTime,
		Msg:       msg,
		MsgType:   MSG_TYPE_CHAT,
		Initiator: initiator,
	}

	if visitor != nil {
		userCopy := visitor.User.Copy()
		chatMsg.UserId = userCopy.Uid
		chatMsg.UserName = userCopy.Name
		chatMsg.Affiliate = userCopy.Affiliate
	}

	if kefu != nil {
		userCopy := kefu.User.Copy()
		chatMsg.KefuId = userCopy.Uid
		chatMsg.KefuName = userCopy.Name
		chatMsg.KefuNick = userCopy.Nick
	}

	return chatMsg
}

func (this *ChatMsg) Json() string {
	return Json(this)
}

func NewSysMsg(msg string, toUser *ChatUser) *ChatMsg {
	userCopy := toUser.User.Copy()

	chatMsg := &ChatMsg{
		MsgTime:   NowStr(),
		MsgType:   MSG_TYPE_SYS,
		UserId:    userCopy.Uid,
		UserName:  userCopy.Name,
		Affiliate: userCopy.Affiliate,
		Msg:       msg,
	}

	return chatMsg
}

func NewNotifyMsg(msg string) string {
	chatMsg := &ChatMsg{
		MsgTime: NowStr(),
		MsgType: MSG_TYPE_SYS,
		Msg:     msg,
	}

	return chatMsg.Json()
}

type VisitorStsMsg struct {
	MsgTime      string `json:"msg_time"`
	MsgType      uint   `json:"msg_type"`
	Uid          uint64 `json:"user_id"`
	Name         string `json:"user_name"`
	Affiliate    string `json:"affiliate"`
	Status       uint   `json:"status"`
	RealKefuId   uint64 `json:"kefu_id"`
	RealKefuName string `json:"kefu_name"`
	ChatStatus   uint   `json:"chat_sts"`
	ScLevel      uint   `json:"sc_level"`
	DepositCnt   uint   `json:"deposit_cnt"`
}
