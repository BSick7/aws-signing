package signing

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// The payloadHash is the hex encoded SHA-256 hash of the request payload, and
// must be provided. Even if the request has no payload (aka body). If the
// request has no payload you should use the hex encoded SHA-256 of an empty
// string as the payloadHash value.
func payloadHash(body io.ReadSeeker) (string, error) {
	if body == nil {
		return emptyStringSHA256, nil
	}

	hash := sha256.New()
	start, err := body.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", err
	}
	defer func() {
		// ensure error is return if unable to seek back to start if payload
		_, err = body.Seek(start, io.SeekStart)
	}()

	io.Copy(hash, body)
	return hex.EncodeToString(hash.Sum(nil)), nil
}
