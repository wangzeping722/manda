package store

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/pkg/errors"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Store struct {
	// 存储数据文件的目录
	RaftDir string
	// raft 监听的端口号
	RaftBind string

	// raft 实例
	raft *raft.Raft

	fsm *fsm

	logger *log.Logger
}

func NewStore(raftDir, raftBind string) (*Store, error) {
	fsm, err := NewFSM(raftDir)
	if err != nil {
		return nil, err
	}

	return &Store{
		logger:   log.New(os.Stderr, "[store] ", log.LstdFlags),
		fsm:      fsm,
		RaftDir:  raftDir,
		RaftBind: raftBind,
	},nil
}

func (s *Store) Open(boostrap bool, localId string) error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localId)
	config.SnapshotThreshold = 1024
	addr, err := net.ResolveTCPAddr("tcp", s.RaftBind)
	if err != nil {
		return nil
	}

	transport, err := raft.NewTCPTransport(s.RaftBind, addr, 3, 10 * time.Second, os.Stderr)
	if err != nil {
		return err
	}

	ss, err := raft.NewFileSnapshotStore(s.RaftDir, 2, os.Stderr)
	if err != nil {
		return  err
	}

	boltDB,err := raftboltdb.NewBoltStore(filepath.Join(s.RaftDir, "raft.db"))
	if err != nil {
		return err
	}

	r, err := raft.NewRaft(config, s.fsm, boltDB, boltDB, ss, transport)
	if err != nil {
		return err
	}

	s.raft = r
	if boostrap {
		configuration := raft.Configuration{Servers: []raft.Server{
			{ID: config.LocalID, Address: transport.LocalAddr()},
		}}
		s.raft.BootstrapCluster(configuration)
	}

	return nil
}

var ErrNotLeader error = errors.New("not leader")

func (s *Store) Get(key string) (string, error) {
	return s.fsm.Get(key)
}

func (s *Store) Set(key, value string) error {
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	c := NewSetCommand(key, value)
	msg, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(msg, 10*time.Second)

	return f.Error()
}

func (s *Store) Delete(key string) error {
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	c := NewDeleteCommand(key)

	msg, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(msg, 10*time.Second)

	return f.Error()
}

func (s *Store) Join(nodeId, addr string) error {
	s.logger.Printf("received join request for remote node %s, addr %s", nodeId, addr)
	cf := s.raft.GetConfiguration()
	if err := cf.Error(); err != nil {
		s.logger.Printf("failed to get raft configuration")
		return err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(nodeId) {
			s.logger.Printf("node %s already joined raft cluster", nodeId)
			return nil
		}
	}

	f := s.raft.AddVoter(raft.ServerID(nodeId),raft.ServerAddress(addr), 0, 0)
	if err := f.Error(); err != nil {
		return err
	}

	s.logger.Printf("node %s at %s joined successfully", nodeId, addr)

	return nil
}

func (s *Store) Leave(nodeID string) error {
	s.logger.Printf("received leave request for remote node %s", nodeID)

	cf := s.raft.GetConfiguration()

	if err := cf.Error(); err != nil {
		s.logger.Printf("failed to get raft configuration")
		return err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(nodeID) {
			f := s.raft.RemoveServer(server.ID, 0, 0)
			if err := f.Error(); err != nil {
				s.logger.Printf("failed to remove server %s", nodeID)
				return err
			}

			s.logger.Printf("node %s leaved successfully", nodeID)
			return nil
		}
	}

	s.logger.Printf("node %s not exists in raft group", nodeID)

	return nil
}

func (s *Store) Snapshot() error {
	s.logger.Printf("doing snapshot mannually")
	f := s.raft.Snapshot()
	return f.Error()
}