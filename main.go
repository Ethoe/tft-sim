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
		models.StatHealth:       1440,
		models.StatAttackDamage: 90,
		models.StatAbilityPower: 0,
		models.StatAttackSpeed:  .8, // attacks per second
		models.StatArmor:        30,
		models.StatMagicResist:  30,
		models.StatMana:         50, // Reduced mana cost for testing
		models.StatCritChance:   0.25,
		models.StatCritDamage:   .4,
	}

	newAbility := models.Ability{
		Name:                 "Transcendent State",
		BaseDamage:           85,
		DamageType:           models.DamageTypePhysical,
		CastTime:             4 * time.Second,
		IsAoE:                true,
		IsAutoAttackModifier: true,
		OnCastComplete: func(u *models.Unit, targets []*models.Target) {
			// Apply an attack speed buff when ability completes
			attackSpeedBuff := models.NewBuff("Transcendent Haste", 6*time.Second).
				AddStatBonus(models.StatAttackSpeed, 0.5). // +50% attack speed
				SetCallbacks(
					func(u *models.Unit) {
						fmt.Printf("[Buff Applied] %s gains Transcendent Haste (+50%% attack speed)\n", u.Name)
					},
					nil,
					func(u *models.Unit) {
						fmt.Printf("[Buff Expired] %s loses Transcendent Haste\n", u.Name)
					},
					nil,
				)
			// Use current time from stats (set by simulator)
			u.BuffManager.ApplyBuff(attackSpeedBuff, u.Stats.CurrentTime)
		},
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
	if deathblade, exists := items.Get("Deathblade"); exists {
		unit.AddItem(deathblade)
	}

	if guinsoos, exists := items.Get("Guinsoos"); exists {
		unit.AddItem(guinsoos)
	}

	if jg, exists := items.Get("JG"); exists {
		unit.AddItem(jg)
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

	// Print active buffs at the end
	fmt.Println("\nActive Buffs at Simulation End:")
	activeBuffs := unit.BuffManager.GetActiveBuffs(sim.Time)
	if len(activeBuffs) == 0 {
		fmt.Println("  None")
	} else {
		for _, buff := range activeBuffs {
			remaining := buff.RemainingDuration(sim.Time)
			fmt.Printf("  %s: %.1fs remaining (Stacks: %d)\n",
				buff.Name, remaining.Seconds(), buff.CurrentStacks)
		}
	}

	fmt.Printf("\nStats: %v\n", results.Stats)

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
