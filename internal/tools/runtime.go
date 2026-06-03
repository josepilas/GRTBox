package tools

import "strings"

const SupportedTCRuntimeVersion = "1.0.0"

func IsLogAction(action string) bool {
	return strings.HasPrefix(action, "log.")
}

func IsTCRuntime(runtime string) bool {
	return runtime == "tc"
}
