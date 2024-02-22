package lib

/**
 * @Author: PFinal南丞
 * @Author: lampxiezi@163.com
 * @Date: 2024/2/6
 * @Desc:
 * @Project: pf_douying
 */
import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
)

// MessageChan 创建一个 存储消息的通道
var MessageChan = make(chan string, 50)

// StateChan 创建一个 存储状态的通道
var StateChan = make(chan string, 1)

func generateMsToken(length int) string {
	randomStr := ""
	baseStr := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789=_"
	_len := len(baseStr) - 1
	for i := 0; i < length; i++ {
		randomStr += string(baseStr[rand.Intn(_len)])
	}
	return randomStr
}

func generateTtwid() string {
	url := "https://live.douyin.com/"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request: ", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request: ", err)
	}
	defer resp.Body.Close()

	return resp.Cookies()[0].Value
}

type DouyinLiveWebFetcher struct {
	ttwid     string
	roomID    string
	liveID    string
	liveURL   string
	userAgent string
	ws        *websocket.Conn
}

func NewDouyinLiveWebFetcher(liveID string) *DouyinLiveWebFetcher {
	return &DouyinLiveWebFetcher{
		liveID:    liveID,
		liveURL:   "https://live.douyin.com/",
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

func (d *DouyinLiveWebFetcher) Start() {
	d.connectWebSocket()
	defer d.Stop()
}

func (d *DouyinLiveWebFetcher) Stop() {
	_ = d.ws.Close()
}

func (d *DouyinLiveWebFetcher) Ttwid() string {
	if d.ttwid != "" {
		return d.ttwid
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", d.liveURL, nil)
	if err != nil {
		log.Fatal("Error creating request: ", err)
	}
	req.Header.Set("User-Agent", d.userAgent)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request: ", err)
	}
	defer resp.Body.Close()

	d.ttwid = resp.Cookies()[0].Value
	return d.ttwid
}

func (d *DouyinLiveWebFetcher) RoomID() string {
	if d.roomID != "" {
		return d.roomID
	}
	url := d.liveURL + d.liveID

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request: ", err)
	}
	req.Header.Set("User-Agent", d.userAgent)
	req.Header.Set("Cookie", fmt.Sprintf("ttwid=%s&msToken=%s; __ac_nonce=0123407cc00a9e438deb4", d.Ttwid(), generateMsToken(107)))
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request: ", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body: ", err)
	}
	match := regexp.MustCompile(`roomId\\":\\"(\d+)\\"`).FindStringSubmatch(string(body))
	if match == nil || len(match) < 2 {
		log.Fatal("No match found for roomId")
	}

	d.roomID = match[1]
	return d.roomID
}

func (d *DouyinLiveWebFetcher) connectWebSocket() {
	d.RoomID()
	wss := "wss://webcast3-ws-web-lq.douyin.com/webcast/im/push/v2/?" +
		"app_name=douyin_web&version_code=180800&webcast_sdk_version=1.3.0&update_version_code=1.3.0" +
		"&compress=gzip" +
		"&internal_ext=internal_src:dim|wss_push_room_id:" + d.roomID +
		"|wss_push_did:" + d.roomID +
		"|dim_log_id:202302171547011A160A7BAA76660E13ED|fetch_time:1676620021641|seq:1|wss_info:0-1676" +
		"620021641-0-0|wrds_kvs:WebcastRoomStatsMessage-1676620020691146024_WebcastRoomRankMessage-167661" +
		"9972726895075_AudienceGiftSyncData-1676619980834317696_HighlightContainerSyncData-2&cursor=t-1676" +
		"620021641_r-1_d-1_u-1_h-1" +
		"&host=https://live.douyin.com&aid=6383&live_id=1" +
		"&did_rule=3&debug=false&endpoint=live_pc&support_wrds=1&" +
		"im_path=/webcast/im/fetch/&user_unique_id=" + d.roomID +
		"&device_platform=web&cookie_enabled=true&screen_width=1440&screen_height=900&browser_language=zh&" +
		"browser_platform=MacIntel&browser_name=Mozilla&" +
		"browser_version=5.0%20(Macintosh;%20Intel%20Mac%20OS%20X%2010_15_7)%20AppleWebKit/537.36%20(KHTML,%20" +
		"like%20Gecko)%20Chrome/110.0.0.0%20Safari/537.36&" +
		"browser_online=true&tz_name=Asia/Shanghai&identity=audience&room_id=" + d.roomID +
		"&heartbeatDuration=0&signature=00000000"
	dialer := websocket.DefaultDialer
	header := http.Header{"Cookie": []string{fmt.Sprintf("ttwid=%s", d.Ttwid())}, "User-Agent": []string{d.userAgent}}
	c, _, err := dialer.Dial(wss, header)
	if err != nil {
		log.Fatal("WebSocket connection error: ", err)
	}
	defer func(c *websocket.Conn) {
		_ = c.Close()
	}(c)
	d.ws = c
	d.wsOnOpen()
	d.wsLoop()
}

func (d *DouyinLiveWebFetcher) wsOnOpen() {
	fmt.Println("WebSocket connected.")
	MessageChan <- "连接直播间成功"
}

func (d *DouyinLiveWebFetcher) wsLoop() {
	for {
		// 监听状态通道 如果状态通道中有消息 则停止循环
		select {
		case <-StateChan:
			fmt.Println("停了")
			//d.Stop()
			//return
		default:
			// 继续循环
		}
		_, message, err := d.ws.ReadMessage()
		// fmt.Println(message)
		if err != nil {
			log.Println("WebSocket read error: ", err)
			break
		}

		p := &PushFrame{} // Parse PushFrame from message
		_ = proto.Unmarshal(message, p)
		if err != nil {
			log.Fatal("Error unmarshaling push frame:", err)
		}
		response := &Response{} // Parse Response from package payload
		//fmt.Println(p.Payload)
		//err = proto.Unmarshal(p.Payload, response)
		//if err != nil {
		//	log.Fatal("Error unmarshaling response:", err)
		//}
		// 使用 gzip 包中的 NewReader 函数创建一个解压缩器
		reader, err := gzip.NewReader(bytes.NewReader(p.Payload))
		if err != nil {
			log.Fatal("Error creating gzip reader:", err)
		}
		defer reader.Close()
		// 读取解压后的数据到一个 bytes.Buffer 中
		var uncompressedData bytes.Buffer
		_, err = io.Copy(&uncompressedData, reader)
		if err != nil {
			log.Fatal("Error reading uncompressed data:", err)
		}
		// 解析响应数据
		err = proto.Unmarshal(uncompressedData.Bytes(), response)
		if response.NeedAck {
			ack := &PushFrame{ // Construct ack
				LogId:       p.LogId,
				PayloadType: "ack",
				Payload:     []byte(response.InternalExt),
			}
			ackBytes, err := proto.Marshal(ack)
			if err != nil {
				log.Println("Error marshaling ack message:", err)
				return
			}
			d.wsWrite(ackBytes)
		}
		for _, msg := range response.MessagesList {
			// fmt.Println(msg.Method)
			switch msg.Method {
			case "WebcastChatMessage":
				d.parseChatMsg(msg.Payload)
			case "WebcastGiftMessage":
				d.parseGiftMsg(msg.Payload)
			case "WebcastLikeMessage":
				d.parseLikeMsg(msg.Payload)
			case "WebcastMemberMessage":
				d.parseMemberMsg(msg.Payload)
			case "WebcastSocialMessage":
				d.parseSocialMsg(msg.Payload)
			case "WebcastRoomUserSeqMessage":
				d.parseRoomUserSeqMsg(msg.Payload)
			case "WebcastFansclubMessage":
				d.parseFansclubMsg(msg.Payload)
			case "WebcastControlMessage":
				d.parseControlMsg(msg.Payload)
			}
		}
	}
}

func (d *DouyinLiveWebFetcher) wsWrite(message []byte) {
	err := d.ws.WriteMessage(websocket.BinaryMessage, message)
	if err != nil {
		log.Println("WebSocket write error: ", err)
	}
}

func (d *DouyinLiveWebFetcher) parseChatMsg(payload []byte) {
	// '聊天消息'
	//fmt.Println("Chat Message")
	message := &ChatMessage{}
	err := proto.Unmarshal(payload, message)
	if err != nil {
		log.Fatal("Failed to parse ChatMessage:", err)
	}
	// fmt.Println("【聊天msg】", message)
	// SendMsg := fmt.Sprintf("【聊天msg】: %s", message)
	// MessageChan <- SendMsg
}

func (d *DouyinLiveWebFetcher) parseGiftMsg(payload []byte) {
	// 礼物消息
	//fmt.Println("Gift Message")
	message := &GiftMessage{}
	err := proto.Unmarshal(payload, message)
	if err != nil {
		log.Fatal("Failed to parse GiftMessage:", err)
	}
	// SendMsg := fmt.Sprintf("【礼物msg】: %s", message)
	// fmt.Println("礼物msg】", message)
	// MessageChan <- SendMsg
}

func (d *DouyinLiveWebFetcher) parseLikeMsg(payload []byte) {
	// 点赞消息
	fmt.Println("Like Message")
}

func (d *DouyinLiveWebFetcher) parseMemberMsg(payload []byte) {
	// 进入直播间消息
	//fmt.Println("Member Message")
	message := &MemberMessage{}
	err := proto.Unmarshal(payload, message)
	if err != nil {
		log.Fatal("Failed to parse GiftMessage:", err)
	}
	fmt.Println("【进场msg】" + message.GetUser().NickName)
	SendMsg := fmt.Sprintf("【进场msg】: %s", message.GetUser().NickName)
	MessageChan <- SendMsg
}

func (d *DouyinLiveWebFetcher) parseSocialMsg(payload []byte) {
	// Parse SocialMessage from payload
	// fmt.Println("Social Message")
}

func (d *DouyinLiveWebFetcher) parseRoomUserSeqMsg(payload []byte) {
	// Parse RoomUserSeqMessage from payload
	// fmt.Println("Room User Sequence Message")
}

func (d *DouyinLiveWebFetcher) parseFansclubMsg(payload []byte) {
	// Parse FansclubMessage from payload
	// fmt.Println("Fansclub Message")
}

func (d *DouyinLiveWebFetcher) parseControlMsg(payload []byte) {
	// Parse ControlMessage from payload
	// fmt.Println("Control Message")
}
