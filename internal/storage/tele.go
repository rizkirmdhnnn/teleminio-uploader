package storage

import (
	"os"
	"path/filepath"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"
	boltstor "github.com/gotd/contrib/bbolt"
	"github.com/gotd/contrib/pebble"
	"github.com/gotd/td/telegram"
	"go.etcd.io/bbolt"
)

// Setup initializes all storage components
type Setup struct {
	SessionDir     string
	SessionStorage *telegram.FileSessionStorage
	PeerDB         *pebble.PeerStorage
	StateStorage   *boltstor.State
}

// NewStorage sets up all storage components
func NewStorage(phone string) (*Setup, error) {
	// Setting up session storage directory
	sessionDir := "session"
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return nil, err
	}

	// Session storage for auth data
	sessionStorage := &telegram.FileSessionStorage{
		Path: filepath.Join(sessionDir, "session.json"),
	}

	// Peer storage for caching and updates
	db, err := pebbledb.Open(filepath.Join(sessionDir, "peers.pebble.db"), &pebbledb.Options{})
	if err != nil {
		return nil, errors.Wrap(err, "create pebble storage")
	}
	peerDB := pebble.NewPeerStorage(db)

	// State storage for updates recovery
	boltdb, err := bbolt.Open(filepath.Join(sessionDir, "updates.bolt.db"), 0666, nil)
	if err != nil {
		return nil, errors.Wrap(err, "create bolt storage")
	}
	stateStorage := boltstor.NewStateStorage(boltdb)

	return &Setup{
		SessionDir:     sessionDir,
		SessionStorage: sessionStorage,
		PeerDB:         peerDB,
		StateStorage:   stateStorage,
	}, nil
}
