package wsl

import (
	"bufio"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func IsWSLFeatureEnabled() bool {
	return SilentExec("wsl", "--set-default-version", "2") == nil
}

func IsWSLInstalled() bool {
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

func SilentExec(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
