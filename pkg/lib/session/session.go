package session

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
)

type Type string

const (
	TypeIdentityProvider Type = "idp"
	TypeOfflineGrant     Type = "offline_grant"
)

type Session interface {
	SessionID() string
	SessionType() Type

	GetClientID() string
	GetCreatedAt() time.Time
	GetAccessInfo() *access.Info
	GetDeviceInfo() (map[string]interface{}, bool)

	GetAuthenticationInfo() authenticationinfo.T

	ToAPIModel() *model.Session
}

type DeleteReason string

const (
	DeleteReasonLogout DeleteReason = "logout"
	DeleteReasonRevoke DeleteReason = "revoke"
)

type CreateReason string

const (
	CreateReasonSignup         CreateReason = "signup"
	CreateReasonLogin          CreateReason = "login"
	CreateReasonPromote        CreateReason = "promote"
	CreateReasonReauthenticate CreateReason = "reauthenticate"
)
