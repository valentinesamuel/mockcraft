package aviation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/valentinesamuel/mockcraft/internal/generators/interfaces"
	"github.com/valentinesamuel/mockcraft/internal/registry"
)

// AviationGenerator provides aviation-specific data generation
type AviationGenerator struct {
	rand *rand.Rand
}

// GenerateByType implements interfaces.Generator.
func (g *AviationGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
	if err := g.validateParameters(dataType, params); err != nil {
		return nil, err
	}

	var result interface{}
	switch dataType {
	case "aircraft_type":
		result = g.GenerateAircraftType()
	case "flight_number":
		result = g.GenerateFlightNumber()
	case "airport_code":
		result = g.GenerateAirportCode()
	case "gate_number":
		result = g.GenerateGateNumber()
	case "seat_number":
		result = g.GenerateSeatNumber()
	case "flight_status":
		result = g.GenerateFlightStatus()
	case "baggage_claim":
		result = g.GenerateBaggageClaim()
	case "flight_schedule":
		result = g.GenerateFlightSchedule()
	case "flight_info":
		result = g.GenerateFlightInfo()
	default:
		return nil, fmt.Errorf("unknown aviation type: %s", dataType)
	}

	// Check if output format is specified
	if format, ok := params["format"]; ok && format == "json" {
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal to JSON: %v", err)
		}
		return string(jsonBytes), nil
	}

	return result, nil
}

func (g *AviationGenerator) validateParameters(dataType string, params map[string]interface{}) error {
	// Define valid data types for aviation generator
	validTypes := map[string]bool{
		"aircraft_type":   true,
		"flight_number":   true,
		"airport_code":    true,
		"gate_number":     true,
		"seat_number":     true,
		"flight_status":   true,
		"baggage_claim":   true,
		"flight_schedule": true,
		"flight_info":     true,
	}

	if !validTypes[dataType] {
		return fmt.Errorf("unsupported aviation data type: %s", dataType)
	}

	return nil
}

// NewAviationGenerator creates a new AviationGenerator instance
func NewAviationGenerator() *AviationGenerator {
	return &AviationGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Auto-register when package is imported
func init() {
	registry.Register("aviation", func() interfaces.Generator {
		return NewAviationGenerator()
	})
}

// SetSeed implements interfaces.Generator
func (g *AviationGenerator) SetSeed(seed int64) {
	g.rand = rand.New(rand.NewSource(seed))
}

// GenerateAircraftType generates a random aircraft type
func (g *AviationGenerator) GenerateAircraftType() Aircraft {
	return ValidAircraftTypes[g.rand.Intn(len(ValidAircraftTypes))]
}

// GenerateFlightNumber generates a random flight number
func (g *AviationGenerator) GenerateFlightNumber() Flight {
	airlines := []string{"AA", "UA", "DL", "BA", "LH", "AF", "KL", "SQ", "EK", "QF"}
	airline := airlines[g.rand.Intn(len(airlines))]
	number := g.rand.Intn(9000) + 1000
	return Flight(fmt.Sprintf("%s%d", airline, number))
}

// GenerateAirportCode generates a random airport code
func (g *AviationGenerator) GenerateAirportCode() Airport {
	return ValidAirportCodes[g.rand.Intn(len(ValidAirportCodes))]
}

// GenerateGateNumber generates a random gate number
func (g *AviationGenerator) GenerateGateNumber() Gate {
	terminals := []string{"A", "B", "C", "D", "E", "F", "G", "H", "T"}
	terminal := terminals[g.rand.Intn(len(terminals))]
	number := g.rand.Intn(99) + 1
	return Gate(fmt.Sprintf("%s%d", terminal, number))
}

// GenerateSeatNumber generates a random seat number
func (g *AviationGenerator) GenerateSeatNumber() Seat {
	rows := g.rand.Intn(40) + 1
	seats := []string{"A", "B", "C", "D", "E", "F"}
	seat := seats[g.rand.Intn(len(seats))]
	return Seat(fmt.Sprintf("%d%s", rows, seat))
}

// GenerateFlightStatus generates a random flight status
func (g *AviationGenerator) GenerateFlightStatus() Status {
	return ValidFlightStatuses[g.rand.Intn(len(ValidFlightStatuses))]
}

// GenerateBaggageClaim generates a random baggage claim number
func (g *AviationGenerator) GenerateBaggageClaim() Baggage {
	return ValidBaggageClaims[g.rand.Intn(len(ValidBaggageClaims))]
}

// GenerateFlightSchedule generates a flight schedule
func (g *AviationGenerator) GenerateFlightSchedule() Schedule {
	departure := time.Now().Add(time.Duration(g.rand.Intn(24)) * time.Hour)
	arrival := departure.Add(time.Duration(g.rand.Intn(12)+1) * time.Hour)

	return Schedule{
		DepartureTime: departure.Format("15:04"),
		ArrivalTime:   arrival.Format("15:04"),
		DepartureDate: departure.Format("2006-01-02"),
		ArrivalDate:   arrival.Format("2006-01-02"),
	}
}

// GenerateFlightInfo generates complete flight information
func (g *AviationGenerator) GenerateFlightInfo() FlightInfo {
	return FlightInfo{
		FlightNumber: g.GenerateFlightNumber(),
		AircraftType: g.GenerateAircraftType(),
		Origin:       g.GenerateAirportCode(),
		Destination:  g.GenerateAirportCode(),
		Gate:         g.GenerateGateNumber(),
		Status:       g.GenerateFlightStatus(),
		BaggageClaim: g.GenerateBaggageClaim(),
		Schedule:     g.GenerateFlightSchedule(),
	}
}

// GetAvailableTypes returns all available aviation-related types
func (g *AviationGenerator) GetAvailableTypes() []string {
	return []string{
		"aircraft_type",
		"flight_number",
		"airport_code",
		"gate_number",
		"seat_number",
		"flight_status",
		"baggage_claim",
		"flight_schedule",
		"flight_info",
	}
}
