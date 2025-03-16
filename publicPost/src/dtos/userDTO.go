package dtos

type UserDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponseDTO struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
