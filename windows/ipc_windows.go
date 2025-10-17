//go:build windows

// copied from github.com/wailsapp/wails
// MIT License

// Copyright (c) 2018-Present Lea Anthony

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package windows

import (
	"log"
	"os"
	"syscall"
	"unsafe"

	"github.com/tadvi/winc/w32"
	sys "golang.org/x/sys/windows"
)

type COPYDATASTRUCT struct {
	dwData uintptr
	cbData uint32
	lpData unsafe.Pointer
}

// WMCOPYDATA_SINGLE_INSTANCE_DATA we define our own type for WM_COPYDATA message
const WMCOPYDATA_SINGLE_INSTANCE_DATA = 1542

func sendMessage(hwnd w32.HWND, data string) {
	pCopyData := new(COPYDATASTRUCT)
	pCopyData.dwData = WMCOPYDATA_SINGLE_INSTANCE_DATA
	// Convert Go string to a null-terminated UTF-16 byte slice for Windows API.
	utf16Data, err := syscall.UTF16FromString(data)
	if err != nil {
		log.Fatalf("UTF16FromString failed: %v", err)
	}
	pCopyData.cbData = uint32(len(utf16Data) * 2) // Size in bytes
	pCopyData.lpData = unsafe.Pointer(&utf16Data[0])

	w32.SendMessage(hwnd, w32.WM_COPYDATA, 0, uintptr(unsafe.Pointer(pCopyData)))
}

// SetupSingleInstance single instance Windows app
func SetupSingleInstance(uniqueId string, secondInstanceBuffer chan<- string) bool {
	id := "convert4share-" + uniqueId

	className := id + "-sic"
	windowName := id + "-siw"
	mutexName := id + "-sim"

	mutexNamePtr, _ := syscall.UTF16PtrFromString(mutexName)
	_, err := sys.CreateMutex(nil, false, mutexNamePtr)

	if err != nil && err == sys.ERROR_ALREADY_EXISTS {
		// app is already running
		classNamePtr, _ := syscall.UTF16PtrFromString(className)
		windowNamePtr, _ := syscall.UTF16PtrFromString(windowName)

		// Use syscall to find the window handle directly from user32.dll
		user32 := syscall.NewLazyDLL("user32.dll")
		findWindow := user32.NewProc("FindWindowW")

		ret, _, callErr := findWindow.Call(
			uintptr(unsafe.Pointer(classNamePtr)),
			uintptr(unsafe.Pointer(windowNamePtr)),
		)

		hwnd := w32.HWND(ret)

		if hwnd != 0 {
			// Pass all file arguments to the running instance.
			// os.Args[0] is the program name, so we skip it.
			for _, arg := range os.Args[1:] {
				// We only send file paths, not commands like "install".
				sendMessage(hwnd, arg)
			}
			// exit second instance of app after sending message
			os.Exit(0)
		} else {
			// If FindWindow fails, we can't send the message.
			// Log the error and let the new instance start.
			log.Printf("could not find existing window, starting new instance. error: %v", callErr)
		}
		// if we got any other unknown error we will just start new application instance
	} else {
		go createEventTargetWindow(className, windowName, secondInstanceBuffer)
		return true
	}
	return false
}

func createEventTargetWindow(className string, windowName string, secondInstanceBuffer chan<- string) {
	// callback handler in the event target window
	wndProc := func(
		hwnd w32.HWND, msg uint32, wparam unsafe.Pointer, lparam unsafe.Pointer,
	) w32.LRESULT {
		if msg == w32.WM_COPYDATA {
			pCopyData := (*COPYDATASTRUCT)(lparam)

			if pCopyData.dwData == WMCOPYDATA_SINGLE_INSTANCE_DATA {
				// The data is a pointer to a UTF-16 string.
				// Create a slice of uint16s from the raw pointer.
				utf16Slice := unsafe.Slice((*uint16)(unsafe.Pointer(pCopyData.lpData)), pCopyData.cbData/2)
				filePath := syscall.UTF16ToString(utf16Slice)

				// Send the file path to the main goroutine
				secondInstanceBuffer <- filePath
			}

			return w32.LRESULT(0)
		}

		return w32.DefWindowProc(hwnd, msg, uintptr(wparam), uintptr(lparam))
	}

	classNamePtr, _ := syscall.UTF16PtrFromString(className)
	windowNamePtr, _ := syscall.UTF16PtrFromString(windowName)

	var class w32.WNDCLASSEX
	class.Size = uint32(unsafe.Sizeof(class))
	class.Style = 0
	class.WndProc = syscall.NewCallback(wndProc)
	class.ClsExtra = 0
	class.WndExtra = 0
	class.Instance = w32.GetModuleHandle("")
	class.Icon = 0
	class.Cursor = 0
	class.Background = 0
	class.MenuName = nil
	class.ClassName = classNamePtr
	class.IconSm = 0

	if w32.RegisterClassEx(&class) == 0 {
		log.Fatalf("RegisterClassEx failed: %v", syscall.GetLastError())
	}

	// create event window that will not be visible for user
	hwnd := w32.CreateWindowEx(
		0,
		classNamePtr,
		windowNamePtr,
		0,
		w32.CW_USEDEFAULT,
		w32.CW_USEDEFAULT,
		w32.CW_USEDEFAULT,
		w32.CW_USEDEFAULT,
		w32.HWND_MESSAGE,
		0,
		w32.GetModuleHandle(""),
		nil,
	)

	if hwnd == 0 {
		log.Fatalf("CreateWindowEx failed: %v", syscall.GetLastError())
	}

	// Run a message loop for this window to process incoming messages.
	// This is crucial for SendMessage to work across processes.
	var msg w32.MSG
	for w32.GetMessage(&msg, hwnd, 0, 0) > 0 {
		w32.TranslateMessage(&msg)
		w32.DispatchMessage(&msg)
	}
}
