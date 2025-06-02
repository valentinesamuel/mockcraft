package health

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/valentinesamuel/mockcraft/internal/generators/interfaces"
)

// HealthGenerator provides health-specific data generation
type HealthGenerator struct {
	generator interfaces.Generator
}

// NewHealthGenerator creates a new HealthGenerator instance
func NewHealthGenerator(generator interfaces.Generator) *HealthGenerator {
	return &HealthGenerator{
		generator: generator,
	}
}

// GenerateByType generates health-related data based on type string and parameters
func (g *HealthGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
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

func init() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
}
