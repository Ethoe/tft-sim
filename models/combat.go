package models

import (
	"math/rand"
	"time"
)

type CritTracker struct {
	TotalCrits   int
	TotalAttacks int
	CritStreak   int
	LastCritTime time.Duration
	RNG          *rand.Rand
}

func NewCritTracker() *CritTracker {
	return &CritTracker{
		RNG:          rand.New(rand.NewSource(time.Now().UnixNano())),
		TotalCrits:   0,
		TotalAttacks: 0,
		CritStreak:   0,
	}
}

func (ct *CritTracker) RollCrit(critChance float64) bool {
	ct.TotalAttacks++
	if critChance >= 1.0 || (critChance > 0 && ct.RNG.Float64() < critChance) {
		ct.TotalCrits++
		ct.CritStreak++
		return true
	}
	ct.CritStreak = 0
	return false
}

func CalculateCritDamage(baseDamage, critDamageMultiplier float64, isAbility bool, hasJeweledGauntlet bool) float64 {
	if !isAbility || hasJeweledGauntlet {
		return baseDamage * critDamageMultiplier
	}
	return baseDamage
}

func CalculateDamage(attacker *Unit, target *Target, resistance float64, baseDamage float64, canCrit bool) (float64, bool) {
	// Get attacker stats
	critChance := attacker.Stats.Get(StatCritChance)
	critDamage := 1.0 + attacker.Stats.Get(StatCritDamage)

	totalDamage := baseDamage

	// Apply crit
	isCrit := attacker.CritTracker.RollCrit(critChance)
	if isCrit && canCrit {
		totalDamage *= critDamage
	}

	// Apply target resistances
	var damageReduction float64
	if resistance >= 0 {
		damageReduction = resistance / (100 + resistance)
	} else {
		damageReduction = 2 - (resistance / (100 - resistance))
	}

	// Apply damage reduction
	totalDamage *= (1 - target.DamageReduction)

	// Final damage after armor
	finalDamage := totalDamage * (1 - damageReduction)

	return finalDamage, isCrit
}

// CalculateTrueDamage calculates true damage which ignores all resistances
func CalculateTrueDamage(attacker *Unit, target *Target, baseDamage float64, canCrit bool) (float64, bool) {
	// Get attacker stats
	critChance := attacker.Stats.Get(StatCritChance)
	critDamage := 1.0 + attacker.Stats.Get(StatCritDamage)

	// Calculate base damage
	totalDamage := baseDamage

	// Apply crit (true damage can crit)
	isCrit := attacker.CritTracker.RollCrit(critChance)
	if isCrit && canCrit {
		totalDamage *= critDamage
	}

	// True damage ignores armor, MR, and damage reduction
	// No resistance calculations needed

	return totalDamage, isCrit
}
