package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/guths/greenlight-api/internal/data"
	"github.com/guths/greenlight-api/internal/validator"
)

func (app *application) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	v := validator.New()

	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if user.Activated {
		v.AddError("email", "user has already been activated")
		app.failedValidationResponse(w, r, v.Errors)
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
		}

		err := app.mailer.Send(user.Email, "token_activation.tmpl", data)

		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	envelope := envelope{"message": "an email will be sent to your containing activation instructions"}

	err = app.writeJSON(w, http.StatusAccepted, envelope, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
