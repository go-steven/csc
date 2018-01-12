package chat

var Prompts = map[string]string{
	"MSG_WELCOME":      "欢迎您，我是在线客服，很高兴为您服务，请问有什么可以为您解答的吗？",
	"MSG_USER_WAITING": "您好！客户[%s]已分配给您，等待您的处理！",

	"MSG_LIMIT_TOO_SHORT": "输入信息不能为空.",
	"MSG_LIMIT_TOO_LONG":  "输入信息过长.",
	"MSG_LIMIT_TOO_FAST":  "输入信息过快，请稍后...",

	"KEFU_TRANSFERED_TO":   "您好！%s已把客户[%s]转移给您，请尽快处理！.",
	"KEFU_TRANSFERED_USER": "客服已转接，接下来由%s为您服务.",
	"KEFU_TRANSFERED_FROM": "客户已转移给客服[%s].",
	"KEFU_STOP_CHAT":       "客服[%s]结束了当前会话.",

	"KEFU_NOT_FOUND": "目前没有客服在线，请稍候.",
	"KEFU_ONLINE":    "%s已上线，有什么需要为您服务的吗?",
	"KEFU_OFFLINE":   "%s已离开",

	"INVALID_SESSION_PARAMS": "建立会话的参数错误",
	"KEFU_IS_OFFLINE":        "客服处于离线状态，请上线.",
	"VISITOR_IS_OFFLINE":     "该客户不在线.",
	"OTHER_KEFU_PROCESSED":   "客服[%s]已在处理该客户的在线咨询.",
	"SESSION_CLOSED":         "当前会话结束.",
}
