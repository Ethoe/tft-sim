package sim

import (
	"fmt"
	"math"
	"tft-sim/models"
	"time"
)

type SimulationConfig struct {
	Duration     time.Duration
	TickInterval time.Duration
	Targets      []*models.Target
	Verbose      bool
}

type SimulationResult struct {
	TotalDamage    float64
	DPS            float64
	DamageByType   map[models.DamageType]float64
	DamageBySource map[string]float64
	DamageLog      []models.DamageEvent
	TimeToKill     map[string]time.Duration
	FinalHealth    map[string]float64
	Stats          map[string]interface{}
	AttackCount    int
	AbilityCount   int
	CritRate       float64
}

type Simulator struct {
	Unit       *models.Unit
	Targets    []*models.Target
	Config     SimulationConfig
	Time       time.Duration
	IsRunning  bool
	Results    SimulationResult
	GainedMana float64
	LastSecond float64
	ManaLocked bool
}

func NewSimulator(unit *models.Unit, targets []*models.Target) *Simulator {
	return &Simulator{
		Unit:    unit,
		Targets: targets,
		Config: SimulationConfig{
			Duration:     30 * time.Second,
			TickInterval: 17 * time.Millisecond, // 60fps
			Verbose:      true,
		},
		Results: SimulationResult{
			DamageByType: make(map[models.DamageType]float64),
			TimeToKill:   make(map[string]time.Duration),
			FinalHealth:  make(map[string]float64),
		},
	}
}

func (s *Simulator) Run() SimulationResult {
	s.Time = 0
	s.LastSecond = 0
	s.GainedMana = 0
	s.IsRunning = true
	s.ManaLocked = false
	s.Unit.AttackTimer = 0

	// Initialize kill tracking
	for _, target := range s.Targets {
		s.Results.TimeToKill[target.Name] = -1
	}

	for s.Time < s.Config.Duration && s.IsRunning {
		s.tick()
		s.Time += s.Config.TickInterval

		// Check if all targets are dead
		allDead := true
		for _, target := range s.Targets {
			if target.CurrentHP > 0 {
				allDead = false
				break
			}
		}

		if allDead {
			break
		}
	}

	// Calculate final results
	s.calculateResults()

	return s.Results
}

func (s *Simulator) tick() {
	// 1. Handle ongoing casts
	if s.Unit.State == models.UnitStateCasting && s.Unit.CastingCtx != nil {
		if s.Time >= s.Unit.CastingCtx.EndTime {
			s.Unit.CompleteCast(s.Time)
		} else {
			// Still casting, check for mana gain
			s.onSecond()
			return // Can't do anything else while casting
		}
	}

	// 2. Check for ability cast
	if s.Unit.CanCastAbility() && s.Unit.CurrentMana >= s.Unit.Ability.ManaCost {
		targets := s.findAbilityTargets()
		if len(targets) > 0 {
			s.startAbilityCast(targets)
			return // Started casting, wait for next tick
		}
	}

	// 3. Check for auto attack
	if s.Unit.CanAutoAttack(s.Time) {
		s.performAutoAttack()
		s.Unit.NextAttackTime = s.Time + s.Unit.GetAttackInterval()
	}

	// 4. New Second
	s.onSecond()
}

func (s *Simulator) onSecond() {
	if int(s.Time.Seconds()) <= int(math.Floor(s.LastSecond)) {
		return
	}

	// Gain Mana
	if s.Unit.State != models.UnitStateCasting || (s.Unit.CastingCtx != nil && s.Unit.CastingCtx.CanGainMana) {
		manaRegen := s.Unit.Stats.Get(models.StatManaRegen)
		s.Unit.CurrentMana += manaRegen
	}

	// Second Effects
	for _, item := range s.Unit.Items {
		if item.OnSecondEffect != nil {
			item.OnSecondEffect(s.Unit)
		}
	}

	s.LastSecond = s.Time.Seconds()
}

func (s *Simulator) startAbilityCast(targets []*models.Target) {
	s.Unit.StartCastingAbility(s.Time, targets)

	if s.Config.Verbose {
		fmt.Printf("[%.2fs] %s starts casting %s (cost: %.0f mana)\n",
			s.Time.Seconds(), s.Unit.Name, s.Unit.Ability.Name, s.Unit.Ability.ManaCost)
	}
}

func (s *Simulator) performAutoAttack() {
	// Find alive target
	target := s.findTarget()
	if target == nil {
		return
	}

	// Calculate damage
	damage := s.Unit.GetAttackDamage()
	actualDamage, isDead := target.TakeDamage(damage, models.DamageTypePhysical)

	// Log damage
	event := models.DamageEvent{
		Timestamp:  s.Time,
		Damage:     actualDamage,
		DamageType: models.DamageTypePhysical,
		IsAbility:  false,
		TargetName: target.Name,
	}
	s.Unit.DamageLog = append(s.Unit.DamageLog, event)
	s.Unit.TotalDamage += actualDamage

	// Apply on-hit effects
	for _, item := range s.Unit.Items {
		if item.OnHitEffect != nil {
			item.OnHitEffect(s.Unit, target, actualDamage)
		}
	}

	// Gain mana from auto attack
	s.Unit.GainMana(true, 0)

	// Record kill time
	if isDead && s.Results.TimeToKill[target.Name] == -1 {
		s.Results.TimeToKill[target.Name] = s.Time
	}

	if s.Config.Verbose {
		fmt.Printf("[%.2fs] %s auto attacks %s for %.1f damage (%.1f HP remaining)\n",
			s.Time.Seconds(), s.Unit.Name, target.Name, actualDamage, target.CurrentHP)
	}
}

func (s *Simulator) castAbility() {
	if s.Unit.Ability.ManaCost > 0 {
		s.Unit.CurrentMana -= s.Unit.Ability.ManaCost
	}

	// Find target(s)
	targets := s.findAbilityTargets()

	for _, target := range targets {
		damage := s.calculateAbilityDamage()
		actualDamage, isDead := target.TakeDamage(damage, s.Unit.Ability.DamageType)

		// Log damage
		event := models.DamageEvent{
			Timestamp:  s.Time,
			Damage:     actualDamage,
			DamageType: s.Unit.Ability.DamageType,
			IsAbility:  true,
			TargetName: target.Name,
		}
		s.Unit.DamageLog = append(s.Unit.DamageLog, event)
		s.Unit.TotalDamage += actualDamage

		// Record kill time
		if isDead && s.Results.TimeToKill[target.Name] == -1 {
			s.Results.TimeToKill[target.Name] = s.Time
		}

		if s.Config.Verbose {
			fmt.Printf("[%.2fs] %s casts %s on %s for %.1f damage (%.1f HP remaining)\n",
				s.Time.Seconds(), s.Unit.Name, s.Unit.Ability.Name, target.Name, actualDamage, target.CurrentHP)
		}
	}
}

func (s *Simulator) calculateAbilityDamage() float64 {
	baseDamage := s.Unit.Ability.BaseDamage
	if baseDamage == 0 {
		// Use AD scaling for physical abilities, AP for magic
		if s.Unit.Ability.DamageType == models.DamageTypePhysical {
			baseDamage = s.Unit.Stats.Get(models.StatAttackDamage)
		} else {
			baseDamage = s.Unit.Stats.Get(models.StatAttackDamage) * 0.5 // Default scaling
			ap := s.Unit.Stats.Get(models.StatAbilityPower)
			baseDamage += baseDamage * (ap / 100) // AP scaling
		}
	}

	return baseDamage
}

func (s *Simulator) findTarget() *models.Target {
	for _, target := range s.Targets {
		if target.CurrentHP > 0 {
			return target
		}
	}
	return nil
}

func (s *Simulator) findAbilityTargets() []*models.Target {
	if s.Unit.Ability.IsAoE {
		// Return all alive targets for AoE
		var alive []*models.Target
		for _, target := range s.Targets {
			if target.CurrentHP > 0 {
				alive = append(alive, target)
			}
		}
		return alive
	}

	// Single target
	target := s.findTarget()
	if target != nil {
		return []*models.Target{target}
	}
	return nil
}

func (s *Simulator) calculateResults() {
	s.Results.TotalDamage = s.Unit.TotalDamage
	s.Results.DPS = s.Unit.TotalDamage / s.Time.Seconds()
	s.Results.DamageLog = s.Unit.DamageLog

	// Calculate damage by type
	for _, event := range s.Unit.DamageLog {
		s.Results.DamageByType[event.DamageType] += event.Damage
	}

	// Record final health
	for _, target := range s.Targets {
		s.Results.FinalHealth[target.Name] = target.CurrentHP
	}
}
