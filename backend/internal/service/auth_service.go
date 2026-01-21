package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"gin-quickstart/backend/internal/domain"
	"gin-quickstart/backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService は認証関連のビジネスロジックを提供
type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtSecret        []byte
}

// NewAuthService はAuthServiceの新しいインスタンスを作成
func NewAuthService(userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, jwtSecret []byte) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
	}
}

// AppClaims はJWTトークンのクレーム
type AppClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Signup は新しいユーザーを登録
func (s *AuthService) Signup(ctx context.Context, input domain.SignupInput) (domain.User, error) {
	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
	}

	user := domain.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
	}

	createdUser, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	return createdUser, nil
}

// Login はユーザー認証を行い、JWTトークンを生成
func (s *AuthService) Login(ctx context.Context, input domain.LoginInput) (string, error) {
	user, err := s.userRepo.FindUserByEmail(ctx, input.Email)
	if err != nil {
		return "", ErrUnauthorized
	}

	// パスワードの検証
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return "", ErrUnauthorized
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

	return nil, ErrUnauthorized
}

// GenerateRefreshToken はリフレッシュトークンを生成して保存
func (s *AuthService) GenerateRefreshToken(ctx context.Context, userID int) (string, error) {
	// ランダムなトークン文字列を生成（32バイト = 256ビット）
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("トークン生成に失敗しました: %w", err)
	}

	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)

	// リフレッシュトークンをデータベースに保存（有効期限: 7日間）
	refreshToken := domain.RefreshToken{
		Token:     tokenString,
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	_, err = s.refreshTokenRepo.CreateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", fmt.Errorf("リフレッシュトークンの保存に失敗しました: %w", err)
	}

	return tokenString, nil
}

// RefreshAccessToken はリフレッシュトークンを使って新しいアクセストークンを生成
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// リフレッシュトークンを検証
	storedToken, err := s.refreshTokenRepo.FindRefreshTokenByToken(ctx, refreshToken)
	if err != nil {
		return "", ErrUnauthorized
	}

	// 有効期限チェック
	if time.Now().After(storedToken.ExpiresAt) {
		return "", ErrUnauthorized
	}

	// ユーザー情報を取得
	user, err := s.userRepo.FindUserByID(ctx, storedToken.UserID)
	if err != nil {
		return "", ErrUnauthorized
	}

	// 新しいアクセストークンを生成
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

// RevokeRefreshToken はリフレッシュトークンを無効化
func (s *AuthService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	return s.refreshTokenRepo.RevokeRefreshToken(ctx, refreshToken)
}
