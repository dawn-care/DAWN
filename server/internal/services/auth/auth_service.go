package auth

import (
	"context"
	"errors"
	"fmt"
	caregiverModel "server/internal/models/user/caregiver"
	"server/internal/services/user/caregiver"
	"server/internal/services/user/carepartner"
)

type AuthService struct {
	Google       GoogleAuthService
	Caregivers   *caregiver.CaregiverService
	Carepartners *carepartner.CarepartnerService
	Sessions     SessionService
	// LoginCodes is nil when email sign-in is not set up (no mail server or no secret)
	LoginCodes LoginCodeService
}

func NewAuthService(
	google GoogleAuthService,
	caregivers *caregiver.CaregiverService,
	carepartners *carepartner.CarepartnerService,
	sessions SessionService,
) *AuthService {
	return &AuthService{Google: google, Caregivers: caregivers, Carepartners: carepartners, Sessions: sessions}
}

// WithLoginCodes turns on sign-in with an emailed code
func (a *AuthService) WithLoginCodes(codes LoginCodeService) *AuthService {
	a.LoginCodes = codes
	return a
}

// ErrEmailLoginDisabled means sign-in with an emailed code is not set up on this server (503)
var ErrEmailLoginDisabled = errors.New("sign-in with an emailed code is not available")

// RequestEmailCode emails a sign-in code. It does the same whether or not the address has an
// account, so it cannot be used to find out who has one.
func (a *AuthService) RequestEmailCode(ctx context.Context, address, ip string) error {
	if a.LoginCodes == nil {
		return ErrEmailLoginDisabled
	}
	return a.LoginCodes.RequestCode(ctx, address, ip)
}

// LoginWithEmailCode signs in the caregiver who owns the address, creating them if this is
// their first sign-in, once the code is right. It returns the token for their cookie.
func (a *AuthService) LoginWithEmailCode(ctx context.Context, address, code string) (string, error) {
	if a.LoginCodes == nil {
		return "", ErrEmailLoginDisabled
	}
	if err := a.LoginCodes.VerifyCode(ctx, address, code); err != nil {
		return "", err
	}

	user, err := a.Caregivers.FindOrCreateFromEmail(address)
	if err != nil {
		return "", err
	}

	return a.StartSession(ctx, user.ID)
}

// ErrGoogleTokenRejected means Google did not accept the ID token, so the login is refused (401).
// Any other login error is a failure on our side (500).
var ErrGoogleTokenRejected = errors.New("google rejected the ID token")

func (a *AuthService) LoginWithGoogle(ctx context.Context, idToken string) (string, error) {
	googleUser, err := a.Google.Verify(ctx, idToken)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGoogleTokenRejected, err)
	}

	user, err := a.Caregivers.FindOrCreateFromGoogle(googleUser)
	if err != nil {
		return "", err
	}

	return a.StartSession(ctx, user.ID)
}

// StartSession signs a caregiver in and returns the token for their cookie. Every way of
// logging in ends here, so they all get the same kind of session.
func (a *AuthService) StartSession(ctx context.Context, caregiverID int) (string, error) {
	return a.Sessions.Create(ctx, caregiverID)
}

// LogoutEverywhere ends every session of the caregiver the token belongs to, on all their devices.
// A token that is unknown or already expired no longer says whose sessions to end, so nothing
// else is ended and that is not an error; the caller still clears the cookie.
func (a *AuthService) LogoutEverywhere(ctx context.Context, token string) error {
	// Find, not Validate: renewing a session that is about to be deleted would be a wasted write
	session, err := a.Sessions.Find(ctx, token)
	if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionExpired) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("looking up the session to log out: %w", err)
	}
	if err := a.Sessions.DeleteAllForUser(ctx, session.UserID); err != nil {
		return fmt.Errorf("ending the sessions of caregiver %d: %w", session.UserID, err)
	}
	return nil
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
