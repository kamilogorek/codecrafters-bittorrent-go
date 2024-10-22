package main

import (
	"bytes"
	"io"
	"net"
)

type TCPClient struct {
	net.Conn
}

func NewTCPClient(address string) (*TCPClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &TCPClient{
		Conn: conn,
	}, nil
}

func (c *TCPClient) Receive(bufSize int) (int, []byte, error) {
	var received int
	buffer := bytes.NewBuffer(nil)
	for {
		chunk := make([]byte, bufSize)
		read, err := c.Conn.Read(chunk)
		if err != nil && err != io.EOF {
			return received, buffer.Bytes(), err
		}
		received += read
		buffer.Write(chunk[:read])

		if read == 0 || read < bufSize {
			break
		}
	}
	return received, buffer.Bytes(), nil
}

func (c *TCPClient) ReceiveExact(expectedSize int) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	for {
		chunk := make([]byte, expectedSize)
		io.ReadFull(c.Conn, chunk)
		buffer.Write(chunk)
		if buffer.Len() == expectedSize {
			break
		}
	}
	return buffer.Bytes(), nil
}
