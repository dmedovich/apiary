package openapi

import (
	"testing"

	"github.com/honeynil/apiary/internal/parser"
)

func TestOperationSecurity(t *testing.T) {
	t.Run("absent inherits global (nil)", func(t *testing.T) {
		if got := operationSecurity(nil); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("none is an empty non-nil slice", func(t *testing.T) {
		got := operationSecurity([]string{"none"})
		reqs, ok := got.([]SecurityRequirement)
		if !ok {
			t.Fatalf("expected []SecurityRequirement, got %T", got)
		}
		if got == nil || len(reqs) != 0 {
			t.Errorf("expected empty non-nil slice (renders as security: []), got %v", reqs)
		}
	})

	t.Run("single scheme", func(t *testing.T) {
		reqs := operationSecurity([]string{"bearer"}).([]SecurityRequirement)
		if len(reqs) != 1 {
			t.Fatalf("expected 1 requirement, got %d", len(reqs))
		}
		if _, ok := reqs[0]["bearer"]; !ok {
			t.Errorf("expected bearer requirement, got %v", reqs[0])
		}
	})

	t.Run("none only special as sole element", func(t *testing.T) {
		reqs := operationSecurity([]string{"none", "bearer"}).([]SecurityRequirement)
		if len(reqs) != 2 {
			t.Errorf("expected 2 requirements (no none-shortcut), got %d", len(reqs))
		}
	})
}

func TestNeedsRequestBody(t *testing.T) {
	twoFields := &parser.TypeInfo{Fields: []*parser.FieldInfo{{Name: "A"}, {Name: "B"}}}

	cases := []struct {
		name       string
		typeInfo   *parser.TypeInfo
		bodyFields int
		want       bool
	}{
		{"has body fields", twoFields, 1, true},
		{"opaque type (nil typeInfo)", nil, 0, true},
		{"path/query/header only, no body fields", twoFields, 0, false},
		{"empty struct, no body fields", &parser.TypeInfo{}, 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := needsRequestBody(c.typeInfo, c.bodyFields); got != c.want {
				t.Errorf("needsRequestBody(%v, %d) = %v, want %v",
					c.typeInfo, c.bodyFields, got, c.want)
			}
		})
	}
}
