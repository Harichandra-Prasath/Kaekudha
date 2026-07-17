package session

import (
	"net"

	"github.com/Harichandra-Prasath/Kaekudha/internal/protocol"
)

type Session struct {
	sessionId protocol.ID
	hostId    protocol.ID
	clients   map[protocol.ID]*User
}

func CreateNewSession(sessionId protocol.ID, hostId protocol.ID, hostAddr *net.UDPAddr) (*Session, *User) {
	s := &Session{sessionId: sessionId, hostId: hostId}
	host := &User{id: hostId, session: s, addr: hostAddr}
	c := map[protocol.ID]*User{
		hostId: host,
	}
	s.clients = c
	return s, host
}

func (S *Session) AddUser(id protocol.ID, addr *net.UDPAddr) *User {
	user := &User{id: id, addr: addr, session: S}
	S.clients[id] = user
	return user
}

func (S *Session) GetUsers() map[protocol.ID]*User {
	return S.clients
}
