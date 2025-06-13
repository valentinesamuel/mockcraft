package aviation

import (
	"fmt"
)

// Aircraft represents an aircraft type
type Aircraft string

// Flight represents a flight number
type Flight string

// Airport represents an airport code
type Airport string

// Gate represents a gate number
type Gate string

// Seat represents a seat number
type Seat string

// Status represents a flight status
type Status string

// Baggage represents a baggage claim number
type Baggage string

// Schedule represents flight schedule information
type Schedule struct {
	DepartureTime string `json:"departure_time"`
	ArrivalTime   string `json:"arrival_time"`
	DepartureDate string `json:"departure_date"`
	ArrivalDate   string `json:"arrival_date"`
}

// String returns a formatted string representation of Schedule
func (s Schedule) String() string {
	return fmt.Sprintf("Departure: %s on %s\nArrival: %s on %s",
		s.DepartureTime,
		s.DepartureDate,
		s.ArrivalTime,
		s.ArrivalDate)
}

// FlightInfo represents complete flight information
type FlightInfo struct {
	FlightNumber Flight   `json:"flight_number"`
	AircraftType Aircraft `json:"aircraft_type"`
	Origin       Airport  `json:"origin"`
	Destination  Airport  `json:"destination"`
	Gate         Gate     `json:"gate"`
	Status       Status   `json:"status"`
	BaggageClaim Baggage  `json:"baggage_claim"`
	Schedule     Schedule `json:"schedule"`
}

// String returns a formatted string representation of FlightInfo
func (f FlightInfo) String() string {
	return fmt.Sprintf(`Flight Information:
Flight Number: %s
Aircraft: %s
Route: %s → %s
Gate: %s
Status: %s
Baggage Claim: %s
Schedule:
  %s`,
		f.FlightNumber,
		f.AircraftType,
		f.Origin,
		f.Destination,
		f.Gate,
		f.Status,
		f.BaggageClaim,
		f.Schedule.String())
}

// Valid aircraft types
var ValidAircraftTypes = []Aircraft{
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

// Valid flight statuses
var ValidFlightStatuses = []Status{
	"On Time",
	"Delayed",
	"Boarding",
	"Departed",
	"Arrived",
	"Cancelled",
	"Diverted",
	"Gate Changed",
}

// Valid airport codes
var ValidAirportCodes = []Airport{
	"JFK", "LAX", "ORD", "ATL", "LHR", "CDG", "FRA", "AMS", "SIN", "DXB",
	"HKG", "NRT", "SYD", "MEL", "SFO", "DFW", "DEN", "SEA", "MIA", "BOS",
}

// Valid baggage claim carousels
var ValidBaggageClaims = []Baggage{
	"A", "B", "C", "D", "E",
}
