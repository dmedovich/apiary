// Package gin is the same task-manager example, but with gin handlers.
//
// It is a separate nested module (its own go.mod) so the real gin dependency
// stays out of the main apiary module. Generate the spec with:
//
//	cd testdata/gin && apiary -security bearer -title "Task Manager API (gin)" -version "1.0.0" -out ../../docs/tasks_gin.yaml .
package gin

import "github.com/gin-gonic/gin"

// apiary:operation POST /api/v1/auth/login
// summary: Log in
// description: Returns a JWT for a username and password.
// tags: auth
// security: none
// request: LoginRequest
// response: LoginResponse
// errors: 400,401,500
func Login(c *gin.Context) {}

// apiary:operation POST /api/v1/auth/refresh
// summary: Refresh token
// tags: auth
// security: bearer
// request: RefreshRequest
// response: LoginResponse
// errors: 401,500
func Refresh(c *gin.Context) {}

// apiary:operation GET /api/v1/tasks
// summary: List tasks
// description: Supports filtering by status and priority, and pagination.
// tags: tasks
// request: ListTasksRequest
// response: ListTasksResponse
// errors: 400,401,500
func ListTasks(c *gin.Context) {}

// apiary:operation GET /api/v1/tasks/{id}
// summary: Get task by ID
// tags: tasks
// request: GetTaskRequest
// response: TaskDTO
// errors: 401,404,500
func GetTask(c *gin.Context) {}

// apiary:operation POST /api/v1/tasks
// summary: Create task
// tags: tasks
// request: CreateTaskRequest
// response: TaskDTO
// errors: 400,401,422,500
func CreateTask(c *gin.Context) {}

// apiary:operation PUT /api/v1/tasks/{id}
// summary: Update task
// tags: tasks
// request: UpdateTaskRequest
// response: TaskDTO
// errors: 400,401,403,404,422,500
func UpdateTask(c *gin.Context) {}

// apiary:operation DELETE /api/v1/tasks/{id}
// summary: Delete task
// description: Only the task creator or an administrator.
// tags: tasks
// request: DeleteTaskRequest
// response: DeleteTaskResponse
// errors: 401,403,404,500
func DeleteTask(c *gin.Context) {}

// apiary:operation GET /api/v1/tasks/{task_id}/comments
// summary: List task comments
// tags: comments
// request: ListCommentsRequest
// response: ListCommentsResponse
// errors: 401,404,500
func ListComments(c *gin.Context) {}

// apiary:operation POST /api/v1/tasks/{task_id}/comments
// summary: Add comment
// tags: comments
// request: AddCommentRequest
// response: CommentDTO
// errors: 400,401,404,500
func AddComment(c *gin.Context) {}
