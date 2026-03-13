package cookie

import "testing"

func TestEncodeDecodeValueRoundTrip(t *testing.T) {
	type payload struct {
		Foo string
		Bar int
	}

	src := payload{Foo: "baz", Bar: 42}

	encoded, err := EncodeValue(src)
	if err != nil {
		t.Fatalf("EncodeValue returned error: %v", err)
	}

	var dst payload
	if err := DecodeValue(encoded, &dst); err != nil {
		t.Fatalf("DecodeValue returned error: %v", err)
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := []byte("hello")
	key := []byte("dummy-key")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	out, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}

	if string(out) != string(plaintext) {
		t.Fatalf("expected %q, got %q", string(plaintext), string(out))
	}
}
