// Package authtest has test doubles for the auth services, shared by the tests of several
// packages so they cannot drift apart.
package authtest

import "context"

// FakeCodes is a LoginCodeService that records what it was asked
type FakeCodes struct {
	Requested  []string // "address|ip"
	Verified   []string // "address|code|ip"
	RequestErr error
	// VerifyErr is returned without running onAccepted, like a wrong code
	VerifyErr error
	// AcceptedErr is what onAccepted returned the last time the code was accepted
	AcceptedErr error
}

func (f *FakeCodes) RequestCode(ctx context.Context, address, ip string) error {
	f.Requested = append(f.Requested, address+"|"+ip)
	return f.RequestErr
}

func (f *FakeCodes) VerifyCode(ctx context.Context, address, code, ip string, onAccepted func() error) error {
	f.Verified = append(f.Verified, address+"|"+code+"|"+ip)
	if f.VerifyErr != nil {
		return f.VerifyErr
	}
	f.AcceptedErr = onAccepted()
	return f.AcceptedErr
}
