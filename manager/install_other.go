// +build !windows,!linux,!darwin

package manager

import (
	"git.rootprojects.org/root/go-serviceman/service"
)

func Render(c *service.Service) ([]byte, error) {
	return nil, nil
}

func install(c *service.Service) error {
	return nil, nil
}
