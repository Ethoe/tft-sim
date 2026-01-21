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

func (t *Target) TakeDamage(damage float64, damageType DamageType) float64 {
	t.CurrentHP -= damage

	if t.IsDead() {
		t.CurrentHP = 0
	}

	return damage
}

func (t *Target) IsDead() bool {
	return t.CurrentHP <= 0
}
