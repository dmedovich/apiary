// Package dto holds request/response types shared across the app, kept
// separate from the handler package on purpose -- this is the layout that
// used to force a duplicate types file next to the handler.
package dto

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32" example:"larry_somik"`
	Email    string `json:"email" validate:"required,email" example:"larry@example.com"`
}

type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}
