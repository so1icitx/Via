package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/atilatair/realput-bg/backend/internal/service/auth"
)

func signUpError(err error) (status int, message, code string) {
	if errors.Is(err, auth.ErrEmailExists) {
		return http.StatusConflict, "email already exists", "EMAIL_EXISTS"
	}
	if isPublicValidation(err) {
		return http.StatusBadRequest, err.Error(), "VALIDATION_ERROR"
	}
	return http.StatusBadRequest, "sign up failed", "SIGN_UP_FAILED"
}

func signInError(err error) (status int, message, code string) {
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return http.StatusUnauthorized, "invalid credentials", "INVALID_CREDENTIALS"
	}
	return http.StatusBadRequest, "sign in failed", "SIGN_IN_FAILED"
}

func updateProfileError(err error) (status int, message, code string) {
	if isPublicValidation(err) {
		return http.StatusBadRequest, err.Error(), "VALIDATION_ERROR"
	}
	return http.StatusBadRequest, "profile update failed", "UPDATE_FAILED"
}

func isPublicValidation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "required") || strings.Contains(msg, "at least")
}