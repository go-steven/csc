安装：
go get github.com/gin-gonic/gin
go get github.com/gin-gonic/contrib/static
go get github.com/igm/sockjs-go/sockjs(目前使用master分支)
go get github.com/nu7hatch/gouuid

部署：
github.com/XiBao/csc

测试启动：：
cd github.com/XiBao/csc
cd server/bin/csc 
go build
./csc --log=$LOG_FILE --port=$PORT(默认10080)

测试客户端：
http://csc.xibao100.com/client/visitor.html?uid=$UID
http://csc.xibao100.com/client/kefu.html?kefuid=$KEFUID

相关接口：
一. 建立与服务端的TCP连接：
1. http://csc.xibao100.com/chat/visitor/$UID/
建立客户与服务器的连接
2. http://csc.xibao100.com/chat/kefu/$KEFUID/
建立客服与服务器的连接

二. 会话操作
1. http://csc.xibao100.com/rest/chat/transfer?uid=$UID&fromkefu=$FROMKEFUID&tokefu=$TOKEFUID&callback=$CALLBACK
客服（会话）转移
2. http://csc.xibao100.com/rest/chat/stop?uid=$UID&kefuid=$KEFUID&callback=$CALLBACK
结束（客服）会话

三. 客服操作
1. http://csc.xibao100.com/rest/user/kefu?kefuid=$KEFUID&online=$ONLINE&callback=$CALLBACK
   $ONLINE: 1-上线，0：下线
客服上线/离线
2. http://csc.xibao100.com/rest/user/manage?kefuid=$KEFUID&auto=$AUTO&callback=$CALLBAC
   $AUTO: 1-自动分配，0：非自动分配
客服设置自动分配标志
3. http://csc.xibao100.com/rest/user/call?kefuid=$KEFUID&uid=$UID&callback=$CALLBAC
客服主动呼叫

四. 查询
1. http://csc.xibao100.com/rest/user/list?role=$ROLE&chatsts=&callback=$CALLBACK
查询在线客服(role=2)/客户列表(role=1), role=1时chatsts参数可选（取值范围：1，2，3，可以用逗号分隔多个值）
取值含义见：CHAT_STS_NO, CHAT_STS_ON, CHAT_STS_WAIT uint = 1, 2, 3

2. http://csc.xibao100.com/rest/user/info?kefuid=$KEFUID&callback=$CALLBACK
查询在线客服信息

五. 常量定义
```go```
const (
    VISITOR, KEFU   uint = 1, 2
    ONLINE, OFFLINE uint = 0, 1

    ALLOWED_MSG_MAX_LEN  = 500
    ALLOWED_MSG_MAX_RATE = 5 // per second

    ONE_KEFU_MAX_VISITORS = 10

    MSG_TYPE_CHAT, MSG_TYPE_SYS uint = 1, 9

    MSG_VISITOR_STS uint = 21

    CHAT_STS_NO, CHAT_STS_ON, CHAT_STS_WAIT uint = 1, 2, 3
)
```
