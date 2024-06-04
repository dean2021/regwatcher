//go:build windows

package regwatcher

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
)

const (
	// HKeyClassesRoot represents HKEY_CLASSES_ROOT hive
	HKeyClassesRoot = windows.Handle(syscall.HKEY_CLASSES_ROOT)

	// HKeyCurrentUser represents HKEY_CURRENT_USER hive
	HKeyCurrentUser = windows.Handle(syscall.HKEY_CURRENT_USER)

	// HKeyLocalMachine represents HKEY_LOCAL_MACHINE hive
	HKeyLocalMachine = windows.Handle(syscall.HKEY_LOCAL_MACHINE)

	// HKeyUsers represents HKEY_USERS hive
	HKeyUsers = windows.Handle(syscall.HKEY_USERS)

	// HKeyCurrentConfig represents HKEY_CURRENT_CONFIG hive
	HKeyCurrentConfig = windows.Handle(syscall.HKEY_CURRENT_CONFIG)

	// HKeyPerformanceData represents HKEY_PERFORMANCE_DATA hive
	HKeyPerformanceData = windows.Handle(syscall.HKEY_PERFORMANCE_DATA)

	// Infinity is infinite timeout
	Infinity = 0xFFFFFFFF
)

var dwFilter = windows.REG_NOTIFY_CHANGE_NAME |
	windows.REG_NOTIFY_CHANGE_ATTRIBUTES |
	windows.REG_NOTIFY_CHANGE_LAST_SET |
	windows.REG_NOTIFY_CHANGE_SECURITY

type Watcher struct {
	hEvent  windows.Handle
	timeout int
	hKey    windows.Handle
}

func (w *Watcher) Create(hMainKey windows.Handle, regPath string) error {
	var hEvent windows.Handle
	var hPath, _ = syscall.UTF16PtrFromString(regPath) //"Software\\Microsoft\\Windows\\CurrentVersion\\Run")
	err := windows.RegOpenKeyEx(hMainKey, hPath, 0, syscall.KEY_NOTIFY, &w.hKey)
	if err != nil {
		return err
	}
	hEvent, err = windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return err
	}
	w.hEvent = hEvent
	return nil
}

func (w *Watcher) Watch() (bool, error) {
	err := windows.RegNotifyChangeKeyValue(w.hKey, true, uint32(dwFilter), w.hEvent, true)
	if err != nil {
		return false, fmt.Errorf("RegNotifyChangeKeyValue failed:%v", err.Error())
	}
	dwEvent, err := windows.WaitForSingleObject(w.hEvent, windows.INFINITE)
	if err != nil {
		return false, err
	}
	err = windows.ResetEvent(w.hEvent)
	if err != nil {
		return false, err
	}
	switch dwEvent {
	case uint32(windows.WAIT_TIMEOUT):
		return false, errors.New("WaitForSingleObject failed: Timeout")
	case uint32(windows.WAIT_OBJECT_0):
		return true, nil
	case uint32(windows.WAIT_ABANDONED):
		return false, errors.New("WaitForSingleObject return WAIT_ABANDONED")
	case uint32(windows.WAIT_FAILED):
		return false, errors.New("WaitForSingleObject return WAIT_FAILED")
	}
	return false, errors.New("unknown error")
}

func (w *Watcher) Close() error {
	err := windows.CloseHandle(w.hEvent)
	if err != nil {
		return fmt.Errorf("CloseHandle failed:%s", err.Error())
	}
	err = windows.CloseHandle(w.hKey)
	if err != nil {
		return fmt.Errorf("CloseHandle failed:%s", err.Error())
	}
	return nil
}

func NewWatcher(hMainKey windows.Handle, regPath string, timeout int) (*Watcher, error) {
	if timeout == 0 {
		timeout = windows.INFINITE
	}
	w := &Watcher{timeout: timeout}
	err := w.Create(hMainKey, regPath)
	if err != nil {
		return nil, err
	}
	return w, nil
}
