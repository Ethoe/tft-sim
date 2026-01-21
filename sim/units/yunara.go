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
			func(u *models.Unit, t *models.Target, f float64, b bool) (float64, models.DamageType, bool) {
				if b {
					t.TakeDamage(f*0.3, models.DamageTypeTrue)
					return f * 0.3, models.DamageTypeTrue, b
				}
				return 0, models.DamageTypeTrue, b
			},
		)

	// Apply the buff
	u.BuffManager.ApplyBuff(buff, u.Stats.CurrentTime)
}

// createLaserAttackOverride creates a function that overrides auto-attacks with lasers
func createLaserAttackOverride(u *models.Unit, starLevel int, baseDamage, damageReduction float64) func(*models.Unit, *models.Target) float64 {
	return func(attacker *models.Unit, target *models.Target) float64 {
		// Assume only one target

		// Calculate laser damage
		return CalculateLaserDamage(attacker, target, baseDamage, starLevel)
	}
}

// CalculateLaserDamage calculates damage for Yone's Transcendent State lasers
// Returns total damage dealt and whether any laser crit
func CalculateLaserDamage(attacker *models.Unit, target *models.Target, baseDamage float64, starLevel int) float64 {
	// Calculate AD scaling
	bonusAD := attacker.Stats.GetBonus(models.StatAttackDamage)

	// Calculate base laser damage
	laserBaseDamage := baseDamage * (bonusAD + 1)

	// Apply damage reduction for each target passed through
	damageMultiplier := 1.0

	// Apply the piercing damage reduction
	finalDamage := laserBaseDamage * damageMultiplier

	return finalDamage
}
