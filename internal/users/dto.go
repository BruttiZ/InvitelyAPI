package users

type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
}
