package items

import (
	"tft-sim/models"
)

func init() {
	Register(models.Item{
		Name:        "IE",
		Description: "Abilities can critically strike.",
		Stats: map[models.StatType]float64{
			models.StatCritChance:   0.35,
			models.StatAttackDamage: 0.35,
		},
		Unique:           true,
		AllowAbilityCrit: true,
	})
}
