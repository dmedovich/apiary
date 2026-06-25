package parser

import "testing"

func TestLeadingDocComment(t *testing.T) {
	t.Run("strips func name and splits summary/description", func(t *testing.T) {
		lines := []string{
			"CreateUser registers a new account.",
			"It validates the email and sends a confirmation.",
			"apiary:operation POST /users",
			"tags: users",
		}
		summary, desc := leadingDocComment(lines, "CreateUser")
		if summary != "Registers a new account." {
			t.Errorf("summary = %q", summary)
		}
		if desc != "It validates the email and sends a confirmation." {
			t.Errorf("description = %q", desc)
		}
	})

	t.Run("no prose before marker yields empty", func(t *testing.T) {
		lines := []string{"apiary:operation GET /x", "summary: explicit"}
		summary, desc := leadingDocComment(lines, "Handler")
		if summary != "" || desc != "" {
			t.Errorf("expected empty, got %q / %q", summary, desc)
		}
	})

	t.Run("line not starting with func name is kept verbatim", func(t *testing.T) {
		lines := []string{"Handles the login flow.", "apiary:operation POST /login"}
		summary, _ := leadingDocComment(lines, "Login")
		if summary != "Handles the login flow." {
			t.Errorf("summary = %q", summary)
		}
	})
}

func TestDefaultOperationID(t *testing.T) {
	cases := []struct {
		receiver string
		funcName string
		want     string
	}{
		{"TaskHandler", "List", "taskList"},
		{"CommentHandler", "List", "commentList"},
		{"AuthHandler", "Login", "authLogin"},
		{"UserController", "Create", "userCreate"},
		{"AccountService", "Get", "accountGet"},
		{"", "ListUsers", "listUsers"}, // free function
		{"Handler", "Ping", "ping"},    // receiver is only the suffix → drop it
		{"API", "Health", "aPIHealth"}, // no known suffix, kept as-is
	}
	for _, c := range cases {
		if got := defaultOperationID(c.receiver, c.funcName); got != c.want {
			t.Errorf("defaultOperationID(%q, %q) = %q, want %q", c.receiver, c.funcName, got, c.want)
		}
	}
}
