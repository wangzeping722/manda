package config

type Config struct {
	Desc       string
	NodeId     string
	Listen     string
	RaftListen string
	RaftDir    string
	Join       string
}

func NewConfig(listen, raftDir, raftListen, nodeId, join string) *Config {
	return &Config{
		NodeId:     nodeId,
		Listen:     listen,
		RaftListen: raftListen,
		RaftDir:    raftDir,
		Join:       join,
	}
}
