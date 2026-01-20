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

func CalculatePhysicalDamage(attacker *Unit, target *Target, baseDamage float64) (float64, bool) {
	// Get attacker stats
	ad := attacker.Stats.Get(StatAttackDamage)
	critChance := attacker.Stats.Get(StatCritChance)
	critDamage := 1.0 + attacker.Stats.Get(StatCritDamage)

	totalDamage := baseDamage + ad

	// Apply crit
	isCrit := attacker.CritTracker.RollCrit(critChance)
	if isCrit {
		totalDamage *= critDamage
	}

	// Apply target resistances
	armor := target.Stats.Get(StatArmor)

	var damageReduction float64
	if armor >= 0 {
		damageReduction = armor / (100 + armor)
	} else {
		damageReduction = 2 - (armor / (100 - armor))
	}

	// Apply damage reduction
	totalDamage *= (1 - target.DamageReduction)

	// Final damage after armor
	finalDamage := totalDamage * (1 - damageReduction)

	return finalDamage, isCrit
}

func CalculateMagicDamage(attacker *Unit, target *Target, baseDamage float64, apRatio float64) (float64, bool) {
	// Get attacker stats
	ap := attacker.Stats.Get(StatAbilityPower)
	critChance := attacker.Stats.Get(StatCritChance)
	critDamage := 1.0 + attacker.Stats.Get(StatCritDamage)

	// Check if abilities can crit
	canAbilityCrit := false
	for _, item := range attacker.Items {
		if item.AllowAbilityCrit {
			canAbilityCrit = true
			break
		}
	}

	// Calculate base damage
	apDamage := ap * apRatio
	totalDamage := baseDamage + apDamage

	// Apply crit if allowed
	isCrit := false
	if canAbilityCrit && attacker.CritTracker.RollCrit(critChance) {
		totalDamage *= critDamage
		isCrit = true
	}

	// Apply target resistances
	mr := target.Stats.Get(StatMagicResist)

	var damageReduction float64
	if mr >= 0 {
		damageReduction = mr / (100 + mr)
	} else {
		damageReduction = 2 - (mr / (100 - mr))
	}

	// Apply damage reduction
	totalDamage *= (1 - target.DamageReduction)

	// Final damage after MR
	finalDamage := totalDamage * (1 - damageReduction)

	return finalDamage, isCrit
}
