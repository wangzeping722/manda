package store

// DB 实例, 定义了操作数据的接口
type DB interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	Delete(key []byte) error
	SnapshotItems() <-chan DataItem

	Close()
}

type DataItem interface {}
