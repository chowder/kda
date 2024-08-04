package validator

type Validator interface {
	Validate(username, password string) bool
}
