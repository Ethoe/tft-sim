package items

import (
	"fmt"
	"tft-sim/models"
	"time"
)

func init() {
	Register(models.Item{
		Name:        "Strikers",
		Description: "+10% Attack Speed, +150 Health, +20% Critical Strike Chance, +10% Damage Amp. Critical Strikes grant 5% Damage Amp for 5 seconds, stacking up to 4 times.",
		Stats: map[models.StatType]float64{
			models.StatAttackSpeed: 0.10,
			models.StatHealth:      150.0,
			models.StatCritChance:  0.20,
			models.StatDamageAmp:   0.10,
		},
		OnHitEffect: func(itemInstance *models.ItemInstance, target *models.Target, damage float64) {
			unit := itemInstance.Owner

			// Check if the last attack was a critical strike
			// The CritTracker is updated before OnHitEffect is called
			if unit.CritTracker != nil && unit.CritTracker.CritStreak > 0 {
				currentTime := unit.Stats.CurrentTime

				itemIndex := -1
				for i := range unit.Items {
					if unit.Items[i].UniqueName == itemInstance.UniqueName {
						itemIndex = i
						break
					}
				}

				if itemIndex == -1 {
					return
				}

				buffName := fmt.Sprintf("Strikers Damage Amp%d", itemIndex)

				// Apply DamageAmpBuff with 5% damage amp for 5 seconds, stacking up to 4 times
				damageAmpBuff := models.NewBuff(buffName, 5*time.Second)
				damageAmpBuff.AddStatBonus(models.StatDamageAmp, 0.05)
				damageAmpBuff.SetStacking(4, models.StackBehaviorAdditive)

				// Apply the buff
				unit.BuffManager.ApplyBuff(damageAmpBuff, currentTime)
			}
		},
		Unique: false,
	})
}
