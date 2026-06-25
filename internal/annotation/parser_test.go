package annotation_test

import (
	"strings"
	"testing"

	"github.com/yaop-labs/apiary/internal/annotation"
)

func TestParse_Full(t *testing.T) {
	lines := []string{
		"apiary:operation POST /api/v1/auth/telegram",
		"summary: Authenticate via Telegram",
		"description: Accepts Telegram WebApp initData",
		"tags: auth, users",
		"errors: 400,401,403,500",
	}

	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if op.Method != "POST" {
		t.Errorf("expected POST, got %s", op.Method)
	}
	if op.Path != "/api/v1/auth/telegram" {
		t.Errorf("expected /api/v1/auth/telegram, got %s", op.Path)
	}
	if op.Summary != "Authenticate via Telegram" {
		t.Errorf("unexpected summary: %q", op.Summary)
	}
	if op.Description != "Accepts Telegram WebApp initData" {
		t.Errorf("unexpected description: %q", op.Description)
	}
	if len(op.Tags) != 2 || op.Tags[0] != "auth" || op.Tags[1] != "users" {
		t.Errorf("unexpected tags: %v", op.Tags)
	}
	if len(op.Errors) != 4 {
		t.Errorf("expected 4 error codes, got %d: %v", len(op.Errors), op.Errors)
	}
	if op.Errors[0].Code != 400 {
		t.Errorf("expected first error code 400, got %d", op.Errors[0].Code)
	}
}

func TestParse_ErrorsWithCustomSchema(t *testing.T) {
	lines := []string{
		"apiary:operation POST /api/v1/users",
		"errors: 400 ValidationError, 401, 500",
	}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if len(op.Errors) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(op.Errors))
	}
	if op.Errors[0].Code != 400 || op.Errors[0].Schema != "ValidationError" {
		t.Errorf("expected {400 ValidationError}, got %+v", op.Errors[0])
	}
	if op.Errors[1].Code != 401 || op.Errors[1].Schema != "" {
		t.Errorf("expected {401 }, got %+v", op.Errors[1])
	}
	if op.Errors[2].Code != 500 || op.Errors[2].Schema != "" {
		t.Errorf("expected {500 }, got %+v", op.Errors[2])
	}
}

func TestParse_UnknownKeyWarns(t *testing.T) {
	lines := []string{
		"apiary:operation GET /x",
		"summry: typo here",
	}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if len(op.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(op.Warnings), op.Warnings)
	}
	if !strings.Contains(op.Warnings[0], "summry") {
		t.Errorf("warning should name the bad key, got %q", op.Warnings[0])
	}
}

func TestParse_ProseWithColonDoesNotWarn(t *testing.T) {
	lines := []string{
		"apiary:operation GET /x",
		"summary: real summary",
		"Note: this handler is internal",
		"See https://example.com for details",
	}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if len(op.Warnings) != 0 {
		t.Errorf("expected no warnings for prose, got %v", op.Warnings)
	}
}

func TestParse_NoMarker(t *testing.T) {
	lines := []string{
		"summary: Some function",
		"description: Does something",
	}
	_, ok := annotation.Parse(lines)
	if ok {
		t.Fatal("expected parse to fail without apiary:operation marker")
	}
}

func TestParse_MethodUppercased(t *testing.T) {
	lines := []string{"apiary:operation get /health"}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if op.Method != "GET" {
		t.Errorf("expected GET, got %s", op.Method)
	}
}

func TestParse_MinimalMarker(t *testing.T) {
	lines := []string{"apiary:operation DELETE /api/v1/users/{id}"}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if op.Method != "DELETE" {
		t.Errorf("expected DELETE, got %s", op.Method)
	}
	if op.Path != "/api/v1/users/{id}" {
		t.Errorf("unexpected path: %s", op.Path)
	}
	if op.Summary != "" || op.Description != "" || len(op.Tags) != 0 || len(op.Errors) != 0 {
		t.Error("expected empty optional fields")
	}
}

func TestParse_DescriptionWithColon(t *testing.T) {
	lines := []string{
		"apiary:operation POST /api/v1/foo",
		"description: Returns key: value pairs",
	}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if op.Description != "Returns key: value pairs" {
		t.Errorf("unexpected description: %q", op.Description)
	}
}

func TestParse_SecurityBearer(t *testing.T) {
	lines := []string{
		"apiary:operation GET /api/v1/me",
		"summary: Current user",
		"security: bearer",
	}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if len(op.Security) != 1 || op.Security[0] != "bearer" {
		t.Errorf("expected security [bearer], got %v", op.Security)
	}
}

func TestParse_SecurityNone(t *testing.T) {
	lines := []string{
		"apiary:operation POST /api/v1/auth/login",
		"security: none",
	}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if len(op.Security) != 1 || op.Security[0] != "none" {
		t.Errorf("expected security [none], got %v", op.Security)
	}
}

func TestParse_SecurityMultiple(t *testing.T) {
	lines := []string{
		"apiary:operation GET /api/v1/admin",
		"security: bearer, apikey",
	}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if len(op.Security) != 2 {
		t.Errorf("expected 2 security schemes, got %v", op.Security)
	}
}

func TestParse_SecurityNilWhenAbsent(t *testing.T) {
	lines := []string{
		"apiary:operation GET /api/v1/items",
		"summary: List items",
	}
	op, ok := annotation.Parse(lines)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if op.Security != nil {
		t.Errorf("expected nil security (inherit global), got %v", op.Security)
	}
}
