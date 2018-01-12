document.getElementById("output").value = ""

var uid = getParameter("uid")
var websock_url = csc_server + '/chat/visitor/' + uid + '/'
//alert(websock_url)
var sock = new SockJS(websock_url)

sock.onopen = function() {
    var myDate = new Date();
	console.log('connection open at: ' + myDate.toLocaleTimeString());
	document.getElementById("status").innerHTML = "connected";
	document.getElementById("send").disabled=false;
};

sock.onmessage = function(e) {
    var json = eval("("+e.data+")");
    if (json.msg_type == '1' || json.msg_type == '9') {
        if (json.initiator == 1) {
            document.getElementById("output").value += "[" + json.msg_time + " " + json.user_name + " " + json.msg_type + "]" + json.msg + "\n";
        } else {
            document.getElementById("output").value += "[" + json.msg_time + " " + json.kefu_nick + " " + json.msg_type + "]" + json.msg + "\n";
        }
    } else {
        document.getElementById("output").value += e.data + "\n";
    }
};

sock.onclose = function() {
    var myDate = new Date();
	console.log('connection closed at: ' + myDate.toLocaleTimeString());
	document.getElementById("status").innerHTML = "disconnected";
	document.getElementById("send").disabled=true;
};

//window.open('http://www.baidu.com')
