package net

import (
	"encoding/json"
	"fmt"
	"log"
	"chatroom/util"
)

// 消息处理
func (c *Client) onMessage(data []byte) {
	// 解析API操作
	message := ClientMessage{}
	var err error
	// 如果是二进制数据，则需要解析处理，第一位是op操作符，剩余的是内容
	if c.frameIsBinary {
		op := ClientAction(data[0])
		content := data[1:]
		message.Op = op
		err = json.Unmarshal(content, &message.Data)
		log.Println("Binary data processing ", "op=", op, string(content))
	} else {
		err = json.Unmarshal(data, &message)
	}
	if err == nil {
		fmt.Println("Processing commands ", message)
		//将操作记录到日志,保证可追溯
		log.Println("uid:",c.uid,",op:",message.Op,"cmd:",message.Data)
		if c.uid == -1 {
			switch message.Op {
			case Login:
				if c.uid == -1 {
					loginData := message.Data.(map[string]any)
					userName := loginData["UserName"]
					if userName == nil {
						c.SendError(LOGIN_ERROR, "Need to provide username")
						return
					}
					name := userName.(string)
					// 只需要用户名即可登录
					if (len(c.name) > 0 && c.name != name) {
						c.SendError(LOGIN_ERROR, "User Changed.Please logout first.")
						return 
					}
					c.name = name
					passwd := ""
					userData := CurrentServer.usersSQL.login(c, passwd)
					log.Println("Login succeeded:", userData)
					c.SendToUserOp(&ClientMessage{
						Op: Login,
						Data: map[string]any{
							"Code": 0,
							"Uid": userData.uid,
							"Name": userName,
							"Msg": "Login Success",
						},
					})
				} else {
					c.SendError(LOGIN_ERROR, "Logged in")
				}
			default:
				c.SendError(OP_ERROR, "Login Invalid operation instruction: "+fmt.Sprint(message.Op))
			}
			return
		}
		fmt.Println("Normal Message!", message.Op)
		switch message.Op {
		case Message:
			// 接收到消息
			fmt.Println("RECV COMMON MSG: ", message.Data)
			
		case CreateRoom:
			if (c.room != nil) {
				c.SendError(ENTER_ROOM_ERROR, "Each user can only join to one room.Please exit room first.")
				return 
			}
			// 创建一个房间
			room := CurrentServer.CreateRoom(c)
			log.Println("Start creating rooms ", room)
			if room != nil {
				c.room = room
				// 创建成功
				c.SendToUserOp(&ClientMessage{
					Op: CreateRoom,
					Data: map[string]any{
						"Code" : 0,
						"RoomId" : room.id,
						"Msg" : "Create Room Success!",
					},
				})
			} else {
				// 创建失败
				c.SendError(CREATE_ROOM_ERROR, "Room already exists, unable to create")
			}
		case JoinRoom:
			rcvData := message.Data.(map[string]any)
			if (c.room != nil) {
				c.SendToUserOp(&ClientMessage{
					Op: JoinRoom,
					Data: map[string]any{
						"Code" : -1,
						"Msg" : "Each user can only join to one room.Please exit room first.",
					},
				})
				return 
			}
			//读取string类型，实际上通常意义上的ID并不一定必须是数字
			roomIdany := rcvData["RoomId"]
			if roomIdany == nil {
				c.SendToUserOp(&ClientMessage{
					Op: JoinRoom,
					Data: map[string]any{
						"Code" : -1,
						"Msg" : "Need to provide roomId.",
					},
				})
				return
			}
			//fmt.Println(roomIdany)
			roomId := util.Str2int(roomIdany.(string))
			if (roomId <= 0) {
				c.SendToUserOp(&ClientMessage{
					Op: JoinRoom,
					Data: map[string]any{
						"Code" : -1,
						"Msg" : "Room not found.",
					},
				})
				return
			}
			ret := CurrentServer.JoinRoom(c, roomId)
			if (ret) {
				c.SendToUserOp(&ClientMessage{
					Op:   JoinRoom,
					Data: map[string]any{
						"Code" : 0,
						"Msg" : "Join room " + string(roomId) + " success.",
					},
				})
			} else {
				c.SendToUserOp(&ClientMessage{
					Op: JoinRoom,
					Data: map[string]any{
						"Code" : -1,
						"Msg" : "Room not found.",
					},
				})
			}
		case ExitRoom:
			if ( c.room == nil ) {
				msg := "Exit Room Success"
				c.SendToUserOp(&ClientMessage{
					Op:   ExitRoom,
					Data: map[string]any{
						"Code" : 0,
						"Msg" : msg,
					},
				})
				return 
			} else {
				lastRoomId := c.room.id
				c.room.ExitClient(c)
				msg := "Exit Room " + string(lastRoomId) + " Success"
				c.room = nil
				c.SendToUserOp(&ClientMessage{
					Op:   ExitRoom,
					Data: map[string]any{
						"Code" : 0,
						"Msg" : msg,
					},
				})
			}
		case GetRoomInfo:
			// 获取房间信息
			if c.room != nil {
				c.SendToUserOp(&ClientMessage{
					Op:   GetRoomInfo,
					Data: c.room.GetRoomData(),
				})
			} else {
				c.SendError(GET_ROOM_ERROR, "No room information exists")
			}
		case StartFrameSync:
			// 开始帧同步
			if c.room != nil {
				c.room.StartFrameSync()
				c.SendToUserOp(&ClientMessage{
					Op: StartFrameSync,
				})
			} else {
				c.SendError(START_FRAME_SYNC_ERROR, "Unable to start frame synchronization because the room does not exist")
			}
		case StopFrameSync:
			// 开始停止帧同步
			if c.room != nil {
				c.room.StopFrameSync()
				c.SendToUserOp(&ClientMessage{
					Op: StopFrameSync,
				})
			} else {
				c.SendError(STOP_FRAME_SYNC_ERROR, "Unable to stop frame synchronization because the room does not exist")
			}
		case UploadFrame:
			if c.room != nil && c.room.frameSync {
				// 缓存到用户数据中
				mapdata, err := message.Data.(map[string]any)
				fdata := FrameData{
					Time: int64(mapdata["Time"].(float64)),
					Data: mapdata["Data"],
				}
				log.Println("Frame synchronization data:", fdata, err)
				if err {
					// 验证是否操作数据是否已无效
					if !isInvalidData(&fdata) {
						c.frames.Push(fdata)
						c.SendToUserOp(&ClientMessage{
							Op: UploadFrame,
						})
					} else {
						c.SendError(UPLOAD_FRAME_ERROR, "Upload frame data timestamp error")
					}
				}
			} else {
				c.SendError(UPLOAD_FRAME_ERROR, "Error in uploading frame synchronization data")
			}
		case RoomMessage:
			// 转发房间信息
			if c.room != nil {
				c.room.SendToAllUserOp(&message, c)
				c.SendToUserOp(&ClientMessage{
					Op: RoomMessage,
				})
			} else {
				c.SendError(SEND_ROOM_ERROR, "Room does not exist")
			}
		case GetRoomList:
			lst := CurrentServer.GetRoomList()
			//jsonData,_ := json.Marshal(lst.List)
			c.SendToUserOp(&ClientMessage{
				Op: GetRoomList,
				Data: map[string]any{
					"Code": 0,
					"Data": lst.List,
				},
			})
		default:
			c.SendError(OP_ERROR, "Cmd Invalid operation instruction:"+fmt.Sprint(message.Op))
		}
	} else {
		fmt.Println("Failed to process command ", string(data), err.Error())
	}
}
