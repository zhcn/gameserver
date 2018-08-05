package login

import (
	"fmt"
	"time"
)

type UserId uint32
type RoomId uint32

const (
	USER_IDLE = 0
	USER_PLAY
	USER_WATCH
	USER_WAITTOSTART
)

type LoginHandler struct {
	loginUserInfos map[UserId]UserInfo // userId -> UserInfo
	roomInfos      map[RoomId]RoomInfo
}

type UserInfo struct {
	userId    UserId
	status    uint16
	coins     uint32
	lastLogin time.Time
	roomId    uint32
}

type RoomInfo struct {
	rootId RoomId
	users  []UserInfo
}

func (lh *LoginHandler) login(userInfo UserInfo) {
	loginUserInfos[userInfo.userId] = userInfo
}

func (lh *LoginHandler) checkUserHeart() {

}

func (lh *LoginHandler) initLogin() {

}
