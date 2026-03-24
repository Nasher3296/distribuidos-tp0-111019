package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	MsgTypeBatch        byte = 0x01
	MsgTypeDone         byte = 0x02
	MsgTypeQueryWinners byte = 0x03
)

const (
	MsgTypeAckOk    byte = 0x10
	MsgTypeAckError byte = 0x11
	MsgTypeWinners  byte = 0x12
)

const (
	RecordSeparator byte = '\n'
	FieldSeparator  byte = ','
)

//   [1 byte: message type][2 bytes uint16 BE: payload length][payload]

func Send(conn net.Conn, msgType byte, data []byte) error {
	header := make([]byte, 3)
	header[0] = msgType
	binary.BigEndian.PutUint16(header[1:], uint16(len(data)))
	return writeAll(conn, append(header, data...))
}

func SendBatch(conn net.Conn, records [][]byte) error {
	payload := bytes.Join(records, []byte{RecordSeparator})
	return Send(conn, MsgTypeBatch, payload)
}

func RecvMessage(conn net.Conn) (byte, []byte, error) {
	typeBuf := make([]byte, 1)
	if err := readAll(conn, typeBuf); err != nil {
		return 0, nil, err
	}
	msgType := typeBuf[0]

	lenBuf := make([]byte, 2)
	if err := readAll(conn, lenBuf); err != nil {
		return 0, nil, err
	}
	length := binary.BigEndian.Uint16(lenBuf)

	var payload []byte
	if length > 0 {
		payload = make([]byte, length)
		if err := readAll(conn, payload); err != nil {
			return 0, nil, err
		}
	}
	return msgType, payload, nil
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
