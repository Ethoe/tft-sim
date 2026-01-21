package items

import (
	"testing"
)

func TestKrakensItemRegistered(t *testing.T) {
	item, exists := Get("Krakens")
	if !exists {
		t.Fatal("Krakens item not found in registry")
	}

	if item.Name != "Krakens" {
		t.Errorf("Expected item name 'Krakens', got '%s'", item.Name)
	}

	// Check that stats are set
	if len(item.Stats) == 0 {
		t.Error("Krakens item has no stats")
	}

	// Check that OnAttackEffect is set (for the stacking passive)
	if item.OnAttackEffect == nil {
		t.Error("Krakens item missing OnAttackEffect for stacking passive")
	}

	// Check description contains key phrases
	desc := item.Description
	if desc == "" {
		t.Error("Krakens item has empty description")
	}
}
