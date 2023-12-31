package net

import (
	"time"
	"sync"
	// "fmt"
	// "encoding/json"
	"chatroom/util"
)

type Room struct {
	id         int
	master     int       // 房主的ID
	rwRoomMtx  *sync.RWMutex  // 房间操作锁
	users      util.Array    // 房间用户
	frameSync  bool          // 是否开启帧同步
	interval   time.Duration // 帧同步的间隔
	lock       bool          // 房间是否锁定（如果游戏已经开始，则会锁定房间，直到游戏结束，如果用户离线，不会立即退出房间，需要通过`ExitRoom`才能退出房间）
	frameDatas []any         // 房间帧数据
	cacheId    int           // 房间已缓存的时间轴Id
}

// 是否为无效房间
func (r *Room) isInvalidRoom() bool {
	r.rwRoomMtx.RLock()
	defer r.rwRoomMtx.RUnlock()
	len := r.users.Length()
	hasOnline := false
	for _, v := range r.users.List {
		if v.(*Client).online {
			hasOnline = true
			break;
		}
	}
	return len == 0 || !hasOnline
}

// 房间的帧同步实现
func onRoomFrame(r *Room) {
	r.rwRoomMtx.Lock()
	defer r.rwRoomMtx.Unlock()
	for {
		if !r.frameSync || r.isInvalidRoom() {
			// 帧同步停止，或者房间已经不存在用户时
			util.Log("Room Stop Frame Synchronization")
			break
		}
		frameData := map[int][]any{}
		// 收集房间的所有用户操作
		for _, v := range r.users.List {
			c := v.(*Client)
			a := frameData[c.uid]
			for _, v2 := range c.frames.List {
				f := v2.(FrameData)
				a = append(a, f.Data)
			}
			if a != nil {
				frameData[c.uid] = a
			}
			c.frames.List = []any{}
		}

		// 缓存数据
		r.cacheId++
		r.frameDatas = append(r.frameDatas, frameData)

		// 发送帧数据到客户端
		for _, v := range r.users.List {
			c := v.(*Client)
			c.SendToUserOp(&ClientMessage{
				Op: FData,
				Data: map[string]any{
					"t": r.cacheId,
					"d": frameData,
				},
			})
		}
		// 帧同步发送间隔
		time.Sleep(r.interval)
	}
}

// 启动帧同步
func (r *Room) StartFrameSync() {
	r.rwRoomMtx.Lock()
	defer r.rwRoomMtx.Unlock()
	if r.frameSync {
		return
	}
	r.frameSync = true
	r.lock = true
	go onRoomFrame(r)
}

// 停止帧同步
func (r *Room) StopFrameSync() {
	r.rwRoomMtx.Lock()
	defer r.rwRoomMtx.Unlock()
	r.frameSync = false
	r.lock = false
}

// 给房间的所有用户发送消息
func (r *Room) SendToAllUser(data []byte) {
	r.rwRoomMtx.Lock()
	defer r.rwRoomMtx.Unlock()
	for _, v := range r.users.List {
		v.(*Client).SendToUser(data)
	}
}

// 给房间的所有用户发送消息
func (r *Room) SendToAllUserOp(data *ClientMessage, igoneClient *Client) {
	r.rwRoomMtx.RLock()
	defer r.rwRoomMtx.RUnlock()
	for _, v := range r.users.List {
		if v != igoneClient {
			v.(*Client).SendToUserOp(data)
		}
	}
}

// 加入用户
func (r *Room) JoinClient(client *Client) {
	if client.room == nil {
		r.rwRoomMtx.Lock()
		r.users.Push(client)
		r.rwRoomMtx.Unlock()
		client.room = r
		roomInfo := r.GetRoomData()
		client.SendToUserOp(&ClientMessage{
			Op:   GetRoomInfo,
			Data: roomInfo,
		})
		// 其他用户通知房间更新
		r.onRoomChanged()
	}
}

func (r *Room) onRoomChanged() {
	r.SendToAllUserOp(&ClientMessage{
		Op: ChangedRoom,
	}, nil)
}

// 用户退出
func (r *Room) ExitClient(client *Client) {
	r.rwRoomMtx.Lock()
	defer r.rwRoomMtx.Unlock()
	if client.room.id == r.id {
		r.users.Remove(client)
		client.room = nil
		if r.users.Length() == 0 {
			// 房间已经不存在用户了，则删除当前房间
			CurrentServer.DeleteRoom(r.id)
		} else {
			// 如果用户仍然存在时，如果是房主掉线，则需要更换房主。不管房主是否更换，都需要通知客户端用户重新更新房间信息
			if r.master == client.uid {
				r.master = r.users.List[0].(*Client).uid
			}
			// 通知更新房间信息
			r.rwRoomMtx.Unlock()
			r.onRoomChanged()
			r.rwRoomMtx.Lock()
		}
	}
}

// 获取房间信息
func (r *Room) GetRoomData() any {
	r.rwRoomMtx.RLock()
	defer r.rwRoomMtx.RUnlock()
	data := map[string]any{}
	data["id"] = r.id
	data["master"] = r.master
	users := util.Array{}
	for _, v := range r.users.List {
		users.Push(v.(*Client).GetUserData())
	}
	data["users"] = users
	return data
}
