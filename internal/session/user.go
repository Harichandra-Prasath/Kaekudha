package session

import (
	"net"

	"github.com/Harichandra-Prasath/Kaekudha/internal/protocol"
)

type User struct {
	id      protocol.ID
	addr    *net.UDPAddr
	session *Session
}

func (U *User) GetSession() *Session {
	return U.session
}

func (U *User) GetAddr() *net.UDPAddr {
	return U.addr
}
