# Data inventory: caregiver accounts and sign-in

What the server stores about caregivers and signing in, and which outside services receive any of it.
It does not yet cover carepartner data, photos, videos or audio. Written for ticket #442 (sign in with an
emailed code); keep it up to date when sign-in changes.

## What we store

| Where | What | Comes from | How long |
|---|---|---|---|
| `Caregiver.name` | The caregiver's name | Their Google profile, or the part of their email before the `@` for a first email sign-in | Until the account is deleted |
| `Caregiver.email` | Email address, lowercase | Google (only a Gmail address, or a Google Workspace account's address), or typed for an email sign-in | Until the account is deleted |
| `Caregiver.google_id` | Google's permanent ID for the account | Google sign-in | Until the account is deleted |
| `Caregiver.role` | `caregiver` or `ADMIN` | Set by the server | Until the account is deleted |
| `Session` | A SHA-256 hash of the session cookie's token, the caregiver's ID, and creation, renewal and expiry times | Every sign-in | 90 days after the last renewal, and 1 year at most; or until they log out everywhere, or the account is deleted |
| `LoginCode` | The email, a keyed hash of the code, how many guesses were made at it, and creation, expiry and used times | Each emailed-code request that was emailed | See "Retention" below |
| `LoginCodeRequest` | The email, a keyed hash of the requesting IP address, and the time | Each emailed-code request that was emailed | See "Retention" below |
| `LoginCodeGuess` | A keyed hash of the IP address, and the time | Each wrong code entered | See "Retention" below |

Not stored: passwords (there are none), sign-in codes in plain text, and IP addresses in plain text in the database.

**Retention of codes, requests and guesses.** A code stops working after 10 minutes, or after one use. The rows are
deleted when they are more than 24 hours old, but only after someone asks for a code and the request goes through (a
refused or failed request does not clean up). There is no scheduled cleanup, so on a quiet server old rows can stay
longer than 24 hours.

## Server logs

- The request log records the method, path, client IP address and port, status and time of every request, in plain
  text, for every route. It does not record request bodies, so no email address or code is in it.
- Sign-in error lines can include text from the mail server, which may contain the recipient's address.
- Sign-in codes are never logged in production. In development with no mail server set up, the email, with its code,
  is printed to the log instead of sent.
- How long logs are kept is up to the host (Docker or the machine running the server). Nothing in this project sets it.

## Outside services

| Service | What it receives | Why | Status |
|---|---|---|---|
| Google (sign-in) | The caregiver signing in with their Google account. We get back their name, email and Google ID. | Google sign-in | In use |
| Mail server (SMTP): Google (Gmail) | The recipient's email address, the sender address, and the full text of each sign-in email, which includes the code | Sending emailed codes | **Production:** the project team's Gmail account (`smtp.gmail.com`), through an app password. **Development:** none by default (the email is printed to the server log); a developer who sets `SMTP_*` sends through the same Gmail account. |
| The recipient's own mail provider | The same email | Delivering it | Whatever provider the caregiver uses |

The mail provider sees caregiver email addresses, so it needs the compliance review below. A personal Gmail account is
covered by Google's consumer terms, which offer no business associate agreement; a Google Workspace account can sign one.
Personal Gmail also limits sending to about 500 emails a day.

## Security measures

- **Session cookie:** `HttpOnly`, `SameSite=Strict`, and `Secure` outside development. The token is only in the browser;
  the database holds a hash.
- **Cross-site requests:** the sign-in and logout POST routes refuse requests that a page on another site makes
  through the browser (`middleware.CrossOriginProtection`).
- **Codes:** 6 digits from the system's secure random source, valid 10 minutes, one use, and a newer code replaces an
  older one. Five wrong guesses burn a code. A code is stored only once the mail server accepts the email.
- **Rate limits:** 3 codes per email per 15 minutes, 10 per email per day, 10 per IP address per hour, and 20 wrong codes
  per IP address per hour across all addresses. Someone who knows a caregiver's address can still use up that address's
  10 codes a day, which stops the caregiver getting a code until the window passes; this is an accepted trade-off of
  emailed codes.
- **Linking Google and email accounts:** a Google sign-in is matched to an email account only by a Gmail address or a
  Google Workspace address. Google never re-checks other addresses, which may have passed to someone else.
- **Hashing:** codes and IP addresses are stored as HMAC-SHA256 under a secret (`LOGIN_CODE_SECRET`, at least 32
  characters), so a copy of the database alone cannot be used to find a code or see who asked. bcrypt is not used,
  because it is not a FIPS-approved algorithm.
- **Email in transit:** the server only sends over TLS and refuses a mail server that does not offer it, in development
  too.
- **Secrets stay out of URLs:** emails and codes only travel in request bodies. There are no magic links.
- **Passwords:** none are stored anywhere. The ticket notes this keeps caregiver logins out of NY SHIELD's
  email-plus-password category; compliance should confirm that.

## Open questions for St. Ann's compliance team

1. Is signing in with an emailed code acceptable as authentication?
2. What requirements apply to the email vendor (vendor review, a business associate agreement)?
3. Are the session length (90 days idle, 1 year at most) and the retention above acceptable?
4. Should Go's FIPS 140-3 mode (`GODEBUG=fips140=on`) be turned on? Only approved algorithms are used for the new
   code, but the server does not claim FIPS certification.
5. Is there a required log retention period?

If the answer to question 1 is no, storing caregiver emails and the code tables would need to be reworked.
