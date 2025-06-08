package health

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/valentinesamuel/mockcraft/internal/generators/interfaces"
	"github.com/valentinesamuel/mockcraft/internal/registry"
)

// HealthGenerator provides health-specific data generation
type HealthGenerator struct {
	rand *rand.Rand
}

// GenerateByType implements interfaces.Generator.
func (g *HealthGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
	if err := g.validateParameters(dataType, params); err != nil {
		return nil, err
	}

	switch dataType {
	case "blood_type":
		return g.GenerateBloodType(), nil
	case "medical_condition":
		return g.GenerateMedicalCondition(), nil
	case "medication":
		return g.GenerateMedication(), nil
	case "symptom":
		return g.GenerateSymptom(), nil
	case "diagnosis":
		return g.GenerateDiagnosis(), nil
	case "allergy":
		return g.GenerateAllergy(), nil
	case "lab_result":
		return g.GenerateLabResult(), nil
	case "vital_sign":
		return g.GenerateVitalSigns(), nil
	case "medical_record":
		return g.GenerateMedicalRecord(), nil
	default:
		return nil, fmt.Errorf("unknown health type: %s", dataType)
	}

}

func (g *HealthGenerator) validateParameters(dataType string, params map[string]interface{}) error {
	// Define valid data types for health generator
    validTypes := map[string]bool{
        "blood_type":        true,
        "medical_condition": true,
        "medication":        true,
        "symptom":           true,
        "diagnosis":         true,
        "allergy":           true,
        "lab_result":        true,
        "vital_sign":        true,
        "medical_record":    true,
    }

    if !validTypes[dataType] {
        return fmt.Errorf("unsupported health data type: %s", dataType)
    }

    // Add any parameter validation logic here if needed
    // For now, we'll accept any parameters
    return nil
}

// NewHealthGenerator creates a new HealthGenerator instance
func NewHealthGenerator() *HealthGenerator {
	return &HealthGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

}

// Auto-register when package is imported
func init() {
	registry.Register("health", func() interfaces.Generator {
		return NewHealthGenerator() // Now this matches the constructor
	})
}

// SetSeed implements interfaces.Generator
func (g *HealthGenerator) SetSeed(seed int64) {
	g.rand = rand.New(rand.NewSource(seed))
}

// GenerateBloodType generates a random blood type
func (g *HealthGenerator) GenerateBloodType() BloodType {
	return ValidBloodTypes[rand.Intn(len(ValidBloodTypes))]
}

// GenerateMedicalCondition generates a random medical condition
func (g *HealthGenerator) GenerateMedicalCondition() MedicalCondition {
	return CommonConditions[rand.Intn(len(CommonConditions))]
}

// GenerateMedication generates a random medication
func (g *HealthGenerator) GenerateMedication() Medication {
	return CommonMedications[rand.Intn(len(CommonMedications))]
}

// GenerateSymptom generates a random symptom
func (g *HealthGenerator) GenerateSymptom() Symptom {
	return CommonSymptoms[rand.Intn(len(CommonSymptoms))]
}

// GenerateDiagnosis generates a random diagnosis
func (g *HealthGenerator) GenerateDiagnosis() Diagnosis {
	return CommonDiagnoses[rand.Intn(len(CommonDiagnoses))]
}

// GenerateAllergy generates a random allergy
func (g *HealthGenerator) GenerateAllergy() Allergy {
	return CommonAllergies[rand.Intn(len(CommonAllergies))]
}

// GenerateLabValue generates a lab value within normal range
func (g *HealthGenerator) GenerateLabValue(min, max float64, unit string) LabValue {
	value := min + rand.Float64()*(max-min)
	return LabValue{
		Value: value,
		Unit:  unit,
	}
}

// GenerateLabResult generates a complete set of lab results
func (g *HealthGenerator) GenerateLabResult() LabResult {
	return LabResult{
		Glucose:     g.GenerateLabValue(NormalLabRanges.Glucose.Min, NormalLabRanges.Glucose.Max, UnitMgDL),
		Cholesterol: g.GenerateLabValue(NormalLabRanges.Cholesterol.Min, NormalLabRanges.Cholesterol.Max, UnitMgDL),
		Hemoglobin:  g.GenerateLabValue(NormalLabRanges.Hemoglobin.Min, NormalLabRanges.Hemoglobin.Max, UnitGDL),
	}
}

// GenerateBloodPressure generates a blood pressure reading within normal range
func (g *HealthGenerator) GenerateBloodPressure() BloodPressure {
	return BloodPressure{
		Systolic:  rand.Intn(NormalVitalRanges.BloodPressure.SystolicMax-NormalVitalRanges.BloodPressure.SystolicMin) + NormalVitalRanges.BloodPressure.SystolicMin,
		Diastolic: rand.Intn(NormalVitalRanges.BloodPressure.DiastolicMax-NormalVitalRanges.BloodPressure.DiastolicMin) + NormalVitalRanges.BloodPressure.DiastolicMin,
		Unit:      UnitMmHg,
	}
}

// GenerateVitalSigns generates a complete set of vital signs
func (g *HealthGenerator) GenerateVitalSigns() VitalSign {
	return VitalSign{
		BloodPressure: g.GenerateBloodPressure(),
		HeartRate: LabValue{
			Value: float64(rand.Intn(NormalVitalRanges.HeartRate.Max-NormalVitalRanges.HeartRate.Min) + NormalVitalRanges.HeartRate.Min),
			Unit:  UnitBPM,
		},
		Temperature: LabValue{
			Value: NormalVitalRanges.Temperature.Min + rand.Float64()*(NormalVitalRanges.Temperature.Max-NormalVitalRanges.Temperature.Min),
			Unit:  UnitFahrenheit,
		},
		RespiratoryRate: LabValue{
			Value: float64(rand.Intn(NormalVitalRanges.RespiratoryRate.Max-NormalVitalRanges.RespiratoryRate.Min) + NormalVitalRanges.RespiratoryRate.Min),
			Unit:  UnitBreathsMin,
		},
	}
}

// GenerateMedicalRecord generates a complete medical record
func (g *HealthGenerator) GenerateMedicalRecord() MedicalRecord {
	// Generate a random number of conditions (1-3)
	numConditions := rand.Intn(3) + 1
	conditions := make([]MedicalCondition, numConditions)
	for i := 0; i < numConditions; i++ {
		conditions[i] = g.GenerateMedicalCondition()
	}

	// Generate a random number of medications (0-3)
	numMeds := rand.Intn(4)
	medications := make([]Medication, numMeds)
	for i := 0; i < numMeds; i++ {
		medications[i] = g.GenerateMedication()
	}

	// Generate a random number of allergies (0-2)
	numAllergies := rand.Intn(3)
	allergies := make([]Allergy, numAllergies)
	for i := 0; i < numAllergies; i++ {
		allergies[i] = g.GenerateAllergy()
	}

	// Generate a random number of symptoms (1-3)
	numSymptoms := rand.Intn(3) + 1
	symptoms := make([]Symptom, numSymptoms)
	for i := 0; i < numSymptoms; i++ {
		symptoms[i] = g.GenerateSymptom()
	}

	// Generate a random number of diagnoses (1-2)
	numDiagnoses := rand.Intn(2) + 1
	diagnoses := make([]Diagnosis, numDiagnoses)
	for i := 0; i < numDiagnoses; i++ {
		diagnoses[i] = g.GenerateDiagnosis()
	}

	return MedicalRecord{
		PatientID:       fmt.Sprintf("P%08d", rand.Intn(100000000)),
		BloodType:       g.GenerateBloodType(),
		Conditions:      conditions,
		Medications:     medications,
		Allergies:       allergies,
		CurrentSymptoms: symptoms,
		Diagnoses:       diagnoses,
		LabResults:      g.GenerateLabResult(),
		VitalSigns:      g.GenerateVitalSigns(),
		LastUpdated:     time.Now().Format(time.RFC3339),
	}
}

// GetAvailableTypes returns all available health-related types
func (g *HealthGenerator) GetAvailableTypes() []string {
	return []string{
		"blood_type",
		"medical_condition",
		"medication",
		"symptom",
		"diagnosis",
		"allergy",
		"lab_result",
		"vital_sign",
		"medical_record",
	}
}
