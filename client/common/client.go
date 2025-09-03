package common

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ===================================================
// Config & Client
// ===================================================

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	active bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		active: true,
	}
	return client
}

// ===================================================
// Signal handlers
// ===================================================

func (c *Client) handleSigterm() {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGTERM)

	go func() {
		_ = <-channel
		log.Infof("action: handleSigterm | result: in_progress | client_id: %v", c.config.ID)
		c.active = false
		signal.Stop(channel)
		close(channel)
		log.Infof("action: handleSigterm | result: success | client_id: %v", c.config.ID)
	}()
}

// ===================================================
// Connection
// ===================================================

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// ===================================================
// Writer/Reader helpers
// ===================================================

func writeU32(w io.Writer, n uint32) error {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], n)
	_, err := w.Write(b[:])
	return err
}

func readU32(r io.Reader) (uint32, error) {
	var n uint32
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func writeAll(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		p = p[n:]
	}
	return nil
}

func readExact(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

func sendFramed(conn net.Conn, payload []byte) error {
	if err := writeU32(conn, uint32(len(payload))); err != nil {
		return err
	}
	return writeAll(conn, payload)
}

func recvFramed(conn net.Conn) ([]byte, error) {
	n, err := readU32(conn)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("empty payload")
	}
	return readExact(conn, int(n))
}

// ===================================================
// Messages
// ===================================================

func BuildBetMessageFromEnv() *BetMessage {
	return &BetMessage{
		Agency:    os.Getenv("CLI_ID"),
		FirstName: os.Getenv("FIRST_NAME"),
		LastName:  os.Getenv("LAST_NAME"),
		Document:  os.Getenv("DOCUMENT"),
		Birthdate: os.Getenv("BIRTH_DATE"),
		Number:    os.Getenv("NUMBER"),
	}
}

func (c *Client) sendBetAndGetResponse(bet *BetMessage) (string, error) {
	betMessageSerialized := bet.Serialize()

	if err := sendFramed(c.conn, []byte(betMessageSerialized)); err != nil {
		return "", err
	}
	resp, err := recvFramed(c.conn)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// ===================================================
// Main loop
// ===================================================

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {

	c.handleSigterm()

	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount && c.active; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		// Create message
		betMessage := BuildBetMessageFromEnv()

		response, err := c.sendBetAndGetResponse(betMessage)
		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return
		}

		if MatchesAckMessage(response) {
			log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
				betMessage.Document, betMessage.Number)
		}

		// Close socket
		c.conn.Close()

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			response,
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
