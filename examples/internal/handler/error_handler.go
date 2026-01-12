package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/repository"
	"gin-quickstart/examples/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// AppHandler はエラーを返すハンドラー関数の型定義
type AppHandler func(c *gin.Context) error

// ErrorHandler はAppHandlerをgin.HandlerFuncに変換し、エラーハンドリングを提供
func ErrorHandler(handler AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := handler(c); err != nil {
			log.Printf("エラーが発生しました: %v", err)

			// バリデーションエラーの場合
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation failed",
					"details": err.Error(),
				})
				return
			}

			// PostgreSQLのユニーク制約違反エラーの場合
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" { // ユニーク制約違反のエラーコード
					c.JSON(http.StatusConflict, gin.H{
						"error":   "Conflict",
						"details": "Resource with this information already exists",
					})
					return
				}
			}

			// Repository層のエラー処理
			switch {
			case errors.Is(err, repository.ErrNotFound):
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Not Found",
				})
				return
			case errors.Is(err, repository.ErrConflict):
				c.JSON(http.StatusConflict, gin.H{
					"error": "Conflict",
				})
				return
			}

			// Service層のエラー処理
			switch {
			case errors.Is(err, service.ErrUnauthorized):
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Unauthorized",
					"message": "Invalid email or password",
				})
				return
			case errors.Is(err, service.ErrForbidden):
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Forbidden",
				})
				return
			}

			// Domain層のエラー処理
			switch {
			case errors.Is(err, domain.ErrInvalidInput):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid Input",
				})
				return
			}

			// データベース関連のエラー
			if errors.Is(err, sql.ErrNoRows) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				// 認証エラーやデータ未発見エラーの場合
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Unauthorized",
					"message": "Invalid email or password",
				})
				return
			}

			// その他の予期せぬエラーの場合
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server Error",
			})
		}
	}
}
