package services

import (
	"errors"

	"vehicle-management-system/config"
	"vehicle-management-system/database"
	"vehicle-management-system/models"
	"vehicle-management-system/utils"
)

type AuthService struct {
	jwtConfig *config.JWTConfig
}

func NewAuthService(jwtConfig *config.JWTConfig) *AuthService {
	return &AuthService{
		jwtConfig: jwtConfig,
	}
}

func (s *AuthService) Login(username, password string) (*models.User, string, error) {
	var user models.User

	// Find user by username
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	// Check password
	if !utils.CheckPassword(password, user.Password) {
		return nil, "", errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, string(user.Role), s.jwtConfig)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &user, token, nil
}

func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User

	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
