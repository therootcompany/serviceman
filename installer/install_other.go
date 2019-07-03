// +build !windows,!linux,!darwin

package installer

import (
	"git.rootprojects.org/root/go-serviceman/service"
)

func install(c *service.Service) error {
	return nil, nil
}
