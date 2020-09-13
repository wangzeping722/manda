package store

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	pb "github.com/wangzeping722/manda/proto"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type fsm struct {
	db DB

	logger *log.Logger
}

func NewFSM(path string) (*fsm, error) {
	db, err := NewBadgerDB(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init fsm")
	}

	return &fsm{
		logger: log.New(os.Stderr, "[fsm] ", log.LstdFlags),
		db:     db,
	}, nil
}

func (f *fsm) Get(key string) (string, error) {
	v, err := f.db.Get([]byte(key))
	if err != nil {
		return "", err
	}

	return string(v), nil
}

func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic("failed to unmarshal raft log")
	}

	switch strings.ToLower(c.Op) {
	case Set:
		return f.applySet(c.Key, c.Value)
	case Delete:
		return f.applyDelete(c.Key)
	default:
		f.logger.Printf("unkonw command type")
		return nil
	}
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return newFsmSnapshot(f.db), nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	f.logger.Printf("Restore snapshot from FSMSnapshot")
	defer rc.Close()

	readBuf, err := ioutil.ReadAll(rc)
	if err != nil {
		f.logger.Printf("failed restore")
		return err
	}

	protoBuf := proto.NewBuffer(readBuf)
	for {
		item := pb.KVItem{}
		err := protoBuf.DecodeMessage(&item)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				break
			}

			f.logger.Printf("DecodeMessage failed %s", err.Error())
			return err
		}
	}

	f.logger.Print("restore success")
	return nil
}

func (f *fsm) applySet(key, value string) interface{} {
	return f.db.Set([]byte(key), []byte(value))
}

func (f *fsm) applyDelete(key string) interface{} {
	return f.db.Delete([]byte(key))
}

func (f *fsm) Close() error {
	f.db.Close()
	return nil
}
