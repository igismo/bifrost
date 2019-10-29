
//=============================================================================
// FILE NAME: expMessages.go
// DESCRIPTION:
// Utilities for creating, marshaling and unmarshaling messages used by the
// Experiment Master and Experiment Agents.
// Contains description of all possible messages related to Experiments
//
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  01012018  Initial design     goran@usa.net
//================================================================================

package main

import (
	"net"
	"bytes"
	"io"
)

//=============================================================================
// Function:    expMessage
// Description:
// Input:
// Output:
// Error Conditions:
//      None [or state condition for each possible error]
//=============================================================================
func readFully(conn net.Conn) ([]byte, error) {
	defer conn.Close()

	result := bytes.NewBuffer(nil)
	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		result.Write(buf[0:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return result.Bytes(), nil
}
