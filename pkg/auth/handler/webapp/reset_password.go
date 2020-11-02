package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebResetPasswordHTML = template.RegisterHTML(
	"web/reset_password.html",
	components...,
)

const ResetPasswordRequestSchema = "ResetPasswordRequestSchema"

var ResetPasswordSchema = validation.NewMultipartSchema("").
	Add(ResetPasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"code": { "type": "string" },
				"x_password": { "type": "string" },
				"x_confirm_password": { "type": "string" }
			},
			"required": ["code", "x_password", "x_confirm_password"]
		}
	`).Instantiate()

func ConfigureResetPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/reset_password")
}

type ResetPasswordHandler struct {
	Database       *db.Handle
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
	WebApp         WebAppService
	PasswordPolicy PasswordPolicy
}

func (h *ResetPasswordHandler) MakeIntent(r *http.Request) *webapp.Intent {
	return &webapp.Intent{
		RedirectURI: "/reset_password/success",
		KeepState:   true,
		Intent:      intents.NewIntentResetPassword(),
	}
}

func (h *ResetPasswordHandler) GetData(r *http.Request, state *webapp.State) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}
	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		anyError,
		viewmodels.GetDefaultPasswordPolicyViewModelOptions(),
	)
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	return data, nil
}

type ResetPasswordInput struct {
	Code     string
	Password string
}

// GetCode implements InputResetPassword.
func (i *ResetPasswordInput) GetCode() string {
	return i.Code
}

// GetNewPassword implements InputResetPassword.
func (i *ResetPasswordInput) GetNewPassword() string {
	return i.Password
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	intent := h.MakeIntent(r)

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			state, err := h.WebApp.GetState(StateID(r))
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebResetPasswordHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				err = ResetPasswordSchema.PartValidator(ResetPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				code := r.Form.Get("code")
				newPassword := r.Form.Get("x_password")
				confirmPassword := r.Form.Get("x_confirm_password")
				err = pwd.ConfirmPassword(newPassword, confirmPassword)
				if err != nil {
					return
				}

				input = &ResetPasswordInput{
					Code:     code,
					Password: newPassword,
				}
				return
			})
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
