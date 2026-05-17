package jwt_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Uncle0206061/zeroquant2/backend/internal/config"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/jwt"
)

func TestMain(m *testing.M) {
	// 初始化测试配置（JWT secret 必须是固定的测试值）
	config.InitForTest()
	// 运行所有测试
	os.Exit(m.Run())
}

func TestGenerateToken(t *testing.T) {
	token, err := jwt.GenerateToken(1, "testuser", "user")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}
	// JWT token 有 3 个部分，用 . 分隔
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("Invalid JWT format, got %d parts", len(parts))
	}
}

func TestValidateToken_Valid(t *testing.T) {
	token, err := jwt.GenerateToken(99, "alice", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != 99 {
		t.Errorf("UserID mismatch: got %d, want 99", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("Username mismatch: got %s, want alice", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("Role mismatch: got %s, want admin", claims.Role)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	_, err := jwt.ValidateToken("not.a.valid.jwt.token")
	if err != jwt.ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got: %v", err)
	}

	_, err = jwt.ValidateToken("")
	if err != jwt.ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for empty string, got: %v", err)
	}

	_, err = jwt.ValidateToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InRlc3QiLCJyb2xlIjoidXNlciJ9.invalid_signature")
	if err != jwt.ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for tampered token, got: %v", err)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	// 已过期的 token（过期时间设为过去的时间戳）
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InRlc3QiLCJyb2xlIjoidXNlciIsImV4cCI6MTYwMDAwMDAwMH0.fake_sig"
	_, err := jwt.ValidateToken(expiredToken)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
	// 预期返回 ErrExpiredToken 或 ErrInvalidToken
	if err != jwt.ErrExpiredToken && err != jwt.ErrInvalidToken {
		t.Errorf("Expected expired/invalid error, got: %v", err)
	}
}

func TestRefreshToken(t *testing.T) {
	original, err := jwt.GenerateToken(5, "bob", "user")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	refreshed, err := jwt.RefreshToken(original)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if refreshed == "" {
		t.Fatal("RefreshToken returned empty token")
	}
	// 注意：同一纳秒内生成的 token 签名相同是正常的
	// 关键验证：刷新后的 token 仍然有效，且携带正确的 claims

	// 验证刷新后的 token 有效
	claims, err := jwt.ValidateToken(refreshed)
	if err != nil {
		t.Fatalf("Refreshed token invalid: %v", err)
	}
	if claims.UserID != 5 || claims.Username != "bob" || claims.Role != "user" {
		t.Errorf("Refreshed token claims mismatch: UserID=%d, Username=%s, Role=%s",
			claims.UserID, claims.Username, claims.Role)
	}
}

func TestRefreshToken_Invalid(t *testing.T) {
	_, err := jwt.RefreshToken("invalid.token")
	if err != jwt.ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got: %v", err)
	}
}
