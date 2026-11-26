//go:build windows

package dialog

import (
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	ole32                = windows.NewLazySystemDLL("ole32.dll")
	shell32              = windows.NewLazySystemDLL("shell32.dll")
	procCoInitializeEx   = ole32.NewProc("CoInitializeEx")
	procCoUninitialize   = ole32.NewProc("CoUninitialize")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")
	procCoTaskMemFree    = ole32.NewProc("CoTaskMemFree")
)

const (
	coinitApartmentThreaded = 0x2
	clsctxInprocServer      = 0x1
	fosPickFolders          = 0x00000020
	fosForceFileSystem      = 0x00000040
	sigdnFileSysPath        = 0x80058000
)

var (
	clsidFileOpenDialog = windows.GUID{
		Data1: 0xDC1C5A9C, Data2: 0xE88A, Data3: 0x4DDE,
		Data4: [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7},
	}
	iidIFileOpenDialog = windows.GUID{
		Data1: 0xD57C7288, Data2: 0xD4AD, Data3: 0x4768,
		Data4: [8]byte{0xBE, 0x02, 0x9D, 0x96, 0x95, 0x32, 0xD9, 0x60},
	}
)

type comObj struct {
	vtbl *iFileOpenDialogVtbl
}

type iFileOpenDialogVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	Show                uintptr
	SetFileTypes        uintptr
	SetFileTypeIndex    uintptr
	GetFileTypeIndex    uintptr
	Advise              uintptr
	Unadvise            uintptr
	SetOptions          uintptr
	GetOptions          uintptr
	SetDefaultFolder    uintptr
	SetFolder           uintptr
	GetFolder           uintptr
	GetCurrentSelection uintptr
	SetFileName         uintptr
	GetFileName         uintptr
	SetTitle            uintptr
	SetOkButtonLabel    uintptr
	SetFileNameLabel    uintptr
	GetResult           uintptr
	AddPlace            uintptr
	SetDefaultExtension uintptr
	Close               uintptr
	SetClientGuid       uintptr
	ClearClientData     uintptr
	SetFilter           uintptr
	GetResults          uintptr
	GetSelectedItems    uintptr
}

type iShellItem struct {
	vtbl *iShellItemVtbl
}

type iShellItemVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	BindToHandler  uintptr
	GetParent      uintptr
	GetDisplayName uintptr
	GetAttributes  uintptr
	Compare        uintptr
}

// PickFolder opens the modern Windows folder picker dialog without spawning any console window.
func PickFolder(title string) string {
	resCh := make(chan string, 1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		hr, _, _ := procCoInitializeEx.Call(0, coinitApartmentThreaded)
		if hr == 0 || hr == 1 {
			defer procCoUninitialize.Call()
		}

		var dlg *comObj
		hr, _, _ = procCoCreateInstance.Call(
			uintptr(unsafe.Pointer(&clsidFileOpenDialog)),
			0,
			clsctxInprocServer,
			uintptr(unsafe.Pointer(&iidIFileOpenDialog)),
			uintptr(unsafe.Pointer(&dlg)),
		)
		if hr != 0 || dlg == nil {
			resCh <- ""
			return
		}
		defer syscall.SyscallN(dlg.vtbl.Release, uintptr(unsafe.Pointer(dlg)))

		var opts uint32
		_, _, _ = syscall.SyscallN(dlg.vtbl.GetOptions, uintptr(unsafe.Pointer(dlg)), uintptr(unsafe.Pointer(&opts)))
		opts |= fosPickFolders | fosForceFileSystem
		_, _, _ = syscall.SyscallN(dlg.vtbl.SetOptions, uintptr(unsafe.Pointer(dlg)), uintptr(opts))

		if title != "" {
			if ptr, err := windows.UTF16PtrFromString(title); err == nil {
				_, _, _ = syscall.SyscallN(dlg.vtbl.SetTitle, uintptr(unsafe.Pointer(dlg)), uintptr(unsafe.Pointer(ptr)))
			}
		}

		hr, _, _ = syscall.SyscallN(dlg.vtbl.Show, uintptr(unsafe.Pointer(dlg)), 0)
		if hr != 0 {
			resCh <- ""
			return
		}

		var item *iShellItem
		hr, _, _ = syscall.SyscallN(dlg.vtbl.GetResult, uintptr(unsafe.Pointer(dlg)), uintptr(unsafe.Pointer(&item)))
		if hr != 0 || item == nil {
			resCh <- ""
			return
		}
		defer syscall.SyscallN(item.vtbl.Release, uintptr(unsafe.Pointer(item)))

		var pszPath *uint16
		hr, _, _ = syscall.SyscallN(item.vtbl.GetDisplayName, uintptr(unsafe.Pointer(item)), sigdnFileSysPath, uintptr(unsafe.Pointer(&pszPath)))
		if hr != 0 || pszPath == nil {
			resCh <- ""
			return
		}
		defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(pszPath)))

		resCh <- windows.UTF16PtrToString(pszPath)
	}()

	return <-resCh
}
