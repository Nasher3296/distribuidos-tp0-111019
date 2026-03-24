package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func Send(conn net.Conn, data []byte) error {
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, uint16(len(data)))
	return writeAll(conn, append(header, data...))
}

func SendBatch(conn net.Conn, records [][]byte) error {
	payload := bytes.Join(records, []byte("\n"))
	return Send(conn, payload)
}

func RecvAck(conn net.Conn) error {
	buf := make([]byte, 1)
	if err := readAll(conn, buf); err != nil {
		return err
	}
	if buf[0] != 0x00 {
		return fmt.Errorf("server responded with error status")
	}
	return nil
}

func writeAll(conn net.Conn, data []byte) error {
	written := 0
	for written < len(data) {
		n, err := conn.Write(data[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}

func readAll(conn net.Conn, buf []byte) error {
	total := 0
	for total < len(buf) {
		n, err := conn.Read(buf[total:])
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("connection closed after reading %d/%d bytes", total, len(buf))
			}
			return err
		}
		total += n
	}
	return nil
}
