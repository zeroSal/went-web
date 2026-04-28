package user

import "testing"

func TestClaims_GetUsername(t *testing.T) {
	tests := []struct {
		name     string
		claims   Claims
		expected string
	}{
		{
			name: "username key",
			claims: Claims{
				"username": "john",
			},
			expected: "john",
		},
		{
			name: "sub key",
			claims: Claims{
				"sub": "user123",
			},
			expected: "user123",
		},
		{
			name:     "empty",
			claims:   Claims{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.claims.GetUsername(); got != tt.expected {
				t.Errorf("GetUsername() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClaims_GetRoles(t *testing.T) {
	tests := []struct {
		name     string
		claims   Claims
		expected []string
	}{
		{
			name: "roles array",
			claims: Claims{
				"roles": []interface{}{"admin", "user"},
			},
			expected: []string{"admin", "user"},
		},
		{
			name: "single role",
			claims: Claims{
				"role": "admin",
			},
			expected: []string{"admin"},
		},
		{
			name:     "empty",
			claims:   Claims{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.claims.GetRoles()
			if len(got) != len(tt.expected) {
				t.Errorf("GetRoles() = %v, want %v", got, tt.expected)
				return
			}
			for i, r := range got {
				if r != tt.expected[i] {
					t.Errorf("GetRoles()[%d] = %v, want %v", i, r, tt.expected[i])
				}
			}
		})
	}
}

func TestClaims_HasRole(t *testing.T) {
	claims := Claims{
		"roles": []interface{}{"admin", "user"},
	}

	if !claims.HasRole("admin") {
		t.Error("expected HasRole(admin) = true")
	}
	if !claims.HasRole("user") {
		t.Error("expected HasRole(user) = true")
	}
	if claims.HasRole("guest") {
		t.Error("expected HasRole(guest) = false")
	}
}
