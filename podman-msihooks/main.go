package main

import (
	"C"
	"bufio"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const KernelWarning = "WSL Kernel installation did not complete successfully. " +
	"Podman machine will attempt to install this at a later time. " +
	"You can also manually complete the installation using the " +
	"\"wsl --update\" command."

//export CheckWSL
func CheckWSL(hInstall uint32) uint32 {
	installed := isWSLInstalled()
	feature := isWSLFeatureEnabled()
	setMsiProperty(hInstall, "HAS_WSL", strBool(installed))
	setMsiProperty(hInstall, "HAS_WSLFEATURE", strBool(feature))

	return 0
}

func setMsiProperty(hInstall uint32, name string, value string) {
	nameW, _ := syscall.UTF16PtrFromString(name)
	valueW, _ := syscall.UTF16PtrFromString(value)

	msi := syscall.NewLazyDLL("msi")
	proc := msi.NewProc("MsiSetPropertyW")
	_, _, _ = proc.Call(uintptr(hInstall), uintptr(unsafe.Pointer(nameW)), uintptr(unsafe.Pointer(valueW)))

}
func strBool(val bool) string {
	if val {
		return "1"
	}

	return "0"
}

func isWSLFeatureEnabled() bool {
	return silentExec(0, "wsl", "--set-default-version", "2") == nil
}

func isWSLInstalled() bool {
	cmd := exec.Command("wsl", "--status")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	out, err := cmd.StdoutPipe()
	cmd.Stderr = nil
	if err != nil {
		return false
	}
	if err = cmd.Start(); err != nil {
		return false
	}
	scanner := bufio.NewScanner(transform.NewReader(out, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))
	result := true
	for scanner.Scan() {
		line := scanner.Text()
		// Windows 11 does not set an error exit code when a kernel is not avail
		if strings.Contains(line, "kernel file is not found") {
			result = false
			break
		}
	}
	if err := cmd.Wait(); !result || err != nil {
		return false
	}

	return true
}

func silentExec(hInstall uint32, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func main() {}
