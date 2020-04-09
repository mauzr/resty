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

package raspivid

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

type Configuration struct {
	On                       bool
	Width, Height, Framerate int
	Flip                     [2]bool
	Exposure                 string
}

type Data struct {
	Data []byte
	Err  error
}

func (c Configuration) arguments() ([]string, error) {
	if !c.On {
		return nil, fmt.Errorf("not on")
	}
	if c.Width <= 0 {
		return nil, fmt.Errorf("invalid width")
	}
	if c.Height <= 0 {
		return nil, fmt.Errorf("invalid height")
	}
	if c.Framerate <= 0 {
		return nil, fmt.Errorf("invalid framerate")
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
		return nil, fmt.Errorf("illegal exposure mode value")
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
	fmt.Printf("configuring raspivid with %v\n", args)
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

type Request struct {
	Configuration Configuration
	Response      chan<- error
}

func handleCommand(stdout, stderr io.Reader, request *Request, requests <-chan Request, data chan<- Data) *Request {
	for {
		dataBuffer := make([]byte, 4096)
		n, err := stdout.Read(dataBuffer)
		if err == io.EOF {
			errorBuffer := make([]byte, 4096)
			n, err = stderr.Read(errorBuffer)
			if err == nil {
				err = fmt.Errorf(string(errorBuffer[:n]))
			}
		}
		switch {
		case request == nil && err == nil:
			select {
			case nextRequest, ok := <-requests:
				switch {
				case !ok:
					return nil
				case cap(nextRequest.Response) < 1:
					close(nextRequest.Response)
					panic(fmt.Errorf("received blocking channel for response"))
				}
				return &nextRequest
			case data <- Data{dataBuffer[:n], nil}:
			}
		case request != nil && err == nil:
			close(request.Response)
			request = nil
			select {
			case nextRequest, ok := <-requests:
				switch {
				case !ok:
					return nil
				case cap(nextRequest.Response) < 1:
					close(nextRequest.Response)
					panic(fmt.Errorf("received blocking channel for response"))
				}
				return &nextRequest
			case data <- Data{dataBuffer[:n], nil}:
			}
		case request == nil && err != nil:
			data <- Data{nil, fmt.Errorf("raspivid failed: %s", err)}
			return nil
		case request != nil && err != nil:
			request.Response <- err
			close(request.Response)
			return nil
		}
	}
}

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
					panic(fmt.Errorf("received blocking channel for response"))
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
