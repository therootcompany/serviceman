// +build !windows,!linux,!darwin

package manager

import (
	"git.rootprojects.org/root/go-serviceman/service"
)

func install(c *service.Service) error {
	return nil, nil
}
