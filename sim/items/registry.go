package items

import (
	"sync"
	"tft-sim/models"
)

var (
	registry = make(map[string]models.Item)
	mu       sync.RWMutex
)

func Register(item models.Item) {
	mu.Lock()
	defer mu.Unlock()
	registry[item.Name] = item
}

func Get(name string) (models.Item, bool) {
	mu.RLock()
	defer mu.RUnlock()
	item, exists := registry[name]
	return item, exists
}

func GetAll() map[string]models.Item {
	mu.RLock()
	defer mu.RUnlock()

	// Return a copy
	copy := make(map[string]models.Item)
	for k, v := range registry {
		copy[k] = v
	}
	return copy
}
