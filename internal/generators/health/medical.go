package health

import (
	"github.com/brianvoe/gofakeit/v6"
)

// Systolic generates a systolic blood pressure value
func Systolic() int {
	return gofakeit.IntRange(90, 140)
}

// Diastolic generates a diastolic blood pressure value
func Diastolic() int {
	return gofakeit.IntRange(60, 90)
}

// BloodPressureUnit returns the unit for blood pressure
func BloodPressureUnit() string {
	return "mmHg"
}

// HeartRate generates a heart rate value
func HeartRate() int {
	return gofakeit.IntRange(60, 100)
}

// HeartRateUnit returns the unit for heart rate
func HeartRateUnit() string {
	return "bpm"
}

// Temperature generates a body temperature value
func Temperature() float64 {
	return gofakeit.Float64Range(97.0, 99.0)
}

// TemperatureUnit returns the unit for temperature
func TemperatureUnit() string {
	return "°F"
}

// RespiratoryRate generates a respiratory rate value
func RespiratoryRate() int {
	return gofakeit.IntRange(12, 20)
}

// RespiratoryRateUnit returns the unit for respiratory rate
func RespiratoryRateUnit() string {
	return "breaths/min"
}

// VitalSigns generates all vital signs as a map
func VitalSigns() map[string]interface{} {
	return map[string]interface{}{
		"blood_pressure": map[string]interface{}{
			"systolic":  Systolic(),
			"diastolic": Diastolic(),
			"unit":      BloodPressureUnit(),
		},
		"heart_rate": map[string]interface{}{
			"value": HeartRate(),
			"unit":  HeartRateUnit(),
		},
		"temperature": map[string]interface{}{
			"value": Temperature(),
			"unit":  TemperatureUnit(),
		},
		"respiratory_rate": map[string]interface{}{
			"value": RespiratoryRate(),
			"unit":  RespiratoryRateUnit(),
		},
	}
}
