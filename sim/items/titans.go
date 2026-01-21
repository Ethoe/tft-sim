package items

import (
	"tft-sim/models"
)

func init() {
	Register(models.Item{
		Name:        "Titans",
		Description: "Grants 10% Attack Speed and 20 Armor. After dealing damage, gain a stack up to 25 times. Each stack gives 2% Attack Damage and 2% Ability Power. At 25 stacks also gain 10% damage amp.",
		Stats: map[models.StatType]float64{
			models.StatAttackSpeed: 0.10,
			models.StatArmor:       20.0,
		},
		OnHitEffect: func(unit *models.Unit, target *models.Target, damage float64) {
			// Create or refresh Titans buff
			currentTime := unit.Stats.CurrentTime

			// Create a buff with 1 stack worth of bonuses
			buff := models.NewBuff("Titans Resolve", 0)
			buff.SetStacking(25, models.StackBehaviorAdditive)

			// Each stack gives 2% AD (as bonus) and 2% AP (as multiplier)
			// For AD: bonus of 0.02 per stack (since AD uses base * (1 + bonus))
			// For AP: multiplier of 0.02 per stack (since AP uses (base + bonus) * multiplier)
			buff.AddStatBonus(models.StatAttackDamage, 0.02)
			buff.AddStatBonus(models.StatAbilityPower, 0.02)

			// Apply the buff - the buff manager will handle stacking
			unit.BuffManager.ApplyBuff(buff, currentTime)

			// Check if we have 25 stacks and add damage amp
			activeBuffs := unit.BuffManager.GetActiveBuffs(currentTime)
			for _, activeBuff := range activeBuffs {
				if activeBuff.Name == "Titans Resolve" && activeBuff.CurrentStacks >= 25 {
					// Ensure damage amp is applied (as multiplier)
					if activeBuff.StatMultipliers[models.StatDamageAmp] < 0.10 {
						activeBuff.AddStatMultiplier(models.StatDamageAmp, 0.10)
					}
				}
			}
		},
		Stacking:  true,
		MaxStacks: 25,
	})
}
