package auth

import (
	"context"
	"errors"
	caregiverModel "server/internal/models/user/caregiver"
	"server/internal/services/user/caregiver"
	"server/internal/services/user/carepartner"
)

type AuthService struct {
	Google       GoogleAuthService
	Caregivers   *caregiver.CaregiverService
	Carepartners *carepartner.CarepartnerService
	Sessions     SessionService
}

func NewAuthService(
	google GoogleAuthService,
	caregivers *caregiver.CaregiverService,
	carepartners *carepartner.CarepartnerService,
	sessions SessionService,
) *AuthService {
	return &AuthService{google, caregivers, carepartners, sessions}
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

func (a *AuthService) UpdateCarepartnerID(ctx context.Context, userID int, carepartnerID int) error {
	// Find the caregiver to update info on
	user, err := a.Carepartners.GetCarepartnerByID(carepartnerID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// TODO - check that the carepartner actually exists (don't want to yet for testing purposes)
	/*
			partner, err := a.Caregivers.GetCaregiverByCarepartnerID(carepartnerID)

			if err != nil {
		    	return err
			}
			if partner == nil {
		    	return errors.New("invalid pairing code")
			}
	*/

	// Register their pairing code/care giver ID
	user.CaregiverID = userID

	// Update the record
	return a.Carepartners.UpdateCarepartner(user)
}
