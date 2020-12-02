package main

import (
	"fmt"

	"git.rootprojects.org/root/serviceman/service"
)

func printLogMessage(conf *service.Service) {
	fmt.Printf("If all went well the logs should have been created at:\n\n\t%s\n", conf.Logdir)
}
