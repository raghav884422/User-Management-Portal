package models

import "time"

// UserResponse is the API response model that includes the dynamically calculated age.
type UserResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Dob  string `json:"dob"`
	Age  int    `json:"age,omitempty"`
}

// CreateUserRequest is the request body for creating a new user.
type CreateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	Dob  string `json:"dob" validate:"required"`
}

// UpdateUserRequest is the request body for updating an existing user.
type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	Dob  string `json:"dob" validate:"required"`
}

// PaginatedUsersResponse wraps a list of users with pagination metadata.
type PaginatedUsersResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// ErrorResponse is the standard error response body.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// CalculateAge computes the age in years from a date of birth to today.
func CalculateAge(dob time.Time) int {
	now := time.Now().UTC()
	years := now.Year() - dob.Year()

	// Adjust if the birthday hasn't occurred yet this year
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		years--
	}
	return years
}
