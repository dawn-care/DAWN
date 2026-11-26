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

	return googleUserFromClaims(payload.Subject, payload.Claims), nil
}

// googleUserFromClaims reads the claims of a verified Google ID token. Google can leave out
// email and name (and sends email_verified as a boolean, or as text in older tokens), so a
// missing or odd claim gives an empty value instead of a panic.
func googleUserFromClaims(subject string, claims map[string]any) *user.GoogleUser {
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	hostedDomain, _ := claims["hd"].(string)

	verified := false
	switch v := claims["email_verified"].(type) {
	case bool:
		verified = v
	case string:
		verified = v == "true"
	}

	return &user.GoogleUser{
		ID:            subject,
		Email:         email,
		Name:          name,
		EmailVerified: verified && email != "",
		HostedDomain:  hostedDomain,
	}
}
