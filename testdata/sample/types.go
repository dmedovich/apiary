package sample

type TelegramAuthRequest struct {
	InitData string `json:"init_data" validate:"required" doc:"initData from the Telegram WebApp" example:"query_id=AAH4pFf..."`
}

type TelegramAuthResponse struct {
	User      UserDTO `json:"user"`
	ExpiresIn int     `json:"expires_in" doc:"Token TTL in seconds" example:"3600"`
	IsNewUser bool    `json:"is_new_user"`
}

type GetUserRequest struct {
	ID int64 `path:"id" validate:"required" example:"42"`
}

type UserResponse struct {
	User UserDTO `json:"user"`
}

type CreateUserRequest struct {
	Username  string `json:"username"   validate:"required,min=3,max=32" example:"larry_somik"`
	Email     string `json:"email"      validate:"required,email" example:"larry@example.com"`
	FirstName string `json:"first_name" example:"Larry"`
	LastName  string `json:"last_name"  example:"Somik"`
	Role      string `json:"role"       validate:"oneof=admin user guest" doc:"User role" example:"user"`
	Age       *int   `json:"age,omitempty" validate:"gte=13,lte=120" doc:"Age (optional)"`
}

type ListProductsRequest struct {
	Currency string `header:"X-Currency" doc:"Price currency (ISO 4217)" example:"USD"`
	Page     int    `query:"page"        default:"1"  example:"1"`
	PageSize int    `query:"page_size"   default:"20" example:"20"`
	Search   string `query:"search"      example:"laptop"`
}

type ListProductsResponse struct {
	Products []ProductDTO `json:"products"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

type HealthRequest struct{}

type HealthResponse struct {
	Status  string `json:"status"  example:"ok"`
	Version string `json:"version" example:"1.2.3"`
}

type UserDTO struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ProductDTO struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       float64         `json:"price"    example:"1990.00"`
	InStock     bool            `json:"in_stock"`
	Category    ProductCategory `json:"category" doc:"Product category"`
}

type ProductCategory string

const (
	CategoryElectronics ProductCategory = "electronics"
	CategoryClothing    ProductCategory = "clothing"
	CategoryFood        ProductCategory = "food"
	CategoryBooks       ProductCategory = "books"
)
