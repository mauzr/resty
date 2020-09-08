/*
Copyright 2019 Alexander Sowitzki.

GNU Affero General Public License version 3 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://opensource.org/licenses/AGPL-3.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package raspivid interfaces with the raspivid program to grab a camera feed.
package raspivid

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

// Configuration specifies parameters for raspivid.
type Configuration struct {
	On                       bool
	Width, Height, Framerate int
	Flip                     [2]bool
	Exposure                 string
}

// Data is a chunk of data read from the raspivid stdout.
type Data struct {
	Data []byte
	Err  error
}

// Request is a request to start a raspivid instance.
type Request struct {
	// Configuration for the raspivid program.
	Configuration Configuration
	// Response receives errors encountered by the raspivid manager. this channel must have a capacity greater 1 or the manager will panic.
	Response chan<- error
}

// ErrConfiguration means that an invalid configuration was passed.
var ErrConfiguration = errors.New("invalid configuration")

// ErrProcess means that raspivid broke for some reason.
var ErrProcess = errors.New("raspivid process failed")

func (c Configuration) arguments() ([]string, error) {
	if !c.On {
		return nil, fmt.Errorf("%w: not on", ErrConfiguration)
	}
	if c.Width <= 0 {
		return nil, fmt.Errorf("%w: invalid width", ErrConfiguration)
	}
	if c.Height <= 0 {
		return nil, fmt.Errorf("%w: invalid height", ErrConfiguration)
	}
	if c.Framerate <= 0 {
		return nil, fmt.Errorf("%w: invalid framerate", ErrConfiguration)
	}
	width := strconv.Itoa(c.Width)
	height := strconv.Itoa(c.Height)
	framerate := strconv.Itoa(c.Framerate)

	exvalid := false
	for _, ex := range []string{"off", "auto", "night", "nightpreview", "backlight", "spotlight", "sports", "snow", "beach", "verylong", "fixedfps", "antishake", "fireworks"} {
		if ex == c.Exposure {
			exvalid = true

			break
		}
	}
	if !exvalid {
		return nil, fmt.Errorf("%w: illegal exposure mode", ErrConfiguration)
	}

	exposure := c.Exposure
	args := []string{"--timeout", "0", "--output", "-", "--profile", "baseline", "--width", width, "--height", height, "--framerate", framerate, "--exposure", exposure}
	if c.Flip[0] {
		args = append(args, "--hflip")
	}
	if c.Flip[1] {
		args = append(args, "--vflip")
	}

	return args, nil
}

func startCmd(args []string) (cmd *exec.Cmd, stdout, stderr io.Reader, err error) {
	cmd = exec.Command("raspivid", args...)
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err = cmd.StderrPipe()
	if err != nil {
		return
	}
	err = cmd.Start()

	return
}

func stopCmd(cmd *exec.Cmd) (err error) {
	err = cmd.Process.Kill()
	if err != nil {
		err = cmd.Wait()
	}

	return err
}

func handleStreaming(requests <-chan Request, readAmount int, dataBuffer []byte, data chan<- Data) (nextRequest *Request, hasNextRequest bool) {
	select {
	case next, hasNext := <-requests:
		if hasNext && cap(next.Response) < 1 {
			close(next.Response)
			panic("received blocking channel for response")
		}
		nextRequest = &next
		hasNextRequest = hasNext
	case data <- Data{dataBuffer[:readAmount], nil}:
		hasNextRequest = true
	}

	return
}

func handleStreamingStart(request Request, requests <-chan Request, readAmount int, readError error, dataBuffer []byte, data chan<- Data) (nextRequest *Request, hasNextRequest bool) {
	defer close(request.Response)
	if readError == nil {
		select {
		case next, hasNext := <-requests:
			if hasNext && cap(next.Response) < 1 {
				close(next.Response)
				panic("received blocking channel for response")
			}
			nextRequest = &next
			hasNextRequest = hasNext
		case data <- Data{dataBuffer[:readAmount], nil}:
			hasNextRequest = true
		}
	} else {
		request.Response <- readError
	}

	return
}

func handleCommand(stdout, stderr io.Reader, request *Request, requests <-chan Request, data chan<- Data) *Request {
	for {
		dataBuffer := make([]byte, 4096)
		n, err := stdout.Read(dataBuffer)
		if errors.Is(err, io.EOF) {
			errorBuffer := make([]byte, 4096)
			n, err = stderr.Read(errorBuffer)
			if err == nil {
				err = fmt.Errorf("%w: %s", ErrProcess, errorBuffer[:n])
			}
		}
		switch {
		case request != nil:
			next, hasNext := handleStreamingStart(*request, requests, n, err, dataBuffer, data)
			if !hasNext {
				return nil
			}
			if next != nil {
				return next
			}
			request = nil
		case err == nil:
			next, hasNext := handleStreaming(requests, n, dataBuffer, data)
			if !hasNext {
				return nil
			}
			if next != nil {
				return next
			}
		case err != nil:
			data <- Data{nil, fmt.Errorf("raspivid failed: %w", err)}

			return nil
		}
	}
}

// New creates a new manager for a raspivid source.
func New(requests <-chan Request) <-chan Data {
	data := make(chan Data)

	go func() {
		defer close(data)

		var request *Request
		for {
			if request == nil {
				newRequest, ok := <-requests
				switch {
				case !ok:
					return
				case cap(newRequest.Response) < 1:
					close(newRequest.Response)
					panic("received blocking channel for response")
				}
				request = &newRequest
			}
			args, err := request.Configuration.arguments()
			if err != nil {
				request.Response <- err
				close(request.Response)

				continue
			}
			cmd, stdout, stderr, err := startCmd(args)
			if err != nil {
				request.Response <- err
				close(request.Response)

				continue
			}
			request = handleCommand(stdout, stderr, request, requests, data)
			if err = stopCmd(cmd); err != nil {
				data <- Data{nil, err}
			}
		}
	}()

	return data
}

/*
-b, --bitrate	: Set bitrate. Use bits per second (e.g. 10MBits/s would be -b 10000000)
-g, --intra	: Specify the intra refresh period (key frame rate/GoP size). Zero to produce an initial I-frame and then just P-frames.
-pf, --profile	: Specify H264 profile to use for encoding
-qp, --qp	: Quantisation parameter. Use approximately 10-40. Default 0 (off)
-ih, --inline	: Insert inline headers (SPS, PPS) to stream
-wr, --wrap	: In segment mode, wrap any numbered filename back to 1 when reach number
-sn, --start	: In segment mode, start with specified segment number
-sp, --split	: In wait mode, create new output file for each start event
-if, --irefresh	: Set intra refresh type (cyclic,adaptive,both,cyclicrows)
-fl, --flush	: Flush buffers in order to decrease latency
-cd, --codec	: Specify the codec to use - H264 (default) or MJPEG
-lev, --level	: Specify H264 level to use for encoding (4,4.1,4.2)
-stm, --spstimings	: Add in h.264 sps timings
-sl, --slices	: Horizontal slices per frame. Default 1 (off)
-v, --verbose	: Output verbose information during run
-md, --mode	: Force sensor mode. 0=auto. See docs for other modes available
-sh, --sharpness	: Set image sharpness (-100 to 100)
-co, --contrast	: Set image contrast (-100 to 100)
-br, --brightness	: Set image brightness (0 to 100)
-sa, --saturation	: Set image saturation (-100 to 100)
-ISO, --ISO	: Set capture ISO
-vs, --vstab	: Turn on video stabilisation
-ev, --ev	: Set EV compensation - steps of 1/6 stop
-ex, --exposure	: Set exposure mode off,auto,night,nightpreview,backlight,spotlight,sports,snow,beach,verylong,fixedfps,antishake,fireworks
-fli, --flicker	: Set flicker avoid mode off,auto,50hz,60hz
-awb, --awb	: Set AWB mode off,auto,sun,cloud,shade,tungsten,fluorescent,incandescent,flash,horizon,greyworld
-ifx, --imxfx	: Set image effect none,negative,solarise,sketch,denoise,emboss,oilpaint,hatch,gpen,pastel,watercolour,film,blur,saturation,colourswap,washedout,posterise,colourpoint,colourbalance,cartoon
-cfx, --colfx	: Set colour effect (U:V)
-mm, --metering	: Set metering mode average,spot,backlit,matrix
-hf, --hflip	: Set horizontal flip
-vf, --vflip	: Set vertical flip
-roi, --roi	: Set region of interest (x,y,w,d as normalised coordinates [0.0-1.0])
-ss, --shutter	: Set shutter speed in microseconds
-awbg, --awbgains	: Set AWB gains - AWB mode must be off
-drc, --drc	: Set DRC Level off,low,med,high
*/
