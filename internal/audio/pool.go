package audio

import "sync"

var PcmPool = sync.Pool{
	New: func() any {
		b := make([]byte, FRAMECOUNT*2)
		return &b
	},
}
