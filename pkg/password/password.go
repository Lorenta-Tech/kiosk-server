package password

import "golang.org/x/crypto/bcrypt"

func Hash(rawPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(rawPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func Verify(hash string, rawPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(rawPassword),
	)
}
