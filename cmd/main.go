package cmd

import (
	"flag"
	"github.com/wangzeping722/manda/config"
	"github.com/wangzeping722/manda/server"
	"os"
	"os/signal"
	"syscall"
)

var (
	listen     string
	raftDir    string
	raftListen string
	nodeId     string
	join       string
)

func init() {
	flag.StringVar(&listen, "listen", ":5379", "server listen port")
	flag.StringVar(&raftDir, "raftDir", "./data", "raft data directory")
	flag.StringVar(&raftListen, "raftlisten", ":15379", "raft bus transport bind address")
	flag.StringVar(&nodeId, "nodeid", "", "")
	flag.StringVar(&join, "join", "", "join to already exist cluster")
}

func main() {
	flag.Parse()

	c := config.NewConfig(listen, raftDir, raftListen, nodeId, join)
	app := server.NewApp(c)

	quitCh := make(chan os.Signal, 0)

	signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quitCh
}
