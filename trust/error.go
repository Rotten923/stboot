package trust

// Error reports problems regarding signature verification.
type Error string

// Error implements error interface.
func (e Error) Error() string {
	return string(e)
}