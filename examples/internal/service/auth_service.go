package service

import (
	"fmt"
	"time"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService は認証関連のビジネスロジックを提供
type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

// NewAuthService はAuthServiceの新しいインスタンスを作成
func NewAuthService(userRepo repository.UserRepository, jwtSecret []byte) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// AppClaims はJWTトークンのクレーム
type AppClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Signup は新しいユーザーを登録
func (s *AuthService) Signup(input domain.SignupInput) (domain.User, error) {
	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
	}

	user := domain.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
	}

	createdUser, err := s.userRepo.CreateUser(user)
	if err != nil {
		return domain.User{}, err
	}

	return createdUser, nil
}

// Login はユーザー認証を行い、JWTトークンを生成
func (s *AuthService) Login(input domain.LoginInput) (string, error) {
	user, err := s.userRepo.FindUserByEmail(input.Email)
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	// パスワードの検証
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	// JWTトークンの生成
	claims := AppClaims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken はJWTトークンを解析してクレームを返す
func (s *AuthService) ParseToken(tokenString string) (*AppClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("予期しない署名方法: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AppClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, domain.ErrUnauthorized
}
