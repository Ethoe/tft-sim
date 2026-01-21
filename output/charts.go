package output

import (
	"fmt"
	"sort"
	"tft-sim/models"
	"tft-sim/sim"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// GenerateDamageChart creates a PNG chart of cumulative damage over time
func GenerateDamageChart(results sim.SimulationResult, filename string) error {
	p := plot.New()
	p.Title.Text = "Damage Over Time"
	p.X.Label.Text = "Time (seconds)"
	p.Y.Label.Text = "Cumulative Damage"

	// Prepare data points for cumulative damage
	pts := make(plotter.XYs, len(results.DamageOverTime))
	for i, dot := range results.DamageOverTime {
		pts[i].X = dot.Timestamp.Seconds()
		pts[i].Y = dot.CumulativeDamage
	}

	line, err := plotter.NewLine(pts)
	if err != nil {
		return fmt.Errorf("failed to create line plot: %w", err)
	}
	line.Color = plotutil.Color(0)
	line.Width = vg.Points(1.5)

	p.Add(line)
	p.Legend.Add("Total Damage", line)

	// Save to file
	if err := p.Save(8*vg.Inch, 6*vg.Inch, filename); err != nil {
		return fmt.Errorf("failed to save chart: %w", err)
	}

	return nil
}

// GenerateDamageByTypeChart creates a multi-series chart showing damage by type over time
func GenerateDamageByTypeChart(results sim.SimulationResult, filename string) error {
	p := plot.New()
	p.Title.Text = "Damage by Type Over Time"
	p.X.Label.Text = "Time (seconds)"
	p.Y.Label.Text = "Cumulative Damage"

	// Initialize cumulative damage by type
	cumulativeByType := map[models.DamageType]float64{
		models.DamageTypePhysical: 0,
		models.DamageTypeMagic:    0,
		models.DamageTypeTrue:     0,
	}

	// Prepare data points for each damage type
	physicalPts := make(plotter.XYs, 0, len(results.DamageOverTime))
	magicPts := make(plotter.XYs, 0, len(results.DamageOverTime))
	truePts := make(plotter.XYs, 0, len(results.DamageOverTime))

	for _, dot := range results.DamageOverTime {
		// Update cumulative damage for each type
		for dmgType, amount := range dot.DamageByType {
			cumulativeByType[dmgType] += amount
		}

		// Add points for each type
		t := dot.Timestamp.Seconds()
		physicalPts = append(physicalPts, plotter.XY{X: t, Y: cumulativeByType[models.DamageTypePhysical]})
		magicPts = append(magicPts, plotter.XY{X: t, Y: cumulativeByType[models.DamageTypeMagic]})
		truePts = append(truePts, plotter.XY{X: t, Y: cumulativeByType[models.DamageTypeTrue]})
	}

	// Create lines for each damage type
	physicalLine, err := plotter.NewLine(physicalPts)
	if err != nil {
		return fmt.Errorf("failed to create physical damage line: %w", err)
	}
	physicalLine.Color = plotutil.Color(0) // Blue
	physicalLine.Width = vg.Points(1.5)

	magicLine, err := plotter.NewLine(magicPts)
	if err != nil {
		return fmt.Errorf("failed to create magic damage line: %w", err)
	}
	magicLine.Color = plotutil.Color(1) // Red
	magicLine.Width = vg.Points(1.5)

	trueLine, err := plotter.NewLine(truePts)
	if err != nil {
		return fmt.Errorf("failed to create true damage line: %w", err)
	}
	trueLine.Color = plotutil.Color(2) // Green
	trueLine.Width = vg.Points(1.5)

	p.Add(physicalLine, magicLine, trueLine)
	p.Legend.Add("Physical Damage", physicalLine)
	p.Legend.Add("Magic Damage", magicLine)
	p.Legend.Add("True Damage", trueLine)
	p.Legend.Top = true

	// Save to file
	if err := p.Save(10*vg.Inch, 8*vg.Inch, filename); err != nil {
		return fmt.Errorf("failed to save chart: %w", err)
	}

	return nil
}

// GenerateDPSChart creates a chart showing damage per second over time
func GenerateDPSChart(results sim.SimulationResult, filename string, windowSize float64) error {
	if windowSize <= 0 {
		windowSize = 1.0 // Default 1-second window
	}

	p := plot.New()
	p.Title.Text = fmt.Sprintf("Damage Per Second (%.1fs window)", windowSize)
	p.X.Label.Text = "Time (seconds)"
	p.Y.Label.Text = "DPS"

	if len(results.DamageOverTime) == 0 {
		return fmt.Errorf("no damage data available")
	}

	// Calculate DPS using sliding window
	dpsPts := make(plotter.XYs, 0)
	events := results.DamageOverTime

	// Sort events by timestamp (should already be sorted, but ensure)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp < events[j].Timestamp
	})

	for i := 0; i < len(events); i++ {
		currentTime := events[i].Timestamp.Seconds()
		windowStart := currentTime - windowSize
		if windowStart < 0 {
			windowStart = 0
		}

		// Calculate damage in the window
		windowDamage := 0.0
		for j := 0; j <= i; j++ {
			if events[j].Timestamp.Seconds() >= windowStart {
				windowDamage += events[j].InstantDamage
			}
		}

		dps := windowDamage / windowSize
		dpsPts = append(dpsPts, plotter.XY{X: currentTime, Y: dps})
	}

	line, err := plotter.NewLine(dpsPts)
	if err != nil {
		return fmt.Errorf("failed to create DPS line: %w", err)
	}
	line.Color = plotutil.Color(3) // Purple
	line.Width = vg.Points(1.5)

	p.Add(line)
	p.Legend.Add("DPS", line)

	// Save to file
	if err := p.Save(8*vg.Inch, 6*vg.Inch, filename); err != nil {
		return fmt.Errorf("failed to save chart: %w", err)
	}

	return nil
}

// GenerateComparisonChart creates a chart comparing cumulative damage over time for multiple builds
func GenerateComparisonChart(results []sim.SimulationResult, labels []string, filename string) error {
	if len(results) != len(labels) {
		return fmt.Errorf("number of results (%d) must match number of labels (%d)", len(results), len(labels))
	}

	if len(results) == 0 {
		return fmt.Errorf("no results provided for comparison")
	}

	p := plot.New()
	p.Title.Text = "Build Comparison: Damage Over Time"
	p.X.Label.Text = "Time (seconds)"
	p.Y.Label.Text = "Cumulative Damage"
	p.Legend.Top = true

	// Create a line for each build
	for i, result := range results {
		if len(result.DamageOverTime) == 0 {
			continue // Skip builds with no damage data
		}

		// Prepare data points for cumulative damage
		pts := make(plotter.XYs, len(result.DamageOverTime))
		for j, dot := range result.DamageOverTime {
			pts[j].X = dot.Timestamp.Seconds()
			pts[j].Y = dot.CumulativeDamage
		}

		line, err := plotter.NewLine(pts)
		if err != nil {
			return fmt.Errorf("failed to create line for build %d: %w", i, err)
		}

		// Use different colors for each build
		line.Color = plotutil.Color(i)
		line.Width = vg.Points(1.5)

		p.Add(line)
		p.Legend.Add(labels[i], line)
	}

	// Save to file
	if err := p.Save(10*vg.Inch, 8*vg.Inch, filename); err != nil {
		return fmt.Errorf("failed to save comparison chart: %w", err)
	}

	return nil
}
