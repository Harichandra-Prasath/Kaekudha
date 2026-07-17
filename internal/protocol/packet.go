package protocol

import "encoding/binary"

type Event uint8

const (
	CREATE_SESSION Event = 1
	JOIN_SESSION   Event = 2
	SEND           Event = 3
)

const HEADER_SIZE = 13

type ID [8]byte

func (id ID) String() string {
	return string(id[:])
}

type VoicePacket struct {
	Buf []byte
}

func NewVoicePocket(size int) *VoicePacket {
	return &VoicePacket{
		Buf: make([]byte, size),
	}
}

func (v *VoicePacket) SetHeader(event Event, id ID) {
	copy(v.Buf[:8], id[:])
	v.Buf[8] = byte(event)
}

func (v *VoicePacket) SetPayloadLength(len uint32) {
	binary.BigEndian.PutUint32(v.Buf[9:13], len)
}
