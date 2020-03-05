// Copyright 2020 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package forwarder

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
)

var _ protocol = &tcpProtocol{}

type tcpProtocol struct {
	conn net.Conn
}

func (c *tcpProtocol) makeRequest(ctx context.Context, req *request) (string, error) {
	var outBuffer bytes.Buffer
	outBuffer.WriteString(fmt.Sprintf("[%d] Url=%s\n", req.RequestID, req.URL))

	if req.Message != "" {
		outBuffer.WriteString(fmt.Sprintf("[%d] Echo=%s\n", req.RequestID, req.Message))
	}

	// Apply per-request timeout to calculate deadline for reads/writes.
	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	// Apply the deadline to the connection.
	deadline, _ := ctx.Deadline()
	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		return outBuffer.String(), err
	}
	if err := c.conn.SetReadDeadline(deadline); err != nil {
		return outBuffer.String(), err
	}

	// Make sure the client write something to the buffer
	message := "HelloWorld"
	if req.Message != "" {
		message = req.Message
	}

	_, err := c.conn.Write([]byte(message))
	if err != nil {
		return outBuffer.String(), err
	}

	resp := make([]byte, 1024)
	n, err := c.conn.Read(resp)
	if err != nil {
		return outBuffer.String(), err
	}

	outBuffer.WriteString(fmt.Sprintf("[%d] Read %d bytes\n", req.RequestID, n))

	for _, line := range strings.Split(string(resp), "\n") {
		if line != "" {
			outBuffer.WriteString(fmt.Sprintf("[%d body] %s\n", req.RequestID, line))
		}
	}

	return outBuffer.String(), nil
}

func (c *tcpProtocol) Close() error {
	c.conn.Close()
	return nil
}
