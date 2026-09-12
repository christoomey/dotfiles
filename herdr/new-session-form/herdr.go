package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"time"
)

// herdrCall sends one request over herdr's socket API (newline-delimited
// JSON, one request per connection) and decodes the result into out. The
// CLI wraps most methods, but not workspace.move, so the sidebar talks to the
// socket directly.
func herdrCall(method string, params any, out any) error {
	conn, err := net.DialTimeout("unix", herdrSocketPath(), 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	req := struct {
		ID     string `json:"id"`
		Method string `json:"method"`
		Params any    `json:"params"`
	}{ID: "form", Method: method, Params: params}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return err
	}
	var resp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(line, &resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return errors.New(resp.Error.Code + ": " + resp.Error.Message)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(resp.Result, out)
}

func herdrSocketPath() string {
	if p := os.Getenv("HERDR_SOCKET_PATH"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "herdr", "herdr.sock")
}
