package models

type Target struct {
	Name            string
	Stats           Stats
	DamageReduction float64 // Percentage (0-1)
	CurrentHP       float64
	MaxHP           float64
}

func NewTarget(name string, hp, armor, mr float64) *Target {
	t := &Target{
		Name:            name,
		Stats:           NewStats(),
		CurrentHP:       hp,
		MaxHP:           hp,
		DamageReduction: 0,
	}

	t.Stats.SetBase(StatHealth, hp)
	t.Stats.SetBase(StatArmor, armor)
	t.Stats.SetBase(StatMagicResist, mr)

	return t
}

func (t *Target) TakeDamage(damage float64, damageType DamageType) (float64, bool) {
	// Apply damage reduction
	damage = damage * (1 - t.DamageReduction)

	// Apply resistances
	var resistance float64
	switch damageType {
	case DamageTypePhysical:
		resistance = t.Stats.Get(StatArmor)
	case DamageTypeMagic:
		resistance = t.Stats.Get(StatMagicResist)
	default:
		resistance = 0
	}

	// TFT damage reduction formula (simplified)
	var damageReduction float64
	if resistance >= 0 {
		damageReduction = resistance / (100 + resistance)
	} else {
		damageReduction = 2 - (resistance / (100 - resistance))
	}
	actualDamage := damage * (1 - damageReduction)

	t.CurrentHP -= actualDamage
	isDead := t.CurrentHP <= 0

	if isDead {
		t.CurrentHP = 0
	}

	return actualDamage, isDead
}
