package auth

import (
	"context"
	caregiverModel "server/internal/models/user/caregiver"
	"server/internal/services/user/caregiver"
)

type AuthService struct {
	Google     GoogleAuthService
	Caregivers *caregiver.CaregiverService
	Sessions   SessionService
}

func NewAuthService(
	google GoogleAuthService,
	caregivers *caregiver.CaregiverService,
	sessions SessionService,
) *AuthService {
	return &AuthService{google, caregivers, sessions}
}

func (a *AuthService) LoginWithGoogle(ctx context.Context, idToken string) (string, error) {
	googleUser, err := a.Google.Verify(ctx, idToken)
	if err != nil {
		return "", err
	}

	user, err := a.Caregivers.FindOrCreateFromGoogle(googleUser)
	if err != nil {
		return "", err
	}

	// creates the session and returns session ID
	return a.Sessions.Create(ctx, user.ID)
}

func (s *AuthService) GetUserByID(ctx context.Context, id int) (*caregiverModel.Caregiver, error) {
	return s.Caregivers.GetCaregiverByID(id)
}
