package email

import (
	"fmt"
	"html"
	"time"
)

// LoginCodeSubject is the subject of every sign-in code email
const LoginCodeSubject = "Your D.A.W.N. sign-in code"

// LoginCodeMessage writes the sign-in code email in plain text and HTML. The code can be
// copied from either, and both say how long it works and what to do if it was not asked for.
func LoginCodeMessage(code string, validFor time.Duration) (text, htmlBody string) {
	minutes := int(validFor.Round(time.Minute) / time.Minute)
	if minutes < 1 {
		minutes = 1
	}
	unit := "minutes"
	if minutes == 1 {
		unit = "minute"
	}

	text = fmt.Sprintf(`Your D.A.W.N. sign-in code is: %s

It expires in %d %s. Enter it on the sign-in page to finish signing in.

If you did not ask for this code, you can ignore this email. Nobody can sign in without the code.
`, code, minutes, unit)

	htmlBody = fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, Helvetica, sans-serif; color: #1f2937; line-height: 1.5;">
<p>Your D.A.W.N. sign-in code is:</p>
<p style="font-size: 32px; font-weight: bold; letter-spacing: 6px; margin: 16px 0;">%s</p>
<p>It expires in %d %s. Enter it on the sign-in page to finish signing in.</p>
<p style="color: #6b7280; font-size: 14px;">If you did not ask for this code, you can ignore this email. Nobody can sign in without the code.</p>
</body>
</html>
`, html.EscapeString(code), minutes, unit)
	return text, htmlBody
}
