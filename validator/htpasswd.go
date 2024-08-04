package validator

import (
	"fmt"
	"github.com/tg123/go-htpasswd"
)

type HtpasswdValidator struct {
	auth *htpasswd.File
}

func NewHtpasswdValidator(htpasswdFile string) (*HtpasswdValidator, error) {
	auth, err := htpasswd.New(htpasswdFile, htpasswd.DefaultSystems, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create htpasswd validator: %w", err)
	}

	return &HtpasswdValidator{auth: auth}, nil
}

func (v *HtpasswdValidator) Validate(username, password string) bool {
	return v.auth.Match(username, password)
}
