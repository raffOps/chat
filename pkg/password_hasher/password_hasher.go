package password_hasher

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}
