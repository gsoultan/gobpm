package platformuser

import "github.com/gsoultan/metis/server/domains/entities"

// ListAccountsResponse carries every platform account with its roles.
type ListAccountsResponse struct {
	Accounts []entities.PlatformUser `json:"accounts"`
	Roles    []entities.PlatformRole `json:"roles"`
	Err      error                   `json:"err,omitzero"`
}

func (r ListAccountsResponse) Failed() error { return r.Err }

// SaveAccountRequest creates or updates an account.
//
// The id decides which: absent means create, and only then is the password
// read. Changing a password is not this call — it is the account holder's own
// act, or an explicit reset, and folding it in here would let a form that
// happened to carry a stale value overwrite a credential nobody meant to touch.
type SaveAccountRequest struct {
	ID          string   `json:"id,omitzero"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name,omitzero"`
	DisplayName string   `json:"display_name,omitzero"`
	Email       string   `json:"email,omitzero"`
	Roles       []string `json:"roles,omitzero"`

	// Password is only read on create, and is never returned.
	Password string `json:"password,omitzero"`
}

type SaveAccountResponse struct {
	ID  string `json:"id,omitzero"`
	Err error  `json:"err,omitzero"`
}

func (r SaveAccountResponse) Failed() error { return r.Err }

// DeleteAccountRequest names an account to remove.
type DeleteAccountRequest struct {
	ID string `json:"id"`
}

type DeleteAccountResponse struct {
	Err error `json:"err,omitzero"`
}

func (r DeleteAccountResponse) Failed() error { return r.Err }

// SetRolesRequest replaces what an account may do.
type SetRolesRequest struct {
	ID    string   `json:"id"`
	Roles []string `json:"roles"`
}

type SetRolesResponse struct {
	Err error `json:"err,omitzero"`
}

func (r SetRolesResponse) Failed() error { return r.Err }
