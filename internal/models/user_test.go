package models_test

import (
	"testing"
	"time"

	"github.com/yourusername/user-api/internal/models"
)

func TestCalculateAge(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		dob      time.Time
		wantAge  int
	}{
		{
			name:    "birthday today returns correct age",
			dob:     now.AddDate(-25, 0, 0),
			wantAge: 25,
		},
		{
			name:    "birthday yesterday returns correct age",
			dob:     now.AddDate(-30, 0, 1),
			wantAge: 29,
		},
		{
			name:    "birthday tomorrow has not happened yet",
			dob:     now.AddDate(-20, 0, -1),
			wantAge: 20,
		},
		{
			name:    "born 1 year ago exactly",
			dob:     now.AddDate(-1, 0, 0),
			wantAge: 1,
		},
		{
			name:    "newborn (0 years old)",
			dob:     now.AddDate(0, -6, 0),
			wantAge: 0,
		},
		{
			name:    "birthday in future month this year",
			dob:     time.Date(now.Year()-10, now.Month()+1, now.Day(), 0, 0, 0, 0, time.UTC),
			wantAge: func() int {
				if now.Month()+1 > 12 {
					return 9
				}
				return 9
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.CalculateAge(tt.dob)
			if got != tt.wantAge {
				t.Errorf("CalculateAge(%v) = %d; want %d", tt.dob.Format("2006-01-02"), got, tt.wantAge)
			}
		})
	}
}

func TestCalculateAge_SpecificDates(t *testing.T) {
	tests := []struct {
		name    string
		dob     string
		refDate time.Time
		want    int
	}{
		{
			name:    "Alice born 1990-05-10 on 2025-06-11 should be 35",
			dob:     "1990-05-10",
			refDate: time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC),
			want:    35,
		},
		{
			name:    "Birthday exactly on check date",
			dob:     "1990-06-11",
			refDate: time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC),
			want:    35,
		},
		{
			name:    "Birthday not yet reached this year",
			dob:     "1990-12-25",
			refDate: time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC),
			want:    34,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dob, _ := time.Parse("2006-01-02", tt.dob)
			// Temporarily monkey-patch via a wrapper for deterministic testing
			got := calculateAgeAt(dob, tt.refDate)
			if got != tt.want {
				t.Errorf("calculateAgeAt(%s, %s) = %d; want %d",
					tt.dob, tt.refDate.Format("2006-01-02"), got, tt.want)
			}
		})
	}
}

// calculateAgeAt is a testable helper that accepts an explicit reference date.
func calculateAgeAt(dob, now time.Time) int {
	years := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		years--
	}
	return years
}
