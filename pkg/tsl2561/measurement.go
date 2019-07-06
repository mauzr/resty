package tsl2561

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"

	"mauzr.eqrx.net/go/pkg/i2c"
)

// Measurement contains the compensated measurement of a TSL2561 and its timestamp
type Measurement struct {
	Illuminance float64
	Timestamp   int64
}

type channelInput struct {
	CH0 uint16
	CH1 uint16
}

func read(device i2c.Device) (input channelInput, err error) {
	defer func() {
		if lerr := device.Write([]byte{0x80, 0x00}).Execute(); err == nil {
			err = lerr
		}
	}()
	err = i2c.Execute([]i2c.Action{
		device.Write([]byte{0x80, 0x03}),
		device.Write([]byte{0x81, 0x02}),
	})
	if err != nil {
		return
	}

	time.Sleep(500000)
	var data [4]byte
	err = device.WriteRead([]byte{0x0c}, data[:]).Execute()
	if err != nil {
		return
	}
	buf := bytes.NewReader(data[:])
	err = binary.Read(buf, binary.LittleEndian, &input)
	return
}

func calculate(input channelInput) (m Measurement) {
	m.Timestamp = time.Now().Unix()

	ch0, ch1 := float64(input.CH0), float64(input.CH1)
	if ch0 == 0 || ch0 > 65000 || ch1 > 65000 {
		m.Illuminance = math.NaN()
	} else {
		ratio := ch1 / ch0
		if ratio >= 0 && ratio <= 0.50 {
			m.Illuminance = 0.0304*ch0 - math.Pow(0.062*ch0*ratio, 1.4)
		} else if ratio <= 0.61 {
			m.Illuminance = 0.0224*ch0 - 0.031*ch1
		} else if ratio <= 0.80 {
			m.Illuminance = 0.0128*ch0 - 0.0153*ch1
		} else if ratio <= 1.30 {
			m.Illuminance = 0.00146*ch0 - 0.00112*ch1
		}
		m.Illuminance *= 16
	}

	return
}

// AttachAndMeasure connects to a TSL2561 behind an I2C bus, receive the
// reading and converts it into a measurment
func AttachAndMeasure(path string, addr i2c.DeviceAddress) (m Measurement, err error) {
	var device i2c.Device
	device, err = i2c.AttachDevice(path, addr)
	defer func() {
		if lerr := device.Close(); err == nil {
			err = lerr
		}
	}()
	if err == nil {
		var input channelInput
		input, err = read(device)
		if err == nil {
			m = calculate(input)
		}
	}
	return
}
