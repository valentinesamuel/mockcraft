package health

// BloodType represents a blood type with Rh factor
type BloodType string

// MedicalCondition represents a medical condition or disease
type MedicalCondition string

// Medication represents a prescribed medication
type Medication string

// Symptom represents a patient symptom
type Symptom string

// Diagnosis represents a medical diagnosis
type Diagnosis string

// Allergy represents a patient allergy
type Allergy string

// LabValue represents a single laboratory test value with its unit
type LabValue struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// LabResult represents a collection of laboratory test results
type LabResult struct {
	Glucose     LabValue `json:"glucose"`
	Cholesterol LabValue `json:"cholesterol"`
	Hemoglobin  LabValue `json:"hemoglobin"`
}

// BloodPressure represents systolic and diastolic blood pressure readings
type BloodPressure struct {
	Systolic  int    `json:"systolic"`
	Diastolic int    `json:"diastolic"`
	Unit      string `json:"unit"`
}

// VitalSign represents a collection of vital signs
type VitalSign struct {
	BloodPressure   BloodPressure `json:"blood_pressure"`
	HeartRate       LabValue      `json:"heart_rate"`
	Temperature     LabValue      `json:"temperature"`
	RespiratoryRate LabValue      `json:"respiratory_rate"`
}

// MedicalRecord represents a complete medical record
type MedicalRecord struct {
	PatientID       string             `json:"patient_id"`
	BloodType       BloodType          `json:"blood_type"`
	Conditions      []MedicalCondition `json:"conditions"`
	Medications     []Medication       `json:"medications"`
	Allergies       []Allergy          `json:"allergies"`
	CurrentSymptoms []Symptom          `json:"current_symptoms"`
	Diagnoses       []Diagnosis        `json:"diagnoses"`
	LabResults      LabResult          `json:"lab_results"`
	VitalSigns      VitalSign          `json:"vital_signs"`
	LastUpdated     string             `json:"last_updated"`
}

// Constants for common medical units
const (
	UnitMgDL       = "mg/dL"
	UnitMmolL      = "mmol/L"
	UnitGDL        = "g/dL"
	UnitMmHg       = "mmHg"
	UnitBPM        = "bpm"
	UnitFahrenheit = "°F"
	UnitBreathsMin = "breaths/min"
)

// Valid blood types
var ValidBloodTypes = []BloodType{
	"A+", "A-", "B+", "B-", "AB+", "AB-", "O+", "O-",
}

// Common medical conditions
var CommonConditions = []MedicalCondition{
	"Hypertension",
	"Type 2 Diabetes",
	"Asthma",
	"Arthritis",
	"Migraine",
	"Anxiety",
	"Depression",
	"Hypothyroidism",
	"GERD",
	"Osteoporosis",
	"Fibromyalgia",
	"Sleep Apnea",
}

// Common medications
var CommonMedications = []Medication{
	"Lisinopril",
	"Metformin",
	"Albuterol",
	"Ibuprofen",
	"Sumatriptan",
	"Sertraline",
	"Levothyroxine",
	"Omeprazole",
	"Alendronate",
	"Gabapentin",
	"CPAP",
	"Atorvastatin",
}

// Common symptoms
var CommonSymptoms = []Symptom{
	"Fever",
	"Cough",
	"Headache",
	"Fatigue",
	"Shortness of breath",
	"Chest pain",
	"Nausea",
	"Dizziness",
	"Joint pain",
	"Rash",
	"Sore throat",
	"Muscle aches",
}

// Common diagnoses
var CommonDiagnoses = []Diagnosis{
	"Common Cold",
	"Influenza",
	"Pneumonia",
	"Bronchitis",
	"Urinary Tract Infection",
	"Gastroenteritis",
	"Sinusitis",
	"Conjunctivitis",
	"Otitis Media",
	"Pharyngitis",
}

// Common allergies
var CommonAllergies = []Allergy{
	"Penicillin",
	"Peanuts",
	"Shellfish",
	"Latex",
	"Pollen",
	"Dust mites",
	"Pet dander",
	"Sulfa drugs",
	"Eggs",
	"Tree nuts",
	"Soy",
	"Wheat",
}

// Normal ranges for lab values
var NormalLabRanges = struct {
	Glucose     struct{ Min, Max float64 }
	Cholesterol struct{ Min, Max float64 }
	Hemoglobin  struct{ Min, Max float64 }
}{
	Glucose:     struct{ Min, Max float64 }{70, 140},
	Cholesterol: struct{ Min, Max float64 }{125, 200},
	Hemoglobin:  struct{ Min, Max float64 }{12, 17},
}

// Normal ranges for vital signs
var NormalVitalRanges = struct {
	BloodPressure struct {
		SystolicMin, SystolicMax   int
		DiastolicMin, DiastolicMax int
	}
	HeartRate       struct{ Min, Max int }
	Temperature     struct{ Min, Max float64 }
	RespiratoryRate struct{ Min, Max int }
}{
	BloodPressure: struct {
		SystolicMin, SystolicMax   int
		DiastolicMin, DiastolicMax int
	}{
		SystolicMin:  90,
		SystolicMax:  120,
		DiastolicMin: 60,
		DiastolicMax: 80,
	},
	HeartRate: struct{ Min, Max int }{
		Min: 60,
		Max: 100,
	},
	Temperature: struct{ Min, Max float64 }{
		Min: 97.0,
		Max: 99.0,
	},
	RespiratoryRate: struct{ Min, Max int }{
		Min: 12,
		Max: 20,
	},
}

// ======================

// package health

// // Blood type definitions
// type BloodType struct {
//     Type   string `json:"type"`
//     RhType string `json:"rh_type"`
// }

// var ValidBloodTypes = []BloodType{
//     {Type: "A", RhType: "positive"},
//     {Type: "A", RhType: "negative"},
//     {Type: "B", RhType: "positive"},
//     {Type: "B", RhType: "negative"},
//     {Type: "AB", RhType: "positive"},
//     {Type: "AB", RhType: "negative"},
//     {Type: "O", RhType: "positive"},
//     {Type: "O", RhType: "negative"},
// }

// // Medical condition
// type MedicalCondition struct {
//     Name        string `json:"name"`
//     Severity    string `json:"severity"`
//     Description string `json:"description"`
// }

// var CommonConditions = []MedicalCondition{
//     {Name: "Hypertension", Severity: "moderate", Description: "High blood pressure"},
//     {Name: "Diabetes", Severity: "moderate", Description: "Blood sugar regulation disorder"},
//     {Name: "Asthma", Severity: "mild", Description: "Respiratory condition"},
//     {Name: "Arthritis", Severity: "mild", Description: "Joint inflammation"},
// }

// // Medication
// type Medication struct {
//     Name     string `json:"name"`
//     Dosage   string `json:"dosage"`
//     Frequency string `json:"frequency"`
// }

// var CommonMedications = []Medication{
//     {Name: "Lisinopril", Dosage: "10mg", Frequency: "once daily"},
//     {Name: "Metformin", Dosage: "500mg", Frequency: "twice daily"},
//     {Name: "Albuterol", Dosage: "90mcg", Frequency: "as needed"},
//     {Name: "Ibuprofen", Dosage: "200mg", Frequency: "as needed"},
// }

// // Symptom
// type Symptom struct {
//     Name        string `json:"name"`
//     Severity    string `json:"severity"`
//     Duration    string `json:"duration"`
// }

// var CommonSymptoms = []Symptom{
//     {Name: "Headache", Severity: "mild", Duration: "2 hours"},
//     {Name: "Fatigue", Severity: "moderate", Duration: "1 day"},
//     {Name: "Shortness of breath", Severity: "moderate", Duration: "30 minutes"},
//     {Name: "Joint pain", Severity: "mild", Duration: "ongoing"},
// }

// // Diagnosis
// type Diagnosis struct {
//     Code        string `json:"code"`
//     Description string `json:"description"`
//     Date        string `json:"date"`
// }

// var CommonDiagnoses = []Diagnosis{
//     {Code: "I10", Description: "Essential hypertension", Date: "2024-01-15"},
//     {Code: "E11", Description: "Type 2 diabetes mellitus", Date: "2024-02-20"},
//     {Code: "J45", Description: "Asthma", Date: "2024-03-10"},
//     {Code: "M79.3", Description: "Panniculitis", Date: "2024-04-05"},
// }

// // Allergy
// type Allergy struct {
//     Allergen string `json:"allergen"`
//     Severity string `json:"severity"`
//     Reaction string `json:"reaction"`
// }

// var CommonAllergies = []Allergy{
//     {Allergen: "Peanuts", Severity: "severe", Reaction: "Anaphylaxis"},
//     {Allergen: "Penicillin", Severity: "moderate", Reaction: "Rash"},
//     {Allergen: "Shellfish", Severity: "mild", Reaction: "Hives"},
//     {Allergen: "Pollen", Severity: "mild", Reaction: "Sneezing"},
// }

// // Lab value structure
// type LabValue struct {
//     Value float64 `json:"value"`
//     Unit  string  `json:"unit"`
// }

// // Lab result structure
// type LabResult struct {
//     Glucose     LabValue `json:"glucose"`
//     Cholesterol LabValue `json:"cholesterol"`
//     Hemoglobin  LabValue `json:"hemoglobin"`
// }

// // Blood pressure structure
// type BloodPressure struct {
//     Systolic  int    `json:"systolic"`
//     Diastolic int    `json:"diastolic"`
//     Unit      string `json:"unit"`
// }

// // Vital signs structure
// type VitalSign struct {
//     BloodPressure   BloodPressure `json:"blood_pressure"`
//     HeartRate       LabValue      `json:"heart_rate"`
//     Temperature     LabValue      `json:"temperature"`
//     RespiratoryRate LabValue      `json:"respiratory_rate"`
// }

// // Medical record structure
// type MedicalRecord struct {
//     PatientID       string             `json:"patient_id"`
//     BloodType       BloodType          `json:"blood_type"`
//     Conditions      []MedicalCondition `json:"conditions"`
//     Medications     []Medication       `json:"medications"`
//     Allergies       []Allergy          `json:"allergies"`
//     CurrentSymptoms []Symptom          `json:"current_symptoms"`
//     Diagnoses       []Diagnosis        `json:"diagnoses"`
//     LabResults      LabResult          `json:"lab_results"`
//     VitalSigns      VitalSign          `json:"vital_signs"`
//     LastUpdated     string             `json:"last_updated"`
// }

// // Units
// const (
//     UnitMgDL         = "mg/dL"
//     UnitGDL          = "g/dL"
//     UnitMmHg         = "mmHg"
//     UnitBPM          = "bpm"
//     UnitFahrenheit   = "°F"
//     UnitBreathsMin   = "breaths/min"
// )

// // Normal ranges
// var NormalLabRanges = struct {
//     Glucose     struct{ Min, Max float64 }
//     Cholesterol struct{ Min, Max float64 }
//     Hemoglobin  struct{ Min, Max float64 }
// }{
//     Glucose:     struct{ Min, Max float64 }{70, 100},
//     Cholesterol: struct{ Min, Max float64 }{125, 200},
//     Hemoglobin:  struct{ Min, Max float64 }{12, 16},
// }

// var NormalVitalRanges = struct {
//     BloodPressure struct{ SystolicMin, SystolicMax, DiastolicMin, DiastolicMax int }
//     HeartRate     struct{ Min, Max int }
//     Temperature   struct{ Min, Max float64 }
//     RespiratoryRate struct{ Min, Max int }
// }{
//     BloodPressure:   struct{ SystolicMin, SystolicMax, DiastolicMin, DiastolicMax int }{90, 120, 60, 80},
//     HeartRate:       struct{ Min, Max int }{60, 100},
//     Temperature:     struct{ Min, Max float64 }{97.8, 99.1},
//     RespiratoryRate: struct{ Min, Max int }{12, 20},
// }