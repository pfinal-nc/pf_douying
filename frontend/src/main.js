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

function stop() {
    console.log("stop!")
    window.go.main.App.LiveRoom().then(result => {
        start_status = false;
    }).catch(err => {
        console.log(err);
    }).finally(() => {
        console.log("finished!")
    });
}

setInterval(function () {
    console.log(start_status)
    if (start_status) {
        window.go.main.App.GetRoomMsg().then(result => {
            console.log(result);
            // 渲染到 页面
            if (result) {
                var tr = '<tr class="relative transform scale-100 text-xs py-1">\n' +
                    '                    <td class="pl-5 px-2 py-2 whitespace-no-wrap">\n' +
                    '                        <div class="leading-5 text-write">' + result +
                    '</div>\n' +
                    '                    </td>\n' +
                    '                </tr>'
                $(".messages").append(tr)
                $('#journal-scroll').scrollTop($("#journal-scroll")[0].scrollHeight);
            }
        }).catch(err => {
            console.log(err);
        }).finally(() => {
            console.log("finished!")
        });
    }

}, 1000);