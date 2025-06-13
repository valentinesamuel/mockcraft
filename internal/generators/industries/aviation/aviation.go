package aviation

import (
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

// AircraftType generates a random aircraft type
func AircraftType() string {
	types := []string{
		"Boeing 737",
		"Boeing 747",
		"Boeing 777",
		"Boeing 787",
		"Airbus A320",
		"Airbus A330",
		"Airbus A350",
		"Airbus A380",
		"Embraer E190",
		"Bombardier CRJ900",
	}
	return types[gofakeit.IntRange(0, len(types)-1)]
}

// FlightNumber generates a random flight number
func FlightNumber() string {
	airlines := []string{"AA", "UA", "DL", "BA", "LH", "AF", "KL", "SQ", "EK", "QF"}
	airline := airlines[gofakeit.IntRange(0, len(airlines)-1)]
	number := gofakeit.IntRange(100, 9999)
	return fmt.Sprintf("%s%d", airline, number)
}

// AirportCode generates a random airport code
func AirportCode() string {
	codes := []string{
		"JFK", "LAX", "ORD", "ATL", "LHR", "CDG", "FRA", "AMS", "SIN", "DXB",
		"HKG", "NRT", "SYD", "MEL", "SFO", "DFW", "DEN", "SEA", "MIA", "BOS",
	}
	return codes[gofakeit.IntRange(0, len(codes)-1)]
}

// GateNumber generates a random gate number
func GateNumber() string {
	terminals := []string{"A", "B", "C", "D", "E", "F", "G", "H", "T"}
	terminal := terminals[gofakeit.IntRange(0, len(terminals)-1)]
	number := gofakeit.IntRange(1, 99)
	return fmt.Sprintf("%s%d", terminal, number)
}

// SeatNumber generates a random seat number
func SeatNumber() string {
	rows := gofakeit.IntRange(1, 40)
	seats := []string{"A", "B", "C", "D", "E", "F"}
	seat := seats[gofakeit.IntRange(0, len(seats)-1)]
	return fmt.Sprintf("%d%s", rows, seat)
}

// FlightStatus generates a random flight status
func FlightStatus() string {
	statuses := []string{
		"On Time",
		"Delayed",
		"Boarding",
		"Departed",
		"Arrived",
		"Cancelled",
		"Diverted",
		"Gate Changed",
	}
	return statuses[gofakeit.IntRange(0, len(statuses)-1)]
}

// BaggageClaim generates a random baggage claim number
func BaggageClaim() string {
	carousels := []string{"A", "B", "C", "D", "E"}
	carousel := carousels[gofakeit.IntRange(0, len(carousels)-1)]
	return carousel
}

// FlightSchedule generates a flight schedule
func FlightSchedule() map[string]interface{} {
	departure := gofakeit.Date()
	arrival := departure.Add(time.Duration(gofakeit.IntRange(1, 12)) * time.Hour)

	return map[string]interface{}{
		"departure_time": departure.Format("15:04"),
		"arrival_time":   arrival.Format("15:04"),
		"departure_date": departure.Format("2006-01-02"),
		"arrival_date":   arrival.Format("2006-01-02"),
	}
}
