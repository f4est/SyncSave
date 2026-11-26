//go:build windows

package tray

import (
	"runtime"
	"runtime/debug"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wmUser          = 0x0400
	wmTrayCallback  = wmUser + 1
	wmCommand       = 0x0111
	wmDestroy       = 0x0002
	wmLButtonUp     = 0x0202
	wmLButtonDblClk = 0x0203
	wmRButtonUp     = 0x0205
	wmSetIcon       = 0x0080

	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004
	nifInfo    = 0x00000010

	niifInfo = 0x00000001

	tpmLeftAlign   = 0x0000
	tpmBottomAlign = 0x0020
	tpmRightButton = 0x0002

	mfString    = 0x00000000
	mfChecked   = 0x00000008
	mfSeparator = 0x00000800

	idiApplication = 32512

	cmdShow        = 1001
	cmdSyncAll     = 1002
	cmdTogglePause = 1003
	cmdBackups     = 1004
	cmdAutostart   = 1005
	cmdExit        = 1006
)

var (
	user32                    = windows.NewLazySystemDLL("user32.dll")
	shell32                   = windows.NewLazySystemDLL("shell32.dll")
	kernel32                  = windows.NewLazySystemDLL("kernel32.dll")
	procRegisterClassExW      = user32.NewProc("RegisterClassExW")
	procCreateWindowExW       = user32.NewProc("CreateWindowExW")
	procDefWindowProcW        = user32.NewProc("DefWindowProcW")
	procGetMessageW           = user32.NewProc("GetMessageW")
	procTranslateMessage      = user32.NewProc("TranslateMessage")
	procDispatchMessageW      = user32.NewProc("DispatchMessageW")
	procPostQuitMessage       = user32.NewProc("PostQuitMessage")
	procLoadIconW             = user32.NewProc("LoadIconW")
	procCreatePopupMenu       = user32.NewProc("CreatePopupMenu")
	procAppendMenuW           = user32.NewProc("AppendMenuW")
	procSetMenuDefaultItem    = user32.NewProc("SetMenuDefaultItem")
	procTrackPopupMenu        = user32.NewProc("TrackPopupMenu")
	procDestroyMenu           = user32.NewProc("DestroyMenu")
	procSetForegroundWnd      = user32.NewProc("SetForegroundWindow")
	procGetCursorPos          = user32.NewProc("GetCursorPos")
	procPostMessageW          = user32.NewProc("PostMessageW")
	procSendMessageW          = user32.NewProc("SendMessageW")
	procShellNotifyIconW      = shell32.NewProc("Shell_NotifyIconW")
	procGetModuleHandleW      = kernel32.NewProc("GetModuleHandleW")
	procGetCurrentProcess     = kernel32.NewProc("GetCurrentProcess")
	procSetProcessWorkingSet  = kernel32.NewProc("SetProcessWorkingSetSize")
)

type wndClassExW struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type point struct {
	x, y int32
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type notifyIconDataW struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         windows.GUID
	hBalloonIcon     uintptr
}

// AcquireSingleInstance creates a named Windows mutex and returns a release func.
// Returns (nil, false) if another instance is already running.
func AcquireSingleInstance(name string) (func(), bool) {
	ptr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return func() {}, true
	}
	h, err := windows.CreateMutex(nil, false, ptr)
	if err != nil {
		if err == windows.ERROR_ALREADY_EXISTS {
			if h != 0 {
				_ = windows.CloseHandle(h)
			}
			return nil, false
		}
		return func() {}, true
	}
	return func() {
		_ = windows.CloseHandle(h)
	}, true
}

// TrimMemory releases unused Go heap pages and trims the Windows process working set
// when the GUI window is minimized to the tray.
func TrimMemory() {
	debug.FreeOSMemory()
	hProc, _, _ := procGetCurrentProcess.Call()
	if hProc != 0 {
		minusOne := ^uintptr(0)
		_, _, _ = procSetProcessWorkingSet.Call(hProc, minusOne, minusOne)
	}
}

// Callbacks wires tray menu actions back to the application.
type Callbacks struct {
	OnShow          func()
	OnSyncAll       func()
	OnTogglePause   func()
	OnOpenBackups   func()
	OnToggleStartup func()
	OnExit          func()
	IsSingleClick   func() bool
	IsPaused        func() bool
	IsAutostart     func() bool
}

// Tray manages a Win32 system tray icon in a dedicated OS thread.
type Tray struct {
	mu    sync.Mutex
	hwnd  uintptr
	hIcon uintptr
	nid   notifyIconDataW
	cb    Callbacks
}

// Start creates the system tray icon on its own locked OS thread.
func Start(tooltip string, cb Callbacks) *Tray {
	t := &Tray{cb: cb}
	ready := make(chan struct{})

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		hInst, _, _ := procGetModuleHandleW.Call(0)
		className, _ := windows.UTF16PtrFromString("SyncSaveTrayWindowClass")
		hIcon := buildAppIcon()
		if hIcon == 0 {
			hIcon, _, _ = procLoadIconW.Call(0, uintptr(idiApplication))
		}
		t.hIcon = hIcon

		wndProc := syscall.NewCallback(func(hwnd uintptr, uMsg uint32, wParam, lParam uintptr) uintptr {
			switch uMsg {
			case wmTrayCallback:
				switch uint32(lParam) {
				case wmLButtonUp:
					if t.cb.IsSingleClick != nil && t.cb.IsSingleClick() && t.cb.OnShow != nil {
						t.cb.OnShow()
					}
				case wmLButtonDblClk:
					if t.cb.OnShow != nil {
						t.cb.OnShow()
					}
				case wmRButtonUp:
					t.showMenu(hwnd)
				}
				return 0
			case wmCommand:
				switch int(wParam & 0xffff) {
				case cmdShow:
					if t.cb.OnShow != nil {
						t.cb.OnShow()
					}
				case cmdSyncAll:
					if t.cb.OnSyncAll != nil {
						t.cb.OnSyncAll()
					}
				case cmdTogglePause:
					if t.cb.OnTogglePause != nil {
						t.cb.OnTogglePause()
					}
				case cmdBackups:
					if t.cb.OnOpenBackups != nil {
						t.cb.OnOpenBackups()
					}
				case cmdAutostart:
					if t.cb.OnToggleStartup != nil {
						t.cb.OnToggleStartup()
					}
				case cmdExit:
					if t.cb.OnExit != nil {
						t.cb.OnExit()
					}
				}
				return 0
			case wmDestroy:
				_, _, _ = procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&t.nid)))
				_, _, _ = procPostQuitMessage.Call(0)
				return 0
			}
			ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(uMsg), wParam, lParam)
			return ret
		})

		wc := wndClassExW{
			cbSize:        uint32(unsafe.Sizeof(wndClassExW{})),
			lpfnWndProc:   wndProc,
			hInstance:     hInst,
			hIcon:         hIcon,
			lpszClassName: className,
		}
		_, _, _ = procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

		hwnd, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(className)),
			uintptr(unsafe.Pointer(className)),
			0, 0, 0, 0, 0,
			0, 0, hInst, 0,
		)
		t.hwnd = hwnd

		t.nid = notifyIconDataW{
			cbSize:           uint32(unsafe.Sizeof(notifyIconDataW{})),
			hWnd:             hwnd,
			uID:              1,
			uFlags:           nifMessage | nifIcon | nifTip,
			uCallbackMessage: wmTrayCallback,
			hIcon:            hIcon,
		}
		copyUTF16(t.nid.szTip[:], tooltip)
		_, _, _ = procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&t.nid)))
		close(ready)

		var m msg
		for {
			r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
			if int32(r) <= 0 {
				break
			}
			_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			_, _, _ = procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
		}
	}()

	<-ready
	return t
}

// ApplyWindowIcon sets the custom SyncSave icon on a Win32 HWND (e.g., WebView2 window).
func (t *Tray) ApplyWindowIcon(hwnd unsafe.Pointer) {
	if t == nil || t.hIcon == 0 || hwnd == nil {
		return
	}
	_, _, _ = procSendMessageW.Call(uintptr(hwnd), wmSetIcon, 0, t.hIcon) // ICON_SMALL
	_, _, _ = procSendMessageW.Call(uintptr(hwnd), wmSetIcon, 1, t.hIcon) // ICON_BIG
}

// ShowBalloon displays a Windows tray notification balloon.
func (t *Tray) ShowBalloon(title, message string) {
	if t == nil || t.hwnd == 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nid.uFlags = nifInfo
	t.nid.dwInfoFlags = niifInfo
	copyUTF16(t.nid.szInfoTitle[:], title)
	copyUTF16(t.nid.szInfo[:], message)
	_, _, _ = procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&t.nid)))
}

func (t *Tray) showMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer procDestroyMenu.Call(hMenu)

	appendItem := func(flags uintptr, id uintptr, label string) {
		ptr, _ := windows.UTF16PtrFromString(label)
		_, _, _ = procAppendMenuW.Call(hMenu, flags, id, uintptr(unsafe.Pointer(ptr)))
	}

	appendItem(mfString, cmdShow, "Открыть SyncSave")
	_, _, _ = procSetMenuDefaultItem.Call(hMenu, cmdShow, 0, 0)
	_, _, _ = procAppendMenuW.Call(hMenu, mfSeparator, 0, 0)

	appendItem(mfString, cmdSyncAll, "Синхронизировать всё сейчас")

	pauseFlags := uintptr(mfString)
	if t.cb.IsPaused != nil && t.cb.IsPaused() {
		pauseFlags |= mfChecked
	}
	appendItem(pauseFlags, cmdTogglePause, "Пауза авто-синхронизации")

	appendItem(mfString, cmdBackups, "Открыть папку бэкапов")
	_, _, _ = procAppendMenuW.Call(hMenu, mfSeparator, 0, 0)

	autostartFlags := uintptr(mfString)
	if t.cb.IsAutostart != nil && t.cb.IsAutostart() {
		autostartFlags |= mfChecked
	}
	appendItem(autostartFlags, cmdAutostart, "Автозапуск с Windows")
	_, _, _ = procAppendMenuW.Call(hMenu, mfSeparator, 0, 0)

	appendItem(mfString, cmdExit, "Выход")

	var pt point
	_, _, _ = procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	_, _, _ = procSetForegroundWnd.Call(hwnd)
	_, _, _ = procTrackPopupMenu.Call(hMenu, tpmLeftAlign|tpmBottomAlign|tpmRightButton, uintptr(pt.x), uintptr(pt.y), 0, hwnd, 0)
}

// Stop removes the tray icon and exits the tray message loop.
func (t *Tray) Stop() {
	if t != nil && t.hwnd != 0 {
		_, _, _ = procPostMessageW.Call(t.hwnd, wmDestroy, 0, 0)
	}
}

func copyUTF16(dst []uint16, s string) {
	u, err := windows.UTF16FromString(s)
	if err != nil {
		return
	}
	n := len(u)
	if n > len(dst)-1 {
		n = len(dst) - 1
	}
	copy(dst[:n], u[:n])
	dst[n] = 0
}
