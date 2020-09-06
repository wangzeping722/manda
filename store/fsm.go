package store

import (
	"github.com/pkg/errors"
	"log"
	"os"
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
		db: db,
	}, nil
}

