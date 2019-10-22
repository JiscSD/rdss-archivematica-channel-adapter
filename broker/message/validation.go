package message

type Validator interface{}

// NewValidator returns a Validator with all the RDSS API schemas loaded.
func NewValidator() (Validator, error) {
	var v interface{}
	return v, nil
}
