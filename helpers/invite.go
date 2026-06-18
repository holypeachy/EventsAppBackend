package helpers

import "crypto/rand"

const InviteCodeLength int = 8

const inviteCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func GenerateNewInviteCode(length int) (string, error) {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	result := make([]byte, length)

	for i, b := range bytes {
		result[i] = inviteCharset[int(b)%len(inviteCharset)]
	}

	return string(result), nil
}
