package store

import "github.com/hashicorp/raft"

type Store struct {
	// 存储数据文件的目录
	RaftDir string
	// raft 监听的端口号
	RaftBind string

	// raft 实例
	raft *raft.Raft

	fsm *raft.FSM


}
