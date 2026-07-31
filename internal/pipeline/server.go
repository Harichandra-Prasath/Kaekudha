package pipeline

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/Harichandra-Prasath/Kaekudha/internal/protocol"
	"github.com/Harichandra-Prasath/Kaekudha/internal/session"
)

const (
	MAX_MESSAGE_BYTES = 100
	SERVER_ID         = "SERVER"
)

type Server struct {
	conn         *net.UDPConn
	serverPacket *protocol.Packet
	serverID     protocol.ID
	clients      map[protocol.ID]*session.User
	sessions     map[protocol.ID]*session.Session
}

func NewServer(host string) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, fmt.Errorf("resolving udp host: %v", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("listening on addr: %s", err)
	}
	var id protocol.ID
	copy(id[:], []byte(SERVER_ID))
	return &Server{clients: map[protocol.ID]*session.User{}, sessions: map[protocol.ID]*session.Session{}, conn: conn, serverPacket: protocol.NewPacket(MAX_MESSAGE_BYTES + protocol.HEADER_SIZE), serverID: id}, nil
}

func (S *Server) Start() {
	buffer := make([]byte, protocol.HEADER_SIZE+MAX_AUDIO_BYTES)
	for {
		n, clientAddr, err := S.conn.ReadFromUDP(buffer)
		if err != nil {
			panic(fmt.Sprintf("error reading from conn: %v", err))
		}
		event := buffer[8]
		switch protocol.Event(event) {
		case protocol.CREATE_SESSION:
			slog.Info("CREATE_SESSION Recived", "clientAddr", clientAddr.String())
			S.handleSessionCreation(buffer[:n], clientAddr)
		case protocol.JOIN_SESSION:
			slog.Info("JOIN_SESSION Recived", "clientAddr", clientAddr.String())
			S.handleSessionJoin(buffer[:n], clientAddr)
		case protocol.SEND:
			S.handleAudio(buffer[:n])
		case protocol.END:
			S.handleEnd(buffer[:n])
		}

	}
}

func (S *Server) writeResponse(addr *net.UDPAddr, msg string) {
	S.serverPacket.SetHeader(protocol.RESP, S.serverID)
	S.serverPacket.SetPayloadLength(uint32(len(msg)))
	copy(S.serverPacket.Buf[protocol.HEADER_SIZE:], []byte(msg))
	S.conn.WriteToUDP(S.serverPacket.Buf[:], addr)
}

func (S *Server) handleSessionCreation(buf []byte, addr *net.UDPAddr) {
	userId := buf[:8]
	sessionId := buf[protocol.HEADER_SIZE : protocol.HEADER_SIZE+8]
	s, user := session.CreateNewSession(protocol.ID(sessionId), protocol.ID(userId), addr)

	if _, ok := S.sessions[protocol.ID(sessionId)]; ok {
		slog.Warn("Session Exists with the name already")
		S.writeResponse(addr, "ERR_DUPLICATE_SESSION")
	}

	S.sessions[protocol.ID(sessionId)] = s
	S.clients[protocol.ID(userId)] = user

	slog.Info("New Session Created", "session", string(sessionId))
	S.writeResponse(addr, "")
}

func (S *Server) handleSessionJoin(buf []byte, addr *net.UDPAddr) {
	clientId := buf[:8]
	sessionId := buf[protocol.HEADER_SIZE : protocol.HEADER_SIZE+8]
	session, ok := S.sessions[protocol.ID(sessionId)]
	if !ok {
		slog.Warn("Trying to Join Non-Exisiting Session", "client", string(clientId), "clientAddr", addr.String())
		S.writeResponse(addr, "ERR_INVALID_SESSION")
		return
	}
	_, ok = S.clients[protocol.ID(clientId)]
	if ok {
		slog.Warn("User already exist with the same Id", "client", string(clientId))
		S.writeResponse(addr, "ERR_DUPLICATE_USER")
		return
	}

	user := session.AddUser(protocol.ID(clientId), addr)
	S.clients[protocol.ID(clientId)] = user

	slog.Info("User Added to Session", "session", string(sessionId), "client", string(clientId))
	S.writeResponse(addr, "")
}

func (S *Server) handleAudio(buf []byte) {
	clientId := buf[:8]
	user, ok := S.clients[protocol.ID(clientId)]
	if !ok {
		return
	}
	session := user.GetSession()
	sessionUsers := session.GetUsers()

	for id, sessionUser := range sessionUsers {
		if id != protocol.ID(clientId) {
			_, err := S.conn.WriteToUDP(buf, sessionUser.GetAddr())
			if err != nil {
				slog.Error("Error writing to client", "client", id.String())
			}
		}
	}
}

func (S *Server) handleEnd(buf []byte) {
	clientId := buf[:8]
	user, ok := S.clients[protocol.ID(clientId)]
	if !ok {
		slog.Warn("Client not Present in the server", "client", string(clientId))
		return
	}
	session := user.GetSession()
	session.RemoveUser(protocol.ID(clientId))
	delete(S.clients, user.GetId())
	slog.Info("User Removed from Session", "Client", user.GetId().String(), "Session", session.GetId().String())
	if len(session.GetUsers()) == 0 {
		slog.Info("No Clients Present on Session. Removing the Session...", "Session", session.GetId().String())
		delete(S.sessions, session.GetId())
	}
	S.writeResponse(user.GetAddr(), "")
}
