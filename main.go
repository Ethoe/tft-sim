package main

import (
	"fmt"
	"tft-sim/models"
	"tft-sim/sim"
	"tft-sim/sim/items"
	"time"
)

func main() {
	stage := 2

	// Create a test unit with base stats
	baseStats := map[models.StatType]float64{
		models.StatHealth:       1000,
		models.StatAttackDamage: 100,
		models.StatAbilityPower: 0,
		models.StatAttackSpeed:  1.0, // attacks per second
		models.StatArmor:        30,
		models.StatMagicResist:  30,
		models.StatMana:         50,
		models.StatCritChance:   0.00,
		models.StatCritDamage:   0,
	}

	newAbility := models.Ability{
		Name:         "Random Damage Ability",
		BaseDamage:   200,
		DamageType:   models.DamageTypePhysical,
		ManaCost:     5000000,
		CastTime:     300 * time.Millisecond,
		IsAoE:        false,
		IsAutoAttack: false,
	}

	newUnit := models.Unit{
		Name:         "Yunara",
		UnitRole:     models.RoleAttackMarksman,
		StarLevel:    1,
		CurrentMana:  0,
		AttackTimer:  0,
		AttackWindup: 20 * time.Millisecond,
	}

	unit := models.NewUnit(newUnit, newAbility, baseStats, stage)

	// Add items
	// if deathblade, exists := items.Get("Deathblade"); exists {
	// 	unit.AddItem(deathblade)
	// }

	if guinsoos, exists := items.Get("Guinsoos"); exists {
		unit.AddItem(guinsoos)
	}

	// Add augment
	combatTraining := models.Augments["Magic Wand"]
	unit.AddAugment(combatTraining)

	// Create targets
	targets := []*models.Target{
		models.NewTarget("Frontline Tank", 2000, 100, 50),
		//models.NewTarget("Backline Carry", 1200, 30, 30),
	}

	// Create and run simulation
	sim := sim.NewSimulator(unit, targets)
	results := sim.Run()

	// Print results
	fmt.Println("\n=== Simulation Results ===")
	fmt.Printf("Total Damage: %.1f\n", results.TotalDamage)
	fmt.Printf("DPS: %.1f\n", results.DPS)
	fmt.Printf("Simulation Duration: %.2fs\n", sim.Time.Seconds())

	fmt.Printf("Stats: %v\n", results.Stats)

	fmt.Println("\nDamage by Type:")
	for dmgType, amount := range results.DamageByType {
		fmt.Printf("  %v: %.1f (%.1f%%)\n", dmgType, amount, (amount/results.TotalDamage)*100)
	}

	fmt.Println("\nTarget Status:")
	for name, health := range results.FinalHealth {
		ttk := results.TimeToKill[name]
		if ttk > 0 {
			fmt.Printf("  %s: Killed at %.2fs\n", name, ttk.Seconds())
		} else {
			fmt.Printf("  %s: %.1f HP remaining\n", name, health)
		}
	}

	// Print damage timeline (first 10 events)
	fmt.Println("\nFirst 10 Damage Events:")
	for i, event := range results.DamageLog {
		if i >= 10 {
			break
		}
		eventType := "Auto"
		if event.IsAbility {
			eventType = "Ability"
		}
		fmt.Printf("  [%.2fs] %s %s on %s for %.1f damage\n",
			event.Timestamp.Seconds(), eventType, unit.Name, event.TargetName, event.Damage)
	}
}
