package store

import (
	"github.com/dgraph-io/badger/v2"
	"log"
	"os"
)

//
type BadgerDB struct {
	path   string
	db     *badger.DB
	logger *log.Logger
}

type KVItem struct {
	key   []byte
	value []byte
	err   error
}

func (i *KVItem) IsFinished() bool {
	return i.err == ErrIterFinished
}

func NewBadgerDB(path string) (*BadgerDB, error) {
	opts := badger.DefaultOptions(path)
	opts.SyncWrites = false
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerDB{
		path:   path,
		db:     db,
		logger: log.New(os.Stderr, "[db_badger] ", log.LstdFlags),
	}, nil
}

func (b BadgerDB) Get(key []byte) ([]byte, error) {
	var value []byte

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		value, err = item.ValueCopy(nil)

		return err
	})

	if err != nil {
		return nil, err
	}
	return value, nil
}

func (b BadgerDB) Set(key, value []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (b BadgerDB) Delete(key []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (b BadgerDB) SnapshotItems() <-chan DataItem {
	ch := make(chan DataItem, 1024)

	go b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		keyCount := 0
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()

			v, err := item.ValueCopy(nil)
			kvItem := &KVItem{
				key:   k,
				value: v,
				err:   err,
			}
			ch <- kvItem
			keyCount++
			if err != nil {
				b.logger.Printf("failed to save key:%s", string(k))
				return err
			}
		}

		kvi := &KVItem{err: ErrIterFinished}
		ch <- kvi
		b.logger.Printf("Sanpshot total %d keys", keyCount)
		return nil
	})
	return ch
}

func (b BadgerDB) Close() {
	b.db.Close()
}
