package items

import (
	"tft-sim/models"
)

func init() {
	Register(models.Item{
		Name:        "Red",
		Description: "burn item",
		Stats: map[models.StatType]float64{
			models.StatAttackSpeed: .45,
			models.StatDamageAmp:   .06,
		},
		Unique: false,
	})
}
