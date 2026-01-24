package items

import (
	"tft-sim/models"
)

func init() {
	Register(models.Item{
		Name:        "Mittens",
		Description: "artifact item",
		Stats: map[models.StatType]float64{
			models.StatAttackSpeed: .65,
			models.StatDamageAmp:   .15,
		},
		Unique: false,
	})
}
