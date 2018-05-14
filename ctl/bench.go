// Copyright 2017 Pilosa Corp.
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

package ctl

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/pilosa/pilosa"
	"github.com/pilosa/pilosa/internal"
	"github.com/pilosa/pilosa/server"
	"github.com/pkg/errors"
)

// BenchCommand represents a command for benchmarking index operations.
type BenchCommand struct {
	// Destination host and port.
	Host string

	// Name of the index & frame to execute against.
	Index string
	Frame string

	// Type of operation and number to execute.
	Op string
	N  int

	// Standard input/output
	*pilosa.CmdIO

	TLS server.TLSConfig
}

// NewBenchCommand returns a new instance of BenchCommand.
func NewBenchCommand(stdin io.Reader, stdout, stderr io.Writer) *BenchCommand {
	return &BenchCommand{
		CmdIO: pilosa.NewCmdIO(stdin, stdout, stderr),
	}
}

// Run executes the bench command.
func (cmd *BenchCommand) Run(ctx context.Context) error {
	// Create a client to the server.
	client, err := CommandClient(cmd)
	if err != nil {
		return errors.Wrap(err, "creating client")
	}

	switch cmd.Op {
	case "set-bit":
		return cmd.runSetBit(ctx, client)
	case "":
		return errors.New("op required")
	default:
		return fmt.Errorf("unknown bench op: %q", cmd.Op)
	}
}

// runSetBit executes a benchmark of random SetBit() operations.
func (cmd *BenchCommand) runSetBit(ctx context.Context, client pilosa.ExternalClient) error {
	if cmd.N == 0 {
		return errors.New("operation count required")
	} else if cmd.Index == "" {
		return pilosa.ErrIndexRequired
	} else if cmd.Frame == "" {
		return pilosa.ErrFrameRequired
	}

	const maxRowID = 1000
	const maxColumnID = 100000

	startTime := time.Now()

	// Execute operation continuously.
	for i := 0; i < cmd.N; i++ {
		rowID := rand.Intn(maxRowID)
		columnID := rand.Intn(maxColumnID)

		queryRequest := &internal.QueryRequest{
			Query:  fmt.Sprintf(`SetBit(row=%d, frame="%s", col=%d)`, rowID, cmd.Frame, columnID),
			Remote: false,
		}
		if _, err := client.Query(ctx, cmd.Index, queryRequest); err != nil {
			return err
		}
	}

	// Print results.
	elapsed := time.Since(startTime)
	fmt.Fprintf(cmd.Stdout, "Executed %d operations in %s (%0.3f op/sec)\n", cmd.N, elapsed, float64(cmd.N)/elapsed.Seconds())

	return nil
}

func (cmd *BenchCommand) TLSHost() string {
	return cmd.Host
}

func (cmd *BenchCommand) TLSConfiguration() server.TLSConfig {
	return cmd.TLS
}
