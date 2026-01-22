package items

import (
	"fmt"
	"tft-sim/models"
)

func init() {
	Register(models.Item{
		Name:        "Krakens",
		Description: "+10% Attack Damage, +10% Attack Speed, +20 Magic Resist. Attacks grant 3.5% stacking Attack Damage, up to 15 attacks. After 15 attacks, gain 30% Attack Speed for the rest of combat.",
		Stats: map[models.StatType]float64{
			models.StatAttackDamage: 0.1,  // 10% attack damage
			models.StatAttackSpeed:  0.1,  // 10% attack speed
			models.StatMagicResist:  20.0, // 20 magic resist
		},
		OnAttackEffect: func(itemInstance *models.ItemInstance) {
			unit := itemInstance.Owner
			currentTime := unit.Stats.CurrentTime

			// Find which Krakens item this is
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

			buffName := fmt.Sprintf("Krakens%d", itemIndex)

			// Create a buff for this Krakens item
			buff := models.NewBuff(buffName, 0)
			buff.SetStacking(15, models.StackBehaviorAdditive)

			// Each application adds 3.5% AD (will stack additively up to 15 times)
			buff.AddStatBonus(models.StatAttackDamage, 0.035)

			// Check current stack count to see if we should add AS bonus
			currentStacks := 0
			for _, existingBuff := range unit.BuffManager.GetActiveBuffs(currentTime) {
				if existingBuff.Name == buffName {
					currentStacks = existingBuff.CurrentStacks
					break
				}
			}

			// If we're about to reach 15 stacks (currently at 14), add AS bonus
			// Only add it once when transitioning from 14 to 15 stacks
			if currentStacks == 14 {
				buff.AddStatBonus(models.StatAttackSpeed, 0.15)
			}

			// Apply the buff (will refresh and stack if already exists)
			unit.BuffManager.ApplyBuff(buff, currentTime)

			// Update item instance stacks
			for _, activeBuff := range unit.BuffManager.GetActiveBuffs(currentTime) {
				if activeBuff.Name == buffName {
					itemInstance.Stacks = activeBuff.CurrentStacks
					break
				}
			}
		},
		Unique: false,
	})
}
