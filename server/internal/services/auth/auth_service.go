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
	// LoginCodes sends and checks emailed sign-in codes. main always sets it (WithLoginCodes);
	// the server does not start without a working email setup.
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

// WithLoginCodes sets the service that sends and checks emailed sign-in codes
func (a *AuthService) WithLoginCodes(codes LoginCodeService) *AuthService {
	a.LoginCodes = codes
	return a
}

// RequestEmailCode emails a sign-in code. It does the same whether or not the address has an
// account, so it cannot be used to find out who has one.
func (a *AuthService) RequestEmailCode(ctx context.Context, address, ip string) error {
	return a.LoginCodes.RequestCode(ctx, address, ip)
}

// LoginWithEmailCode signs in the caregiver who owns the address, creating them if this is
// their first sign-in, once the code is right. It returns the token for their cookie.
// The code is used up only once the session exists, so if finding the caregiver or starting
// the session fails, the same code can be entered again.
func (a *AuthService) LoginWithEmailCode(ctx context.Context, address, code, ip string) (string, error) {
	var token string
	err := a.LoginCodes.VerifyCode(ctx, address, code, ip, func() error {
		user, err := a.Caregivers.FindOrCreateFromEmail(address)
		if err != nil {
			return fmt.Errorf("finding or creating the caregiver for email sign-in: %w", err)
		}
		token, err = a.StartSession(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("starting a session for caregiver %d: %w", user.ID, err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return token, nil
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
