package net

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"strings"
	"time"
	"chatroom/util"
	"os"
	//"encoding/binary"
	//"io/ioutil"
	"fmt"
	"log"
)

type ClientAction int

const (
	Error          ClientAction = -1 // 通用错误，发生错误时，Data请传递`ClientError`结构体
	Message        ClientAction = 0  // 普通消息
	CreateRoom     ClientAction = 1  // 创建房间
	JoinRoom       ClientAction = 2  // 加入房间
	ChangedRoom    ClientAction = 3  // 房间信息变更
	GetRoomInfo 	 ClientAction = 4  // 获取房间信息
	StartFrameSync ClientAction = 5  // 开启帧同步
	StopFrameSync  ClientAction = 6  // 停止帧同步
	UploadFrame    ClientAction = 7  // 上传帧同步数据
	Login          ClientAction = 8  // 登陆用户
	FData          ClientAction = 9  // 帧数据
	RoomMessage    ClientAction = 10 // 发送房间消息
	GetRoomList    ClientAction = 11 // 获取房间列表
	ExitRoom    	 ClientAction = 12 // 退出房间
)

type ClientMessage struct {
	Op   ClientAction
	Data any
}

type ClientError struct {
	Code ClientErrorCode
	Msg  string
}

type ClientErrorCode int

const (
	NORMAL_SUCCESS		   ClientErrorCode = 0	  // 正常无错误
	CREATE_ROOM_ERROR      ClientErrorCode = 1001 // 创建房间错误
	ENTER_ROOM_ERROR       ClientErrorCode = 1002 // 进入房间错误
	START_FRAME_SYNC_ERROR ClientErrorCode = 1003 // 启动帧同步错误
	STOP_FRAME_SYNC_ERROR  ClientErrorCode = 1004 // 停止帧同步错误
	UPLOAD_FRAME_ERROR     ClientErrorCode = 1005 // 上传帧同步数据错误
	LOGIN_ERROR            ClientErrorCode = 1006 // 登陆失败
	LOGIN_OUT_ERROR        ClientErrorCode = 1007 // 在别处登陆事件
	OP_ERROR               ClientErrorCode = 1008 // 无效的操作指令
	SEND_ROOM_ERROR        ClientErrorCode = 1009 // 发送房间消息错误
	GET_ROOM_ERROR				 ClientErrorCode = 1009 // 获取房间信息错误
)

type Opcode int

const (
	Continuation Opcode = 0x00
	Text         Opcode = 0x01
	Binary       Opcode = 0x02
	Close        Opcode = 0x08
	Ping         Opcode = 0x09
	Pong         Opcode = 0x0A
)

type State int

const (
	Handshake       State = iota // 握手状态
	Head                         // 读取Head
	HeadExtraLength              // 读取内容长度
	HeadExtraMask                // 读取掩码
	Body                         // 读取内容
)

type Client struct {
	net.Conn
	websocket     bool           // 是否使用webscoket协议
	handshakeData string         // 握手信息
	data          util.Bytes     // 缓存数据
	isFinal       bool           // 是否最终包
	opcode        Opcode         // 操作符
	frameIsBinary bool           // 是否二进制数据
	partialLength int            // 内容长度
	isMasked      bool           // 是否存在掩码
	state         State          // 状态码
	length        int            // 长度
	mask          []byte         // 掩码数据
	lastPong      int64          // 上一次心跳时间
	room          *Room          // 房间（每个用户同一时间只会进入到一个房间中）
	userData      map[string]any // 用户自定义数据
	frames        util.Array     // 用户帧同步缓存操作
	uid           int            // 用户ID
	name          string         // 用户名称
	passwd        string         // 用户密码
	online        bool           // 是否在线
}

// 发送数据给所有人
func (c *Client) SendToAllUser(data []byte) {
	for _, v := range CurrentServer.users.List {
		v.(*Client).SendToUser(data)
	}
}

func (c *Client) GenFilePath() string {
	if (c == nil ) {
		return "data/00/00/100.log"
	}
	datapath := "data/" + util.GetSubPath(uint64(c.uid))
	util.EnsureDir(datapath)
	file := datapath + "/"  + util.Num2str(c.uid) + ".log"
	return file
}

func (c *Client) WriteToFile( msg string) {
	file := c.GenFilePath()
	file6, err := os.OpenFile(file, os.O_RDWR | os.O_SYNC | os.O_APPEND | os.O_CREATE, 0766)
	if err != nil {
		log.Println("Error Can't open the file.")

	}
	file6.WriteString(msg + "\n")
	file6.Close()
}

func (c *Client) RemoveLogFile() {
	file := c.GenFilePath()
	err := os.Remove(file)
	if err != nil {
		log.Println("Remove file Error!", err)
	}
}

// 单独发送数据到当前用户
func (c *Client) SendToUser(data []byte) {
	c.Write(data)
}

// 发送客户端数据到当前用户
func (c *Client) SendToUserOp(data *ClientMessage) {
	v, err := json.Marshal(data)
	
	if err == nil {
		// 发送
		// log.Println("发送内容：", string(v))
		bdata := prepareFrame(v, Text, true)
		c.SendToUser(bdata.Data)
	}
}

// 发送telnet的响应消息到客户端
func (c *Client) SendTelnetMsgToUser(data *ClientMessage) {
	

	messg :=  data.Data.(string)
	data.Data = messg

	v, err := json.Marshal(data)
	if err == nil {
		// 发送
		// log.Println("发送内容：", string(v))
		bdata := prepareFrame(v, Text, true)
		c.SendToUser(bdata.Data)
	}
}

// 发送二进制文件到客户端
func (c *Client) SendBinaryFileToUser(data *ClientMessage) {
	file := data.Data.(string)

	/*
	fbytes, err := ioutil.ReadFile(file)
    if err != nil {
        fmt.Println("read fail", err)
    }
	if err == nil {
		// 发送
		bdata := prepareFrame(fbytes, Binary, false)
		c.SendToUser(bdata.Data)
	}*/


    f, err := os.Open(file)
    if err != nil {
        fmt.Println("read fail")
        return 
    }

    defer f.Close()

    buf := make([]byte, 1024)
	lastn := 0
    for {
        n, err := f.Read(buf)
        if err != nil && err != io.EOF{
            fmt.Println("read buf fail", err)
            return 
        }

		// sendbuf := make([]byte, 1028)
		// binary.BigEndian.PutUint32(sendbuf[0:4], uint32(n))
		// sendbuf = append(sendbuf, buf[0:1024]...)
		// fmt.Println(n)

        if n == 0 {
			if (lastn == 1024) {
				sendbuf := []byte{0x03}
				bdata := prepareFrame(sendbuf, Binary, false)
				c.SendToUser(bdata.Data)
			}
            break
        }else{
			lastn = n
			bdata := prepareFrame(buf[0:n], Binary, true)
			c.SendToUser(bdata.Data)
		}
    }

}

// 数据缓存处理
func (c *Client) onData(data []byte) {
	c.data.WriteBytes(data)
	if c.state == Handshake {
		// 接收到结束符
		cdata := c.data.ReadUTFString(c.data.ByteLength())
		c.handshakeData += cdata
		index := strings.Index(c.handshakeData, "\r\n\r\n")
		if index != -1 {
			// 开始握手
			c.handshake(c.handshakeData)
		}
	} else {
		// todo 这里需要解析websocket的数据结构
		data, ok := ReadWebSocketData(c)
		if ok {
			log.Println(string(data))
		}
	}
}

// 发送一个WebSocket包
func (c *Client) WriteWebSocketData(data []byte, opcode Opcode) {
	var dataContent = prepareFrame(data, opcode, true).Data
	c.SendToUser(dataContent)
	log.Println("Length sent", len(dataContent))
}

// 包装成一个WebSocket包
func prepareFrame(data []byte, opcode Opcode, isFinal bool) util.Bytes {
	newdata := util.Bytes{Data: []byte{}}
	var isMasked = false // All clientes messages must be masked: http://tools.ietf.org/html/rfc6455#section-5.1
	var mask = generateMask()
	var sizeMask = 0x00
	if isMasked {
		sizeMask = 0x80
	}
	var sizeFinal = 0x00
	if isFinal {
		sizeFinal = 0x80
	}
	newdata.Write(int(opcode) | sizeFinal)
	var byteLength = len(data)
	if byteLength < 126 {
		newdata.Write(byteLength | sizeMask)
	} else if byteLength < 65536 {
		newdata.Write(126 | sizeMask)
		newdata.WriteShort(byteLength)
	} else {
		newdata.Write(127 | sizeMask)
		newdata.Write(0)
		newdata.Write(byteLength)
	}
	if isMasked {
		for i := 0; i < 4; i++ {
			newdata.Data = append(newdata.Data, mask[i])
		}
		maskdata := applyMask(data, mask[:])
		newdata.WriteBytes(maskdata)
	} else {
		newdata.WriteBytes(data)
	}
	return newdata
}

func generateMask() [4]byte {
	var maskData = [4]byte{}
	maskData[0] = byte(rand.Intn(256))
	maskData[1] = byte(rand.Intn(256))
	maskData[2] = byte(rand.Intn(256))
	maskData[3] = byte(rand.Intn(256))
	return maskData
}

// 读取WebSocket的数据包
func ReadWebSocketData(c *Client) ([]byte, bool) {
	var byteLength = c.data.ByteLength()
	switch c.state {
	case Head:
		// 字节少于2的时候，意味着数据不足
		if byteLength < 2 {
			return nil, false
		}
		b0 := c.data.ReadInt()
		b1 := c.data.ReadInt()
		c.isFinal = ((b0 >> 7) & 1) != 0
		c.opcode = Opcode(((b0 >> 0) & 0xF))
		if c.opcode == Text {
			c.frameIsBinary = false
		} else if c.opcode == Binary {
			c.frameIsBinary = true
		}
		c.partialLength = ((b1 >> 0) & 0x7F)
		c.isMasked = ((b1 >> 7) & 1) != 0

		// util.Log(b0, b1)
		// util.Log("c.isFinal=", c.isFinal)
		// util.Log("c.isMasked=", c.isMasked)
		// util.Log("c.opcode=", c.opcode)
		// util.Log("c.partialLength=", c.partialLength)
		c.state = HeadExtraLength
	case HeadExtraLength:
		if c.partialLength == 126 {
			if byteLength < 2 {
				return nil, false
			}
			c.length = c.data.ReadUnsignedShort()
		} else if c.partialLength == 127 {
			if byteLength < 8 {
				return nil, false
			}
			var tmp = c.data.ReadUnsignedInt()
			if tmp != 0 {
				return nil, false
			}
			c.length = c.data.ReadUnsignedInt()
		} else {
			c.length = c.partialLength
		}
		c.state = HeadExtraMask

		// util.Log("c.length=", c.length)
	case HeadExtraMask:
		if c.isMasked {
			if byteLength < 4 {
				return nil, false
			}
			c.mask = c.data.ReadBytes(4)
			// util.Log("c.mask=", c.mask)
		}
		c.state = Body
	case Body:
		// util.Log("len=", byteLength, c.length)
		if byteLength < c.length {
			return nil, false
		}
		data := c.data.ReadBytes(c.length)
		switch c.opcode {
		case Binary, Text, Continuation:
			log.Println("do c.opcode")
			if c.isFinal {
				if c.isMasked {
					data = applyMask(data, c.mask)
				}
			}
			log.Println(string(data))
			// 回复一句话
			// c.WriteWebSocketData([]byte("我是来自服务器的消息"), Text)
			c.onMessage(data)
		case Ping:
			c.WriteWebSocketData(data, Pong)
		case Pong:
			c.lastPong = time.Now().Unix()
		case Close:
			data = applyMask(data, c.mask)
			log.Println("interrupt:", string(data))
			c.Close()
		}
		c.state = Head
	default:
		return nil, false
	}

	byteLength = c.data.ByteLength()
	if byteLength > 0 {
		return ReadWebSocketData(c)
	}
	return nil, false
}

func applyMask(data []byte, mask []byte) []byte {
	var newdata = make([]byte, len(data))
	var makelen = len(mask)
	for i, v := range data {
		newdata[i] = v ^ mask[i%makelen]
	}
	return newdata
}

// 用户离线时触发
func (c *Client) onUserOut() {
	// 如果存在房间时，则需要退出房间
	if c.room != nil {
		log.Println("User logout")
		c.online = false
		c.room.ExitClient(c)
		CurrentServer.Logout(c)
	}
}

func (c *Client) SendError(errCode ClientErrorCode, data string) {
	c.SendToUserOp(&ClientMessage{
		Op: Error,
		Data: ClientError{
			Code: errCode,
			Msg:  data,
		},
	})
}

func (c *Client) SendNormalMessage(data string) {
	c.SendToUserOp(&ClientMessage{
		Op: Message,
		Data: data,
	})
}

// 获取用户数据
func (c *Client) GetUserData() any {
	data := map[string]any{}
	data["uid"] = c.uid
	data["name"] = c.name
	data["data"] = c.userData
	return data
}

// 同意WebSocket握手
func (c *Client) handshake(content string) {
	log.Println("handshake")
	s := strings.Split(content, "\n")
	var secWebSocketKey string
	for _, v := range s {
		keys := strings.Split(v, ":")
		switch keys[0] {
		case "Sec-WebSocket-Key":
			secWebSocketKey = keys[1]
			secWebSocketKey = strings.ReplaceAll(secWebSocketKey, " ", "")
			secWebSocketKey = strings.ReplaceAll(secWebSocketKey, "\n", "")
			secWebSocketKey = strings.ReplaceAll(secWebSocketKey, "\r", "")
		}
	}
	if secWebSocketKey != "" {
		// 同意握手时，返回信息
		base := secWebSocketKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
		t := sha1.New()
		io.WriteString(t, base)
		bs := t.Sum(nil)
		encoded := base64.StdEncoding.EncodeToString(bs)
		handdata := []string{
			"HTTP/1.1 101 Switching Protocols",
			"Upgrade: websocket",
			"Connection: Upgrade",
			"Sec-WebSocket-Accept: " + encoded,
		}
		data := strings.Join(handdata, "\n") + "\r\n\r\n"
		c.SendToUser([]byte(data))
		// 标记握手成功
		c.state = Head
		log.Println("handshake end")
	} else {
		c.Close()
	}
}

// 客户端逻辑处理
func clientHandle(c *Client) {
	defer c.Close()
	defer c.onUserOut()
	defer log.Println("Out user", c.RemoteAddr().String())
	for {
		// 每次客户端读取的数据长度
		var bytes [128]byte
		n, e := c.Read(bytes[:])
		if e != nil {
			break
		}
		if n == 0 {
			continue
		}
		// 缓存数据
		c.onData(bytes[:n])
	}
}

// 创建客户端对象
func CreateClient(c net.Conn) Client {
	log.Println("Join user", c.RemoteAddr().String())
	client := Client{
		Conn:      c,
		websocket: true,
		data:      util.Bytes{Data: []byte{}},
		state:     Handshake,
		userData:  map[string]any{},
		online:    true,
		uid: 			 -1,
	}

	go clientHandle(&client)
	return client
}
