package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/db"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var ErrCollaboratorNotFound = apierrors.NotFound.WithReason("CollaboratorNotFound").New("collaborator not found")
var ErrCollaboratorUnauthorized = apierrors.Unauthorized.WithReason("CollaboratorUnauthorized").New("collaborator unauthorized")
var ErrCollaboratorSelfDeletion = apierrors.Forbidden.WithReason("CollaboratorSelfDeletion").New("cannot remove self from collaborator")

var ErrCollaboratorInvitationNotFound = apierrors.NotFound.WithReason("CollaboratorInvitationNotFound").New("collaborator invitation not found")
var ErrCollaboratorInvitationDuplicate = apierrors.AlreadyExists.WithReason("CollaboratorInvitationDuplicate").New("collaborator invitation duplicate")
var ErrCollaboratorInvitationInvalidCode = apierrors.Invalid.WithReason("CollaboratorInvitationInvalidCode").New("collaborator invitation invalid code")
var ErrCollaboratorInvitationDuplicateCode = apierrors.InternalError.WithReason("CollaboratorInvitationDuplicateCode").New("collaborator invitation duplicate code")

type CollaboratorService struct {
	Context     context.Context
	Clock       clock.Clock
	SQLBuilder  *db.SQLBuilder
	SQLExecutor *db.SQLExecutor
}

func (s *CollaboratorService) selectCollaborator() sq.SelectBuilder {
	return s.SQLBuilder.Select(
		"id",
		"app_id",
		"user_id",
		"created_at",
	).From(s.SQLBuilder.FullTableName("app_collaborator"))
}

func (s *CollaboratorService) selectCollaboratorInvitation() sq.SelectBuilder {
	return s.SQLBuilder.Select(
		"id",
		"app_id",
		"invited_by",
		"invitee_email",
		"code",
		"created_at",
		"expire_at",
	).From(s.SQLBuilder.FullTableName("app_collaborator_invitation"))
}

func (s *CollaboratorService) ListCollaborators(appID string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("app_id = ?", appID)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cs []*model.Collaborator
	for rows.Next() {
		c, err := scanCollaborator(rows)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}

	return cs, nil
}

func (s *CollaboratorService) ListCollaboratorsByUser(userID string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("user_id = ?", userID)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cs []*model.Collaborator
	for rows.Next() {
		c, err := scanCollaborator(rows)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}

	return cs, nil
}

func (s *CollaboratorService) NewCollaborator(appID string, userID string) *model.Collaborator {
	now := s.Clock.NowUTC()
	c := &model.Collaborator{
		ID:        uuid.New(),
		AppID:     appID,
		UserID:    userID,
		CreatedAt: now,
	}
	return c
}

func (s *CollaboratorService) CreateCollaborator(c *model.Collaborator) error {
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Insert(s.SQLBuilder.FullTableName("app_collaborator")).
		Columns(
			"id",
			"app_id",
			"user_id",
			"created_at",
		).
		Values(
			c.ID,
			c.AppID,
			c.UserID,
			c.CreatedAt,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *CollaboratorService) GetCollaborator(id string) (*model.Collaborator, error) {
	q := s.selectCollaborator().Where("id = ?", id)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return scanCollaborator(row)
}

func (s *CollaboratorService) GetCollaboratorByAppAndUser(appID string, userID string) (*model.Collaborator, error) {
	q := s.selectCollaborator().Where("app_id = ? AND user_id = ?", appID, userID)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return scanCollaborator(row)
}

func (s *CollaboratorService) DeleteCollaborator(c *model.Collaborator) error {
	sessionInfo := session.GetValidSessionInfo(s.Context)
	if c.UserID == sessionInfo.UserID {
		return ErrCollaboratorSelfDeletion
	}

	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Delete(s.SQLBuilder.FullTableName("app_collaborator")).
		Where("id = ?", c.ID),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *CollaboratorService) ListInvitations(appID string) ([]*model.CollaboratorInvitation, error) {
	now := s.Clock.NowUTC()
	q := s.selectCollaboratorInvitation().Where("app_id = ? AND expire_at > ?", appID, now)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*model.CollaboratorInvitation
	for rows.Next() {
		i, err := scanCollaboratorInvitation(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *CollaboratorService) SendInvitation(
	appID string,
	inviteeEmail string,
) (*model.CollaboratorInvitation, error) {
	sessionInfo := session.GetValidSessionInfo(s.Context)
	invitedBy := sessionInfo.UserID

	// FIXME(collaborator): Check if the invitee is collaborator
	// It is impossible now because Admin API does not have getUserByClaim

	// Check if the invitee has a pending invitation already.
	invitations, err := s.ListInvitations(appID)
	if err != nil {
		return nil, err
	}
	for _, i := range invitations {
		if i.InviteeEmail == inviteeEmail {
			return nil, ErrCollaboratorInvitationDuplicate
		}
	}

	code := generateCollaboratorInvitationCode()
	now := s.Clock.NowUTC()
	// Expire in 3 days.
	expireAt := now.Add(3 * 24 * time.Hour)

	i := &model.CollaboratorInvitation{
		ID:           uuid.New(),
		AppID:        appID,
		InvitedBy:    invitedBy,
		InviteeEmail: inviteeEmail,
		Code:         code,
		CreatedAt:    now,
		ExpireAt:     expireAt,
	}

	err = s.createCollaboratorInvitation(i)
	if err != nil {
		return nil, err
	}

	// FIXME(collaborator): Send the invitation email

	return i, nil
}

func (s *CollaboratorService) GetInvitation(id string) (*model.CollaboratorInvitation, error) {
	q := s.selectCollaboratorInvitation().Where("id = ?", id)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return scanCollaboratorInvitation(row)
}

func (s *CollaboratorService) DeleteInvitation(i *model.CollaboratorInvitation) error {
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Delete(s.SQLBuilder.FullTableName("app_collaborator_invitation")).
		Where("id = ?", i.ID),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *CollaboratorService) AcceptInvitation(code string) (*model.Collaborator, error) {
	actorID := session.GetValidSessionInfo(s.Context).UserID

	now := s.Clock.NowUTC()
	q := s.selectCollaboratorInvitation().Where("code = ? AND expire_at > ?", code, now)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*model.CollaboratorInvitation
	for rows.Next() {
		i, err := scanCollaboratorInvitation(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	if len(is) <= 0 {
		return nil, ErrCollaboratorInvitationInvalidCode
	}

	if len(is) > 1 {
		return nil, ErrCollaboratorInvitationDuplicateCode
	}

	invitation := is[0]

	// FIXME(collaborator): Check if the inviteeEmail matches actor.
	err = s.DeleteInvitation(invitation)
	if err != nil {
		return nil, err
	}

	collaborator := s.NewCollaborator(invitation.AppID, actorID)
	err = s.CreateCollaborator(collaborator)
	if err != nil {
		return nil, err
	}

	return collaborator, nil
}

func (s *CollaboratorService) createCollaboratorInvitation(i *model.CollaboratorInvitation) error {
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Insert(s.SQLBuilder.FullTableName("app_collaborator_invitation")).
		Columns(
			"id",
			"app_id",
			"invited_by",
			"invitee_email",
			"code",
			"created_at",
			"expire_at",
		).
		Values(
			i.ID,
			i.AppID,
			i.InvitedBy,
			i.InviteeEmail,
			i.Code,
			i.CreatedAt,
			i.ExpireAt,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func scanCollaborator(scan db.Scanner) (*model.Collaborator, error) {
	c := &model.Collaborator{}

	err := scan.Scan(
		&c.ID,
		&c.AppID,
		&c.UserID,
		&c.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCollaboratorNotFound
	} else if err != nil {
		return nil, err
	}

	return c, nil
}

func scanCollaboratorInvitation(scan db.Scanner) (*model.CollaboratorInvitation, error) {
	i := &model.CollaboratorInvitation{}

	err := scan.Scan(
		&i.ID,
		&i.AppID,
		&i.InvitedBy,
		&i.InviteeEmail,
		&i.Code,
		&i.CreatedAt,
		&i.ExpireAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCollaboratorInvitationNotFound
	} else if err != nil {
		return nil, err
	}

	return i, nil
}

func generateCollaboratorInvitationCode() string {
	code := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	return code
}
