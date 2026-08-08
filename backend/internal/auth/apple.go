package auth

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

const appleJWKSURL = "https://appleid.apple.com/auth/keys"
const appleIssuer = "https://appleid.apple.com"

type AppleClaims struct {
	Sub   string
	Email string
}

type appleTokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// appleKeyfunc is created lazily on first use and reused across requests —
// keyfunc handles its own background JWKS refresh, so this should live for
// the process lifetime rather than being rebuilt per verification.
var appleKeyfunc jwt.Keyfunc

func getAppleKeyfunc(ctx context.Context) (jwt.Keyfunc, error) {
	if appleKeyfunc != nil {
		return appleKeyfunc, nil
	}
	k, err := keyfunc.NewDefaultCtx(ctx, []string{appleJWKSURL})
	if err != nil {
		return nil, fmt.Errorf("fetch apple jwks: %w", err)
	}
	appleKeyfunc = k.Keyfunc
	return appleKeyfunc, nil
}

// VerifyAppleIDToken validates an identity token from Sign in with Apple:
// signature against Apple's published JWKS, issuer, expiry, and audience
// (the app's bundle ID / Services ID). Apple only includes the user's name
// in the initial authorization response (not the token) and never resends
// it on subsequent sign-ins, so callers must capture it client-side on
// first sign-in — this only returns what the token itself carries.
func VerifyAppleIDToken(ctx context.Context, rawToken string, audiences []string) (AppleClaims, error) {
	keyfn, err := getAppleKeyfunc(ctx)
	if err != nil {
		return AppleClaims{}, err
	}

	var claims appleTokenClaims
	token, err := jwt.ParseWithClaims(rawToken, &claims, keyfn, jwt.WithIssuer(appleIssuer))
	if err != nil {
		return AppleClaims{}, fmt.Errorf("apple id token validation failed: %w", err)
	}
	if !token.Valid {
		return AppleClaims{}, fmt.Errorf("apple id token invalid")
	}

	audOK := false
	for _, tokenAud := range claims.Audience {
		for _, aud := range audiences {
			if tokenAud == aud {
				audOK = true
			}
		}
	}
	if !audOK {
		return AppleClaims{}, fmt.Errorf("apple id token audience mismatch")
	}
	if claims.Email == "" {
		return AppleClaims{}, fmt.Errorf("apple id token missing email claim")
	}

	return AppleClaims{Sub: claims.Subject, Email: claims.Email}, nil
}
