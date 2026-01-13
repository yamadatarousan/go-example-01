package repository

import (
	"context"

	"gin-quickstart/internal/domain"
)

// ============================================================================
// レイヤードアーキテクチャにおけるインターフェースの役割
// ============================================================================
//
// レイヤードアーキテクチャでは、各層が明確な責務を持ち、依存関係に制約があります：
//
//   ┌─────────────────┐
//   │   Handler層     │  HTTPリクエスト/レスポンス処理
//   ├─────────────────┤
//   │   Service層     │  ビジネスロジック
//   ├─────────────────┤
//   │  Repository層   │  データ永続化（この層）
//   ├─────────────────┤
//   │   Domain層      │  ビジネスの概念
//   └─────────────────┘
//
// **依存のルール: 上の層は下の層にのみ依存できる**
//
// 問題は、「下の層の"何"に依存するか」です。
//
// ============================================================================
// 問題：具体実装に依存すると層が密結合になる
// ============================================================================
//
// インターフェースがない場合、Service層は具体的な実装クラスに依存します：
//
//   // Service層
//   type TodoService struct {
//       todoRepo *todoRepository  // ← 具体的な構造体に依存
//   }
//
//   // Repository層の実装
//   type todoRepository struct {
//       db *sql.DB  // PostgreSQL前提
//   }
//
// この場合、以下の問題が発生します：
//
// 1. **層の独立性が失われる**
//    - Repository層の実装詳細（PostgreSQLを使う）をService層が知っている
//    - データベースを変更すると、Service層も修正が必要になる可能性がある
//
// 2. **層の境界が曖昧になる**
//    - Service層がRepository層の内部構造（db *sql.DB）を知ることができる
//    - 誤ってService層からDBに直接アクセスできてしまう
//
// 3. **テストが困難になる**
//    - Service層のテストで必ずPostgreSQLが必要になる
//    - Repository層なしでService層だけをテストできない
//
// ============================================================================
// 解決：インターフェースで層の境界を定義する
// ============================================================================
//
// インターフェースを使うと、層間の「契約」が明確になります：
//
//   // Repository層：インターフェース（契約）を定義
//   type TodoRepository interface {
//       FindAll(userID int) ([]domain.Todo, error)
//       // これが層の境界線
//   }
//
//   // Service層：インターフェース（抽象）に依存
//   type TodoService struct {
//       todoRepo TodoRepository  // ← 抽象に依存（具体実装を知らない）
//   }
//
//   // Repository層：インターフェースを実装（この層の内部実装）
//   type todoRepository struct {
//       db *sql.DB  // Service層からは見えない
//   }
//
//   func (r *todoRepository) FindAll(userID int) ([]domain.Todo, error) {
//       // PostgreSQL固有の実装
//   }
//
// **これがレイヤードアーキテクチャの核心：層の境界 = インターフェース**
//
// ============================================================================
// レイヤードアーキテクチャにおける3つの重要な効果
// ============================================================================
//
// ### 1. 層の独立性が保たれる
//
// Service層は「TODOを保存/取得する機能がある」ことだけを知り、
// 「どうやって保存するか（PostgreSQL? MySQL? メモリ?）」を知りません。
//
//   // Service層のコード
//   func (s *TodoService) GetTodos(userID int) ([]domain.Todo, error) {
//       return s.todoRepo.FindAll(userID)  // 実装を知らない
//   }
//
// Repository層の実装を変更しても、Service層は影響を受けません。
// これが「層の独立性」です。
//
// ### 2. 層の境界が明確になる
//
// インターフェースに定義されたメソッドだけが、層を越えて呼べる操作です。
//
//   // Repository層の実装
//   type todoRepository struct {
//       db *sql.DB
//   }
//
//   func (r *todoRepository) FindAll(userID int) ([]domain.Todo, error) {
//       // ← これは公開（インターフェースで定義）
//   }
//
//   func (r *todoRepository) optimizeQuery() {
//       // ← これは内部実装（Service層から呼べない）
//   }
//
// Service層は `FindAll()` しか呼べません。
// `optimizeQuery()` や `db` フィールドは見えません。
// これが「層の境界」です。
//
// ### 3. 各層を独立してテスト可能
//
// インターフェースがあることで、Repository層なしでService層をテストできます：
//
//   // Service層のテスト
//   type mockRepository struct{}
//
//   func (m *mockRepository) FindAll(userID int) ([]domain.Todo, error) {
//       return []domain.Todo{{ID: 1, Name: "Test"}}, nil
//   }
//
//   func TestTodoService(t *testing.T) {
//       mockRepo := &mockRepository{}
//       service := NewTodoService(mockRepo)  // Repository層の実装不要
//       todos, err := service.GetTodos(1)
//       // Service層のロジックだけをテスト
//   }
//
// 各層を独立してテストできることで、「どの層にバグがあるか」が特定しやすくなります。
//
// ============================================================================
// なぜHandler層やDomain層ではなく、Repository層でインターフェースを切るのか
// ============================================================================
//
// レイヤードアーキテクチャでは、通常**インフラ層（Repository層）でインターフェースを切ります**。
//
// 理由：
//
// 1. **Repository層は外部システム（DB）に依存する唯一の層**
//    - Domain層：外部依存なし（純粋なビジネスロジック）
//    - Service層：Repositoryに依存するだけ
//    - Repository層：PostgreSQL、MySQL、外部APIなど外部に依存
//    - Handler層：HTTPフレームワーク（Gin）に依存
//
//    外部依存がある層を抽象化することで、上位層（Service）を外部から隔離できます。
//
// 2. **Repository層の実装は変わりやすい**
//    - 開発中：メモリ内実装
//    - テスト：モック実装
//    - 本番：PostgreSQL実装
//
//    インターフェースがあれば、Service層を変更せずに実装を切り替えられます。
//
// 3. **Handler層でもインターフェースを切ることはある**
//    - Handler層をインターフェース化することも可能
//    - ただし、HTTPフレームワークは頻繁に変更しないため、優先度は低い
//    - まずはRepository層（最も変わりやすい層）でインターフェースを切る
//
// ============================================================================
// まとめ：レイヤードアーキテクチャとインターフェース
// ============================================================================
//
// - **インターフェース = 層の境界線**
// - 上位層は「抽象（インターフェース）」に依存し、「具体実装」を知らない
// - これにより各層が独立し、変更の影響が局所化される
// - Repository層でインターフェースを切ることで、外部依存（DB）を隔離する
//
// レイヤードアーキテクチャの目的は「変更に強い設計」です。
// インターフェースは、その目的を実現するための具体的な手段です。
//
// ============================================================================

// TodoRepository はTODOデータアクセスの層境界インターフェース
//
// このインターフェースは、Service層とRepository層の境界を定義します。
// Service層はこのインターフェースを通じてのみRepository層と通信します。
type TodoRepository interface {
	FindAll(userID int) ([]domain.Todo, error)
	FindByID(todoID, userID int) (domain.Todo, error)
	Create(todo domain.Todo) (domain.Todo, error)
	CreateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error)
	UpdateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error)
	DeleteTodoWithAudit(ctx context.Context, todoID, userID int) error
}

// UserRepository はユーザーデータアクセスの層境界インターフェース
//
// このインターフェースは、Service層とRepository層の境界を定義します。
type UserRepository interface {
	CreateUser(user domain.User) (domain.User, error)
	FindUserByEmail(email string) (domain.User, error)
	FindAllUsers() ([]domain.User, error)
}
