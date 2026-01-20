package models

import (
	"time"
)

// StackBehavior defines how buffs stack when reapplied
type StackBehavior int

const (
	StackBehaviorRefresh        StackBehavior = iota // Refresh duration, keep stacks
	StackBehaviorAdditive                            // Add new stack, refresh duration
	StackBehaviorMultiplicative                      // Multiply effects, refresh duration
	StackBehaviorIndependent                         // Create independent instance
)

// Buff represents a temporary effect on a unit
type Buff struct {
	Name        string
	Description string
	Duration    time.Duration
	AppliedTime time.Duration
	Source      interface{} // Unit, Item, or Ability that created it

	// Stat modifications
	StatBonuses     map[StatType]float64
	StatMultipliers map[StatType]float64

	// Behavioral modifications
	ModifiesAutoAttack bool
	AutoAttackOverride func(*Unit, *Target) (float64, bool) // Returns damage, isCrit

	// Callbacks
	OnApply   func(*Unit)
	OnTick    func(*Unit, time.Duration) // Called each tick while active
	OnExpire  func(*Unit)
	OnRefresh func(*Unit, *Buff) // When buff is refreshed/reapplied

	// Stacking
	MaxStacks     int
	CurrentStacks int
	StackBehavior StackBehavior
	IsExpired     bool
}

// NewBuff creates a new buff with default values
func NewBuff(name string, duration time.Duration) *Buff {
	return &Buff{
		Name:            name,
		Duration:        duration,
		StatBonuses:     make(map[StatType]float64),
		StatMultipliers: make(map[StatType]float64),
		MaxStacks:       1,
		CurrentStacks:   1,
		StackBehavior:   StackBehaviorRefresh,
		IsExpired:       false,
	}
}

// AddStatBonus adds a flat stat bonus to the buff
func (b *Buff) AddStatBonus(stat StatType, value float64) *Buff {
	b.StatBonuses[stat] = value
	return b
}

// AddStatMultiplier adds a stat multiplier to the buff
func (b *Buff) AddStatMultiplier(stat StatType, value float64) *Buff {
	b.StatMultipliers[stat] = value
	return b
}

// SetAutoAttackOverride sets a custom auto attack function
func (b *Buff) SetAutoAttackOverride(fn func(*Unit, *Target) (float64, bool)) *Buff {
	b.ModifiesAutoAttack = true
	b.AutoAttackOverride = fn
	return b
}

// SetStacking configures stacking behavior
func (b *Buff) SetStacking(maxStacks int, behavior StackBehavior) *Buff {
	b.MaxStacks = maxStacks
	b.StackBehavior = behavior
	return b
}

// SetCallbacks sets the buff's callback functions
func (b *Buff) SetCallbacks(onApply, onTick, onExpire, onRefresh func(*Unit)) *Buff {
	if onApply != nil {
		b.OnApply = onApply
	}
	if onTick != nil {
		b.OnTick = func(u *Unit, elapsed time.Duration) {
			onTick(u)
		}
	}
	if onExpire != nil {
		b.OnExpire = onExpire
	}
	if onRefresh != nil {
		b.OnRefresh = func(u *Unit, buff *Buff) {
			onRefresh(u)
		}
	}
	return b
}

// IsActive checks if the buff is still active at the given time
func (b *Buff) IsActive(currentTime time.Duration) bool {
	if b.IsExpired {
		return false
	}
	if b.Duration <= 0 {
		return true // Permanent buff
	}
	return currentTime < b.AppliedTime+b.Duration
}

// RemainingDuration calculates remaining duration at given time
func (b *Buff) RemainingDuration(currentTime time.Duration) time.Duration {
	if b.Duration <= 0 || b.IsExpired {
		return 0
	}
	remaining := b.AppliedTime + b.Duration - currentTime
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Refresh refreshes the buff duration and handles stacking
func (b *Buff) Refresh(currentTime time.Duration, newBuff *Buff) {
	b.AppliedTime = currentTime

	switch b.StackBehavior {
	case StackBehaviorAdditive:
		if b.CurrentStacks < b.MaxStacks {
			b.CurrentStacks++
			// Add new buff's stats to existing
			for stat, value := range newBuff.StatBonuses {
				b.StatBonuses[stat] += value
			}
			for stat, value := range newBuff.StatMultipliers {
				b.StatMultipliers[stat] += value
			}
		}
	case StackBehaviorMultiplicative:
		if b.CurrentStacks < b.MaxStacks {
			b.CurrentStacks++
			// Multiply existing stat multipliers
			for stat, value := range newBuff.StatMultipliers {
				if existing, ok := b.StatMultipliers[stat]; ok {
					b.StatMultipliers[stat] = existing * (1 + value)
				} else {
					b.StatMultipliers[stat] = 1 + value
				}
			}
		}
		// For Refresh and Independent, just update duration
	}

	if b.OnRefresh != nil {
		b.OnRefresh(newBuff.Source.(*Unit), b)
	}
}

// Expire marks the buff as expired and triggers OnExpire
func (b *Buff) Expire(unit *Unit) {
	b.IsExpired = true
	if b.OnExpire != nil {
		b.OnExpire(unit)
	}
}

// BuffManager manages active buffs on a unit
type BuffManager struct {
	Unit  *Unit
	Buffs []*Buff
}

// NewBuffManager creates a new buff manager for a unit
func NewBuffManager(unit *Unit) *BuffManager {
	return &BuffManager{
		Unit:  unit,
		Buffs: make([]*Buff, 0),
	}
}

// ApplyBuff applies a new buff to the unit
func (bm *BuffManager) ApplyBuff(buff *Buff, currentTime time.Duration) {
	// Check for existing buff with same name
	for _, existing := range bm.Buffs {
		if existing.Name == buff.Name && !existing.IsExpired {
			existing.Refresh(currentTime, buff)
			return
		}
	}

	// Apply new buff
	buff.AppliedTime = currentTime
	buff.Source = bm.Unit
	bm.Buffs = append(bm.Buffs, buff)

	if buff.OnApply != nil {
		buff.OnApply(bm.Unit)
	}
}

// RemoveBuff removes a buff by name
func (bm *BuffManager) RemoveBuff(name string) {
	for i, buff := range bm.Buffs {
		if buff.Name == name && !buff.IsExpired {
			buff.Expire(bm.Unit)
			// Mark for removal
			bm.Buffs[i].IsExpired = true
		}
	}
	bm.cleanupExpired()
}

// UpdateBuffs updates all buffs and removes expired ones
func (bm *BuffManager) UpdateBuffs(currentTime time.Duration) {
	for _, buff := range bm.Buffs {
		if buff.IsExpired {
			continue
		}

		if !buff.IsActive(currentTime) {
			buff.Expire(bm.Unit)
			continue
		}

		if buff.OnTick != nil {
			buff.OnTick(bm.Unit, currentTime-buff.AppliedTime)
		}
	}
	bm.cleanupExpired()
}

// GetActiveBuffs returns all active buffs
func (bm *BuffManager) GetActiveBuffs(currentTime time.Duration) []*Buff {
	active := make([]*Buff, 0)
	for _, buff := range bm.Buffs {
		if !buff.IsExpired && buff.IsActive(currentTime) {
			active = append(active, buff)
		}
	}
	return active
}

// HasBuff checks if unit has an active buff with given name
func (bm *BuffManager) HasBuff(name string, currentTime time.Duration) bool {
	for _, buff := range bm.Buffs {
		if buff.Name == name && !buff.IsExpired && buff.IsActive(currentTime) {
			return true
		}
	}
	return false
}

// GetBuffStats calculates total stat bonuses from all active buffs
func (bm *BuffManager) GetBuffStats(currentTime time.Duration) (map[StatType]float64, map[StatType]float64) {
	bonuses := make(map[StatType]float64)
	multipliers := make(map[StatType]float64)

	for _, buff := range bm.GetActiveBuffs(currentTime) {
		for stat, value := range buff.StatBonuses {
			bonuses[stat] += value
		}
		for stat, value := range buff.StatMultipliers {
			multipliers[stat] += value
		}
	}

	return bonuses, multipliers
}

// cleanupExpired removes expired buffs from the list
func (bm *BuffManager) cleanupExpired() {
	active := make([]*Buff, 0)
	for _, buff := range bm.Buffs {
		if !buff.IsExpired {
			active = append(active, buff)
		}
	}
	bm.Buffs = active
}

// Predefined buffs for common effects
var (
	// AttackSpeedBuff provides temporary attack speed
	AttackSpeedBuff = func(duration time.Duration, amount float64) *Buff {
		return NewBuff("Attack Speed Boost", duration).
			AddStatBonus(StatAttackSpeed, amount).
			SetCallbacks(
				nil, // onApply
				nil, // onTick
				nil, // onExpire
				nil, // onRefresh
			)
	}

	// DamageAmpBuff increases damage dealt
	DamageAmpBuff = func(duration time.Duration, amount float64) *Buff {
		return NewBuff("Damage Amplification", duration).
			AddStatMultiplier(StatDamageAmp, amount).
			SetCallbacks(
				nil, nil, nil, nil,
			)
	}

	// EmpoweredAutoBuff makes next auto attack deal bonus damage
	EmpoweredAutoBuff = func(duration time.Duration, bonusDamage float64) *Buff {
		var hasTriggered bool
		return NewBuff("Empowered Auto", duration).
			SetAutoAttackOverride(func(u *Unit, t *Target) (float64, bool) {
				if hasTriggered {
					// Use normal auto attack after first empowered one
					return CalculatePhysicalDamage(u, t, 0)
				}
				hasTriggered = true
				damage, isCrit := CalculatePhysicalDamage(u, t, bonusDamage)
				// Remove buff after use
				u.BuffManager.RemoveBuff("Empowered Auto")
				return damage, isCrit
			}).
			SetCallbacks(
				nil, nil, nil, nil,
			)
	}
)
