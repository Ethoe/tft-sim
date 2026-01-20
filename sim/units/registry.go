package units

import (
	"sync"
	"tft-sim/models"
)

// UnitFactory is a function that creates a new unit with the given star level
type UnitFactory func(starLevel int) *models.Unit

var (
	registry = make(map[string]UnitFactory)
	mu       sync.RWMutex
)

// Register registers a unit factory for the given unit name
func Register(name string, factory UnitFactory) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = factory
}

// Get creates a new unit instance with the given name and star level
func Get(name string, starLevel int) (*models.Unit, bool) {
	mu.RLock()
	defer mu.RUnlock()

	factory, exists := registry[name]
	if !exists {
		return nil, false
	}

	return factory(starLevel), true
}

// GetAll returns a copy of all registered unit names
func GetAll() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
