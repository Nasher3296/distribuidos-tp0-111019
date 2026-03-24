package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/bet"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/protocol"
)

var log = logging.MustGetLogger("log")

type ClientConfig struct {
	ID            string
	ServerAddress string
	MaxBatchSize  int
}

type Client struct {
	config ClientConfig
	conn   net.Conn
}

func NewClient(config ClientConfig) *Client {
	return &Client{config: config}
}

func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	if err := c.createClientSocket(); err != nil {
		return
	}
	defer func() {
		c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}()

	go func() {
		<-sigChan
		log.Infof("action: sigterm_received | result: success | client_id: %v", c.config.ID)
		c.conn.Close()
	}()

	if err := c.sendAllBets(); err != nil {
		return
	}
	if err := c.notifyDone(); err != nil {
		return
	}
	c.queryWinners()
}

func (c *Client) sendAllBets() error {
	filePath := fmt.Sprintf("/data/agency-%s.csv", c.config.ID)
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("action: open_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for {
		batch := c.readNextBatch(scanner)
		if len(batch) == 0 {
			break
		}
		if err := c.sendBatch(batch); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) readNextBatch(scanner *bufio.Scanner) []bet.Bet {
	var batch []bet.Bet
	for len(batch) < c.config.MaxBatchSize && scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			continue
		}
		batch = append(batch, bet.Bet{
			Agency:    c.config.ID,
			FirstName: fields[0],
			LastName:  fields[1],
			Document:  fields[2],
			Birthdate: fields[3],
			Number:    fields[4],
		})
	}
	return batch
}

func (c *Client) sendBatch(bets []bet.Bet) error {
	records := make([][]byte, len(bets))
	for i, b := range bets {
		records[i] = b.ToCsvBytes()
	}

	if err := protocol.SendBatch(c.conn, records); err != nil {
		log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}

	if err := c.receiveAck(); err != nil {
		log.Errorf("action: receive_ack | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}

	for _, b := range bets {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			b.Document, b.Number)
	}

	return nil
}

func (c *Client) notifyDone() error {
	if err := protocol.Send(c.conn, protocol.MsgTypeDone, []byte(c.config.ID)); err != nil {
		log.Errorf("action: notify_done | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}
	if err := c.receiveAck(); err != nil {
		log.Errorf("action: notify_done | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}
	return nil
}

func (c *Client) queryWinners() {
	if err := protocol.Send(c.conn, protocol.MsgTypeQueryWinners, []byte(c.config.ID)); err != nil {
		log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	msgType, payload, err := protocol.RecvMessage(c.conn)
	if err != nil {
		log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	if msgType != protocol.MsgTypeWinners {
		log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: unexpected message type %d", c.config.ID, msgType)
		return
	}

	var winners []string
	if len(payload) > 0 {
		winners = strings.Split(string(payload), string([]byte{protocol.RecordSeparator}))
	}
	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(winners))
}

func (c *Client) receiveAck() error {
	msgType, _, err := protocol.RecvMessage(c.conn)
	if err != nil {
		return err
	}
	if msgType != protocol.MsgTypeAckOk {
		return fmt.Errorf("server responded with error status")
	}
	return nil
}
