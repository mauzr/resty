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

// Package influxdb interfaces with influxdb TSDB instance.
package influxdb

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.eqrx.net/mauzr/pkg/rest"
)

// Measurement is a sample set that need to be stored in an influxdb.
type Measurement struct {
	Name      string
	Tags      map[string]string
	Fields    map[string]interface{}
	Timestamp time.Time
}

// Client implements Client.
type Client struct {
	c           rest.Client
	destination string
	token       string
}

// Line returns the measurement as line protocol string.
func (m Measurement) Line() string {
	timestamp := m.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	tags := make([]string, 0, len(m.Tags))
	for key, value := range m.Tags {
		tags = append(tags, fmt.Sprintf("%s=%v", key, value))
	}
	fields := make([]string, 0, len(m.Fields))
	for key, value := range m.Fields {
		fields = append(fields, fmt.Sprintf("%s=%v", key, value))
	}

	return fmt.Sprintf("%s,%s %s %v", m.Name, strings.Join(tags, ","), strings.Join(fields, ","), timestamp.UnixNano())
}

// New creates a new influxdb client.
func New(destination, token string, tls *tls.Config) *Client {
	return &Client{
		rest.NewClient(tls),
		destination,
		token,
	}
}

// Write an influxdb measurement to the database.
func (c Client) Write(ctx context.Context, bucket string, measurements ...Measurement) error {
	url := fmt.Sprintf("%sapi/v2/write?org=eqrx&bucket=%s&precision=ns", c.destination, bucket)
	lines := make([]string, 0, len(measurements))
	for _, m := range measurements {
		lines = append(lines, m.Line())
	}
	body := strings.Join(lines, "\n")

	return c.c.Request(ctx, url, http.MethodPost).Header("Authorization", fmt.Sprintf("Token %s", c.token)).StringBody(body).Send(http.StatusNoContent).Check()
}
