package models

import "encoding/json"

// User represents the USERS table schema.
type User struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	FullName string `json:"fullname"`
	Role     string `json:"role"`
	PIN      string `json:"pin"`
	IsActive int    `json:"is_active"`
}

func (u *User) UnmarshalJSON(data []byte) error {
	type Alias User
	aux := &struct {
		ID          string `json:"id"`
		FullNameAlt string `json:"full_name"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if u.UID == "" && aux.ID != "" {
		u.UID = aux.ID
	}
	if u.FullName == "" && aux.FullNameAlt != "" {
		u.FullName = aux.FullNameAlt
	}
	return nil
}
