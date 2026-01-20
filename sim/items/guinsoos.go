package items

import (
	"tft-sim/models"
)

func init() {
	Register(models.Item{
		Name:        "Guinsoos",
		Description: "Basic Attacks grant +6% bonus Attack Speed for the rest of combat. Stacks with no upper limit.",
		Stats: map[models.StatType]float64{
			models.StatAttackSpeed:  0.10,
			models.StatAbilityPower: 0.10,
		},
		OnSecondEffect: func(u *models.Unit) {
			u.Stats.AddBonus(models.StatAttackSpeed, 0.07)
		},
		Stacking:  true,
		MaxStacks: 10000,
	})
}
