package websocket

import (
	"fmt"
)

func FAQStr() string {
	result := fmt.Sprintf("<p>%s</p>", faqPrompts["QA_HEAD"])
	for i := 0; i < len(faqList); i++ {
		result = fmt.Sprintf("%s<p>%d#.	%s</p>", result, i+1, faqList[fmt.Sprintf("%d#", i+1)]["Q"])
	}
	result = fmt.Sprintf("%s<p>%s</p>", result, faqPrompts["QA_TAIL"])

	return result
}

func FAQCheck(input string) (bool, string) {
	msgLen := len([]rune(input))
	if msgLen == 0 {
		return false, ""
	}

	if msgLen <= 2 {
		input = special_char_mapping(input)
		if input == "?#" {
			return true, FAQStr()
		}
		if len(input) == 1 {
			input = input + "#"
		}
	}

	key, ok := faqKeys[input]
	if !ok {
		return false, ""
	}
	v := faqList[key]
	return true, fmt.Sprintf("<p>%s.	%s</p><p>&nbsp;&nbsp;</p><p>回答：%s<p></p>", key, v["Q"], v["A"])
}

func FAQKefuCheck(input string) (uint, string) {
	input = special_char_mapping(input)
	if input == "?#" {
		return 2, FAQStr()
	}

	v, ok := faqList[input]
	if !ok {
		return 0, ""
	}

	return 1, v["A"]
}

var faqPrompts = map[string]string{
	"QA_HEAD": "常见问题",
	"QA_TAIL": "回复问题编号加#号，比如 <font color='red'>1#</font> ，即可查看相应问题的回答。如还有其他问题，在线客服将会为您服务。",
}

var faqList = map[string]map[string]string{
	"1#": {
		"Q": "我订购了软件为什么还要充值？",
		"A": "亲，我们的软件是免费订购的。投放推广类似于直通车，按点击付费，所以投放之前需要先充值，充值后即可进行自助投放，只有有点击时才会扣钱，您充值的每一分钱都会用于真实的流量推广。",
	},
	"2#": {
		"Q": "投放后我的宝贝展示在哪里？",
		"A": `投放后，您的宝贝将在360购物、百度贴吧、网易、新浪、优酷、爱奇艺等数百家淘外网站的广告位置，根据时段、地域、人群等条件匹配，向潜在目标人群进行推广展示，为您的店铺带来真实流量。
		您可以点击【推广宝贝】页面内的“投放展示位范例”，预览部分广告展示位置。`,
	},
	"3#": {
		"Q": "你们这里是怎么计费的？",
		"A": "我们和直通车一样，按点击计费，只有用户点击您的广告进入店铺时才会扣钱，您可以自主控制投放出价和每日最高消耗金额，把握自己的推广预算。目前我们默认出价为0.6元，但实际每次点击付费会低于0.6元哦。",
	},
	"4#": {
		"Q": "投放后我怎么查看推广效果？",
		"A": "投放后，从第二天起您即可在【我的报表】页面查看到每天的投放成果，您也可以通过观察店铺里的流量变化进行评估。和直通车类似，一般稳定投放一到两周左右看效果最好，如果到时发现投放效果不尽人意，可咨询我们的在线客服，我们的客服可为您提供优化建议。",
	},
	"5#": {
		"Q": "推广的转化率与什么因素有关？",
		"A": "转化率与所投放宝贝的类目、性价比、图片、店铺等级、评分等息息相关，所以，各个店铺的转化率各不相同，没有统一的标准。目前，我们整体转化大概在3%左右。如果您投放后转化率较低，可以咨询在线客服为您提供优化建议，或者通过更换投放宝贝、优化宝贝标题图片等方式，优化投放策略，逐步提升转化率。",
	},
}

var faqKeys = map[string]string{
	"1#": "1#",
	"2#": "2#",
	"3#": "3#",
	"4#": "4#",
	"5#": "5#",

	"1#.	我订购了软件为什么还要充值？": "1#",
	"1#.	我订购了软件为什么还要充值": "1#",
	"我订购了软件为什么还要充值":  "1#",
	"我订购了软件为什么还要充值？": "1#",

	"2#.	投放后我的宝贝展示在哪里？": "2#",
	"2#.	投放后我的宝贝展示在哪里": "2#",
	"投放后我的宝贝展示在哪里":  "2#",
	"投放后我的宝贝展示在哪里？": "2#",

	"3#.	你们这里是怎么计费的？": "3#",
	"3#.	你们这里是怎么计费的": "3#",
	"你们这里是怎么计费的":  "3#",
	"你们这里是怎么计费的？": "3#",

	"4#.	投放后我怎么查看推广效果？": "4#",
	"4#.	投放后我怎么查看推广效果": "4#",
	"投放后我怎么查看推广效果":  "4#",
	"投放后我怎么查看推广效果？": "4#",

	"5#.	推广的转化率与什么因素有关？": "5#",
	"5#.	推广的转化率与什么因素有关": "5#",
	"推广的转化率与什么因素有关":  "5#",
	"推广的转化率与什么因素有关？": "5#",
}

// 特殊字符映射: 全角 --> 半角
var specialCharMap = map[string]string{
	"？": "?",
	"＃": "#",
	"０": "0",
	"１": "1",
	"２": "2",
	"３": "3",
	"４": "4",
	"５": "5",
	"６": "6",
	"７": "7",
	"８": "8",
	"９": "9",
}

func special_char_mapping(s string) (ret string) {
	for _, v := range []rune(s) {
		if val, found := specialCharMap[string(v)]; found {
			ret += val
		} else {
			ret += string(v)
		}
	}

	return
}
