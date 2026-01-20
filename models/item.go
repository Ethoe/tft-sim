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

// Predefined items
var Items = map[string]Item{
	"Rabadons": {
		Name:        "Rabadon's Deathcap",
		Description: "+80 Ability Power",
		Stats: map[StatType]float64{
			StatAbilityPower: 80,
		},
	},
	"Guinsoos": {
		Name:        "Guinsoo's Rageblade",
		Description: "+10% Attack Speed per attack (stacks)",
		Stats: map[StatType]float64{
			StatAttackSpeed: 0.15,
		},
		OnHitEffect: func(u *Unit, t *Target, damage float64) {
			// Stacking attack speed
			u.Stats.AddBonus(StatAttackSpeed, 0.1)
		},
	},
	"Jeweled Gauntlet": {
		Name:        "Jeweled Gauntlet",
		Description: "Abilities can critically strike",
		Stats: map[StatType]float64{
			StatCritChance:   0.20,
			StatAbilityPower: 0.30,
		},
	},
}
