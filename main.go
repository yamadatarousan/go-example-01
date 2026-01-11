package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

type Todo struct {
	ID     int    `json:"id"`
	Name   string `json:"name" binding:"required"`
	UserID int    `json:"user_id"`
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuidObj, _ := uuid.NewRandom()
		requestID := uuidObj.String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

type TodoHandler struct {
	repo *TodoRepository
}

func NewTodoHandler(repo *TodoRepository) *TodoHandler {
	return &TodoHandler{repo: repo}
}

func (h *TodoHandler) getTodos(c *gin.Context) error {
	claims := c.MustGet("claims").(*AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)
	todos, err := h.repo.FindAll(userID)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, todos)
	return nil
}

func (h *TodoHandler) createTodo(c *gin.Context) error {
	var newTodo Todo
	if err := c.ShouldBindJSON(&newTodo); err != nil {
		return err
	}

	claims := c.MustGet("claims").(*AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)
	newTodo.UserID = userID
	createdTodo, err := h.repo.CreateTodoWithAudit(c.Request.Context(), newTodo)
	if err != nil {
		return err
	}
	c.JSON(http.StatusCreated, createdTodo)
	return nil
}

type AppHandler func(c *gin.Context) error

func errorHandler(handler AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := handler(c); err != nil {
			log.Printf("Error occurred: %v", err)

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
						"details": "Todo with this name already exists",
					})
					return
				}
			}

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var db *sql.DB

func initDB() {
	var err error
	// --- PostgreSQLへの接続情報 (DSN: Data Source Name) ---
	// 環境変数から接続情報を取得（Docker環境対応）
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5435")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "todo_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// 接続確認
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Successfully connected to the database")
}

type AuthHandler struct {
	repo *TodoRepository
}

func NewAuthHandler(repo *TodoRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

type SignupInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never return password hash
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

var jwtSecret []byte

type AppClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (h *AuthHandler) signup(c *gin.Context) error {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}
	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
	}

	createdUser, err := h.repo.CreateUser(user)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{"id": createdUser.ID, "email": createdUser.Email, "created_at": createdUser.CreatedAt})
	return nil
}

func (h *AuthHandler) login(c *gin.Context) error {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	user, err := h.repo.FindUserByEmail(input.Email)
	if err != nil {
		return err
	}

	// パスワードの検証
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return err
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
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
	return nil
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// "Bearer <token>" という形式を期待
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			return
		}

		if claims, ok := token.Claims.(*AppClaims); ok && token.Valid {
			c.Set("claims", claims)
			log.Println("Authenticated user ID:", claims.Subject, "Role:", claims.Role)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
	}
}

type AdminHandler struct {
	repo *TodoRepository
}

func NewAdminHandler(repo *TodoRepository) *AdminHandler {
	return &AdminHandler{repo: repo}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "Not an admin"})
			return
		}

		appClaims, ok := claims.(*AppClaims)
		if !ok || appClaims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "Not an admin"})
			return
		}

		c.Next()
	}
}

func (h *AdminHandler) getAllUsers(c *gin.Context) error {
	users, err := h.repo.FindAllUsers()
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, users)
	return nil
}

func (h *TodoHandler) getTodo(c *gin.Context) error {
	// URLパラメータからTODO IDを取得
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return nil
	}

	// ログインユーザーのIDを取得
	claims := c.MustGet("claims").(*AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	// TODOを取得
	todo, err := h.repo.FindByID(todoID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return nil
		}
		return err
	}

	c.JSON(http.StatusOK, todo)
	return nil
}

func (h *TodoHandler) updateTodo(c *gin.Context) error {
	// URLパラメータからTODO IDを取得
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return nil
	}

	// リクエストボディから更新内容を取得
	var updateTodo Todo
	if err := c.BindJSON(&updateTodo); err != nil {
		return err
	}

	// ログインユーザーのIDを取得
	claims := c.MustGet("claims").(*AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	// 更新対象のTODOを設定
	updateTodo.ID = todoID
	updateTodo.UserID = userID

	// TODOを更新
	updatedTodo, err := h.repo.UpdateTodoWithAudit(c.Request.Context(), updateTodo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return nil
		}
		return err
	}

	c.JSON(http.StatusOK, updatedTodo)
	return nil
}

func (h *TodoHandler) deleteTodo(c *gin.Context) error {
	// URLパラメータからTODO IDを取得
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return nil
	}

	// ログインユーザーのIDを取得
	claims := c.MustGet("claims").(*AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	// TODOを削除
	err = h.repo.DeleteTodoWithAudit(c.Request.Context(), todoID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return nil
		}
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
	return nil
}

func main() {
	// JWT秘密鍵を環境変数から読み取る
	jwtSecret = []byte(getEnv("JWT_SECRET", "a-very-secret-key"))

	initDB()

	// --- 依存関係の構築 (DI: Dependency Injection) ---
	// 1. リポジトリのインスタンスを作成
	repo := NewTodoRepository(db)
	// 2. ハンドラのインスタンスを作成し、リポジトリを注入
	todoHandler := NewTodoHandler(repo)
	authHandler := NewAuthHandler(repo)
	adminHandler := NewAdminHandler(repo)

	// Ginのモードを設定します。デフォルトは "debug" モードです。
	router := gin.New()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Authorization", "Content-Type"}
	router.Use(cors.New(config))

	// カスタムログフォーマッタを定義します。
	logFromatter := func(param gin.LogFormatterParams) string {
		requestID := param.Keys["RequestID"]
		return fmt.Sprintf("%s | %s | %s | %3d | %13v | %15s | %s\n",
			param.TimeStamp.Format(time.RFC3339), // timeパッケージの定数を明示的に使用
			requestID,
			param.Method,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Path,
		)
	}
	// ミドルウェアを .Use() で適用します。適用した順に実行されます。
	// 1. Recovery: panicが発生してもサーバーが落ちないようにする。
	router.Use(gin.Recovery())
	// 2. RequestID: これ以降の処理（ロガーなど）で使えるようにIDを生成する。
	router.Use(requestIDMiddleware())
	// 3. Logger: カスタムフォーマットのロガーを適用する。
	router.Use(gin.LoggerWithFormatter(logFromatter))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.POST("/signup", errorHandler(authHandler.signup))
	router.POST("/login", errorHandler(authHandler.login))

	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware())
	{
		v1.GET("/todos", errorHandler(todoHandler.getTodos))
		v1.GET("todos/:id", errorHandler(todoHandler.getTodo))
		v1.POST("/todos", errorHandler(todoHandler.createTodo))
		v1.PUT("/todos/:id", errorHandler(todoHandler.updateTodo))
		v1.DELETE("/todos/:id", errorHandler(todoHandler.deleteTodo))

		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(adminMiddleware())
		{
			adminRoutes.GET("/users", errorHandler(adminHandler.getAllUsers))
		}
	}

	// --- Graceful Shutdownの実装 ---

	// 1. http.Serverを独自に設定
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 2. サーバーをゴルーチンで起動（非同期処理）
	// これにより、サーバーの起動をブロックせずに、後続のシャットダウン処理に進むことができる
	go func() {
		log.Println("Starting server at port 8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 3. 終了シグナルを待機するためのチャネルを作成
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // SIGINT (Ctrl+C) と SIGTERM を監視
	<-quit                                               // シグナルを待機

	log.Println("Shutting down server...")

	// 4. サーバーをシャットダウンするためのコンテキストを作成（ここでは5秒のタイムアウトを設定）
	// 5秒以内に既存のリクエストの処理が終わらなければ、強制的に終了する
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
