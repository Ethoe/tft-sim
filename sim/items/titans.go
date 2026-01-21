package items

import (
	"fmt"
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
		OnHitEffect: func(itemInstance *models.ItemInstance, target *models.Target, damage float64) {
			unit := itemInstance.Owner
			currentTime := unit.Stats.CurrentTime

			// Find the index of this item instance to create unique buff name
			itemIndex := -1
			for i := range unit.Items {
				if unit.Items[i].UniqueName == itemInstance.UniqueName {
					itemIndex = i
					break
				}
			}

			if itemIndex == -1 {
				return // Item not found
			}

			// Create unique buff name for this item instance
			buffName := fmt.Sprintf("Titans Resolve %d", itemIndex)

			// Create a buff with 1 stack worth of bonuses
			buff := models.NewBuff(buffName, 0)
			buff.SetStacking(25, models.StackBehaviorAdditive)

			// Each stack gives 2% AD and 2% AP
			buff.AddStatBonus(models.StatAttackDamage, 0.02)
			buff.AddStatBonus(models.StatAbilityPower, 0.02)

			// Apply the buff - the buff manager will handle stacking
			unit.BuffManager.ApplyBuff(buff, currentTime)

			// Check if we have 25 stacks and add damage amp
			activeBuffs := unit.BuffManager.GetActiveBuffs(currentTime)
			for _, activeBuff := range activeBuffs {
				if activeBuff.Name == buffName && activeBuff.CurrentStacks >= 25 {
					// Ensure damage amp is applied (as multiplier)
					if activeBuff.StatMultipliers[models.StatDamageAmp] < 0.10 {
						activeBuff.AddStatMultiplier(models.StatDamageAmp, 0.10)
					}
				}
			}

			// Update item instance stacks to match buff stacks
			for _, activeBuff := range activeBuffs {
				if activeBuff.Name == buffName {
					itemInstance.Stacks = activeBuff.CurrentStacks
					break
				}
			}
		},
		Stacking:  true,
		MaxStacks: 25,
	})
}
