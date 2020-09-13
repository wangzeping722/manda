package server

import (
	"bufio"
	"bytes"
	"github.com/pkg/errors"
	"github.com/wangzeping722/manda/store"
	"github.com/yongman/go/goredis"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var (
	ErrParams        = errors.New("ERR params invalid")
	ErrRespType      = errors.New("ERR resp type invalid")
	ErrCmdNotSupport = errors.New("ERR command not supported")
)

type Command struct {
	cmd  string
	args [][]byte
}

type Client struct {
	app   *App
	store *store.Store
	cmd   string
	args  [][]byte

	buf     bytes.Buffer
	conn    net.Conn
	rReader *goredis.RespReader
	rWriter *goredis.RespWriter
	logger  *log.Logger
}

func newClient(app *App) *Client {
	return &Client{
		app:    app,
		store:  app.store,
		logger: log.New(os.Stderr, "[client] ", log.LstdFlags),
	}
}

func ClientHandler(conn net.Conn, app *App) {
	c := newClient(app)
	c.conn = conn
	br := bufio.NewReader(conn)
	c.rReader = goredis.NewRespReader(br)

	bw := bufio.NewWriter(conn)
	c.rWriter = goredis.NewRespWriter(bw)

	go c.connHandler()
}

func (c *Client) Resp(resp interface{}) error {
	var err error = nil

	switch v := resp.(type) {
	case []interface{}:
		err = c.rWriter.WriteArray(v)
	case []byte:
		err = c.rWriter.WriteBulk(v)
	case nil:
		err = c.rWriter.WriteBulk(nil)
	case int64:
		err = c.rWriter.WriteInteger(v)
	case string:
		err = c.rWriter.WriteString(v)
	case error:
		err = c.rWriter.WriteError(v)
	default:
		err = ErrRespType
	}

	return err
}

func (c *Client) FlushResp(resp interface{}) error {
	err := c.Resp(resp)
	if err != nil {
		return err
	}
	return c.rWriter.Flush()
}

func (c *Client) connHandler() {
	defer c.conn.Close()

	for {
		c.cmd = ""
		c.args = nil

		req, err := c.rReader.ParseRequest()
		if err != nil && err != io.EOF {
			c.logger.Println(err.Error())
			return
		} else if err != nil {
			return
		}
		err = c.handleRequest(req)
		if err != nil && err != io.EOF {
			c.logger.Println(err.Error())
			return
		}
	}
}

func (c *Client) handleRequest(req [][]byte) error {
	if len(req) == 0 {
		c.cmd = ""
		c.args = nil
	} else {
		c.cmd = strings.ToLower(string(req[0]))
		c.args = req[1:]
	}

	var (
		err error
		v   string
	)

	c.logger.Printf("process %s command", c.cmd)

	switch c.cmd {
	case "get":
		if v, err = c.handleGet(); err == nil {
			c.FlushResp(v)
		}
	case "set":
		if err = c.handleSet(); err == nil {
			c.FlushResp("OK")
		}
	case "del":
		if err = c.handleDel(); err == nil {
			c.FlushResp("OK")
		}
	case "join":
		if err = c.handleJoin(); err == nil {
			c.FlushResp("OK")
		}
	case "leave":
		if err = c.handleLeave(); err == nil {
			c.FlushResp("OK")
		}
	case "ping":
		if len(c.args) != 0 {
			err = ErrParams
		}
		c.FlushResp("PONG")
		err = nil
	case "snapshot":
		if err = c.handleSnapshot(); err == nil {
			c.FlushResp("OK")
		}

	default:
		err = ErrCmdNotSupport
	}
	if err != nil {
		c.FlushResp(err)
	}

	return err
}
