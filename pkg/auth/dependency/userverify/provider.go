package userverify

import (
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
)

type Provider interface {
	CreateVerifyCode(principal *password.Principal) (*VerifyCode, error)
	VerifyUser(
		passwordProvider password.Provider,
		authStore authinfo.Store,
		authInfo *authinfo.AuthInfo,
		code string,
	) (*VerifyCode, error)
}

type providerImpl struct {
	codeGenerator CodeGenerator
	store         Store
	config        config.UserVerificationConfiguration
	time          time.Provider
}

func NewProvider(
	codeGenerator CodeGenerator,
	store Store,
	config config.UserVerificationConfiguration,
	time time.Provider,
) Provider {
	return &providerImpl{
		codeGenerator: codeGenerator,
		store:         store,
		config:        config,
		time:          time,
	}
}

func (provider *providerImpl) CreateVerifyCode(principal *password.Principal) (*VerifyCode, error) {
	_, isValid := provider.config.LoginIDKeys[principal.LoginIDKey]
	if !isValid {
		return nil, skyerr.NewError(skyerr.InvalidArgument, "invalid login ID")
	}

	code := provider.codeGenerator.Generate(principal.LoginIDKey)

	verifyCode := NewVerifyCode()
	verifyCode.UserID = principal.UserID
	verifyCode.LoginIDKey = principal.LoginIDKey
	verifyCode.LoginID = principal.LoginID
	verifyCode.Code = code
	verifyCode.Consumed = false
	verifyCode.CreatedAt = provider.time.NowUTC()

	if err := provider.store.CreateVerifyCode(&verifyCode); err != nil {
		return nil, err
	}

	return &verifyCode, nil
}

func (provider *providerImpl) VerifyUser(
	passwordProvider password.Provider,
	authStore authinfo.Store,
	authInfo *authinfo.AuthInfo,
	code string,
) (*VerifyCode, error) {
	verifyCode, err := provider.store.GetVerifyCodeByUser(authInfo.ID)
	if err != nil {
		return nil, skyerr.NewError(skyerr.InvalidArgument, "invalid verification code")
	}

	if !verifyCode.Check(code) {
		return nil, skyerr.NewError(skyerr.InvalidArgument, "invalid verification code")
	}

	if verifyCode.Consumed {
		return nil, skyerr.NewError(skyerr.InvalidArgument, "invalid verification code")
	}

	principals, err := passwordProvider.GetPrincipalsByLoginID(
		verifyCode.LoginIDKey,
		verifyCode.LoginID,
	)
	if err == nil {
		// filter principals belonging to the user
		userPrincipals := []*password.Principal{}
		for _, principal := range principals {
			if principal.UserID == authInfo.ID {
				userPrincipals = append(userPrincipals, principal)
			}
		}
		principals = userPrincipals
	}

	if err != nil || len(principals) == 0 {
		return nil, skyerr.NewError(
			skyerr.InvalidArgument,
			"the login ID does not belong to the user",
		)
	}

	expiryTime := provider.config.LoginIDKeys[verifyCode.LoginIDKey].Expiry
	expireAt := verifyCode.CreatedAt.Add(gotime.Duration(expiryTime) * gotime.Second)
	if provider.time.NowUTC().After(expireAt) {
		return nil, skyerr.NewError(skyerr.InvalidArgument, "the code has expired")
	}

	err = provider.markUserVerified(passwordProvider, authStore, authInfo, verifyCode)
	if err != nil {
		return nil, err
	}

	return verifyCode, nil
}

func (provider *providerImpl) markUserVerified(
	passwordProvider password.Provider,
	authStore authinfo.Store,
	authInfo *authinfo.AuthInfo,
	verifyCode *VerifyCode,
) (err error) {
	if err = provider.store.MarkConsumed(verifyCode.ID); err != nil {
		return
	}

	principals, err := passwordProvider.GetPrincipalsByUserID(authInfo.ID)
	if err != nil {
		return
	}

	// Update user
	authInfo.VerifyInfo[verifyCode.LoginID] = true
	authInfo.Verified = IsUserVerified(
		authInfo.VerifyInfo,
		principals,
		provider.config.Criteria,
		provider.config.LoginIDKeys,
	)

	if err = authStore.UpdateAuth(authInfo); err != nil {
		return
	}

	return
}
