package models

type Item struct {
	Name             string
	Description      string
	Stats            map[StatType]float64
	OnHitEffect      func(*ItemInstance, *Target, float64)
	OnAttackEffect   func(*ItemInstance)
	OnAbilityCast    func(*ItemInstance)
	OnSecondEffect   func(*ItemInstance)
	OnEquipEffect    func(*ItemInstance, *[]ItemInstance)
	Unique           bool
	AllowAbilityCrit bool
	Stacking         bool
	MaxStacks        int
}

type ItemInstance struct {
	UniqueName string
	Item       Item
	Stacks     int
	Owner      *Unit
}
