package models

import "time"

type StatType int

const (
	StatHealth StatType = iota
	StatArmor
	StatMagicResist
	StatAttackDamage
	StatAbilityPower
	StatAttackSpeed
	StatCritChance
	StatCritDamage
	StatMana
	StatManaRegen
	StatVamp
	StatDamageReduction
	StatDamageAmp
)

const (
	AttackSpeedCap = 5.0
)

type Stats struct {
	Base        map[StatType]float64
	Bonus       map[StatType]float64
	Multipliers map[StatType]float64
	Unit        *Unit         // Reference to unit for buff calculations
	CurrentTime time.Duration // Current simulation time for buff calculations
}

func NewStats() Stats {
	return Stats{
		Base:        make(map[StatType]float64),
		Bonus:       make(map[StatType]float64),
		Multipliers: make(map[StatType]float64),
		Unit:        nil,
		CurrentTime: 0,
	}
}

// SetUnit sets the unit reference for buff calculations
func (s *Stats) SetUnit(unit *Unit) {
	s.Unit = unit
}

// SetCurrentTime sets the current simulation time for buff calculations
func (s *Stats) SetCurrentTime(currentTime time.Duration) {
	s.CurrentTime = currentTime
}

func (s *Stats) GetBonus(stat StatType) float64 {
	if s.Unit != nil && s.Unit.BuffManager != nil {
		buffBonuses, _ := s.Unit.BuffManager.GetBuffStats(s.CurrentTime)
		return s.Bonus[stat] + buffBonuses[stat]
	}

	return s.Bonus[stat]
}

func (s *Stats) Get(stat StatType) float64 {
	base := s.Base[stat]
	bonus := s.Bonus[stat]
	multiplier := s.Multipliers[stat]

	if multiplier == 0 {
		multiplier = 1.0
	}

	// Add buff bonuses if unit is available
	if s.Unit != nil && s.Unit.BuffManager != nil {
		buffBonuses, buffMultipliers := s.Unit.BuffManager.GetBuffStats(s.CurrentTime)
		bonus += buffBonuses[stat]
		multiplier += buffMultipliers[stat]
	}

	result := (base + bonus) * multiplier

	if stat == StatAttackSpeed || stat == StatAttackDamage {
		result = base * (1 + bonus)
		if stat == StatAttackSpeed && result > AttackSpeedCap {
			return AttackSpeedCap
		}
	}

	return result
}

func (s *Stats) SetBase(stat StatType, value float64) {
	s.Base[stat] = value
}

func (s *Stats) AddBonus(stat StatType, value float64) {
	s.Bonus[stat] += value
}

func (s *Stats) AddMultiplier(stat StatType, value float64) {
	s.Multipliers[stat] += value
}
