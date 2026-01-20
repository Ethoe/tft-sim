package models

import (
	"time"
)

type DamageType int

const (
	DamageTypePhysical DamageType = iota
	DamageTypeMagic
	DamageTypeTrue
)

type Role int

const (
	RoleAttackTank Role = iota
	RoleAttackFighter
	RoleAttackMarksman
	RoleAttackCaster
	RoleAttackAssassin
	RoleAttackSpecialist
	RoleHybridFighter
	RoleMagicTank
	RoleMagicFighter
	RoleMagicMarksman
	RoleMagicCaster
	RoleMagicAssassin
	RoleMagicSpecialist
)

type UnitState int

const (
	UnitStateIdle UnitState = iota
	UnitStateAttacking
	UnitStateCasting
	UnitStateChanneling
)

type CastingContext struct {
	Ability       *Ability
	StartTime     time.Duration
	EndTime       time.Duration
	CanGainMana   bool
	CanAutoAttack bool
	Targets       []*Target
	IsAbilityCast bool
}

type Ability struct {
	Name                        string
	BaseDamage                  float64
	ADRatio                     float64
	APRatio                     float64
	DamageType                  DamageType
	CastTime                    time.Duration
	IsAoE                       bool
	IsAutoAttack                bool
	IsAutoAttackModifier        bool
	AllowsManaGainDuringCast    bool
	AllowsAutoAttacksDuringCast bool
	OnCast                      func(*Unit, []*Target)
	OnCastStart                 func(*Unit)
	OnCastComplete              func(*Unit, []*Target)
	CanAbilityCrit              bool
}

type Unit struct {
	Name        string
	Stats       Stats
	UnitRole    Role
	StarLevel   int
	CurrentMana float64

	// State
	State      UnitState
	CastingCtx *CastingContext

	// Attack mechanics
	AttackTimer    time.Duration
	AttackWindup   time.Duration
	NextAttackTime time.Duration

	// Ability
	Ability Ability

	// Items and augments
	Items    []Item
	Augments []Augment

	// Combat tracking
	TotalDamage  float64
	DamageLog    []DamageEvent
	AttackCount  int
	AbilityCount int
	CritTracker  *CritTracker
}
type DamageEvent struct {
	Timestamp  time.Duration
	Damage     float64
	DamageType DamageType
	IsAbility  bool
	TargetName string
	IsCrit     bool
}

func NewUnit(newUnit Unit, newAbility Ability, baseStats map[StatType]float64, stage int) *Unit {
	unit := &Unit{
		Name:           newUnit.Name,
		Stats:          NewStats(),
		UnitRole:       newUnit.UnitRole,
		StarLevel:      newUnit.StarLevel,
		CurrentMana:    newUnit.CurrentMana,
		AttackTimer:    0,
		AttackWindup:   200 * time.Millisecond,
		DamageLog:      make([]DamageEvent, 0),
		CritTracker:    NewCritTracker(),
		NextAttackTime: 0,
	}

	// Set base stats
	for stat, value := range baseStats {
		unit.Stats.SetBase(stat, value)
	}

	switch newUnit.UnitRole {
	case RoleAttackCaster:
	case RoleMagicCaster:
		unit.Stats.SetBase(StatManaRegen, 2)
	}

	unit.Ability = newAbility

	return unit
}

func (u *Unit) GetAttackDamage() float64 {
	baseAD := u.Stats.Get(StatAttackDamage)

	// Apply crit
	critChance := u.Stats.Get(StatCritChance)
	critDamage := u.Stats.Get(StatCritDamage)

	// Simplified crit calculation
	isCrit := critChance > 0 // In real implementation, use RNG
	if isCrit {
		return baseAD * (1 + critDamage)
	}

	return baseAD
}

func (u *Unit) CanAutoAttack(currentTime time.Duration) bool {
	if u.State == UnitStateCasting && !u.CastingCtx.CanAutoAttack {
		return false
	}

	if u.State == UnitStateChanneling {
		return false
	}

	return currentTime >= u.NextAttackTime
}

func (u *Unit) GetAttackSpeed() float64 {
	baseAS := u.Stats.Get(StatAttackSpeed)
	if baseAS >= AttackSpeedCap {
		return AttackSpeedCap
	}
	return baseAS
}

func (u *Unit) CanCastAbility() bool {
	if u.State != UnitStateIdle && u.State != UnitStateAttacking {
		return false
	}

	if u.CurrentMana < u.Stats.Get(StatMana) {
		return false
	}

	return true
}

func (u *Unit) StartCastingAbility(currentTime time.Duration, targets []*Target) {
	u.State = UnitStateCasting
	u.CastingCtx = &CastingContext{
		Ability:       &u.Ability,
		StartTime:     currentTime,
		EndTime:       currentTime + u.Ability.CastTime,
		CanGainMana:   u.Ability.AllowsManaGainDuringCast,
		CanAutoAttack: u.Ability.AllowsAutoAttacksDuringCast,
		Targets:       targets,
		IsAbilityCast: !u.Ability.IsAutoAttack,
	}

	// Spend mana immediately
	if u.Stats.Get(StatMana) > 0 {
		u.CurrentMana -= u.Stats.Get(StatMana)
	}

	// Trigger on-cast-start effects
	if u.Ability.OnCastStart != nil {
		u.Ability.OnCastStart(u)
	}
}

func (u *Unit) CompleteCast(currentTime time.Duration) {
	if u.CastingCtx == nil {
		return
	}

	// Cast the ability
	if u.Ability.OnCast != nil {
		u.Ability.OnCast(u, u.CastingCtx.Targets)
	}

	// Trigger on-cast-complete effects
	if u.Ability.OnCastComplete != nil {
		u.Ability.OnCastComplete(u, u.CastingCtx.Targets)
	}

	// Reset state
	u.State = UnitStateIdle
	u.CastingCtx = nil
}

func (u *Unit) GainMana(fromAutoAttack bool, fromAttack float64) {
	if u.State == UnitStateCasting && u.CastingCtx != nil && !u.CastingCtx.CanGainMana {
		return
	}

	if fromAutoAttack {
		switch u.UnitRole {
		case RoleAttackTank:
		case RoleMagicTank:
			u.CurrentMana += 5
		case RoleAttackCaster:
		case RoleMagicCaster:
			u.CurrentMana += 7
		default:
			u.CurrentMana += 10
		}
	} else {
		if u.UnitRole == RoleAttackTank || u.UnitRole == RoleMagicTank {
			// I don't know how taking damage mana works
			println(fromAttack)
			u.CurrentMana += 5
		}
	}

	maxMana := u.Stats.Get(StatMana)
	if u.CurrentMana > maxMana {
		u.CurrentMana = maxMana
	}
}

func (u *Unit) GetAttackInterval() time.Duration {
	// Attack speed formula: attacks per second = baseAS
	// Interval in milliseconds = 1000 / attacks per second
	as := u.GetAttackSpeed()
	if as <= 0 {
		return 1 * time.Second // Default to 1 attack per second
	}

	intervalMs := 1000.0 / as
	return time.Duration(intervalMs) * time.Millisecond
}

func (u *Unit) AddItem(item Item) {
	u.Items = append(u.Items, item)
	// Apply item stats
	for stat, value := range item.Stats {
		u.Stats.AddBonus(stat, value)
	}

	if item.AllowAbilityCrit {
		if u.Ability.CanAbilityCrit {
			u.Stats.AddBonus(StatCritDamage, .1)
		}
		u.Ability.CanAbilityCrit = item.AllowAbilityCrit
	}

	if item.OnEquipEffect != nil {
		//item.OnEquipEffect(u, &u.Items)
	}
}

func (u *Unit) AddAugment(augment Augment) {
	u.Augments = append(u.Augments, augment)
	// Apply augment stats
	for stat, value := range augment.Stats {
		u.Stats.AddBonus(stat, value)
	}
}
