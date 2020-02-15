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

package influxdb

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.eqrx.net/mauzr/pkg/io/rest"
)

// Measurement is a sample set that need to be stored in an influxdb.
type Measurement struct {
	Name      string
	Tags      map[string]interface{}
	Fields    map[string]interface{}
	Timestamp time.Time
}

type Client interface {
	Write(ctx context.Context, bucket string, measurements ...Measurement) error
}

// client implements Client
type client struct {
	c           http.Client
	destination string
	token       string
}

// prepareMap converts an map into a string suitable for the line protocol.
func prepareMap(input map[string]interface{}) string {
	sets := make([]string, 0, len(input))
	for key, value := range input {
		sets = append(sets, fmt.Sprintf("%s=%v", key, value))
	}
	return strings.Join(sets, ",")
}

// Line returns the measurement as line protocol string.
func (m Measurement) Line() string {
	timestamp := m.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	return fmt.Sprintf("%s,%s %s %v", m.Name, prepareMap(m.Tags), prepareMap(m.Fields), timestamp.UnixNano())
}

// New creates a new influxdb client.
func New(c rest.REST, destination, token string) Client {
	client := client{
		http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					ClientAuth:               tls.RequireAndVerifyClientCert,
					MinVersion:               tls.VersionTLS12,
					PreferServerCipherSuites: true,
				},
			},
		},
		destination,
		token,
	}
	return client
}

// Write an influxdb measurement to the database.
func (c client) Write(ctx context.Context, bucket string, measurements ...Measurement) error {
	url := fmt.Sprintf("%s/api/v2/write?org=eqrx&bucket=%s&precision=ns", c.destination, bucket)
	lines := make([]string, 0, len(measurements))
	for _, m := range measurements {
		lines = append(lines, m.Line())
	}
	body := strings.Join(lines, "\n")
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err == nil {
		request.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))
		var response *http.Response
		response, err = c.c.Do(request)
		switch {
		case err != nil:
		case response.StatusCode != http.StatusNoContent:
			err = fmt.Errorf("%v", response.Status)
		default:
			response.Body.Close()
		}
	}
	return err
}
