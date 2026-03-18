package cookie

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestEncodeDecodeValueRoundTrip(t *testing.T) {
	type payload struct {
		Foo string
		Bar int
	}

	src := payload{Foo: "baz", Bar: 42}
	key := []byte("test-secret-32-bytes-long enough!!")

	encoded, err := EncodeValue(src, key)
	if err != nil {
		t.Fatalf("EncodeValue returned error: %v", err)
	}

	var dst payload
	if err := DecodeValue(encoded, &dst, key); err != nil {
		t.Fatalf("DecodeValue returned error: %v", err)
	}
	if dst.Foo != src.Foo || dst.Bar != src.Bar {
		t.Fatalf("expected %+v, got %+v", src, dst)
	}
}

func TestEncodeDecodeValueRoundTripNoKey(t *testing.T) {
	type payload struct {
		Foo string
	}
	src := payload{Foo: "baz"}
	encoded, err := EncodeValue(src, nil)
	if err != nil {
		t.Fatalf("EncodeValue returned error: %v", err)
	}
	var dst payload
	if err := DecodeValue(encoded, &dst, nil); err != nil {
		t.Fatalf("DecodeValue returned error: %v", err)
	}
	if dst.Foo != src.Foo {
		t.Fatalf("expected %+v, got %+v", src, dst)
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := []byte("hello")
	key := []byte("dummy-key")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("Encrypt should not return plaintext as-is")
	}

	out, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}

	if string(out) != string(plaintext) {
		t.Fatalf("expected %q, got %q", string(plaintext), string(out))
	}
}

func TestSignedManagerWithKeysRotation(t *testing.T) {
	oldSecret := "old-secret"
	newSecret := "new-secret"
	m := NewSignedManagerWithKeys(newSecret, oldSecret)

	// Sign with new secret
	val, err := m.Encode("session-123")
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var s string
	if err := m.Decode(val, &s); err != nil {
		t.Fatalf("Decode (new): %v", err)
	}
	if s != "session-123" {
		t.Fatalf("expected session-123, got %q", s)
	}

	// Value signed with old secret (simulate by creating with old manager)
	mOld := NewSignedManager(oldSecret)
	valOld, _ := mOld.Encode("session-old")
	if err := m.Decode(valOld, &s); err != nil {
		t.Fatalf("Decode (old key): %v", err)
	}
	if s != "session-old" {
		t.Fatalf("expected session-old, got %q", s)
	}
}

// TestDecodeValue_WrongKey verifies that decoding with a different key fails.
func TestDecodeValue_WrongKey(t *testing.T) {
	type payload struct{ Foo string }
	key := []byte("test-secret-32-bytes-long enough!!")
	encoded, err := EncodeValue(payload{Foo: "secret"}, key)
	if err != nil {
		t.Fatalf("EncodeValue: %v", err)
	}
	wrongKey := []byte("other-secret-32-bytes-long enough!!")
	var dst payload
	err = DecodeValue(encoded, &dst, wrongKey)
	if err == nil {
		t.Fatal("expected error when decoding with wrong key")
	}
	if err != ErrDecrypt && err != ErrCodecDecode {
		t.Errorf("expected ErrDecrypt or ErrCodecDecode, got %v", err)
	}
}

// TestDecodeValue_TamperedPayload verifies that tampering with the encrypted payload causes decode to fail.
func TestDecodeValue_TamperedPayload(t *testing.T) {
	key := []byte("test-secret-32-bytes-long enough!!")
	encoded, err := EncodeValue(struct{ X int }{42}, key)
	if err != nil {
		t.Fatalf("EncodeValue: %v", err)
	}
	raw, _ := base64.URLEncoding.DecodeString(encoded)
	if len(raw) < 2 {
		t.Fatal("encoded value too short to tamper")
	}
	raw[0] ^= 0x01
	tampered := base64.URLEncoding.EncodeToString(raw)
	var dst struct{ X int }
	err = DecodeValue(tampered, &dst, key)
	if err == nil {
		t.Fatal("expected error when decoding tampered payload")
	}
}

// TestDecodeValue_InvalidBase64 verifies that invalid base64 returns ErrCodecDecode.
func TestDecodeValue_InvalidBase64(t *testing.T) {
	key := []byte("test-secret-32-bytes-long enough!!")
	var dst struct{ Foo string }
	err := DecodeValue("not-valid-base64!!!", &dst, key)
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
	if err != ErrCodecDecode {
		t.Errorf("expected ErrCodecDecode, got %v", err)
	}
}

// TestDecrypt_WrongKey verifies that Decrypt with wrong key fails.
func TestDecrypt_WrongKey(t *testing.T) {
	key := []byte("encryption-key-32-bytes-long!!!!!")
	plain := []byte("secret")
	ct, err := Encrypt(plain, key)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	wrongKey := []byte("wrong-key-32-bytes-long!!!!!!!!!")
	_, err = Decrypt(ct, wrongKey)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
	if err != ErrDecrypt {
		t.Errorf("expected ErrDecrypt, got %v", err)
	}
}

// TestSignedManager_Decode_TamperedValue verifies that SignedManager.Decode rejects tampered values.
func TestSignedManager_Decode_TamperedValue(t *testing.T) {
	m := NewSignedManager("secret")
	val, err := m.Encode("session-123")
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// Tamper: flip a character in the payload part (before the dot).
	tampered := val
	if idx := bytes.IndexByte([]byte(val), '.'); idx > 0 {
		b := []byte(val)
		b[0] ^= 0x01
		tampered = string(b)
	}
	var s string
	err = m.Decode(tampered, &s)
	if err == nil {
		t.Fatal("expected error when decoding tampered signed value")
	}
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}
