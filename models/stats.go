package models

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
}

func NewStats() Stats {
	return Stats{
		Base:        make(map[StatType]float64),
		Bonus:       make(map[StatType]float64),
		Multipliers: make(map[StatType]float64),
	}
}

func (s *Stats) Get(stat StatType) float64 {
	base := s.Base[stat]
	bonus := s.Bonus[stat]
	multiplier := s.Multipliers[stat]

	if multiplier == 0 {
		multiplier = 1.0
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
