package auth

import (
	"context"
	"os"
	"server/internal/services/user"

	"google.golang.org/api/idtoken"
)

type GoogleAuthService interface {
	Verify(ctx context.Context, idToken string) (*user.GoogleUser, error)
}

type googleAuthService struct{}

func NewGoogleAuthService() GoogleAuthService {
	return &googleAuthService{}
}

func (g *googleAuthService) Verify(ctx context.Context, token string) (*user.GoogleUser, error) {
	payload, err := idtoken.Validate(ctx, token, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		return nil, err
	}

	return &user.GoogleUser{
		ID:    payload.Subject,
		Email: payload.Claims["email"].(string),
		Name:  payload.Claims["name"].(string),
	}, nil
}
