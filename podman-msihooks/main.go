package main

import (
	"C"
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)
import "os/user"

const (
	INSTALLMESSAGE_INFO        = 0x04000000
	INSTALLMESSAGE_PROGRESS    = 0x0A000000
	INSTALLMESSAGE_WARNING     = 0x02000000
	INSTALLMESSAGE_ACTIONSTART = 0x08000000
	MB_ICONWARNING             = 0x00000030
	MB_OK                      = 0x00000000
	MB_DEFBUTTON1              = 0x00000000
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

//export InstallWSLKernel
func InstallWSLKernel(hInstall uint32) uint32 {
	user, _ := user.Current()
	warnBox(hInstall, user.Username)
	result := installWslKernel(hInstall)
	if result != nil {
		log(hInstall, result.Error())
		warnBox(hInstall, KernelWarning)
	}

	return 0
}

func setMsiProperty(hInstall uint32, name string, value string) {
	nameW, _ := syscall.UTF16PtrFromString(name)
	valueW, _ := syscall.UTF16PtrFromString(value)

	msi := syscall.NewLazyDLL("msi")
	proc := msi.NewProc("MsiSetPropertyW")
	_, _, _ = proc.Call(uintptr(hInstall), uintptr(unsafe.Pointer(nameW)), uintptr(unsafe.Pointer(valueW)))

}

func msiCreateRecord(cParams uint32) uint32 {
	msi := syscall.NewLazyDLL("msi")
	proc := msi.NewProc("MsiCreateRecord")
	handle, _, _ := proc.Call(uintptr(cParams))

	return uint32(handle)
}

func msiCloseHandle(handle uint32) {
	msi := syscall.NewLazyDLL("msi")
	proc := msi.NewProc("MsiCloseHandle")
	_, _, _ = proc.Call(uintptr(handle))
}

func msiProcessMessage(hInstall uint32, messageType uint32, record uint32) {
	msi := syscall.NewLazyDLL("msi")
	proc := msi.NewProc("MsiProcessMessage")
	_, _, _ = proc.Call(uintptr(hInstall), uintptr(messageType), uintptr(record))
}

func msiRecordSetString(record uint32, field uint, value string) {
	valueW, _ := syscall.UTF16PtrFromString(value)
	msi := syscall.NewLazyDLL("msi")
	proc := msi.NewProc("MsiRecordSetStringW")
	_, _, _ = proc.Call(uintptr(record), uintptr(field), uintptr(unsafe.Pointer(valueW)))
}

func msiRecordSetInt(record uint32, field uint, value int) {
	msi := syscall.NewLazyDLL("msi")
	proc := msi.NewProc("MsiRecordSetInteger")
	_, _, _ = proc.Call(uintptr(record), uintptr(field), uintptr(value))
}

func log(hInstall uint32, message string) {
	record := msiCreateRecord(1)
	message = "InstallWSLKernel: " + message
	msiRecordSetString(record, 0, message)
	msiProcessMessage(hInstall, INSTALLMESSAGE_INFO, record)
	msiCloseHandle(record)
}

func warnBox(hInstall uint32, message string) {
	record := msiCreateRecord(1)
	msiRecordSetString(record, 0, message)
	msiProcessMessage(hInstall, INSTALLMESSAGE_WARNING|MB_ICONWARNING|MB_OK|MB_DEFBUTTON1, record)
	msiCloseHandle(record)
}

func setupProgressInfo(hInstall uint32, action string, message string) {
	record := msiCreateRecord(4)
	msiRecordSetString(record, 1, action)
	msiRecordSetString(record, 2, message)
	msiProcessMessage(hInstall, INSTALLMESSAGE_ACTIONSTART, record)
	msiCloseHandle(record)
}
func setupProgress(hInstall uint32, total int) {
	record := msiCreateRecord(5)
	msiRecordSetInt(record, 1, 0)
	msiRecordSetInt(record, 2, total)
	msiRecordSetInt(record, 3, 0)
	msiRecordSetInt(record, 4, 0)
	msiProcessMessage(hInstall, INSTALLMESSAGE_PROGRESS, record)
	msiCloseHandle(record)
}

func incProgress(hInstall uint32, progress int) {
	record := msiCreateRecord(5)
	msiRecordSetInt(record, 1, 2)
	msiRecordSetInt(record, 2, progress)
	msiRecordSetInt(record, 3, 0)
	msiRecordSetInt(record, 4, 0)
	msiProcessMessage(hInstall, INSTALLMESSAGE_PROGRESS, record)
	msiCloseHandle(record)
}

func logf(hInstall uint32, format string, args ...interface{}) {
	log(hInstall, fmt.Sprintf(format, args...))
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

type Progress struct {
	hInstall uint32
	total    int
	count    int
}

func (p *Progress) setup(hInstall uint32, total int, name string, desc string) {
	p.hInstall = hInstall
	p.total = total
	p.count = 0
	setupProgressInfo(hInstall, name, desc)
	setupProgress(hInstall, total)
}

func (p *Progress) inc(add int) {
	if p.count+add > p.total {
		add = p.total - p.count
	}
	p.count += add
	incProgress(p.hInstall, add)
}

func installWslKernel(hInstall uint32) error {
	//silentExec(0, "powershell", "powershell")
	log(hInstall, "Installing WSL Kernel Update")
	var (
		desc = "Updating WSL Kernel..."
		err  error
	)
	backoff := 500 * time.Millisecond
	for i := 1; i < 6; i++ {
		progress := Progress{}
		progress.setup(hInstall, 100, "InstallWSLKernel", desc)
		c := make(chan error)
		go func() {
			c <- silentExec(hInstall, "wsl", "--update")
		}()
	loop:
		for {
			select {
			case err = <-c:
				break loop
			case <-time.After(time.Second / 4):
				progress.inc(5)
			}
		}

		if err == nil {
			progress.inc(100)
			break
		}

		warnBox(hInstall, err.Error())

		desc = fmt.Sprintf("Updating WSL Kernel... (Retry %d)", i)

		// In case of unusual circumstances (e.g. race with installer actions)
		// retry a few times
		log(hInstall, "An error occurred attempting the WSL Kernel update, retrying...")
		time.Sleep(backoff)
		backoff *= 2
	}

	if err != nil {
		return fmt.Errorf("could not install WSL Kernel: %w", err)
	}

	return nil
}

func main() {}
