package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/bet"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/protocol"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	Bet           bet.Bet
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration as a parameter
func NewClient(config ClientConfig) *Client {
	return &Client{config: config}
}

// createClientSocket Initializes client socket. In case of failure, error is
// printed in stdout/stderr and exit 1 is returned
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

// Run connects to the server, sends the configured bet, and waits for confirmation
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

	if err := protocol.SendBet(c.conn, c.config.Bet); err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	if err := protocol.RecvAck(c.conn); err != nil {
		log.Errorf("action: receive_ack | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
		c.config.Bet.Document, c.config.Bet.Number)
}
