package main

import (
	"fmt"
	"strings"
	"tft-sim/models"
	"tft-sim/sim"
	"tft-sim/sim/items"
	"tft-sim/sim/units"
)

func main() {
	// Get Yunara unit from registry (1-star)
	unit, exists := units.Get("Yunara", 1)
	if !exists {
		fmt.Println("Error: Yunara unit not found in registry")
		return
	}

	// Add items
	if titans, exists := items.Get("Titans"); exists {
		unit.AddItem(titans)
	}

	if guinsoos, exists := items.Get("Guinsoos"); exists {
		unit.AddItem(guinsoos)
	}

	if ie, exists := items.Get("IE"); exists {
		unit.AddItem(ie)
	}

	// Add augment
	// combatTraining := models.Augments["Magic Wand"]
	// unit.AddAugment(combatTraining)

	// Create targets - add multiple targets to test laser piercing
	targets := []*models.Target{
		models.NewTarget("Frontline Tank", 5000, 30, 50),
	}

	// Create and run simulation
	simulator := sim.NewSimulator(unit, targets)
	results := simulator.Run()

	// Print results
	fmt.Println("\n=== Simulation Results ===")
	fmt.Printf("Total Damage: %.1f\n", results.TotalDamage)
	fmt.Printf("DPS: %.1f\n", results.DPS)

	itemNames := make([]string, len(unit.Items))
	for i, item := range unit.Items {
		itemNames[i] = item.Name
	}
	fmt.Printf("%s %v Star - (%s)\n\n", unit.Name, unit.StarLevel, strings.Join(itemNames, ", "))

	fmt.Printf("Simulation Duration: %.2fs\n", simulator.Time.Seconds())

	// Print active buffs at the end
	fmt.Println("\nActive Buffs at Simulation End:")
	activeBuffs := unit.BuffManager.GetActiveBuffs(simulator.Time)
	if len(activeBuffs) == 0 {
		fmt.Println("  None")
	} else {
		for _, buff := range activeBuffs {
			remaining := buff.RemainingDuration(simulator.Time)
			fmt.Printf("  %s: %.1fs remaining\n",
				buff.Name, remaining.Seconds())
		}
	}

	fmt.Printf("\nStats: %v\n", results.Stats)

	fmt.Println("\nDamage by Type:")
	for dmgType, amount := range results.DamageByType {
		typeName := "Physical"
		switch dmgType {
		case models.DamageTypeMagic:
			typeName = "Magic"
		case models.DamageTypeTrue:
			typeName = "True"
		}
		fmt.Printf("  %s: %.1f (%.1f%%)\n", typeName, amount, (amount/results.TotalDamage)*100)
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

	// Print damage timeline (first 15 events to see laser attacks)
	fmt.Println("\nFirst 15 Damage Events:")
	for i, event := range results.DamageLog {
		if i >= 15 {
			break
		}
		eventType := "Auto"
		if event.IsAbility {
			eventType = "Ability"
		}
		damageType := "Physical"
		switch event.DamageType {
		case models.DamageTypeMagic:
			damageType = "Magic"
		case models.DamageTypeTrue:
			damageType = "True"
		}
		critStr := ""
		if event.IsCrit {
			critStr = " CRIT!"
		}
		fmt.Printf("  [%.2fs] %s %s on %s for %.1f %s damage%s\n",
			event.Timestamp.Seconds(), eventType, unit.Name, event.TargetName, event.Damage, damageType, critStr)
	}

	// Print crit stats
	fmt.Printf("\nCrit Rate: %.1f%% (%d/%d)\n",
		results.CritRate*100, unit.CritTracker.TotalCrits, unit.CritTracker.TotalAttacks)
}
