// Package router is an example task-manager service annotated for apiary.
//
//	apiary -security bearer -title "Task Manager API" -version "1.0.0" -out docs/tasks.yaml ./testdata/router
package router

import "context"

type AuthHandler struct{}
type TaskHandler struct{}
type CommentHandler struct{}

// apiary:operation POST /api/v1/auth/login
// summary: Log in
// description: Returns a JWT for a username and password.
// tags: auth
// security: none
// errors: 400,401,500
func (h *AuthHandler) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	return LoginResponse{}, nil
}

// apiary:operation POST /api/v1/auth/refresh
// summary: Refresh token
// tags: auth
// security: bearer
// errors: 401,500
func (h *AuthHandler) Refresh(ctx context.Context, req RefreshRequest) (LoginResponse, error) {
	return LoginResponse{}, nil
}

// apiary:operation GET /api/v1/tasks
// summary: List tasks
// description: Supports filtering by status and priority, and pagination.
// tags: tasks
// errors: 400,401,500
func (h *TaskHandler) List(ctx context.Context, req ListTasksRequest) (ListTasksResponse, error) {
	return ListTasksResponse{}, nil
}

// apiary:operation GET /api/v1/tasks/{id}
// summary: Get task by ID
// tags: tasks
// errors: 401,404,500
func (h *TaskHandler) Get(ctx context.Context, req GetTaskRequest) (TaskDTO, error) {
	return TaskDTO{}, nil
}

// apiary:operation POST /api/v1/tasks
// summary: Create task
// tags: tasks
// errors: 400,401,422,500
func (h *TaskHandler) Create(ctx context.Context, req CreateTaskRequest) (TaskDTO, error) {
	return TaskDTO{}, nil
}

// apiary:operation PUT /api/v1/tasks/{id}
// summary: Update task
// tags: tasks
// errors: 400,401,403,404,422,500
func (h *TaskHandler) Update(ctx context.Context, req UpdateTaskRequest) (TaskDTO, error) {
	return TaskDTO{}, nil
}

// apiary:operation DELETE /api/v1/tasks/{id}
// summary: Delete task
// description: Only the task creator or an administrator.
// tags: tasks
// errors: 401,403,404,500
func (h *TaskHandler) Delete(ctx context.Context, req DeleteTaskRequest) (DeleteTaskResponse, error) {
	return DeleteTaskResponse{}, nil
}

// apiary:operation GET /api/v1/tasks/{task_id}/comments
// summary: List task comments
// tags: comments
// errors: 401,404,500
func (h *CommentHandler) List(ctx context.Context, req ListCommentsRequest) (ListCommentsResponse, error) {
	return ListCommentsResponse{}, nil
}

// apiary:operation POST /api/v1/tasks/{task_id}/comments
// summary: Add comment
// tags: comments
// errors: 400,401,404,500
func (h *CommentHandler) Add(ctx context.Context, req AddCommentRequest) (CommentDTO, error) {
	return CommentDTO{}, nil
}
