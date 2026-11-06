package user

type GoogleUser struct {
	ID    string
	Email string
	Name  string
	// EmailVerified is true only when Google says it has checked that the person owns Email.
	// Only a verified email may be matched to an existing caregiver.
	EmailVerified bool
}
