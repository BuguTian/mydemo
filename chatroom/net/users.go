package net

import (
	"fmt"
	"chatroom/util"
)
type RegisterUserData struct {
	uid      int
	username string
	passwd   string
	client   *Client
}

// 用户数据数据库
type UserDataSQL struct {
	create_uid_index int
	users            map[int]*RegisterUserData
}

// 登陆角色
func (u *UserDataSQL) login(c *Client, passwd string) *RegisterUserData {
	if ( c == nil ) {
		return nil
	}
	if u.users == nil {
		u.users = map[int]*RegisterUserData{}
	}
	uid := c.uid
	user, ok := u.users[uid]
	fmt.Println("uid=",uid)
	if ok {
		// 判断用户名密码是否正确
    if ( user.passwd != passwd ) {
			//如果密码不对就返回空
			return nil
		}
		// 用户曾经登陆过，需要检测用户是否在线，否则会发生挤出的事件
		if user.client != nil {
			if user.client.room != nil {
				// 如果原本就存在房间时，则需要把用户返回到房间中
				r := user.client.room
				r.ExitClient(user.client)
				r.JoinClient(c)
				util.Log("The user is still in the room, join the room")
			}
			util.Log("Offline handling")
			user.client.SendError(LOGIN_OUT_ERROR, "The user has logged in elsewhere")
			user.client.Close()
		}
	} else {
		// 新用户
		u.create_uid_index++
		uid = u.create_uid_index
		u.users[uid] = &RegisterUserData{
			uid:      u.create_uid_index,
		}
		user = u.users[uid]
	}
	user.client = c
	c.uid = user.uid
	user.username = c.name
	user.passwd = c.passwd
	return user
}
