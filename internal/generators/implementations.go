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

// Parameter-aware generator implementations for VARCHAR support

func (e *Engine) truncateToLength(str string, params map[string]interface{}) string {
	if maxLength, ok := params["max_length"]; ok {
		if maxLenInt, ok := maxLength.(int); ok {
			if len(str) > maxLenInt {
				return str[:maxLenInt]
			}
		}
	}
	return str
}

func (e *Engine) generateFirstNameWithParams(params map[string]interface{}) (interface{}, error) {
	name := e.generateFirstName()
	return e.truncateToLength(name, params), nil
}

func (e *Engine) generateLastNameWithParams(params map[string]interface{}) (interface{}, error) {
	name := e.generateLastName()
	return e.truncateToLength(name, params), nil
}

func (e *Engine) generateEmailWithParams(params map[string]interface{}) (interface{}, error) {
	email := e.generateEmail()
	return e.truncateToLength(email, params), nil
}

func (e *Engine) generatePhoneWithParams(params map[string]interface{}) (interface{}, error) {
	phone := e.generatePhone()
	return e.truncateToLength(phone, params), nil
}

func (e *Engine) generateAddressWithParams(params map[string]interface{}) (interface{}, error) {
	address := e.generateAddress()
	return e.truncateToLength(address, params), nil
}

func (e *Engine) generateCompanyWithParams(params map[string]interface{}) (interface{}, error) {
	company := e.generateCompany()
	return e.truncateToLength(company, params), nil
}

func (e *Engine) generateWordWithParams(params map[string]interface{}) (interface{}, error) {
	word := e.generateWord()
	return e.truncateToLength(word, params), nil
}

// PostgreSQL-specific generator implementations

func (e *Engine) generateJSON(params map[string]interface{}) (interface{}, error) {
	complexity := "simple"
	if comp, ok := params["complexity"].(string); ok {
		complexity = comp
	}
	
	switch complexity {
	case "simple":
		return fmt.Sprintf(`{"name": "%s", "value": %d}`, e.generateWord(), e.rand.Intn(100)), nil
	case "complex":
		return fmt.Sprintf(`{"user": {"name": "%s", "email": "%s"}, "settings": {"theme": "dark", "notifications": true}, "data": [1, 2, 3]}`, 
			e.generateFirstName(), e.generateEmail()), nil
	case "array":
		return `[{"id": 1, "name": "item1"}, {"id": 2, "name": "item2"}]`, nil
	default:
		return `{"default": true}`, nil
	}
}

func (e *Engine) generateJSONB(params map[string]interface{}) (interface{}, error) {
	// JSONB is same as JSON but stored in binary format
	return e.generateJSON(params)
}

func (e *Engine) generateInet() string {
	return fmt.Sprintf("%d.%d.%d.%d", e.rand.Intn(256), e.rand.Intn(256), e.rand.Intn(256), e.rand.Intn(256))
}

func (e *Engine) generateCIDR() string {
	// Generate valid network address with proper subnet calculation
	prefix := e.rand.Intn(25) + 8 // /8 to /32
	
	// Calculate how many bits are for network vs host
	networkBits := prefix
	
	// Generate a random IP and mask it to get the network address
	a := e.rand.Intn(256)
	b := e.rand.Intn(256)
	c := e.rand.Intn(256)
	d := e.rand.Intn(256)
	
	// Convert to 32-bit integer
	ip := uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8 | uint32(d)
	
	// Create network mask
	mask := uint32(0xffffffff) << (32 - networkBits)
	
	// Apply mask to get network address
	networkIP := ip & mask
	
	// Convert back to octets
	na := (networkIP >> 24) & 0xff
	nb := (networkIP >> 16) & 0xff
	nc := (networkIP >> 8) & 0xff
	nd := networkIP & 0xff
	
	return fmt.Sprintf("%d.%d.%d.%d/%d", na, nb, nc, nd, prefix)
}

func (e *Engine) generateMACAddr() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		e.rand.Intn(256), e.rand.Intn(256), e.rand.Intn(256),
		e.rand.Intn(256), e.rand.Intn(256), e.rand.Intn(256))
}

func (e *Engine) generateBytea(params map[string]interface{}) (interface{}, error) {
	size := 16
	if sizeVal, ok := params["size"].(int); ok {
		size = sizeVal
	}
	
	data := make([]byte, size)
	e.rand.Read(data)
	return fmt.Sprintf("\\x%x", data), nil
}

func (e *Engine) generateMoney(params map[string]interface{}) (interface{}, error) {
	min := 0.0
	max := 1000.0
	if minVal, ok := params["min"].(float64); ok {
		min = minVal
	}
	if maxVal, ok := params["max"].(float64); ok {
		max = maxVal
	}
	
	amount := min + e.rand.Float64()*(max-min)
	return fmt.Sprintf("$%.2f", amount), nil
}

func (e *Engine) generateInterval(params map[string]interface{}) (interface{}, error) {
	intervalType := "hours"
	if typeVal, ok := params["type"].(string); ok {
		intervalType = typeVal
	}
	
	switch intervalType {
	case "hours":
		hours := e.rand.Intn(24) + 1
		return fmt.Sprintf("%d hours", hours), nil
	case "days":
		days := e.rand.Intn(365) + 1
		return fmt.Sprintf("%d days", days), nil
	case "months":
		months := e.rand.Intn(12) + 1
		return fmt.Sprintf("%d months", months), nil
	case "years":
		years := e.rand.Intn(10) + 1
		return fmt.Sprintf("%d years", years), nil
	default:
		return "1 hour", nil
	}
}

func (e *Engine) generateSerial() int {
	return e.rand.Intn(1000000) + 1
}

func (e *Engine) generateBigSerial() int64 {
	return e.rand.Int63n(1000000000000) + 1
}

func (e *Engine) generatePoint() string {
	x := e.rand.Float64() * 100
	y := e.rand.Float64() * 100
	return fmt.Sprintf("(%.2f,%.2f)", x, y)
}

func (e *Engine) generateLine() string {
	// Line format: {A,B,C} where Ax + By + C = 0
	a := e.rand.Float64() * 10
	b := e.rand.Float64() * 10
	c := e.rand.Float64() * 10
	return fmt.Sprintf("{%.2f,%.2f,%.2f}", a, b, c)
}

func (e *Engine) generateCircle() string {
	// Circle format: <(x,y),r>
	x := e.rand.Float64() * 100
	y := e.rand.Float64() * 100
	r := e.rand.Float64() * 50
	return fmt.Sprintf("<(%.2f,%.2f),%.2f>", x, y, r)
}

func (e *Engine) generatePolygon() string {
	// Simple triangle polygon
	x1, y1 := e.rand.Float64()*100, e.rand.Float64()*100
	x2, y2 := e.rand.Float64()*100, e.rand.Float64()*100
	x3, y3 := e.rand.Float64()*100, e.rand.Float64()*100
	return fmt.Sprintf("((%.2f,%.2f),(%.2f,%.2f),(%.2f,%.2f))", x1, y1, x2, y2, x3, y3)
}

func (e *Engine) generatePath() string {
	// Path format: [(x1,y1),...,(xn,yn)] or ((x1,y1),...,(xn,yn))
	x1, y1 := e.rand.Float64()*100, e.rand.Float64()*100
	x2, y2 := e.rand.Float64()*100, e.rand.Float64()*100
	x3, y3 := e.rand.Float64()*100, e.rand.Float64()*100
	return fmt.Sprintf("[(%.2f,%.2f),(%.2f,%.2f),(%.2f,%.2f)]", x1, y1, x2, y2, x3, y3)
}

func (e *Engine) generateTSVector() string {
	words := []string{"hello", "world", "test", "example", "data"}
	selectedWords := make([]string, e.rand.Intn(3)+1)
	for i := range selectedWords {
		selectedWords[i] = words[e.rand.Intn(len(words))]
	}
	return strings.Join(selectedWords, " ")
}

func (e *Engine) generateTSQuery() string {
	words := []string{"hello", "world", "test"}
	operators := []string{" & ", " | "}
	word1 := words[e.rand.Intn(len(words))]
	word2 := words[e.rand.Intn(len(words))]
	operator := operators[e.rand.Intn(len(operators))]
	return word1 + operator + word2
}

func (e *Engine) generateHStore(params map[string]interface{}) (interface{}, error) {
	// Simple key-value pairs
	return fmt.Sprintf(`"key1"=>"%s", "key2"=>"%d"`, e.generateWord(), e.rand.Intn(100)), nil
}

func (e *Engine) generateXML() string {
	name := e.generateFirstName()
	return fmt.Sprintf("<person><name>%s</name><age>%d</age></person>", name, e.rand.Intn(80)+18)
}

func (e *Engine) generateArray(params map[string]interface{}) (interface{}, error) {
	elementType := "text"
	minSize := 1
	maxSize := 5
	
	if elemType, ok := params["element_type"].(string); ok {
		elementType = elemType
	}
	if minVal, ok := params["min_size"].(int); ok {
		minSize = minVal
	}
	if maxVal, ok := params["max_size"].(int); ok {
		maxSize = maxVal
	}
	
	size := minSize
	if maxSize > minSize {
		size += e.rand.Intn(maxSize - minSize)
	}
	
	elements := make([]string, size)
	for i := 0; i < size; i++ {
		switch elementType {
		case "integer":
			elements[i] = fmt.Sprintf("%d", e.rand.Intn(1000))
		case "text":
			elements[i] = fmt.Sprintf(`"%s"`, e.generateWord())
		default:
			elements[i] = fmt.Sprintf(`"%s"`, e.generateWord())
		}
	}
	
	return fmt.Sprintf("{%s}", strings.Join(elements, ",")), nil
}

func (e *Engine) generateIntRange() string {
	start := e.rand.Intn(100)
	end := start + e.rand.Intn(100) + 1
	return fmt.Sprintf("[%d,%d)", start, end)
}

func (e *Engine) generateInt8Range() string {
	start := e.rand.Int63n(1000000)
	end := start + e.rand.Int63n(1000000) + 1
	return fmt.Sprintf("[%d,%d)", start, end)
}

func (e *Engine) generateNumRange() string {
	start := e.rand.Float64() * 100
	end := start + e.rand.Float64()*100 + 1
	return fmt.Sprintf("[%.2f,%.2f)", start, end)
}

func (e *Engine) generateTSRange() string {
	now := time.Now()
	start := now.Add(-time.Duration(e.rand.Intn(24)) * time.Hour)
	end := start.Add(time.Duration(e.rand.Intn(12)+1) * time.Hour)
	return fmt.Sprintf("[\"%s\",\"%s\")", start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"))
}

func (e *Engine) generateTSTZRange() string {
	now := time.Now()
	start := now.Add(-time.Duration(e.rand.Intn(24)) * time.Hour)
	end := start.Add(time.Duration(e.rand.Intn(12)+1) * time.Hour)
	return fmt.Sprintf("[\"%s\",\"%s\")", start.Format(time.RFC3339), end.Format(time.RFC3339))
}

func (e *Engine) generateDateRange() string {
	now := time.Now()
	start := now.AddDate(0, 0, -e.rand.Intn(30))
	end := start.AddDate(0, 0, e.rand.Intn(30)+1)
	return fmt.Sprintf("[%s,%s)", start.Format("2006-01-02"), end.Format("2006-01-02"))
}

func (e *Engine) generateBit(params map[string]interface{}) (interface{}, error) {
	length := 8
	if lengthVal, ok := params["length"].(int); ok {
		length = lengthVal
	}
	
	bits := make([]string, length)
	for i := 0; i < length; i++ {
		bits[i] = fmt.Sprintf("%d", e.rand.Intn(2))
	}
	return strings.Join(bits, ""), nil
}

func (e *Engine) generateBox() string {
	// Box format: (x1,y1),(x2,y2)
	x1, y1 := e.rand.Float64()*100, e.rand.Float64()*100
	x2, y2 := x1+e.rand.Float64()*50, y1+e.rand.Float64()*50
	return fmt.Sprintf("(%.2f,%.2f),(%.2f,%.2f)", x1, y1, x2, y2)
}

func (e *Engine) generateLseg() string {
	// Line segment format: [(x1,y1),(x2,y2)]
	x1, y1 := e.rand.Float64()*100, e.rand.Float64()*100
	x2, y2 := e.rand.Float64()*100, e.rand.Float64()*100
	return fmt.Sprintf("[(%.2f,%.2f),(%.2f,%.2f)]", x1, y1, x2, y2)
}

// SQLite-specific generator implementations
// SQLite stores complex types as TEXT, INTEGER, REAL, or BLOB

func (e *Engine) generateSQLiteJSON(params map[string]interface{}) (interface{}, error) {
	// Generate JSON stored as TEXT in SQLite
	complexity := "simple"
	if complexityVal, ok := params["complexity"].(string); ok {
		complexity = complexityVal
	}
	
	switch complexity {
	case "simple":
		return fmt.Sprintf(`{"name": "%s", "value": %d}`, e.generateFirstName(), e.rand.Intn(100)), nil
	case "complex":
		return fmt.Sprintf(`{"user": {"name": "%s", "email": "%s"}, "settings": {"theme": "dark", "notifications": %t}, "data": [%d, %d, %d]}`, 
			e.generateFirstName(), e.generateEmail(), e.rand.Intn(2) == 1, e.rand.Intn(100), e.rand.Intn(100), e.rand.Intn(100)), nil
	case "array":
		items := make([]string, e.rand.Intn(3)+1)
		for i := range items {
			items[i] = fmt.Sprintf(`{"id": %d, "name": "%s"}`, i+1, e.generateWord())
		}
		return fmt.Sprintf(`[%s]`, strings.Join(items, ", ")), nil
	default:
		return fmt.Sprintf(`{"id": %d, "name": "%s"}`, e.rand.Intn(1000), e.generateWord()), nil
	}
}

func (e *Engine) generateSQLiteArray(params map[string]interface{}) (interface{}, error) {
	// Generate array stored as JSON TEXT in SQLite
	elementType := "text"
	minSize := 1
	maxSize := 5
	
	if elemType, ok := params["element_type"].(string); ok {
		elementType = elemType
	}
	if minVal, ok := params["min_size"].(int); ok {
		minSize = minVal
	}
	if maxVal, ok := params["max_size"].(int); ok {
		maxSize = maxVal
	}
	
	size := minSize
	if maxSize > minSize {
		size += e.rand.Intn(maxSize - minSize)
	}
	
	elements := make([]string, size)
	for i := 0; i < size; i++ {
		switch elementType {
		case "integer":
			elements[i] = fmt.Sprintf("%d", e.rand.Intn(1000))
		case "text":
			elements[i] = fmt.Sprintf(`"%s"`, e.generateWord())
		case "boolean":
			elements[i] = fmt.Sprintf("%t", e.rand.Intn(2) == 1)
		default:
			elements[i] = fmt.Sprintf(`"%s"`, e.generateWord())
		}
	}
	
	return fmt.Sprintf("[%s]", strings.Join(elements, ", ")), nil
}

func (e *Engine) generateSQLiteBlob(params map[string]interface{}) (interface{}, error) {
	// Generate binary data for BLOB storage
	size := 16
	if sizeVal, ok := params["size"].(int); ok {
		size = sizeVal
	}
	
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(e.rand.Intn(256))
	}
	
	// Return as hex string for SQLite BLOB
	return fmt.Sprintf("X'%x'", data), nil
}

func (e *Engine) generateSQLiteAutoIncrement() int {
	// SQLite AUTOINCREMENT generates sequential integers
	// For mock data, we'll generate positive integers
	return e.rand.Intn(1000000) + 1
}

func (e *Engine) generateSQLiteDatetime() string {
	// SQLite datetime stored as TEXT in ISO8601 format
	start := time.Now().AddDate(-1, 0, 0)
	end := time.Now()
	duration := end.Sub(start)
	randomDuration := time.Duration(e.rand.Int63n(int64(duration)))
	return start.Add(randomDuration).Format("2006-01-02 15:04:05")
}

func (e *Engine) generateSQLiteDate() string {
	// SQLite date stored as TEXT in ISO8601 format
	start := time.Now().AddDate(-1, 0, 0)
	end := time.Now()
	duration := end.Sub(start)
	randomDuration := time.Duration(e.rand.Int63n(int64(duration)))
	return start.Add(randomDuration).Format("2006-01-02")
}

func (e *Engine) generateSQLiteTime() string {
	// SQLite time stored as TEXT in HH:MM:SS format
	hour := e.rand.Intn(24)
	minute := e.rand.Intn(60)
	second := e.rand.Intn(60)
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

func (e *Engine) generateSQLiteBoolean() int {
	// SQLite boolean stored as INTEGER (0 or 1)
	return e.rand.Intn(2)
}

func (e *Engine) generateSQLiteReal(params map[string]interface{}) (interface{}, error) {
	// Generate REAL (floating point) numbers for SQLite
	min := 0.0
	max := 100.0
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
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, value), nil
}

func (e *Engine) generateSQLiteDecimal(params map[string]interface{}) (interface{}, error) {
	// SQLite decimal stored as REAL or TEXT
	min := 0.0
	max := 100.0
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
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, value), nil
}

func (e *Engine) generateSQLitePoint() string {
	// SQLite geometric point stored as TEXT (JSON format)
	x := e.rand.Float64() * 100
	y := e.rand.Float64() * 100
	return fmt.Sprintf(`{"x": %.2f, "y": %.2f}`, x, y)
}

func (e *Engine) generateSQLitePolygon() string {
	// SQLite polygon stored as TEXT (JSON format)
	numPoints := e.rand.Intn(3) + 3 // 3-5 points
	points := make([]string, numPoints)
	for i := 0; i < numPoints; i++ {
		x := e.rand.Float64() * 100
		y := e.rand.Float64() * 100
		points[i] = fmt.Sprintf(`{"x": %.2f, "y": %.2f}`, x, y)
	}
	return fmt.Sprintf(`{"points": [%s]}`, strings.Join(points, ", "))
}