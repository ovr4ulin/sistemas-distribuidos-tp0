package common

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

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
