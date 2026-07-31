package audio

import (
	"fmt"

	"github.com/gen2brain/malgo"
)

type Player struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device
}

func NewPlayer(inChan <-chan *[]byte) (*Player, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {})
	if err != nil {
		return nil, fmt.Errorf("initialsing new context: %v", err)
	}

	config := malgo.DefaultDeviceConfig(malgo.Playback)
	config.Playback.Format = malgo.FormatS16
	config.Playback.Channels = CHANNELS
	config.SampleRate = SAMPLE_RATE
	config.PeriodSizeInMilliseconds = PERIOD_MS

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: func(pOutputSample, pInputSamples []byte, framecount uint32) {
			select {
			case pcm := <-inChan:
				copy(pOutputSample, *pcm)
				PcmPool.Put(pcm)
			default:
				clear(pOutputSample)
			}
		},
	}

	device, err := malgo.InitDevice(ctx.Context, config, deviceCallbacks)
	if err != nil {
		return nil, fmt.Errorf("initialising new device: %v", err)
	}

	return &Player{ctx: ctx, device: device}, nil
}

func (P *Player) Start(errChan chan<- error) {
	err := P.device.Start()
	if err != nil {
		errChan <- fmt.Errorf("player device start: %v", err)
	}
}

func (P *Player) Stop() error {
	err := P.device.Stop()
	if err != nil {
		return fmt.Errorf("player device stop: %v", err)
	}
	return nil
}
