package store

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	pb "github.com/wangzeping722/manda/proto"
	"log"
	"os"
)

type fsmSnapshot struct {
	db       DB
	keyCount int
	logger   *log.Logger
}

func newFsmSnapshot(db DB) *fsmSnapshot {
	return &fsmSnapshot{
		db: db,
		logger: log.New(os.Stderr, "[fsmSnampshot] ", log.LstdFlags),
	}
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	f.logger.Printf("Start snapshot items")
	defer sink.Close()

	ch := f.db.SnapshotItems()
	for {
		buff := proto.NewBuffer([]byte{})

		data := <-ch
		item := data.(*KVItem)
		if item.IsFinished() {
			break
		}

		pbKvItem := &pb.KVItem{
			Key:   item.key,
			Value: item.value,
		}

		err := buff.EncodeMessage(pbKvItem)
		if err != nil {
			f.logger.Printf("failed to snapshot: %s", err.Error())
			return sink.Cancel()
		}

		if _, err := sink.Write(buff.Bytes()); err != nil {
			return err
		}

		f.keyCount++
	}

	return nil
}

func (f *fsmSnapshot) Release() {
	f.logger.Printf("Persist total %d keys", f.keyCount)
}
