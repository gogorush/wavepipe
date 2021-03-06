package auth

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/mdlayher/wavepipe/data"
)

// tokenAuthenticate uses the token authentication method to log in to the API, returning
// a session user and a pair of client/server errors
func tokenAuthenticate(req *http.Request) (*data.User, *data.Session, error, error) {
	// Token for authentication
	var token string

	// Check for empty authorization header
	if req.Header.Get("Authorization") == "" {
		// If no header, check for credentials via querystring parameters
		token = req.URL.Query().Get("s")
	} else {
		// Fetch credentials from HTTP Basic auth
		tempToken, _, err := basicCredentials(req.Header.Get("Authorization"))
		if err != nil {
			return nil, nil, err, nil
		}

		// Copy credentials
		token = tempToken
	}

	// Check if token is blank
	if token == "" {
		return nil, nil, ErrNoToken, nil
	}

	// Attempt to load session by key
	session := new(data.Session)
	session.Key = token
	if err := session.Load(); err != nil {
		// Check for invalid user
		if err == sql.ErrNoRows {
			return nil, nil, ErrInvalidToken, nil
		}

		// Server error
		return nil, nil, nil, err
	}

	// Attempt to load associated user by user ID from session
	user := new(data.User)
	user.ID = session.UserID
	if err := user.Load(); err != nil {
		// Server error
		return nil, nil, nil, err
	}

	// Update session expiration date by 1 week
	session.Expire = time.Now().Add(7 * 24 * time.Hour).Unix()
	if err := session.Update(); err != nil {
		return nil, nil, nil, err
	}

	// No errors, return session user and session
	return user, session, nil, nil
}
