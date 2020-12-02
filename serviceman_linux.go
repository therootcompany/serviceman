package main

import (
	"fmt"

	"git.rootprojects.org/root/serviceman/manager"
	"git.rootprojects.org/root/serviceman/service"
)

func printLogMessage(conf *service.Service) {
	sudo := ""
	unit := "--unit"
	if conf.System {
		if !manager.IsPrivileged() {
			sudo = "sudo"
		}
	} else {
		unit = "--user-unit"
	}
	fmt.Println("If all went well you should be able to see some goodies in the logs:\n")
	fmt.Printf("\t%sjournalctl -xe %s %s.service\n", sudo, unit, conf.Name)
	if !conf.System {
		fmt.Println("\nIf that's not the case, see https://unix.stackexchange.com/a/486566/45554.")
		fmt.Println("(you may need to run `systemctl restart systemd-journald`)")
	}
}
