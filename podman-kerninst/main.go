package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	MB_ICONWARNING             = 0x00000030
	MB_OK                      = 0x00000000
	MB_DEFBUTTON1              = 0x00000000
)

const KernelWarning = "WSL Kernel installation did not complete successfully. " +
	"Podman machine will attempt to install this at a later time. " +
	"You can also manually complete the installation using the " +
	"\"wsl --update\" command."

func setupLogging(name string) (*eventlog.Log, error) {
		// Reuse the Built-in .NET Runtime Source so that we do not
		// have to provide a messaage table and modify the system
		// event configuration
		log, err := eventlog.Open(".NET Runtime")
		if err != nil {
			return nil, err
		}
	
		logrus.AddHook(NewEventHook(log, name))
		logrus.SetLevel(logrus.InfoLevel)
	
		return log, nil
}
	
func silentExec(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func installWslKernel() error {
	logrus.Info("Installing WSL Kernel Update")
	var (
		err  error
	)
	backoff := 500 * time.Millisecond
	for i := 1; i < 6; i++ {
		err = silentExec("wsl", "--update")

		err = fmt.Errorf("oh no!")
		if err == nil {
			break
		}

		// In case of unusual circumstances (e.g. race with installer actions)
		// retry a few times
		logrus.Warn("An error occurred attempting the WSL Kernel update, retrying...")
		time.Sleep(backoff)
		backoff *= 2
	}

	if err != nil {
		err = fmt.Errorf("could not install WSL Kernel: %w", err)
	}

	return err
}

// Creates an "warn" style pop-up window
func warn(title string, caption string) int {
	format := MB_ICONWARNING|MB_OK|MB_DEFBUTTON1

	user32 := syscall.NewLazyDLL("user32.dll")
	captionPtr, _ := syscall.UTF16PtrFromString(caption)
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	ret, _, _ := user32.NewProc("MessageBoxW").Call(
		uintptr(0),
		uintptr(unsafe.Pointer(captionPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(format))

	return int(ret)
}

func main() {
	args := os.Args
	setupLogging(path.Base(args[0]))
	result := installWslKernel()
	if result != nil {
		logrus.Error(result.Error())
		_ = warn("Podman Setup", KernelWarning)
	}

	logrus.Info("WSL Kernel Update successful")
}