package models

type Augment struct {
	Name        string
	Description string
	Stats       map[StatType]float64
}

// Predefined augments
var Augments = map[string]Augment{
	"Combat Training": {
		Name:        "Combat Training",
		Description: "Your team gains 10 Attack Damage",
		Stats: map[StatType]float64{
			StatAttackDamage: .10,
		},
	},
	"Magic Wand": {
		Name:        "Magic Wand",
		Description: "Your team gains 10 Ability Power",
		Stats: map[StatType]float64{
			StatAbilityPower: .10,
		},
	},
}
