package cookie

// Encrypt encrypts plaintext for use in cookie values.
// TODO: implement using authenticated encryption.
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	return plaintext, nil
}

// Decrypt decrypts a cookie value.
// TODO: implement.
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	return ciphertext, nil
}
