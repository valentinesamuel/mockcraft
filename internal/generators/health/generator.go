package health

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/valentinesamuel/mockcraft/internal/generators/base"
)

// MedicalGenerator extends BaseGenerator with health-specific data generation
type MedicalGenerator struct {
	*base.BaseGenerator
}

// NewMedicalGenerator creates a new medical generator instance
func NewMedicalGenerator() *MedicalGenerator {
	return &MedicalGenerator{
		BaseGenerator: base.NewBaseGenerator(),
	}
}

// GenerateByType generates medical data based on type string and parameters
func (mg *MedicalGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
	switch dataType {
	case "systolic":
		return mg.Faker.IntRange(90, 140), nil
	case "diastolic":
		return mg.Faker.IntRange(60, 90), nil
	case "blood_pressure_unit":
		return "mmHg", nil
	case "heart_rate":
		return mg.Faker.IntRange(60, 100), nil
	case "heart_rate_unit":
		return "bpm", nil
	case "temperature":
		return mg.Faker.Float64Range(97.0, 99.0), nil
	case "temperature_unit":
		return "°F", nil
	case "respiratory_rate":
		return mg.Faker.IntRange(12, 20), nil
	case "respiratory_unit":
		return "breaths/min", nil
	case "blood_type":
		return mg.generateBloodType(), nil
	case "medical_condition":
		return mg.generateMedicalCondition(), nil
	case "medication":
		return mg.generateMedication(), nil
	case "symptom":
		return mg.generateSymptom(), nil
	case "diagnosis":
		return mg.generateDiagnosis(), nil
	case "allergy":
		return mg.generateAllergy(), nil
	case "lab_result":
		return mg.generateLabResult(params), nil
	case "vital_sign":
		return mg.generateVitalSign(params), nil
	case "medical_record":
		return mg.generateMedicalRecord(params), nil
	default:
		return nil, fmt.Errorf("unknown medical type: %s", dataType)
	}
}

// GetAvailableTypes returns all supported medical types
func (mg *MedicalGenerator) GetAvailableTypes() []string {
	baseTypes := mg.BaseGenerator.GetAvailableTypes()
	medicalTypes := []string{
		"systolic",
		"diastolic",
		"blood_pressure_unit",
		"heart_rate",
		"heart_rate_unit",
		"temperature",
		"temperature_unit",
		"respiratory_rate",
		"respiratory_unit",
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
	return append(baseTypes, medicalTypes...)
}

// Private helper methods for medical data generation

func (mg *MedicalGenerator) generateBloodType() BloodType {
	return ValidBloodTypes[rand.Intn(len(ValidBloodTypes))]
}

func (mg *MedicalGenerator) generateMedicalCondition() MedicalCondition {
	return CommonConditions[rand.Intn(len(CommonConditions))]
}

func (mg *MedicalGenerator) generateMedication() Medication {
	return CommonMedications[rand.Intn(len(CommonMedications))]
}

func (mg *MedicalGenerator) generateSymptom() Symptom {
	return CommonSymptoms[rand.Intn(len(CommonSymptoms))]
}

func (mg *MedicalGenerator) generateDiagnosis() Diagnosis {
	return CommonDiagnoses[rand.Intn(len(CommonDiagnoses))]
}

func (mg *MedicalGenerator) generateAllergy() Allergy {
	return CommonAllergies[rand.Intn(len(CommonAllergies))]
}

func (mg *MedicalGenerator) generateLabResult(params map[string]interface{}) LabResult {
	// Get parameters with defaults
	unit := mg.getStringParam(params, "unit", UnitMgDL)
	faker := mg.Faker

	// Generate random lab values within normal ranges
	return LabResult{
		Glucose: LabValue{
			Value: faker.Float64Range(NormalLabRanges.Glucose.Min, NormalLabRanges.Glucose.Max),
			Unit:  unit,
		},
		Cholesterol: LabValue{
			Value: faker.Float64Range(NormalLabRanges.Cholesterol.Min, NormalLabRanges.Cholesterol.Max),
			Unit:  unit,
		},
		Hemoglobin: LabValue{
			Value: faker.Float64Range(NormalLabRanges.Hemoglobin.Min, NormalLabRanges.Hemoglobin.Max),
			Unit:  UnitGDL,
		},
	}
}

func (mg *MedicalGenerator) generateVitalSign(params map[string]interface{}) VitalSign {
	faker := mg.Faker
	return VitalSign{
		BloodPressure: BloodPressure{
			Systolic:  faker.IntRange(NormalVitalRanges.BloodPressure.SystolicMin, NormalVitalRanges.BloodPressure.SystolicMax),
			Diastolic: faker.IntRange(NormalVitalRanges.BloodPressure.DiastolicMin, NormalVitalRanges.BloodPressure.DiastolicMax),
			Unit:      UnitMmHg,
		},
		HeartRate: LabValue{
			Value: float64(faker.IntRange(NormalVitalRanges.HeartRate.Min, NormalVitalRanges.HeartRate.Max)),
			Unit:  UnitBPM,
		},
		Temperature: LabValue{
			Value: faker.Float64Range(NormalVitalRanges.Temperature.Min, NormalVitalRanges.Temperature.Max),
			Unit:  UnitFahrenheit,
		},
		RespiratoryRate: LabValue{
			Value: float64(faker.IntRange(NormalVitalRanges.RespiratoryRate.Min, NormalVitalRanges.RespiratoryRate.Max)),
			Unit:  UnitBreathsMin,
		},
	}
}

func (mg *MedicalGenerator) generateMedicalRecord(params map[string]interface{}) MedicalRecord {
	faker := mg.Faker

	// Generate random number of conditions, medications, etc.
	numConditions := faker.IntRange(0, 3)
	numMedications := faker.IntRange(0, 4)
	numAllergies := faker.IntRange(0, 3)
	numSymptoms := faker.IntRange(0, 4)
	numDiagnoses := faker.IntRange(0, 2)

	// Generate arrays of medical data
	conditions := make([]MedicalCondition, numConditions)
	for i := range conditions {
		conditions[i] = mg.generateMedicalCondition()
	}

	medications := make([]Medication, numMedications)
	for i := range medications {
		medications[i] = mg.generateMedication()
	}

	allergies := make([]Allergy, numAllergies)
	for i := range allergies {
		allergies[i] = mg.generateAllergy()
	}

	symptoms := make([]Symptom, numSymptoms)
	for i := range symptoms {
		symptoms[i] = mg.generateSymptom()
	}

	diagnoses := make([]Diagnosis, numDiagnoses)
	for i := range diagnoses {
		diagnoses[i] = mg.generateDiagnosis()
	}

	return MedicalRecord{
		PatientID:       faker.UUID(),
		BloodType:       mg.generateBloodType(),
		Conditions:      conditions,
		Medications:     medications,
		Allergies:       allergies,
		CurrentSymptoms: symptoms,
		Diagnoses:       diagnoses,
		LabResults:      mg.generateLabResult(params),
		VitalSigns:      mg.generateVitalSign(params),
		LastUpdated:     time.Now().Format(time.RFC3339),
	}
}

// Helper method to get string parameter with default
func (mg *MedicalGenerator) getStringParam(params map[string]interface{}, key string, defaultVal string) string {
	if val, exists := params[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}
