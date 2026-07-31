package session

import (
	"net"

	"github.com/Harichandra-Prasath/Kaekudha/internal/protocol"
)

type Session struct {
	sessionId protocol.ID
	clients   map[protocol.ID]*User
}

func CreateNewSession(sessionId protocol.ID, userId protocol.ID, hostAddr *net.UDPAddr) (*Session, *User) {
	s := &Session{sessionId: sessionId}
	user := &User{id: userId, session: s, addr: hostAddr}
	c := map[protocol.ID]*User{
		userId: user,
	}
	s.clients = c
	return s, user
}

func (S *Session) AddUser(id protocol.ID, addr *net.UDPAddr) *User {
	user := &User{id: id, addr: addr, session: S}
	S.clients[id] = user
	return user
}

func (S *Session) RemoveUser(id protocol.ID) {
	delete(S.clients, id)
}

func (S *Session) GetUsers() map[protocol.ID]*User {
	return S.clients
}

func (S *Session) GetId() protocol.ID {
	return S.sessionId
}
