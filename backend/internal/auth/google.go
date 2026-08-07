package auth

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type GoogleClaims struct {
	Sub   string
	Email string
	Name  string
}

// VerifyGoogleIDToken validates the token's signature, expiry, and audience
// against the app's OAuth client IDs. audiences covers both the iOS and web
// client IDs since expo-auth-session's proxy flow can present either.
func VerifyGoogleIDToken(ctx context.Context, rawToken string, audiences []string) (GoogleClaims, error) {
	var lastErr error
	for _, aud := range audiences {
		payload, err := idtoken.Validate(ctx, rawToken, aud)
		if err != nil {
			lastErr = err
			continue
		}
		email, _ := payload.Claims["email"].(string)
		name, _ := payload.Claims["name"].(string)
		if email == "" {
			return GoogleClaims{}, fmt.Errorf("google id token missing email claim")
		}
		return GoogleClaims{Sub: payload.Subject, Email: email, Name: name}, nil
	}
	return GoogleClaims{}, fmt.Errorf("google id token validation failed: %w", lastErr)
}
