package primitive

import (
	"encoding/json"
	"fmt"
	"time"
)

// BirthdateLayout is the format string for birthdates.
const BirthdateLayout = "2006-01-02"

// Birthdate represents a date of birth.
type Birthdate time.Time

// ParseBirthdate parses a Birthdate from a date string.
func ParseBirthdate(rawDate string) (Birthdate, error) {
	date, err := time.Parse(BirthdateLayout, rawDate)
	if err != nil {
		return Birthdate{}, fmt.Errorf("parse Birthdate: %w", err)
	}

	return Birthdate(date), nil
}

// UnmarshalJSON allows Birthdates to be parsed from JSON payloads.
func (bd *Birthdate) UnmarshalJSON(b []byte) error {
	var rawDate string
	if err := json.Unmarshal(b, &rawDate); err != nil {
		return err
	}

	date, err := ParseBirthdate(rawDate)
	if err != nil {
		return err
	}

	*bd = Birthdate(date)

	return nil
}
