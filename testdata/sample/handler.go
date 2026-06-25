// Package sample shows how to annotate handlers for apiary.
//
//	apiary -security bearer -title "Sample API" -version "0.1.0" -out docs/sample.yaml ./testdata/sample
package sample

import "context"

type AuthHandler struct{}
type UserHandler struct{}
type ProductHandler struct{}
type HealthHandler struct{}

// apiary:operation POST /api/v1/auth/telegram
// summary: Authenticate via Telegram
// description: Accepts initData from the Telegram WebApp and verifies the HMAC-SHA256 signature.
// tags: auth
// security: none
// errors: 400,401,500
func (h *AuthHandler) TelegramAuth(ctx context.Context, req TelegramAuthRequest) (TelegramAuthResponse, error) {
	return TelegramAuthResponse{}, nil
}

// apiary:operation GET /api/v1/users/{id}
// summary: Get user profile
// tags: users
// errors: 401,403,404,500
func (h *UserHandler) GetUser(ctx context.Context, req GetUserRequest) (UserResponse, error) {
	return UserResponse{}, nil
}

// apiary:operation POST /api/v1/users
// summary: Create user
// tags: users
// errors: 400,409,500
func (h *UserHandler) CreateUser(ctx context.Context, req CreateUserRequest) (UserResponse, error) {
	return UserResponse{}, nil
}

// apiary:operation GET /api/v1/products
// summary: List products
// description: Supports pagination and full-text search. Prices are returned in the currency from the X-Currency header.
// tags: products
// errors: 400,500
func (h *ProductHandler) ListProducts(ctx context.Context, req ListProductsRequest) (ListProductsResponse, error) {
	return ListProductsResponse{}, nil
}

// apiary:operation GET /health
// summary: Health check
// tags: infra
// security: none
func (h *HealthHandler) Health(req HealthRequest) (HealthResponse, error) {
	return HealthResponse{Status: "ok"}, nil
}
