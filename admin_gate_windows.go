//go:build windows && !bindings

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

const adminRequiredMessage = "GRTBox must be run as administrator. Please close it and open it again with \"Run as administrator\". Even the basic GRTBox functions depend on administrator access."

func ensureAdministratorOrExit() {
	token, err := windows.OpenCurrentProcessToken()
	if err == nil {
		defer token.Close()
		if token.IsElevated() {
			return
		}
	}

	text := windows.StringToUTF16Ptr(adminRequiredMessage)
	title := windows.StringToUTF16Ptr("Administrator Required")
	windows.MessageBox(0, text, title, windows.MB_OK|windows.MB_ICONERROR)
	os.Exit(1)
}
