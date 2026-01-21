package models

type Item struct {
	Name             string
	Description      string
	Stats            map[StatType]float64
	OnHitEffect      func(*Unit, *Target, float64)
	OnAttackEffect   func(*Unit)
	OnAbilityCast    func(*Unit)
	OnSecondEffect   func(*Unit)
	OnEquipEffect    func(*Unit, *[]Item)
	Unique           bool
	AllowAbilityCrit bool
	Stacking         bool
	MaxStacks        int
}

type ItemInstance struct {
	Item   Item
	Stacks int
	Owner  *Unit
}
