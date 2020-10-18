package ytcompare

import (
	"bytes"
	"compress/gzip"
	"sync"
)

//Store instance
type Store struct {
	Items map[int32][][]byte
	locks []sync.RWMutex
}

//NewStore create a new shards store
func NewStore() *Store {
	locks := make([]sync.RWMutex, 0)
	for i := uint64(0); i < 1000; i++ {
		locks = append(locks, sync.RWMutex{})
	}
	return &Store{Items: make(map[int32][][]byte, 0), locks: locks}
}

//Add add a new shard
func (store *Store) Add(nodeID int32, shard []byte) {
	lock := store.locks[int(nodeID)%len(store.locks)]
	lock.Lock()
	defer lock.Unlock()
	store.Items[nodeID] = append(store.Items[nodeID], shard)
}

//Clear clear items
func (store *Store) Clear() {
	store.Items = make(map[int32][][]byte)
}

//GenerateData compressed compare data for one miner
func (store *Store) GenerateData(nodeID int32) (bytes.Buffer, error) {
	var res bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&res, 7)
	for _, b := range store.Items[nodeID] {
		_, err := gz.Write(b)
		if err != nil {
			return res, err
		}
	}
	gz.Close()
	return res, nil
}
