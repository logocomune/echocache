package echocache

import "math/rand/v2"

// randString generates a random alphanumeric string of the specified length, ensuring the first character is a letter.
func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const letterset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	if length <= 0 {
		return ""
	}
	result := make([]byte, length)
	result[0] = letterset[rand.IntN(len(letterset))] // Ensure first character is a letter
	for i := 1; i < length; i++ {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}
