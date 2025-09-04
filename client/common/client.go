package common

import (
	"bufio"
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
// CSV
// ===================================================

func (c *Client) openCSV(path string) (*os.File, *bufio.Scanner, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	sc := bufio.NewScanner(f)
	return f, sc, nil
}

// ===================================================
// Fase 1: Enviar batches de apuestas
// ===================================================

func (c *Client) BuildBetBatchFromCSV(sc *bufio.Scanner, n int) (*BetBatchMessage, error) {
	bets := make([]BetMessage, 0, n)

	for i := 0; i < n && sc.Scan(); i++ {
		cols := strings.Split(sc.Text(), ",")
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

	if len(bets) == 0 {
		return nil, io.EOF
	}

	return &BetBatchMessage{Bets: bets}, nil
}

func (c *Client) sendBetBatchAndGetResponse(sc *bufio.Scanner, n int) (*AckMessage, error) {
	// Leer N líneas
	BetBatchMessage, err := c.BuildBetBatchFromCSV(sc, n)
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

func (c *Client) sendBetsPhase(sc *bufio.Scanner) error {
	for c.active {
		ack, err := c.sendBetBatchAndGetResponse(sc, c.config.BatchMaxAmount)
		if err == io.EOF {
			break
		}
		if err != nil || !ack.Success {
			log.Errorf("action: apuesta_enviada | result: fail")
			_ = c.conn.Close()
			return err
		}
		time.Sleep(c.config.LoopPeriod)
	}
	log.Infof("action: apuesta_enviada | result: success")
	return nil
}

// ===================================================
// Fase 2: Notificar fin de apuestas y cerrar socket
// ===================================================

func (c *Client) sendEndOfBetsAndClose() error {
	msg := EndOfBetsMessage{Agency: c.config.ID}
	msgSerialized := msg.Serialize()
	if err := sendFramed(c.conn, []byte(msgSerialized)); err != nil {
		_ = c.conn.Close()
		return err
	}
	return c.conn.Close()
}

// ===================================================
// Fase 3: Consulta de ganadores
// ===================================================

func (c *Client) winnersRequestOnce() (pending bool, notif *WinnersNotificationMessage, err error) {
	if err := c.createClientSocket(); err != nil {
		return false, nil, err
	}
	defer c.conn.Close()

	winnerRequestMessage := WinnersRequestMessage{Agency: c.config.ID}
	if err := sendFramed(c.conn, []byte(winnerRequestMessage.Serialize())); err != nil {
		return false, nil, err
	}

	responseEncoded, err := recvFramed(c.conn)
	if err != nil {
		return false, nil, err
	}
	response := string(responseEncoded)

	if MatchesWinnersPendingMessage(response) {
		return true, nil, nil
	}
	if MatchesWinnersNotificationMessage(response) {
		n, derr := DeserializeWinnersNotificationMessage(response)
		if derr != nil {
			return false, nil, derr
		}
		return false, n, nil
	}
	return false, nil, fmt.Errorf("unexpected response to WinnersRequest: %q", response)
}

func (c *Client) pollWinnersUntilReady(pollDelay time.Duration) (*WinnersNotificationMessage, error) {
	for c.active {
		pending, winnerNotificationMessage, err := c.winnersRequestOnce()
		if err != nil {
			return nil, err
		}
		if pending {
			time.Sleep(pollDelay)
			continue
		}
		return winnerNotificationMessage, nil
	}
	return nil, fmt.Errorf("client deactivated")
}

// ===================================================
// Main Loop
// ===================================================

func (c *Client) StartClientLoop() error {
	c.handleSigterm()

	// 1) Abrir CSV
	csvFile, sc, err := c.openCSV("/bets.csv")
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// 2) Abrir conexion y enviar batches
	if err := c.createClientSocket(); err != nil {
		return err
	}
	if err := c.sendBetsPhase(sc); err != nil {
		return err
	}

	// 3) Enviar EndOfBets y cerrar socket
	if err := c.sendEndOfBetsAndClose(); err != nil {
		return err
	}

	// 4) Polling de winners
	const pollDelay = 200 * time.Millisecond
	notif, err := c.pollWinnersUntilReady(pollDelay)
	if err != nil {
		return err
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", notif.Count)
	return nil
}
