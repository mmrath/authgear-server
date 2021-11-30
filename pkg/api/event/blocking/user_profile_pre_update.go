package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserProfilePreUpdate event.Type = "user.profile.pre_update"
)

type UserProfilePreUpdateBlockingEventPayload struct {
	UserRef   model.UserRef `json:"-"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserProfilePreUpdateBlockingEventPayload) BlockingEventType() event.Type {
	return UserProfilePreUpdate
}

func (e *UserProfilePreUpdateBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserProfilePreUpdateBlockingEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *UserProfilePreUpdateBlockingEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserProfilePreUpdateBlockingEventPayload) ApplyMutations(mutations event.Mutations) (event.BlockingPayload, bool) {
	if mutations.User.StandardAttributes != nil {
		copied := *e
		copied.UserModel.StandardAttributes = mutations.User.StandardAttributes
		return &copied, true
	}

	return e, false
}

func (e *UserProfilePreUpdateBlockingEventPayload) GenerateFullMutations() event.Mutations {
	return event.Mutations{
		User: event.UserMutations{
			StandardAttributes: e.UserModel.StandardAttributes,
		},
	}
}

var _ event.BlockingPayload = &UserProfilePreUpdateBlockingEventPayload{}
