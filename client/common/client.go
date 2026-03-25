package common

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/bet"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/protocol"
)

var log = logging.MustGetLogger("log")

type ClientConfig struct {
	ID            string
	ServerAddress string
	Bet           bet.Bet
	LoopAmount    int
	LoopPeriod    time.Duration
}

type Client struct {
	config ClientConfig
	conn   net.Conn
}

func NewClient(config ClientConfig) *Client {
	return &Client{config: config}
}

func (c *Client) createClientSocket() error {
	initial_amount := c.config.LoopAmount
	for c.config.LoopAmount > 0 {
		conn, err := net.Dial("tcp", c.config.ServerAddress)
		if err == nil {
			c.conn = conn
			return nil
		}
		log.Errorf(
			"action: connect | result: fail | client_id: %v | pending_retries: %v | error: %v",
			c.config.ID, c.config.LoopAmount, err,
		)
		time.Sleep(c.config.LoopPeriod)
		c.config.LoopAmount--
	}
	log.Criticalf("action: connect | result: fail | client_id: %v | error: max retries reached", c.config.ID)
	return fmt.Errorf("could not connect after %d attempts", initial_amount)
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

	b := c.config.Bet

	if err := protocol.Send(c.conn, b.ToCsvBytes()); err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if err := protocol.RecvAck(c.conn); err != nil {
		log.Errorf("action: receive_ack | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
		b.Document, b.Number)
}
