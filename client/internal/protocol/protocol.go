package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/bet"
)

func SendBet(conn net.Conn, b bet.Bet) error {
	payload := fmt.Sprintf("%s,%s,%s,%s,%s,%s",
		b.Agency, b.FirstName, b.LastName,
		b.Document, b.Birthdate, b.Number)

	data := []byte(payload)
	msg := make([]byte, 2+len(data))
	binary.BigEndian.PutUint16(msg[:2], uint16(len(data)))
	copy(msg[2:], data)

	return writeAll(conn, msg)
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
