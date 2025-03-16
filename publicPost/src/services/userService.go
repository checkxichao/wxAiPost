package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"publicPost/src/config"
	"publicPost/src/models"
	"publicPost/src/repositories"
)

type UserService struct {
	userRepo   repositories.UserRepository
	tokenRepo  repositories.TokenRepository
	cfg        *config.Config
	wechatRepo repositories.WechatRepository
}

func NewUserService(userRepo repositories.UserRepository, tokenRepo repositories.TokenRepository, cfg *config.Config) *UserService {
	return &UserService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

func (s *UserService) Register(username, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Power:    0,
		Show:     true,
	}
	err = s.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (s *UserService) Delete(id int) error {

	err := s.userRepo.DeleteUser(id)
	if err != nil {
		return err
	}
	return nil
}
func (s *UserService) GetAllUsers(username string) ([]*models.User, error) {
	check := s.userRepo.CheckUserPower(username)
	if !check {
		return nil, errors.New("user not power")
	}

	users, err := s.userRepo.GetAllUser()
	return users, err

}
func (s *UserService) SetUserPower(username string) error {

	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}
	if user.Power > 0 {
		return errors.New("user is power on")
	}

	errs := s.userRepo.SetUserPower(username)
	if errs != nil {
		return errs
	}
	return nil
}
func (s *UserService) UpdateUserPwd(username string, pwd string, power int) error {
	user, err := s.userRepo.FindUserByUsername(username)
	if err != nil {
		return err
	}
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if user.Password == string(hashPwd) {
		return errors.New("密码重复")
	}
	if power < 0 {
		return errors.New("权限输入错误")
	}

	err = s.userRepo.UpdateUserInfo(username, string(hashPwd), power)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) Login(username, password string) (*models.User, string, string, error) {
	user, err := s.userRepo.FindUserByUsername(username)
	if err != nil {
		return nil, "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", "", err
	}

	accessToken, err := GenerateJWT(user.ID, user.Username, []byte(s.cfg.JWTSecret), s.cfg.AccessTokenExpiry)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := GenerateJWT(user.ID, user.Username, []byte(s.cfg.JWTSecret), s.cfg.RefreshTokenExpiry)
	if err != nil {
		return nil, "", "", err
	}

	err = s.tokenRepo.SaveRefreshToken(user.ID, refreshToken, time.Now().Add(s.cfg.RefreshTokenExpiry))
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *UserService) GetWechatInfo() ([]*models.WechatAccount, error) {
	info, err := s.wechatRepo.GetWechatInfo()
	return info, err
}

func (s *UserService) RefreshToken(refreshToken string) (string, string, error) {

	claims, err := ParseJWT(refreshToken, []byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", fmt.Errorf("无效token")
	}

	userID, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return "", "", fmt.Errorf("无效userID token")
	}

	isValid, err := s.tokenRepo.IsRefreshTokenValid(int(userID), refreshToken)
	if err != nil {
		return "", "", err
	}
	if !isValid {
		return "", "", fmt.Errorf("刷新令牌失败")
	}

	newAccessToken, err := GenerateJWT(int(userID), claims.Username, []byte(s.cfg.JWTSecret), s.cfg.AccessTokenExpiry)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := GenerateJWT(int(userID), claims.Username, []byte(s.cfg.JWTSecret), s.cfg.RefreshTokenExpiry)
	if err != nil {
		return "", "", err
	}

	err = s.tokenRepo.RevokeRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	err = s.tokenRepo.SaveRefreshToken(int(userID), newRefreshToken, time.Now().Add(s.cfg.RefreshTokenExpiry))
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *UserService) Logout(accessToken string, refreshToken string) error {

	claims, err := ParseJWT(accessToken, []byte(s.cfg.JWTSecret))
	if err != nil {
		return fmt.Errorf("无效token")
	}

	expiry := time.Unix(claims.ExpiresAt, 0)

	err = s.tokenRepo.BlacklistToken(accessToken, expiry)
	if err != nil {
		return err
	}

	err = s.tokenRepo.RevokeRefreshToken(refreshToken)
	if err != nil {
		return err
	}

	return nil
}
