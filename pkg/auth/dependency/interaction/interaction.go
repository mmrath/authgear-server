package interaction

import (
	"encoding/json"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// Interaction represents an interaction with authenticators/identities, and authentication process.
type Interaction struct {
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	ExpireAt    time.Time `json:"expire_at"`
	SessionID   string    `json:"session_id,omitempty"`
	SessionType string    `json:"session_type,omitempty"`
	ClientID    string    `json:"client_id,omitempty"`

	Intent Intent           `json:"-"`
	Error  *skyerr.APIError `json:"error,omitempty"`

	UserID                      string             `json:"user_id"`
	Identity                    *IdentityInfo      `json:"identity"`
	PrimaryAuthenticator        *AuthenticatorInfo `json:"primary_authenticator"`
	PrimaryAuthenticatorState   map[string]string  `json:"primary_authenticator_state,omitempty"`
	SecondaryAuthenticator      *AuthenticatorInfo `json:"secondary_authenticator"`
	SecondaryAuthenticatorState map[string]string  `json:"secondary_authenticator_state,omitempty"`

	PendingIdentity      *IdentityInfo       `json:"pending_identity,omitempty"`
	PendingAuthenticator *AuthenticatorInfo  `json:"pending_authenticator,omitempty"`
	NewIdentities        []IdentityInfo      `json:"new_identities,omitempty"`
	NewAuthenticators    []AuthenticatorInfo `json:"new_authenticators,omitempty"`
}

func (i *Interaction) IsNewIdentity(id string) bool {
	for _, identity := range i.NewIdentities {
		if identity.ID == id {
			return true
		}
	}
	return false
}

func (i *Interaction) IsNewAuthenticator(id string) bool {
	for _, authenticator := range i.NewAuthenticators {
		if authenticator.ID == id {
			return true
		}
	}
	return false
}

func (i *Interaction) MarshalJSON() ([]byte, error) {
	type jsonInteraction struct {
		*Interaction
		Intent     Intent     `json:"intent"`
		IntentType IntentType `json:"intent_type"`
	}
	ji := jsonInteraction{
		Interaction: i,
		Intent:      i.Intent,
		IntentType:  i.Intent.Type(),
	}
	return json.Marshal(ji)
}

func (i *Interaction) UnmarshalJSON(data []byte) error {
	type jsonInteraction struct {
		*Interaction
		Intent     json.RawMessage `json:"intent"`
		IntentType IntentType      `json:"intent_type"`
	}
	ji := &jsonInteraction{Interaction: i}
	if err := json.Unmarshal(data, ji); err != nil {
		return err
	}

	i.Intent = NewIntent(ji.IntentType)
	if err := json.Unmarshal(ji.Intent, i.Intent); err != nil {
		return err
	}

	return nil
}
