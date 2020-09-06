package server

import (
	"github.com/wangzeping722/manda/config"
	"log"
	"net"
	"os"
)

type App struct {
	listener net.Listener

	// TODO need a db

	log *log.Logger
}

func NewApp(cfg *config.Config) *App {
	var err error
	app := &App{}

	app.log = log.New(os.Stderr, "[server] ", log.LstdFlags)


}
