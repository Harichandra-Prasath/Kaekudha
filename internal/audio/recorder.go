package audio

import (
	"fmt"

	"github.com/gen2brain/malgo"
)

type Recorder struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device
}

const (
	CHANNELS    = 1
	SAMPLE_RATE = 16000
	PERIOD_MS   = 20
	FRAMECOUNT  = 320
)

func NewRecorder(outChan chan<- []byte) (*Recorder, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {})
	if err != nil {
		return nil, fmt.Errorf("initialsing new context: %v", err)
	}

	config := malgo.DefaultDeviceConfig(malgo.Capture)
	config.Capture.Format = malgo.FormatS16
	config.Capture.Channels = CHANNELS
	config.SampleRate = SAMPLE_RATE
	config.PeriodSizeInMilliseconds = PERIOD_MS

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: func(pOutputSample, pInputSamples []byte, framecount uint32) {
			outChan <- pInputSamples
		},
	}

	device, err := malgo.InitDevice(ctx.Context, config, deviceCallbacks)
	if err != nil {
		return nil, fmt.Errorf("initialising new device: %v", err)
	}

	return &Recorder{ctx: ctx, device: device}, nil
}

func (R *Recorder) Start(errChan chan<- error) {
	err := R.device.Start()
	if err != nil {
		errChan <- fmt.Errorf("recording device start: %v", err)
	}
}
