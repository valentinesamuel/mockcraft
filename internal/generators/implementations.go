package generators

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/valentinesamuel/mockcraft/internal/generators/industries/aviation"
	"github.com/valentinesamuel/mockcraft/internal/generators/industries/health"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Base generator implementations using gofakeit
func (e *Engine) generateUUID() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.UUID()
}

func (e *Engine) generateFirstName() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.FirstName()
}

func (e *Engine) generateLastName() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.LastName()
}

func (e *Engine) generateEmail() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.Email()
}

func (e *Engine) generatePhone() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.Phone()
}

func (e *Engine) generateAddress() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.Address().Address
}

func (e *Engine) generateCompany() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.Company()
}

func (e *Engine) generateJobTitle() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.JobTitle()
}

func (e *Engine) generateDate() string {
	start := time.Now().AddDate(-1, 0, 0)
	end := time.Now()
	duration := end.Sub(start)
	randomDuration := time.Duration(e.rand.Int63n(int64(duration)))
	return start.Add(randomDuration).Format("2006-01-02")
}

func (e *Engine) generateDateTime() string {
	start := time.Now().AddDate(-1, 0, 0)
	end := time.Now()
	duration := end.Sub(start)
	randomDuration := time.Duration(e.rand.Int63n(int64(duration)))
	return start.Add(randomDuration).Format(time.RFC3339)
}

func (e *Engine) generateTime() string {
	hour := e.rand.Intn(24)
	minute := e.rand.Intn(60)
	second := e.rand.Intn(60)
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

func (e *Engine) generateTimestamp() time.Time {
	start := time.Now().AddDate(-1, 0, 0)
	end := time.Now()
	duration := end.Sub(start)
	randomDuration := time.Duration(e.rand.Int63n(int64(duration)))
	return start.Add(randomDuration)
}

func (e *Engine) generateNumber(params map[string]interface{}) (interface{}, error) {
	min := 0
	max := 100
	if minVal, ok := params["min"].(int); ok {
		min = minVal
	}
	if maxVal, ok := params["max"].(int); ok {
		max = maxVal
	}
	return min + e.rand.Intn(max-min+1), nil
}

func (e *Engine) generateFloat(params map[string]interface{}) (interface{}, error) {
	min := 0.0
	max := 100.0
	if minVal, ok := params["min"].(float64); ok {
		min = minVal
	}
	if maxVal, ok := params["max"].(float64); ok {
		max = maxVal
	}
	return min + e.rand.Float64()*(max-min), nil
}

func (e *Engine) generateBoolean() bool {
	return e.rand.Float32() < 0.5
}

func (e *Engine) generateEnum(params map[string]interface{}) (interface{}, error) {
	values, ok := params["values"].([]string)
	if !ok {
		return nil, fmt.Errorf("enum generator requires 'values' parameter as a string slice")
	}
	
	if len(values) == 0 {
		return nil, fmt.Errorf("enum generator requires at least one value in 'values' parameter")
	}
	
	// Select a random value from the provided list
	index := e.rand.Intn(len(values))
	return values[index], nil
}

func (e *Engine) generateText(params map[string]interface{}) (interface{}, error) {
	minLength := 10
	maxLength := 100
	if minVal, ok := params["min"].(int); ok {
		minLength = minVal
	}
	if maxVal, ok := params["max"].(int); ok {
		maxLength = maxVal
	}
	
	length := minLength + e.rand.Intn(maxLength-minLength+1)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[e.rand.Intn(len(charset))]
	}
	return string(result), nil
}

func (e *Engine) generateParagraph(params map[string]interface{}) (interface{}, error) {
	count := 1
	if countVal, ok := params["count"].(int); ok {
		count = countVal
	}
	
	sentences := make([]string, count)
	for i := 0; i < count; i++ {
		sentences[i] = e.generateSentence()
	}
	return strings.Join(sentences, " "), nil
}

func (e *Engine) generateSentence() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.Sentence(10)
}

func (e *Engine) generateWord() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.Word()
}

func (e *Engine) generateChar() string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return string(chars[e.rand.Intn(len(chars))])
}

func (e *Engine) generateURL() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.URL()
}

func (e *Engine) generateIP() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.IPv4Address()
}

func (e *Engine) generateDomain() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.DomainName()
}

func (e *Engine) generateUsername() string {
	gofakeit.SetGlobalFaker(gofakeit.New(e.rand.Int63()))
	return gofakeit.Username()
}

func (e *Engine) generatePassword(params map[string]interface{}) (interface{}, error) {
	length := 12
	if lengthVal, ok := params["length"].(int); ok {
		length = lengthVal
	}
	
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	special := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	allChars := lowercase + uppercase + numbers + special
	
	password := make([]byte, length)
	// Ensure at least one from each category
	if length >= 4 {
		password[0] = lowercase[e.rand.Intn(len(lowercase))]
		password[1] = uppercase[e.rand.Intn(len(uppercase))]
		password[2] = numbers[e.rand.Intn(len(numbers))]
		password[3] = special[e.rand.Intn(len(special))]
		
		for i := 4; i < length; i++ {
			password[i] = allChars[e.rand.Intn(len(allChars))]
		}
	} else {
		for i := 0; i < length; i++ {
			password[i] = allChars[e.rand.Intn(len(allChars))]
		}
	}
	
	// Shuffle
	e.rand.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})
	
	return string(password), nil
}

// Aviation generator implementations
func (e *Engine) generateAircraftType() string {
	return string(aviation.ValidAircraftTypes[e.rand.Intn(len(aviation.ValidAircraftTypes))])
}

func (e *Engine) generateFlightNumber() string {
	airlines := []string{"AA", "UA", "DL", "BA", "LH", "AF", "KL", "SQ", "EK", "QF"}
	airline := airlines[e.rand.Intn(len(airlines))]
	number := e.rand.Intn(9000) + 1000
	return fmt.Sprintf("%s%d", airline, number)
}

func (e *Engine) generateAirportCode() string {
	return string(aviation.ValidAirportCodes[e.rand.Intn(len(aviation.ValidAirportCodes))])
}

func (e *Engine) generateGateNumber() string {
	terminals := []string{"A", "B", "C", "D", "E", "F", "G", "H", "T"}
	terminal := terminals[e.rand.Intn(len(terminals))]
	number := e.rand.Intn(99) + 1
	return fmt.Sprintf("%s%d", terminal, number)
}

func (e *Engine) generateSeatNumber() string {
	rows := e.rand.Intn(40) + 1
	seats := []string{"A", "B", "C", "D", "E", "F"}
	seat := seats[e.rand.Intn(len(seats))]
	return fmt.Sprintf("%d%s", rows, seat)
}

func (e *Engine) generateFlightStatus() string {
	return string(aviation.ValidFlightStatuses[e.rand.Intn(len(aviation.ValidFlightStatuses))])
}

func (e *Engine) generateBaggageClaim() string {
	return string(aviation.ValidBaggageClaims[e.rand.Intn(len(aviation.ValidBaggageClaims))])
}

func (e *Engine) generateFlightSchedule() string {
	departure := time.Now().Add(time.Duration(e.rand.Intn(24)) * time.Hour)
	arrival := departure.Add(time.Duration(e.rand.Intn(12)+1) * time.Hour)
	
	schedule := aviation.Schedule{
		DepartureTime: departure.Format("15:04"),
		ArrivalTime:   arrival.Format("15:04"),
		DepartureDate: departure.Format("2006-01-02"),
		ArrivalDate:   arrival.Format("2006-01-02"),
	}
	
	// Convert to JSON string for database storage
	jsonBytes, err := json.Marshal(schedule)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to generate flight schedule: %s"}`, err.Error())
	}
	return string(jsonBytes)
}

func (e *Engine) generateFlightInfo() string {
	flightInfo := aviation.FlightInfo{
		FlightNumber: aviation.Flight(e.generateFlightNumber()),
		AircraftType: aviation.Aircraft(e.generateAircraftType()),
		Origin:       aviation.Airport(e.generateAirportCode()),
		Destination:  aviation.Airport(e.generateAirportCode()),
		Gate:         aviation.Gate(e.generateGateNumber()),
		Status:       aviation.Status(e.generateFlightStatus()),
		BaggageClaim: aviation.Baggage(e.generateBaggageClaim()),
	}
	
	// Convert to JSON string for database storage
	jsonBytes, err := json.Marshal(flightInfo)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to generate flight info: %s"}`, err.Error())
	}
	return string(jsonBytes)
}

// Health generator implementations
func (e *Engine) generateBloodType() string {
	return string(health.ValidBloodTypes[e.rand.Intn(len(health.ValidBloodTypes))])
}

func (e *Engine) generateMedicalCondition() string {
	return string(health.CommonConditions[e.rand.Intn(len(health.CommonConditions))])
}

func (e *Engine) generateMedication() string {
	return string(health.CommonMedications[e.rand.Intn(len(health.CommonMedications))])
}

func (e *Engine) generateSymptom() string {
	return string(health.CommonSymptoms[e.rand.Intn(len(health.CommonSymptoms))])
}

func (e *Engine) generateDiagnosis() string {
	return string(health.CommonDiagnoses[e.rand.Intn(len(health.CommonDiagnoses))])
}

func (e *Engine) generateAllergy() string {
	return string(health.CommonAllergies[e.rand.Intn(len(health.CommonAllergies))])
}

func (e *Engine) generateLabResult() string {
	labResult := health.LabResult{
		Glucose: health.LabValue{
			Value: health.NormalLabRanges.Glucose.Min + e.rand.Float64()*(health.NormalLabRanges.Glucose.Max-health.NormalLabRanges.Glucose.Min),
			Unit:  health.UnitMgDL,
		},
		Cholesterol: health.LabValue{
			Value: health.NormalLabRanges.Cholesterol.Min + e.rand.Float64()*(health.NormalLabRanges.Cholesterol.Max-health.NormalLabRanges.Cholesterol.Min),
			Unit:  health.UnitMgDL,
		},
		Hemoglobin: health.LabValue{
			Value: health.NormalLabRanges.Hemoglobin.Min + e.rand.Float64()*(health.NormalLabRanges.Hemoglobin.Max-health.NormalLabRanges.Hemoglobin.Min),
			Unit:  health.UnitGDL,
		},
	}
	
	// Convert to JSON string for database storage
	jsonBytes, err := json.Marshal(labResult)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to generate lab result: %s"}`, err.Error())
	}
	return string(jsonBytes)
}

func (e *Engine) generateVitalSigns() string {
	vitalSign := health.VitalSign{
		BloodPressure: health.BloodPressure{
			Systolic:  e.rand.Intn(health.NormalVitalRanges.BloodPressure.SystolicMax-health.NormalVitalRanges.BloodPressure.SystolicMin) + health.NormalVitalRanges.BloodPressure.SystolicMin,
			Diastolic: e.rand.Intn(health.NormalVitalRanges.BloodPressure.DiastolicMax-health.NormalVitalRanges.BloodPressure.DiastolicMin) + health.NormalVitalRanges.BloodPressure.DiastolicMin,
			Unit:      health.UnitMmHg,
		},
		HeartRate: health.LabValue{
			Value: float64(e.rand.Intn(health.NormalVitalRanges.HeartRate.Max-health.NormalVitalRanges.HeartRate.Min) + health.NormalVitalRanges.HeartRate.Min),
			Unit:  health.UnitBPM,
		},
		Temperature: health.LabValue{
			Value: health.NormalVitalRanges.Temperature.Min + e.rand.Float64()*(health.NormalVitalRanges.Temperature.Max-health.NormalVitalRanges.Temperature.Min),
			Unit:  health.UnitFahrenheit,
		},
		RespiratoryRate: health.LabValue{
			Value: float64(e.rand.Intn(health.NormalVitalRanges.RespiratoryRate.Max-health.NormalVitalRanges.RespiratoryRate.Min) + health.NormalVitalRanges.RespiratoryRate.Min),
			Unit:  health.UnitBreathsMin,
		},
	}
	
	// Convert to JSON string for database storage
	jsonBytes, err := json.Marshal(vitalSign)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to generate vital signs: %s"}`, err.Error())
	}
	return string(jsonBytes)
}

func (e *Engine) generateMedicalRecord() string {
	// Generate random collections
	numConditions := e.rand.Intn(3) + 1
	conditions := make([]health.MedicalCondition, numConditions)
	for i := 0; i < numConditions; i++ {
		conditions[i] = health.MedicalCondition(e.generateMedicalCondition())
	}
	
	numMeds := e.rand.Intn(4)
	medications := make([]health.Medication, numMeds)
	for i := 0; i < numMeds; i++ {
		medications[i] = health.Medication(e.generateMedication())
	}
	
	numAllergies := e.rand.Intn(3)
	allergies := make([]health.Allergy, numAllergies)
	for i := 0; i < numAllergies; i++ {
		allergies[i] = health.Allergy(e.generateAllergy())
	}
	
	numSymptoms := e.rand.Intn(3) + 1
	symptoms := make([]health.Symptom, numSymptoms)
	for i := 0; i < numSymptoms; i++ {
		symptoms[i] = health.Symptom(e.generateSymptom())
	}
	
	numDiagnoses := e.rand.Intn(2) + 1
	diagnoses := make([]health.Diagnosis, numDiagnoses)
	for i := 0; i < numDiagnoses; i++ {
		diagnoses[i] = health.Diagnosis(e.generateDiagnosis())
	}
	
	medicalRecord := health.MedicalRecord{
		PatientID:       fmt.Sprintf("P%08d", e.rand.Intn(100000000)),
		BloodType:       health.BloodType(e.generateBloodType()),
		Conditions:      conditions,
		Medications:     medications,
		Allergies:       allergies,
		CurrentSymptoms: symptoms,
		Diagnoses:       diagnoses,
		LastUpdated:     time.Now().Format(time.RFC3339),
	}
	
	// Convert to JSON string for database storage
	jsonBytes, err := json.Marshal(medicalRecord)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to generate medical record: %s"}`, err.Error())
	}
	return string(jsonBytes)
}

// MongoDB-specific generator implementations
func (e *Engine) generateMongoDBObjectID() string {
	return primitive.NewObjectID().Hex()
}

func (e *Engine) generateMongoDBDecimal128(params map[string]interface{}) (interface{}, error) {
	min := 0.0
	max := 1000000.0
	precision := 2
	
	if minVal, ok := params["min"].(float64); ok {
		min = minVal
	}
	if maxVal, ok := params["max"].(float64); ok {
		max = maxVal
	}
	if precisionVal, ok := params["precision"].(int); ok {
		precision = precisionVal
	}
	
	value := min + e.rand.Float64()*(max-min)
	multiplier := math.Pow(10, float64(precision))
	value = math.Round(value*multiplier) / multiplier
	
	// Create a Decimal128 from the float value
	bigFloat := big.NewFloat(value)
	bigInt, _ := bigFloat.Int(nil)
	decimal, err := primitive.ParseDecimal128(bigInt.String())
	if err != nil {
		// Fallback to string representation
		return fmt.Sprintf("%.2f", value), nil
	}
	return decimal, nil
}

func (e *Engine) generateMongoDBBinary(params map[string]interface{}) (interface{}, error) {
	subtype := "generic"
	size := 16
	
	if subtypeVal, ok := params["subtype"].(string); ok {
		subtype = subtypeVal
	}
	if sizeVal, ok := params["size"].(int); ok {
		size = sizeVal
	}
	
	switch subtype {
	case "uuid":
		// Generate UUID binary (16 bytes)
		uuid := primitive.NewObjectID()
		return primitive.Binary{Subtype: 0x04, Data: uuid[:]}, nil
	case "md5":
		// Generate MD5 hash (16 bytes)
		data := make([]byte, 32)
		e.rand.Read(data)
		hash := md5.Sum(data)
		return primitive.Binary{Subtype: 0x05, Data: hash[:]}, nil
	case "user_defined":
		// Generate custom binary data
		data := make([]byte, size)
		e.rand.Read(data)
		return primitive.Binary{Subtype: 0x80, Data: data}, nil
	default: // generic
		// Generate generic binary data
		data := make([]byte, size)
		e.rand.Read(data)
		return primitive.Binary{Subtype: 0x00, Data: data}, nil
	}
}

func (e *Engine) generateMongoDBTimestamp() primitive.Timestamp {
	// MongoDB timestamp is an internal type with T (time) and I (increment)
	t := uint32(time.Now().Unix())
	i := uint32(e.rand.Int31())
	return primitive.Timestamp{T: t, I: i}
}

func (e *Engine) generateMongoDBRegex() primitive.Regex {
	patterns := []string{
		"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", // Email
		"^\\d{3}-\\d{2}-\\d{4}$",                            // SSN
		"^\\d{10}$",                                         // Phone
		"^[A-Z]{2}\\d{6}$",                                  // ID
		"^\\w{3,20}$",                                       // Username
	}
	
	options := []string{"i", "m", "s", "x", ""}
	
	pattern := patterns[e.rand.Intn(len(patterns))]
	option := options[e.rand.Intn(len(options))]
	
	return primitive.Regex{Pattern: pattern, Options: option}
}

func (e *Engine) generateMongoDBJavaScript() string {
	jsFunctions := []string{
		"function(x) { return x * 2; }",
		"function(a, b) { return a + b; }",
		"function() { return new Date(); }",
		"function(str) { return str.toUpperCase(); }",
		"function(arr) { return arr.length; }",
	}
	
	return jsFunctions[e.rand.Intn(len(jsFunctions))]
}

func (e *Engine) generateMongoDBMinKey() primitive.MinKey {
	return primitive.MinKey{}
}

func (e *Engine) generateMongoDBMaxKey() primitive.MaxKey {
	return primitive.MaxKey{}
}

func (e *Engine) generateNull() interface{} {
	return nil
}

func (e *Engine) generateRandomInt(params map[string]interface{}) (interface{}, error) {
	min := 0
	max := 100
	if minVal, ok := params["min"].(int); ok {
		min = minVal
	}
	if maxVal, ok := params["max"].(int); ok {
		max = maxVal
	}
	return min + e.rand.Intn(max-min+1), nil
}

func (e *Engine) generateEmbeddedDocument(params map[string]interface{}) (interface{}, error) {
	// This will be handled by the schema parser for nested_fields
	// For now, return an empty document
	return map[string]interface{}{}, nil
}

func (e *Engine) generateArrayOfStrings(params map[string]interface{}) (interface{}, error) {
	minCount := 1
	maxCount := 5
	nestedGenerator := "word"
	
	if minVal, ok := params["min_count"].(int); ok {
		minCount = minVal
	}
	if maxVal, ok := params["max_count"].(int); ok {
		maxCount = maxVal
	}
	if genVal, ok := params["nested_generator"].(string); ok {
		nestedGenerator = genVal
	}
	
	count := minCount + e.rand.Intn(maxCount-minCount+1)
	result := make([]string, count)
	
	for i := 0; i < count; i++ {
		switch nestedGenerator {
		case "allergy":
			result[i] = e.generateAllergy()
		case "airport_code":
			result[i] = e.generateAirportCode()
		case "word":
			result[i] = e.generateWord()
		case "sentence":
			result[i] = e.generateSentence()
		default:
			result[i] = e.generateWord()
		}
	}
	
	return result, nil
}