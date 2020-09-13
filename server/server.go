package server

import (
	"fmt"
	"github.com/wangzeping722/manda/config"
	"github.com/wangzeping722/manda/store"
	"github.com/yongman/go/goredis"
	"log"
	"net"
	"os"
)

type App struct {
	listener net.Listener
	store    *store.Store
	log      *log.Logger
}

func NewApp(cfg *config.Config) *App {
	var err error
	app := &App{}

	app.log = log.New(os.Stderr, "[server] ", log.LstdFlags)
	app.store, err = store.NewStore(cfg.RaftDir, cfg.RaftListen)
	if err != nil {
		app.log.Println(err.Error())
		panic(err.Error())
	}

	bootstrap := cfg.Join == ""
	err = app.store.Open(bootstrap, cfg.NodeId)
	if err != nil {
		app.log.Println(err.Error())
		panic(err.Error())
	}

	if !bootstrap {
		// send join request to node already exists
		rc := goredis.NewClient(cfg.Join, "")
		app.log.Printf("join request send to %s", cfg.Join)
		_, err := rc.Do("join", cfg.RaftListen, cfg.NodeId)
		if err != nil {
			app.log.Println(err)
		}
		rc.Close()
	}

	app.listener, err = net.Listen("tcp", cfg.Listen)
	app.log.Printf("server listen in %s", cfg.Listen)
	if err != nil {
		fmt.Println(err.Error())
	}

	return app
}

func (app *App) Run() {
	for {
		// accept new client connect and perform
		conn, err := app.listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		// handle conn
		ClientHandler(conn, app)
	}
}
