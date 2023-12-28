package net

import (
	"fmt"
	"net"
	"time"
	"sync"
	"chatroom/util"
)

var CurrentServer *Server

type Server struct {
	users      util.Array  // 用户列表
	rwUsersMtx  *sync.RWMutex //房间表操作锁
	rooms      map[int]*Room  // 房间列表
	rwRoomsMtx  *sync.RWMutex //房间表操作锁
	create_uid int         // 房间的创建ID
	usersSQL   UserDataSQL // 用户数据库，管理已注册、登陆的用户基础信息
}

// 开始侦听服务器
func (s *Server) Listen(port int) {
	CurrentServer = s
	s.rooms = make(map[int]*Room)
	s.rwUsersMtx = new(sync.RWMutex)
	s.rwRoomsMtx = new(sync.RWMutex)
	fmt.Println("Server start:0.0.0.0:" + fmt.Sprint(port))
	n, e := net.Listen("tcp", ":" + fmt.Sprint(port))
	if e != nil {
		fmt.Println(e.Error())
	}
	for {
		c, e := n.Accept()
		if e == nil {
			// 将用户写入到用户列表中
			s.users.Push(CreateClient(c))
		}
	}
}

func (s *Server) GetRoomInstanceById( rid int) *Room {
	s.rwRoomsMtx.RLock()
	defer s.rwRoomsMtx.RUnlock()
	val,ok := s.rooms[rid]
	if (ok) {
		return val
	}
	return nil
}

// 创建房间
func (s *Server) CreateRoom(c *Client) *Room {
	if c.room != nil {
		return nil
	}
	s.rwRoomsMtx.Lock()
	defer s.rwRoomsMtx.Unlock()
	s.create_uid++
	interval := 1. / 30.
	interval = float64(time.Second) * interval
	room := Room{
		id:       s.create_uid,
		master:   c.uid,
		interval: time.Duration(interval),
		rwRoomMtx: new(sync.RWMutex),
	}
	s.rooms[room.id] = &room
	room.JoinClient(c)
	
	return &room
}

func (s *Server) DeleteRoom(roomId int) {
	s.rwRoomsMtx.Lock()
	defer s.rwRoomsMtx.Unlock()
	if (roomId != -1) {
		delete(s.rooms, roomId)
	}
}

// 获取 Room 列表
func (s *Server) GetRoomList() util.Array {
	var ret util.Array
	s.rwRoomsMtx.RLock()
	defer s.rwRoomsMtx.RUnlock()
	for _,val := range s.rooms {
		ret.Push(val.id);
	}
	return ret
}

// 加入房间
func (s *Server) JoinRoom(c *Client, roomId int) bool {
	s.rwRoomsMtx.Lock()
	defer s.rwRoomsMtx.Unlock()
	// 如果用户已经在房间中，则无法继续加入
	if c.room != nil {
		return false
	}

	val, ok := s.rooms[roomId]
	if ( ok ) {
		//添加到 room 的 users 中
		s.rwRoomsMtx.Unlock()
		val.JoinClient(c)
		s.rwRoomsMtx.Lock()
		return true
	}
	return false
}

// 退出房间
func (s *Server) ExitRoom(c *Client) {
	if c.room != nil {
		room := s.GetRoomInstanceById(c.room.id)
		if (room != nil) {
			room.ExitClient(c)
		}
	}
}

// 退出登录
func (s *Server) Logout(c *Client) {
	s.rwUsersMtx.Lock()
	// 从服务器列表中删除
	s.users.Remove(c)
	s.rwUsersMtx.Unlock()
}
