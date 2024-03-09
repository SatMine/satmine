// filePath:  glo/store.go
package store

import (
	"satmine/satmine"
	"sync"
)

// Store is a struct for storing global data
type Store struct {
	// Define your global data here
	OrdIdx *satmine.BTOrdIdx
	// Simulated blocks are used only for testing purposes.
	MockBlock *satmine.HookBlock
}

// instance is the private instance of Store
var instance *Store

// once ensures the singleton instance is created only once
var once sync.Once

// Instance returns the single instance of Store
// sync.Once is used to ensure the instance is created only once
func Instance() *Store {
	once.Do(func() {
		instance = &Store{}
	})
	return instance
}

// Init initializes the data of the Store
// This method can be called from outside to set initial values
func (s *Store) Init(OrdIdx *satmine.BTOrdIdx) {
	s.OrdIdx = OrdIdx
}
