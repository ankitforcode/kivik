package serve

import (
	"bytes"
	"encoding/base64"
	"strconv"

	"github.com/flimzy/kivik/authdb"
	"github.com/pkg/errors"
)

// CreateAuthToken hashes a user name, salt, timestamp, and the server secret
// into an authentication token.
func (s *Service) CreateAuthToken(name, salt string, time int64) (string, error) {
	secret := s.getAuthSecret()
	return authdb.CreateAuthToken(name, salt, secret, time), nil
}

// ValidateCookie validates a cookie against a user context.
func (s *Service) ValidateCookie(user *authdb.UserContext, cookie string) (bool, error) {
	name, t, err := DecodeCookie(cookie)
	if err != nil {
		return false, err
	}
	token, err := s.CreateAuthToken(name, user.Salt, t)
	if err != nil {
		return false, err
	}
	return token == cookie, nil
}

// DecodeCookie decodes a Base64-encoded cookie, and returns its component
// parts.
func DecodeCookie(cookie string) (name string, created int64, err error) {
	data, err := base64.RawURLEncoding.DecodeString(cookie)
	if err != nil {
		return "", 0, err
	}
	parts := bytes.SplitN(data, []byte(":"), 3)
	t, err := strconv.ParseInt(string(parts[1]), 16, 64)
	if err != nil {
		return "", 0, errors.Wrap(err, "invalid timestamp")
	}
	return string(parts[0]), t, nil
}
