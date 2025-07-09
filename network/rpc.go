package network

import (
	"encoding/json"
	"net"
)

// Listen decodifica JSON e delega ao handler
func Listen(addr string, handler func(raw []byte, conn net.Conn)) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			var raw json.RawMessage
			if err := json.NewDecoder(c).Decode(&raw); err != nil {
				return
			}
			handler(raw, c)
		}(conn)
	}
}

// Request envia req e espera resp (JSON)
func Request(addr string, req, resp any) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	json.NewEncoder(conn).Encode(req)
	return json.NewDecoder(conn).Decode(resp)
}

// Send envia msg (JSON) sem resposta
func Send(addr string, msg any) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	return json.NewEncoder(conn).Encode(msg)
}
