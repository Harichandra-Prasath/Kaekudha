package protocol

import "encoding/binary"

type Event uint8

const (
	CREATE_SESSION Event = 1
	JOIN_SESSION   Event = 2
	SEND           Event = 3
	RESP           Event = 4
	END            Event = 5
)

const HEADER_SIZE = 13

type ID [8]byte

func (id ID) String() string {
	return string(id[:])
}

type Packet struct {
	Buf []byte
}

func NewPacket(size int) *Packet {
	return &Packet{
		Buf: make([]byte, size),
	}
}

func (v *Packet) SetHeader(event Event, id ID) {
	copy(v.Buf[:8], id[:])
	v.Buf[8] = byte(event)
}

func (v *Packet) SetPayloadLength(len uint32) {
	binary.BigEndian.PutUint32(v.Buf[9:13], len)
}
