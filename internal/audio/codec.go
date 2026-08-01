package audio

import (
	"fmt"
	"math"

	"gopkg.in/hraban/opus.v2"
)

type Codec struct {
	encoder *opus.Encoder
	decoder *opus.Decoder
}

const GAIN = 5

func NewCodec() (*Codec, error) {
	enc, err := opus.NewEncoder(SAMPLE_RATE, CHANNELS, opus.AppVoIP)
	if err != nil {
		return nil, fmt.Errorf("creating new encoder: %v", err)
	}
	dec, err := opus.NewDecoder(SAMPLE_RATE, CHANNELS)
	if err != nil {
		return nil, fmt.Errorf("creating new decoder: %v", err)
	}
	return &Codec{encoder: enc, decoder: dec}, nil
}

func (C *Codec) Encode(pcm []int16, buf []byte) (int, error) {
	n, err := C.encoder.Encode(pcm, buf)
	if err != nil {
		return 0, fmt.Errorf("encoding pcm: %v", err)
	}
	return n, nil
}

func (C *Codec) Decode(buf []byte, pcm []int16) (int, error) {
	n, err := C.decoder.Decode(buf, pcm)
	if err != nil {
		return 0, fmt.Errorf("decoding raw bytes: %v", err)
	}
	return n, nil
}

func (C *Codec) Amplify(pcm []int16) {
	for i := range pcm {
		v := int32(pcm[i]) * GAIN
		if v > math.MaxInt16 {
			v = math.MaxInt16
		}
		if v < math.MinInt16 {
			v = math.MinInt16
		}
		pcm[i] = int16(v)
	}
}
