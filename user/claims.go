package user

import "slices"

type Claims map[string]any

func (c Claims) GetID() any {
	return c["sub"]
}

func (c Claims) GetUsername() string {
	if v, ok := c["username"].(string); ok {
		return v
	}
	if v, ok := c["sub"].(string); ok {
		return v
	}
	return ""
}

func (c Claims) GetRoles() []string {
	if v, ok := c["roles"].([]interface{}); ok {
		roles := make([]string, len(v))
		for i, r := range v {
			if s, ok := r.(string); ok {
				roles[i] = s
			}
		}
		return roles
	}
	if v, ok := c["roles"].([]string); ok {
		return v
	}
	if v, ok := c["role"].(string); ok {
		return []string{v}
	}
	return nil
}

func (c Claims) HasRole(role string) bool {
	roles := c.GetRoles()
	return slices.Contains(roles, role)
}
