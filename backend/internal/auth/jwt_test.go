package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIssueAndValidateStudentToken(t *testing.T) {
	secret := "test-secret"
	studentID := uuid.Must(uuid.Parse("a1b2c3d4-e5f6-7890-abcd-ef1234567890"))

	token, err := IssueStudentToken(secret, studentID, time.Hour)
	if err != nil {
		t.Fatalf("IssueStudentToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ValidateStudentToken(secret, token)
	if err != nil {
		t.Fatalf("ValidateStudentToken: %v", err)
	}
	if claims.Role != RoleStudent {
		t.Errorf("role: got %q, want %q", claims.Role, RoleStudent)
	}
	gotID, err := StudentIDFromClaims(claims)
	if err != nil {
		t.Fatalf("StudentIDFromClaims: %v", err)
	}
	if gotID != studentID {
		t.Errorf("student ID: got %v, want %v", gotID, studentID)
	}
}

func TestValidateStudentToken_wrongSecret(t *testing.T) {
	studentID := uuid.New()
	token, _ := IssueStudentToken("secret1", studentID, time.Hour)

	_, err := ValidateStudentToken("secret2", token)
	if err == nil {
		t.Fatal("expected error when validating with wrong secret")
	}
}

func TestValidateStudentToken_emptySecret(t *testing.T) {
	_, err := IssueStudentToken("", uuid.New(), time.Hour)
	if err == nil {
		t.Fatal("expected error when secret is empty")
	}
}
