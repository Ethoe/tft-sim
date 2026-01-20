package items

import (
	"tft-sim/models"
)

func init() {
	Register(models.Item{
		Name:        "Jeweled Gauntlet",
		Description: "Abilities can critically strike.",
		Stats: map[models.StatType]float64{
			models.StatCritChance:   0.35,
			models.StatAbilityPower: 0.35,
		},
		Unique:           true,
		AllowAbilityCrit: true,
	})
}
