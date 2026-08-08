// Package gincrosspkg demonstrates gin handlers referencing DTOs declared in
// a separate package via request:/response: annotations -- including a
// renamed import alias, the trickiest resolution path.
package gincrosspkg

import (
	d "github.com/yaop-labs/apiary/testdata/gin_crosspkg/dto"
	"github.com/yaop-labs/apiary/testdata/gin_crosspkg/gin"
)

type UserHandler struct{}

// apiary:operation POST /api/v1/users
// summary: Create user
// tags: users
// errors: 400,409,500
// request: d.CreateUserRequest
// response: d.UserResponse
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req d.CreateUserRequest
	_ = c.ShouldBindJSON(&req)
	c.JSON(201, d.UserResponse{})
}

// apiary:operation GET /api/v1/users
// summary: List users
// tags: users
// response: []d.UserResponse
func (h *UserHandler) ListUsers(c *gin.Context) {
	c.JSON(200, []d.UserResponse{})
}
