package auth_test

import (
	"simplehttp/internal/auth"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "normal password", password: "mysecretpassword"},
		{name: "empty password", password: ""},
		{name: "long password", password: "a very long password with spaces and special chars !@#$%^&*()"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auth.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("HashPassword() returned empty hash")
			}
		})
	}
}

func TestHashPasswordIsNonDeterministic(t *testing.T) {
	hash1, err := auth.HashPassword("password")
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}
	hash2, err := auth.HashPassword("password")
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}
	if hash1 == hash2 {
		t.Error("HashPassword() produced identical hashes for the same password; expected unique salts")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mysecretpassword"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() setup failed: %v", err)
	}

	tests := []struct {
		name      string
		password  string
		hash      string
		wantMatch bool
		wantErr   bool
	}{
		{name: "correct password", password: password, hash: hash, wantMatch: true},
		{name: "wrong password", password: "wrongpassword", hash: hash, wantMatch: false},
		{name: "empty password against hash", password: "", hash: hash, wantMatch: false},
		{name: "invalid hash", password: password, hash: "notavalidhash", wantMatch: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := auth.CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if match != tt.wantMatch {
				t.Errorf("CheckPasswordHash() = %v, want %v", match, tt.wantMatch)
			}
		})
	}
}
func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "testsecret"

	tests := []struct {
		name      string
		userID    uuid.UUID
		secret    string
		expiresIn time.Duration
		wantErr   bool
	}{
		{name: "valid token", userID: userID, secret: secret, expiresIn: time.Hour},
		{name: "short-lived token", userID: userID, secret: secret, expiresIn: time.Second},
		{name: "zero duration", userID: userID, secret: secret, expiresIn: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.MakeJWT(tt.userID, tt.secret, tt.expiresIn)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("MakeJWT() returned empty token")
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "testsecret"
	userID := uuid.New()

	validToken, err := auth.MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() setup failed: %v", err)
	}

	expiredToken, err := auth.MakeJWT(userID, secret, -time.Second)
	if err != nil {
		t.Fatalf("MakeJWT() setup failed: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		secret      string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{name: "valid token", tokenString: validToken, secret: secret, wantUserID: userID},
		{name: "expired token", tokenString: expiredToken, secret: secret, wantErr: true},
		{name: "invalid token string", tokenString: "not.a.token", secret: secret, wantErr: true},
		{name: "empty token", tokenString: "", secret: secret, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := auth.ValidateJWT(tt.tokenString, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotID != tt.wantUserID {
				t.Errorf("ValidateJWT() = %v, want %v", gotID, tt.wantUserID)
			}
		})
	}
}
