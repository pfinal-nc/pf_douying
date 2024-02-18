//Global JS function for greeting
let start_status = false;

function greet() {
    let roomId = document.getElementById("room-id").value;
    if (roomId) {
        window.go.main.App.JoinRoom(roomId).then(result => {
            console.log(result);
            // 循环调用 GetRoomMsg 方法 获取消息 渲染到 页面
            start_status = true
        }).catch(err => {
            console.log(err);
        }).finally(() => {
            console.log("finished!")
        });
    } else {
        console.log("请输入房间号");
    }
}

setInterval(function () {
    console.log(start_status)
    if (start_status) {
        window.go.main.App.GetRoomMsg().then(result => {
            console.log(result);
            // 渲染到 页面
            var li = document.createElement("li")
            li.textContent = result
            var message_con = document.getElementById("messages")
            message_con.appendChild(li)

            // 滚动到底部的函数
            function scrollToBottom() {
                message_con.scrollTop = message_con.scrollHeight;
            }
            scrollToBottom();
        }).catch(err => {
            console.log(err);
        }).finally(() => {
            console.log("finished!")
        });
    }

}, 1000);