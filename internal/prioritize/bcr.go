package prioritize

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MitigationStructure defines the effectiveness, cost, and lifespan of a mitigation measure.
type MitigationStructure struct {
	ID                   string  `yaml:"id" json:"id"`
	Name                 string  `yaml:"name" json:"name"`
	Effectiveness        float64 `yaml:"effectiveness" json:"effectiveness"`
	BaseCostUSD          float64 `yaml:"base_cost_usd" json:"base_cost_usd"`
	FencingCostPerMeter  float64 `yaml:"fencing_cost_per_meter" json:"fencing_cost_per_meter"`
	AnnualMaintenancePct float64 `yaml:"annual_maintenance_pct" json:"annual_maintenance_pct"`
	LifespanYears        int     `yaml:"lifespan_years" json:"lifespan_years"`
	Citation             string  `yaml:"citation" json:"citation"`
}

// BCRResult contains financial and collision reduction calculations for a candidate project.
type BCRResult struct {
	CapitalCostUSD            float64
	AnnualMaintenanceUSD      float64
	LifetimeCostUSD           float64
	ExpectedAvoidedCollisions float64
	P10AvoidedCollisions      float64
	P90AvoidedCollisions      float64
	ExpectedSavingsUSD        float64
	P10SavingsUSD             float64
	P90SavingsUSD             float64
	BCR                       float64
	P10BCR                    float64
	P90BCR                    float64
}

// CalculateBCR computes the Benefit-Cost Ratio and uncertainty intervals for a mitigation structure.
// Lifetime cost = Capital cost + (Annual maintenance cost * lifespan years)
// Expected avoided collisions = annual rate * effectiveness * lifespan years
// Expected savings = avoided collisions * avg species cost
// BCR = Expected savings / Lifetime cost
// P10 and P90 BCR are computed from risk uncertainty intervals (P10 and P90 rates).
func CalculateBCR(annualRate, p10Rate, p90Rate, lengthM float64, structure MitigationStructure, avgSpeciesCost float64) BCRResult {
	lifespan := structure.LifespanYears
	if lifespan <= 0 {
		lifespan = 20
	}
	lifespanF := float64(lifespan)

	capitalCost := structure.BaseCostUSD + (structure.FencingCostPerMeter * lengthM)
	if capitalCost < 0 {
		capitalCost = 0
	}
	annualMaint := capitalCost * structure.AnnualMaintenancePct
	lifetimeCost := capitalCost + (annualMaint * lifespanF)

	eff := structure.Effectiveness
	if eff < 0 {
		eff = 0
	}

	avoided := annualRate * eff * lifespanF
	p10Avoided := p10Rate * eff * lifespanF
	p90Avoided := p90Rate * eff * lifespanF

	savings := avoided * avgSpeciesCost
	p10Savings := p10Avoided * avgSpeciesCost
	p90Savings := p90Avoided * avgSpeciesCost

	var bcr, p10BCR, p90BCR float64
	if lifetimeCost > 0 {
		bcr = savings / lifetimeCost
		p10BCR = p10Savings / lifetimeCost
		p90BCR = p90Savings / lifetimeCost
	}

	return BCRResult{
		CapitalCostUSD:            capitalCost,
		AnnualMaintenanceUSD:      annualMaint,
		LifetimeCostUSD:           lifetimeCost,
		ExpectedAvoidedCollisions: avoided,
		P10AvoidedCollisions:      p10Avoided,
		P90AvoidedCollisions:      p90Avoided,
		ExpectedSavingsUSD:        savings,
		P10SavingsUSD:             p10Savings,
		P90SavingsUSD:             p90Savings,
		BCR:                       bcr,
		P10BCR:                    p10BCR,
		P90BCR:                    p90BCR,
	}
}

// PrioritizeConfig holds the configuration for mitigation effectiveness, species costs, and prioritization thresholds.
type PrioritizeConfig struct {
	Structures           map[string]MitigationStructure
	SpeciesCosts         map[string]float64
	DefaultSpeciesCost   float64
	MinPercentile        float64
	DefaultStructureType string
}

// DefaultConfig returns default configurations matching the literature and USDOT guidance.
func DefaultConfig() PrioritizeConfig {
	structures := map[string]MitigationStructure{
		"overpass_with_fencing": {
			ID:                   "overpass_with_fencing",
			Name:                 "Wildlife Overpass with Continuous Fencing",
			Effectiveness:        0.86,
			BaseCostUSD:          3500000.0,
			FencingCostPerMeter:  120.0,
			AnnualMaintenancePct: 0.015,
			LifespanYears:        25,
			Citation:             "Rytwinski, T., et al. (2016). How Effective Are Wildlife Mitigation Measures in Reducing Wildlife-Vehicle Collisions? A Systematic Review and Meta-Analysis. PLoS ONE 11(11): e0165034.",
		},
		"underpass_with_fencing": {
			ID:                   "underpass_with_fencing",
			Name:                 "Wildlife Underpass with Continuous Fencing",
			Effectiveness:        0.83,
			BaseCostUSD:          1800000.0,
			FencingCostPerMeter:  120.0,
			AnnualMaintenancePct: 0.015,
			LifespanYears:        20,
			Citation:             "Clevenger, A.P., & Huijser, M.P. (2011). Wildlife Crossing Structure Handbook: Design and Evaluation in North America. FHWA-CFL/TD-11-003.",
		},
		"fencing_with_jumpouts": {
			ID:                   "fencing_with_jumpouts",
			Name:                 "Wildlife Exclusion Fencing with Escape Jumpouts",
			Effectiveness:        0.78,
			BaseCostUSD:          150000.0,
			FencingCostPerMeter:  120.0,
			AnnualMaintenancePct: 0.02,
			LifespanYears:        20,
			Citation:             "Huijser, M.P., et al. (2009). Cost-Benefit Analyses of Mitigation Measures Aimed at Reducing Collisions with Large Ungulates in the United States and Canada. Transportation Research Record 2140: 18-26.",
		},
		"variable_speed_signs": {
			ID:                   "variable_speed_signs",
			Name:                 "Dynamic / Variable Wildlife Warning Signs",
			Effectiveness:        0.35,
			BaseCostUSD:          60000.0,
			FencingCostPerMeter:  0.0,
			AnnualMaintenancePct: 0.05,
			LifespanYears:        10,
			Citation:             "Huijser, M.P., et al. (2009). Cost-Benefit Analyses of Mitigation Measures Aimed at Reducing Large Ungulate Collisions.",
		},
	}

	speciesCosts := map[string]float64{
		"deer":      14000.0,
		"elk":       45000.0,
		"moose":     83000.0,
		"pronghorn": 16000.0,
	}

	return PrioritizeConfig{
		Structures:         structures,
		SpeciesCosts:       speciesCosts,
		DefaultSpeciesCost: 15000.0,
		MinPercentile:      90.0,
	}
}

// LoadConfigFromFiles loads mitigation effectiveness and species costs from YAML files.
func LoadConfigFromFiles(effPath, speciesPath string) (PrioritizeConfig, error) {
	cfg := DefaultConfig()

	effBytes, err := os.ReadFile(effPath)
	if err != nil {
		return cfg, fmt.Errorf("read mitigation effectiveness file: %w", err)
	}

	var effDoc struct {
		Structures map[string]struct {
			Name                 string  `yaml:"name"`
			Effectiveness        float64 `yaml:"effectiveness"`
			BaseCostUSD          float64 `yaml:"base_cost_usd"`
			FencingCostPerMeter  float64 `yaml:"fencing_cost_per_meter"`
			AnnualMaintenancePct float64 `yaml:"annual_maintenance_pct"`
			LifespanYears        int     `yaml:"lifespan_years"`
			Citation             string  `yaml:"citation"`
		} `yaml:"structures"`
	}
	if err := yaml.Unmarshal(effBytes, &effDoc); err != nil {
		return cfg, fmt.Errorf("unmarshal mitigation effectiveness: %w", err)
	}

	cfg.Structures = make(map[string]MitigationStructure, len(effDoc.Structures))
	for id, s := range effDoc.Structures {
		cfg.Structures[id] = MitigationStructure{
			ID:                   id,
			Name:                 s.Name,
			Effectiveness:        s.Effectiveness,
			BaseCostUSD:          s.BaseCostUSD,
			FencingCostPerMeter:  s.FencingCostPerMeter,
			AnnualMaintenancePct: s.AnnualMaintenancePct,
			LifespanYears:        s.LifespanYears,
			Citation:             s.Citation,
		}
	}

	speciesBytes, err := os.ReadFile(speciesPath)
	if err != nil {
		return cfg, fmt.Errorf("read species costs file: %w", err)
	}

	var speciesDoc struct {
		Costs   map[string]float64 `yaml:"costs"`
		Default float64            `yaml:"default"`
	}
	if err := yaml.Unmarshal(speciesBytes, &speciesDoc); err != nil {
		return cfg, fmt.Errorf("unmarshal species costs: %w", err)
	}

	if speciesDoc.Costs != nil {
		cfg.SpeciesCosts = speciesDoc.Costs
	}
	if speciesDoc.Default > 0 {
		cfg.DefaultSpeciesCost = speciesDoc.Default
	}

	return cfg, nil
}
