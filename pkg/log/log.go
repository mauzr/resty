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

// Package log interfaces with journald or logs to stdout.
package log

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
)

// Logger instance to dump logs into.
type Logger interface {
	Error(message string, args ...interface{})
	Warning(message string, args ...interface{})
	Notice(message string, args ...interface{})
	Informational(message string, args ...interface{})
	Debug(message string, args ...interface{})
	RetainedMessages() []string
	RetainLevel(int)
}

// Root is the default logger for all.
var Root Logger = &systemdLogger{4, []string{}, sync.Mutex{}}

type systemdLogger struct {
	retainLevel int
	messages    []string
	mutex       sync.Mutex
}

var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_]+$`)

func writeField(builder io.Writer, name string, value interface{}) {
	if !validName.MatchString(name) {
		panic(fmt.Sprintf("invalid field name: %v", name))
	}
	name = strings.ToUpper(name)
	valueString := fmt.Sprintf("%v", value)

	if strings.ContainsRune(valueString, '\n') {
		fmt.Fprintf(builder, "%v\n", name)
		if err := binary.Write(builder, binary.LittleEndian, uint64(len(valueString))); err != nil {
			panic(err)
		}
		fmt.Fprintf(builder, "%v\n", valueString)
	} else {
		fmt.Fprintf(builder, "%v=%v\n", name, valueString)
	}
}

func (l *systemdLogger) RetainLevel(level int) {
	l.retainLevel = level
}

func (l *systemdLogger) RetainedMessages() []string {
	l.mutex.Lock()
	lines := l.messages
	l.messages = []string{}
	l.mutex.Unlock()
	return lines
}

func (l *systemdLogger) send(priority int, message string, args []interface{}) {
	builder := strings.Builder{}
	writeField(&builder, "PRIORITY", priority)
	m := fmt.Sprintf(message, args...)
	writeField(&builder, "MESSAGE", m)
	c, err := net.Dial("unixgram", "/run/systemd/journal/socket")
	if err != nil {
		panic(fmt.Sprintf("could not log to journal: %v", err))
	}

	if priority <= l.retainLevel {
		l.mutex.Lock()
		l.messages = append(l.messages, m)
		if len(l.messages) >= 100 {
			l.messages = l.messages[:100]
		}
		l.mutex.Unlock()
	}

	_, err = c.Write([]byte(builder.String()))
	_ = c.Close()
	if err != nil {
		panic(fmt.Sprintf("could not log to journal: %v", err))
	}
}

func (l *systemdLogger) Error(message string, args ...interface{}) {
	l.send(3, message, args)
}

func (l *systemdLogger) Warning(message string, args ...interface{}) {
	l.send(4, message, args)
}

func (l *systemdLogger) Notice(message string, args ...interface{}) {
	l.send(5, message, args)
}

func (l *systemdLogger) Informational(message string, args ...interface{}) {
	l.send(6, message, args)
}

func (l *systemdLogger) Debug(message string, args ...interface{}) {
	l.send(7, message, args)
}
