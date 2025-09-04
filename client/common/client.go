package common

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
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
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
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

func (c *Client) ReadBetBatch(sc *bufio.Scanner, n int) (*BetBatchMessage, error) {
	bets := make([]BetMessage, 0, n)

	for i := 0; i < n && sc.Scan(); i++ {
		cols := strings.Split(sc.Text(), ",")
		log.Infof("action: lectura | result: in_progress | columns: %v %v", cols, i)
		bets = append(bets, BetMessage{
			Agency:    c.config.ID,
			FirstName: cols[0],
			LastName:  cols[1],
			Document:  cols[2],
			Birthdate: cols[3],
			Number:    cols[4],
		})
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	// Si no leíste nada → EOF
	if len(bets) == 0 {
		return nil, io.EOF
	}

	return &BetBatchMessage{Bets: bets}, nil
}

func (c *Client) sendBetBatchAndGetResponse(sc *bufio.Scanner, n int) (*AckMessage, error) {
	// Leer N líneas
	BetBatchMessage, err := c.ReadBetBatch(sc, n)
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}

	// Serializar el mensaje
	BetBatchMessageSerialized := BetBatchMessage.Serialize()

	// Enviar mensaje
	if err := sendFramed(c.conn, []byte(BetBatchMessageSerialized)); err != nil {
		return nil, err
	}

	// Recibir respuesta
	resp, err := recvFramed(c.conn)
	if err != nil {
		return nil, err
	}
	response := string(resp)

	// Checkear ACK
	if !MatchesAckMessage(response) {
		return nil, fmt.Errorf("unexpected response, expected AckMessage, got: %q", response)
	}
	ack, err := DeserializeAckMessage(response)
	if err != nil {
		return nil, err
	}

	return ack, nil
}

// ===================================================
// Main loop
// ===================================================

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() error {
	c.handleSigterm()

	// Abrir archivo CSV
	csvFilename := "/bets.csv"
	csvFile, err := os.Open(csvFilename)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	sc := bufio.NewScanner(csvFile)

	// Abrir conexion
	if err := c.createClientSocket(); err != nil {
		return err
	}
	defer c.conn.Close()

	for c.active {

		ack, err := c.sendBetBatchAndGetResponse(sc, c.config.BatchMaxAmount)

		if err == io.EOF {
			break
		}

		if err != nil || !ack.Success {
			log.Errorf("action: apuesta_enviada | result: fail")
			return err
		}

		time.Sleep(c.config.LoopPeriod)
	}

	log.Infof("action: apuesta_enviada | result: success")
	return nil
}
