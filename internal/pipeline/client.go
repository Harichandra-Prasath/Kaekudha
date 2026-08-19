package pipeline

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"
	"unsafe"

	"github.com/Harichandra-Prasath/Kaekudha/internal/audio"
	"github.com/Harichandra-Prasath/Kaekudha/internal/protocol"
	"github.com/Harichandra-Prasath/Kaekudha/internal/storage"
)

const (
	BUFFER_SIZE     = 50
	MAX_AUDIO_BYTES = 200
)

type Client struct {
	id       protocol.ID
	conn     *net.UDPConn
	handler  *handler
	recorder *audio.Recorder
	player   *audio.Player
	codec    *audio.Codec
	timer    *time.Timer
	errChan  chan error
}

type handler struct {
	outPacket *protocol.Packet
	outChan   chan []byte

	inPacket *protocol.Packet
	inChan   *storage.RingBuffer[*[]byte]

	respChan chan *[]byte
}

func GetNewClient(name string, host string) (*Client, error) {
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, fmt.Errorf("resolving udp addr: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("creating connection: %v", err)
	}
	var id protocol.ID
	copy(id[:], name)

	h := handler{
		outPacket: protocol.NewPacket(MAX_AUDIO_BYTES + protocol.HEADER_SIZE),
		outChan:   make(chan []byte, BUFFER_SIZE),
		inPacket:  protocol.NewPacket(MAX_AUDIO_BYTES + protocol.HEADER_SIZE),
		inChan:    storage.New[*[]byte](BUFFER_SIZE),
		respChan:  make(chan *[]byte, 64),
	}

	rec, err := audio.NewRecorder(h.outChan)
	if err != nil {
		return nil, fmt.Errorf("creating new recorder: %v", err)
	}

	player, err := audio.NewPlayer(h.inChan)
	if err != nil {
		return nil, fmt.Errorf("creating new player: %v", err)
	}

	codec, err := audio.NewCodec()
	if err != nil {
		return nil, fmt.Errorf("creating new coded: %v", err)
	}

	return &Client{
		id:       id,
		conn:     conn,
		handler:  &h,
		recorder: rec,
		player:   player,
		codec:    codec,
		errChan:  make(chan error),
		timer:    time.NewTimer(5 * time.Second),
	}, nil
}

func (C *Client) Start(create string, join string) error {
	go C.checkError()

	slog.Info("Started reading from server...")
	go C.startInflow(C.errChan)

	var id protocol.ID
	var event protocol.Event
	if create != "" {
		event = protocol.CREATE_SESSION
		copy(id[:], create)
	} else {
		event = protocol.JOIN_SESSION
		copy(id[:], join)
	}

	err := C.sendInitEvent(event, id)
	if err != nil {
		return fmt.Errorf("sending init event: %v", err)
	}

	slog.Info("Starting the recorder...")
	go C.recorder.Start(C.errChan)
	slog.Info("Starting the player...")
	go C.player.Start(C.errChan)

	go C.startOutFlow(C.errChan)
	return nil
}

func (C *Client) cleanup() {
	err := C.recorder.Stop()
	if err != nil {
		slog.Error("Error stopping the recorder device", "err", err)
	}
	slog.Info("Recorder Stopped")
	err = C.player.Stop()
	if err != nil {
		slog.Error("Error stopping the player device", "err", err)
	}
	slog.Info("Player Stopped")

	slog.Info("Cleanup Completed")
}

func (C *Client) Stop() {
	C.cleanup()

	packet := C.handler.outPacket
	packet.SetHeader(protocol.END, C.id)
	C.conn.Write(packet.Buf[:protocol.HEADER_SIZE])
	err := C.getServerResp()
	if err != nil {
		slog.Error("Error getting Server Response", "err", err)
	}
}

func (C *Client) checkError() {
	for err := range C.errChan {
		panic(fmt.Sprintf("fatal error in client pipeline: %v", err))
	}
}

func (C *Client) sendInitEvent(event protocol.Event, sessionId protocol.ID) error {
	packet := C.handler.outPacket
	packet.SetHeader(event, C.id)
	n := len(sessionId)
	copy(packet.Buf[protocol.HEADER_SIZE:], sessionId[:])
	packet.SetPayloadLength(uint32(n))
	_, err := C.conn.Write(packet.Buf[:protocol.HEADER_SIZE+n])
	if err != nil {
		return fmt.Errorf("writing init event: %v", err)
	}
	err = C.getServerResp()
	if err != nil {
		return fmt.Errorf("getting server response: %v", err)
	}
	return nil
}

func (C *Client) getServerResp() error {
	// Safeguard
	if !C.timer.Stop() {
		select {
		case <-C.timer.C:
		default:
		}
	}

	C.timer.Reset(5 * time.Second)
	for {
		select {
		case msg := <-C.handler.respChan:
			if len(*msg) == 0 {
				return nil
			}
			return fmt.Errorf("error response from server: %s", string(*msg))
		case <-C.timer.C:
			return fmt.Errorf("timed-out waiting for server")
		}
	}
}

func (C *Client) startOutFlow(errChan chan<- error) {
	for payload := range C.handler.outChan {
		packet := C.handler.outPacket
		packet.SetHeader(protocol.SEND, C.id)
		pcm := unsafe.Slice((*int16)(unsafe.Pointer(&payload[0])), len(payload)/2)
		n, err := C.codec.Encode(pcm, packet.Buf[protocol.HEADER_SIZE:])
		if err != nil {
			errChan <- fmt.Errorf("encoding the samples: %v", err)
		}
		packet.SetPayloadLength(uint32(n))
		_, err = C.conn.Write(packet.Buf[:protocol.HEADER_SIZE+n])
		if err != nil {
			errChan <- fmt.Errorf("writing the data: %v", err)
		}
	}
}

func (C *Client) startInflow(errChan chan<- error) {
	packet := C.handler.inPacket
	samples := make([]int16, audio.FRAMECOUNT)
	for {
		_, err := C.conn.Read(packet.Buf)
		if err != nil {
			errChan <- fmt.Errorf("reading from server: %v", err)
		}

		event := packet.Buf[8]
		switch event {
		case byte(protocol.RESP):
			payloadLen := binary.BigEndian.Uint32(packet.Buf[9:13])
			// One time Allocation
			buff := make([]byte, payloadLen)
			copy(buff[:], packet.Buf[protocol.HEADER_SIZE:protocol.HEADER_SIZE+payloadLen])
			C.handler.respChan <- &buff

		case byte(protocol.SEND):
			payloadLen := binary.BigEndian.Uint32(packet.Buf[9:13])
			n, err := C.codec.Decode(packet.Buf[protocol.HEADER_SIZE:protocol.HEADER_SIZE+payloadLen], samples)
			if err != nil {
				errChan <- fmt.Errorf("decoding the raw data: %v", err)
			}

			buf := audio.PcmPool.Get().(*[]byte)
			buff := *buf

			C.codec.Amplify(samples)

			for i, s := range samples[:n] {
				binary.LittleEndian.PutUint16(buff[i*2:], uint16(s))
			}

			if ok := C.handler.inChan.Push(buf); !ok {
				slog.Info("Dropping packets. Try increasing the buffer size...")
				continue
			}
		case byte(protocol.END):
			slog.Info("END Event Recieved From Server. Exiting...")
			C.cleanup()
			os.Exit(0)
		}

	}
}
