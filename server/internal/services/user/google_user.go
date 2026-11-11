package user

import "strings"

type GoogleUser struct {
	ID    string
	Email string
	Name  string
	// EmailVerified is true only when Google says it has checked that the person owns Email.
	EmailVerified bool
	// HostedDomain is the Google Workspace domain the account belongs to (the "hd" claim),
	// empty for a personal Google account
	HostedDomain string
}

// TrustedEmail is the email when it proves the person owns that address today, and empty
// otherwise. Only a trusted email may be saved on a caregiver or matched to one.
//
// Google checks an address once, when it is added to a Google account, and never again. That
// stays true only for addresses Google itself controls: Gmail, and the domain of a Google
// Workspace organization (which can delete the account when the address is given to someone
// else). A personal Google account made with any other address keeps "verified" even after
// that address passes to a new person, so it is not trusted. This follows Google's advice for
// using the email claim as an identity.
func (g *GoogleUser) TrustedEmail() string {
	if !g.EmailVerified || g.Email == "" {
		return ""
	}
	if g.HostedDomain != "" || strings.HasSuffix(strings.ToLower(strings.TrimSpace(g.Email)), "@gmail.com") {
		return g.Email
	}
	return ""
}
