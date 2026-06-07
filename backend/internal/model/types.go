// Package model defines API contracts and domain types.
package model

import "encoding/json"

// QuestionsRequest is sent by the frontend to generate follow-up questions.
type QuestionsRequest struct {
	Category   string `json:"category"`
	UserInput  string `json:"userInput"`
	Lang       string `json:"lang"`
}

// Question is a single follow-up item in the guidance flow.
type Question struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Type     string   `json:"type"`
	Options  []string `json:"options"`
}

// QuestionsResponse wraps generated questions.
type QuestionsResponse struct {
	Questions []Question `json:"questions"`
}

// ResultsRequest carries answers from the question flow.
type ResultsRequest struct {
	Category  string            `json:"category"`
	UserInput string            `json:"userInput"`
	Answers   map[string]string `json:"answers"`
	Lang      string            `json:"lang"`
}

// ResultItem is one personalized path recommendation.
type ResultItem struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Summary          string `json:"summary"`
	WhyFits          string `json:"whyFits"`
	Opportunities    string `json:"opportunities"`
	HonestAssessment string `json:"honestAssessment"`
	NextSteps        string `json:"nextSteps"`
}

// Opportunity is a concrete program, internship, or event.
type Opportunity struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	HowToApply  string `json:"howToApply"`
	Timing      string `json:"timing"`
}

// ResultsPayload is the structured guidance returned to the UI.
type ResultsPayload struct {
	Intro         string        `json:"intro"`
	Items         []ResultItem  `json:"items"`
	Opportunities []Opportunity `json:"opportunities"`
}

// ErrorResponse is the standard API error envelope.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// AuthUser is the public user profile exposed to the frontend.
type AuthUser struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	Image         string `json:"image,omitempty"`
}

// SessionResponse mirrors the shape expected by the frontend session hook.
type SessionResponse struct {
	User    *AuthUser `json:"user,omitempty"`
	Session *struct {
		ID        string `json:"id"`
		ExpiresAt string `json:"expiresAt"`
	} `json:"session,omitempty"`
}

// SignUpRequest registers a user with email and password.
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// SignInRequest authenticates with email and password.
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateProfileRequest updates mutable profile fields.
type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// SaveResultRequest persists a guidance result for the signed-in user.
type SaveResultRequest struct {
	Query string          `json:"query"`
	Data  ResultsPayload  `json:"data"`
}

// SavedResultRow is a stored result row returned to the client.
type SavedResultRow struct {
	ID        int             `json:"id"`
	Query     string          `json:"query"`
	Data      json.RawMessage `json:"data"`
	CreatedAt string          `json:"createdAt"`
}