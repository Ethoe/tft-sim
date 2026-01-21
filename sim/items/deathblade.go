package items

import (
	"tft-sim/models"
)

func init() {
	Register(models.Item{
		Name:        "Deathblade",
		Description: "+55 Attack Damage",
		Stats: map[models.StatType]float64{
			models.StatAttackDamage: .55,
			models.StatDamageAmp:    .10,
		},
		Unique: false,
	})
}
