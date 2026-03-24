package entities

// Membership represents user-group relationship.
type Membership struct {
	User  *User  `json:"user,omitzero"`
	Group *Group `json:"group,omitzero"`
}
