package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tft-sim/models"
	"tft-sim/output"
	"tft-sim/sim"
	"tft-sim/sim/items"
	"tft-sim/sim/units"
)

// runSimulation runs a simulation with a specific build and returns the results
func runSimulation(buildName string, itemNames []string) (sim.SimulationResult, error) {
	// Get Yunara unit from registry (1-star)
	unit, exists := units.Get("Yunara", 2)
	if !exists {
		return sim.SimulationResult{}, fmt.Errorf("Yunara unit not found in registry")
	}

	// Add items
	for _, itemName := range itemNames {
		if item, exists := items.Get(itemName); exists {
			unit.AddItem(item)
		} else {
			return sim.SimulationResult{}, fmt.Errorf("item %s not found in registry", itemName)
		}
	}

	// Create targets
	targets := []*models.Target{
		models.NewTarget("Frontline Tank", 50000, 100, 50),
	}

	// Create and run simulation
	simulator := sim.NewSimulator(unit, targets)
	results := simulator.Run()

	// Print build summary
	fmt.Printf("\n=== Build: %s ===\n", buildName)
	fmt.Printf("Items: %v\n", itemNames)
	fmt.Printf("Total Damage: %.1f\n", results.TotalDamage)
	fmt.Printf("DPS: %.1f\n", results.DPS)
	fmt.Printf("Simulation Duration: %.2fs\n", simulator.Time.Seconds())

	return results, nil
}

func main() {
	fmt.Println("=== TFT Simulation Build Comparison ===")

	generateIndividual := false

	// Define the two builds to compare
	builds := []struct {
		name      string
		itemNames []string
	}{
		{
			name:      "Yunara - RB Kraken IE",
			itemNames: []string{"Guinsoos", "Krakens", "IE"},
		},
		{
			name:      "Yunara - RB Titans IE",
			itemNames: []string{"Guinsoos", "Titans", "IE"},
		},
		{
			name:      "Yunara - 2Krakens IE",
			itemNames: []string{"Krakens", "Krakens", "IE"},
		},
		{
			name:      "Yunara - 5 DBs",
			itemNames: []string{"Deathblade", "Deathblade", "Deathblade", "Deathblade", "Deathblade"},
		},
	}

	// Run simulations for each build
	var allResults []sim.SimulationResult
	var buildLabels []string

	for _, build := range builds {
		fmt.Printf("\nRunning simulation for: %s\n", build.name)
		results, err := runSimulation(build.name, build.itemNames)
		if err != nil {
			fmt.Printf("Error running simulation for %s: %v\n", build.name, err)
			continue
		}
		allResults = append(allResults, results)
		buildLabels = append(buildLabels, build.name)
	}

	if len(allResults) == 0 {
		fmt.Println("Error: No simulations completed successfully")
		return
	}

	// Print comparison summary
	fmt.Println("\n=== Build Comparison Summary ===")
	for i, result := range allResults {
		fmt.Printf("\nBuild %d: %s\n", i+1, buildLabels[i])
		fmt.Printf("  Total Damage: %.1f\n", result.TotalDamage)
		fmt.Printf("  DPS: %.1f\n", result.DPS)
		fmt.Printf("  Crit Ratio: %.1f%% \n", result.CritRate*100)
		fmt.Printf("  Damage Breakdown:\n")
		for dmgType, amount := range result.DamageByType {
			typeName := "Physical"
			switch dmgType {
			case models.DamageTypeMagic:
				typeName = "Magic"
			case models.DamageTypeTrue:
				typeName = "True"
			}
			percentage := (amount / result.TotalDamage) * 100
			fmt.Printf("    %s: %.1f (%.1f%%)\n", typeName, amount, percentage)
		}
	}

	// Generate comparison chart
	fmt.Println("\n=== Generating Comparison Chart ===")

	// Create output directory if it doesn't exist
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create output directory: %v\n", err)
		// Continue anyway, charts might fail
	}

	// Generate comparison chart
	comparisonChart := filepath.Join(outputDir, "comparisons.png")
	if err := output.GenerateComparisonChart(allResults, buildLabels, comparisonChart); err != nil {
		fmt.Printf("Failed to generate comparison chart: %v\n", err)
	} else {
		fmt.Printf("✓ Build comparison chart saved to: %s\n", comparisonChart)
	}

	// Also generate individual charts for each build
	if generateIndividual {
		for i, result := range allResults {
			buildName := strings.ReplaceAll(strings.ToLower(buildLabels[i]), " + ", "_")
			buildName = strings.ReplaceAll(buildName, " ", "_")

			// Generate cumulative damage chart for this build
			cumulativeChart := filepath.Join(outputDir, fmt.Sprintf("%s_damage_over_time.png", buildName))
			if err := output.GenerateDamageChart(result, cumulativeChart); err != nil {
				fmt.Printf("Failed to generate cumulative damage chart for %s: %v\n", buildLabels[i], err)
			} else {
				fmt.Printf("✓ Individual chart for %s saved to: %s\n", buildLabels[i], cumulativeChart)
			}
		}
	}

	fmt.Println("\nComparison completed successfully!")
}
