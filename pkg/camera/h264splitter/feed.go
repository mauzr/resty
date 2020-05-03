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

// Package h264splitter contains splitters for H264 streams.
package h264splitter

import (
	"bytes"
	"fmt"
	"io"

	"go.eqrx.net/mauzr/pkg/camera/raspivid"
)

// Data is the result of the splitter.
type Data struct {
	// Frame is an individual NAL frame.
	Frame []byte
	// Err is an error that was received from the stream source or occurred in the splitter.
	Err error
}

// New creates a splitter the separates H264 NAL frames.
func New(source <-chan raspivid.Data) <-chan Data {
	feed := make(chan Data)
	go func() {
		defer close(feed)
		defer func() {
			for {
				data, ok := <-source
				if !ok {
					return
				}
				if data.Err != nil {
					feed <- Data{nil, data.Err}
				}
			}
		}()
		frameSeparator := []byte{0, 0, 0, 1}
		workBuffer := &bytes.Buffer{}
		frameBuffer := &bytes.Buffer{}

		for {
			data, ok := <-source
			if !ok {
				return
			}
			if data.Err != nil {
				feed <- Data{nil, data.Err}
				return
			}
			_, _ = workBuffer.Write(data.Data)

			for workBuffer.Len() != 0 {
				var ready bool
				ready, err := processWorkBuffer(frameBuffer, workBuffer, frameSeparator)
				if err != nil {
					feed <- Data{nil, fmt.Errorf("h264splitter: %w", err)}
					return
				}
				if ready {
					feed <- Data{frameBuffer.Bytes(), nil}
					frameBuffer = &bytes.Buffer{}
				}
			}
		}
	}()
	return feed
}

func processWorkBuffer(frameBuffer, workBuffer *bytes.Buffer, frameSeparator []byte) (ready bool, err error) {
	separatorIndex := bytes.Index(workBuffer.Bytes(), frameSeparator)
	switch separatorIndex {
	case 0:
		if frameBuffer.Len() != 0 {
			ready = true
		} else {
			_, err = io.CopyN(frameBuffer, workBuffer, int64(len(frameSeparator)))
		}
	case -1:
		_, err = io.Copy(frameBuffer, workBuffer)
	default:
		_, err = io.CopyN(frameBuffer, workBuffer, int64(separatorIndex))
	}
	return
}
