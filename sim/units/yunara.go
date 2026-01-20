package units

import (
	"fmt"
	"tft-sim/models"
	"time"
)

func init() {
	Register("Yunara", NewYunara)
}

// NewYunara creates a new Yunara unit with the given star level
func NewYunara(starLevel int) *models.Unit {
	// Base stats (same for all star levels for now)
	healths := []float64{800, 1440, 2592}
	ads := []float64{60, 90, 135}

	baseStats := map[models.StatType]float64{
		models.StatHealth:       healths[starLevel-1],
		models.StatAttackDamage: ads[starLevel-1],
		models.StatAbilityPower: 0,
		models.StatAttackSpeed:  .8, // attacks per second
		models.StatArmor:        30,
		models.StatMagicResist:  30,
		models.StatMana:         50, // Mana cost for Transcendent State
		models.StatCritChance:   0.25,
		models.StatCritDamage:   .4,
	}

	// Create ability
	ability := createTranscendentStateAbility(starLevel)

	// Create unit template
	unitTemplate := models.Unit{
		Name:         "Yunara",
		UnitRole:     models.RoleAttackMarksman,
		StarLevel:    starLevel,
		CurrentMana:  0,
		AttackTimer:  0,
		AttackWindup: 20 * time.Millisecond,
	}

	// Stage is hardcoded to 2 for now (as in main.go)
	stage := 2
	unit := models.NewUnit(unitTemplate, ability, baseStats, stage)

	return unit
}

// createTranscendentStateAbility creates Yunara's Transcendent State ability
func createTranscendentStateAbility(starLevel int) models.Ability {
	// Determine values based on star level
	var baseDamage float64
	var damageReductionPerTarget float64
	var attackSpeedBonus float64

	switch starLevel {
	case 1:
		baseDamage = 85
		damageReductionPerTarget = 0.7 // 70%
		attackSpeedBonus = 0.75        // 75%
	case 2:
		baseDamage = 130
		damageReductionPerTarget = 0.7 // 70%
		attackSpeedBonus = 0.75        // 75%
	case 3:
		baseDamage = 450
		damageReductionPerTarget = 0.3 // 30%
		attackSpeedBonus = 3.0         // 300%
	default:
		// Default to 1-star values
		baseDamage = 85
		damageReductionPerTarget = 0.7
		attackSpeedBonus = 0.75
	}

	return models.Ability{
		Name:                        "Transcendent State",
		BaseDamage:                  baseDamage,
		DamageType:                  models.DamageTypePhysical,
		CastTime:                    4 * time.Second,
		IsAoE:                       true,
		IsAutoAttackModifier:        true,
		AllowsManaGainDuringCast:    false,
		AllowsAutoAttacksDuringCast: true,
		OnCastStart: func(u *models.Unit) {
			// Apply Transcendent State buff
			applyTranscendentStateBuff(u, starLevel, baseDamage, damageReductionPerTarget, attackSpeedBonus)
		},
	}
}

// applyTranscendentStateBuff applies the Transcendent State buff to the unit
func applyTranscendentStateBuff(u *models.Unit, starLevel int, baseDamage, damageReduction, attackSpeedBonus float64) {
	// Calculate actual attack speed bonus (AP scaling)
	ap := u.Stats.Get(models.StatAbilityPower)
	actualAttackSpeedBonus := attackSpeedBonus + (ap * attackSpeedBonus)

	// Create the buff
	buff := models.NewBuff("Transcendent State", 4*time.Second).
		AddStatBonus(models.StatAttackSpeed, actualAttackSpeedBonus).
		SetAutoAttackOverride(createLaserAttackOverride(u, starLevel, baseDamage, damageReduction)).
		SetCallbacks(
			func(u *models.Unit) {
				fmt.Printf("[Buff Applied] %s enters Transcendent State (+%.0f%% attack speed)\n",
					u.Name, actualAttackSpeedBonus*100)
			},
			nil,
			func(u *models.Unit) {
				fmt.Printf("[Buff Expired] %s leaves Transcendent State\n", u.Name)
				if u.CastingCtx != nil {
				}
			},
			nil,
		)

	// Apply the buff
	u.BuffManager.ApplyBuff(buff, u.Stats.CurrentTime)
}

// createLaserAttackOverride creates a function that overrides auto-attacks with lasers
func createLaserAttackOverride(u *models.Unit, starLevel int, baseDamage, damageReduction float64) func(*models.Unit, *models.Target) (float64, bool) {
	return func(attacker *models.Unit, target *models.Target) (float64, bool) {
		// For now, simulate single target laser attack
		// In a real implementation, this would hit multiple targets in a line

		// Get all alive targets (simulating line piercing)
		var aliveTargets []*models.Target
		// This is a simplified version - in reality we'd need access to all targets
		// For now, just hit the primary target
		aliveTargets = append(aliveTargets, target)

		// Calculate laser damage
		totalDamage, anyCrit := CalculateLaserDamage(attacker, aliveTargets, baseDamage, starLevel)

		return totalDamage, anyCrit
	}
}

// CalculateLaserDamage calculates damage for Yone's Transcendent State lasers
// baseDamage: 85/130/450 based on star level
// targets: slice of targets in the line (closest to farthest)
// starLevel: 1, 2, or 3 for damage reduction scaling
// Returns total damage dealt and whether any laser crit
func CalculateLaserDamage(attacker *models.Unit, targets []*models.Target, baseDamage float64, starLevel int) (float64, bool) {
	if len(targets) == 0 {
		return 0, false
	}

	totalDamage := 0.0
	anyCrit := false

	// Determine damage reduction per target based on star level
	damageReductionPerTarget := 0.7 // 70% for star levels 1 and 2
	if starLevel == 3 {
		damageReductionPerTarget = 0.3 // 30% for star level 3
	}

	// Calculate AD scaling
	bonusAD := attacker.Stats.GetBonus(models.StatAttackDamage)

	for i, target := range targets {
		if target.CurrentHP <= 0 {
			continue // Skip dead targets
		}

		// Calculate base laser damage (85/130/450 + AD)
		laserBaseDamage := baseDamage * (bonusAD + 1)

		// Apply damage reduction for each target passed through
		damageMultiplier := 1.0
		for j := 0; j < i; j++ {
			damageMultiplier *= (1 - damageReductionPerTarget)
		}

		// Calculate physical damage with the laser's base damage
		physicalDamage, isCrit := models.CalculatePhysicalDamage(attacker, target, laserBaseDamage, attacker.Ability.CanAbilityCrit)

		// Apply the piercing damage reduction
		finalDamage := physicalDamage * damageMultiplier

		if isCrit && attacker.Ability.CanAbilityCrit {
			anyCrit = true
			// Apply 33% bonus true damage on crit
			// True damage bonus is 33% of the crit damage (before armor reduction)
			// and is not subject to further crit
			bonusTrueDamage := finalDamage * 0.33
			// Apply true damage directly (ignores armor/MR/damage reduction)
			//target.TakeDamage(bonusTrueDamage, models.DamageTypeTrue)
			finalDamage += bonusTrueDamage
			// totalDamage += bonusTrueDamage
		}

		// Apply damage to target
		target.TakeDamage(finalDamage, models.DamageTypePhysical)
		totalDamage += finalDamage
	}

	return totalDamage, anyCrit
}
