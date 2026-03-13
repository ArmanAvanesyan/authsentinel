package cookie

// EncodeValue serializes v into a string for cookie storage.
// TODO: implement with encryption and integrity.
func EncodeValue(v any) (string, error) {
	return "", nil
}

// DecodeValue deserializes raw into dst.
// TODO: implement.
func DecodeValue(raw string, dst any) error {
	return nil
}
