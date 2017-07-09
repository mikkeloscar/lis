package lis

// #cgo pkg-config: x11 xscrnsaver
// #include <X11/Xlib.h>
// #include <X11/extensions/scrnsaver.h>
//
// /* wrapper around the DefaultRootWindow macro */
// static Drawable wrap_DefaultRootWindow(Display *diplay) {
// 	return DefaultRootWindow(diplay);
// }
import "C"

import "fmt"

// XIdle returns the xserver idle time in miliseconds.
func XIdle() (uint, error) {
	var eventBase, errorBase C.int
	var info C.XScreenSaverInfo

	display := C.XOpenDisplay(C.CString(""))
	if display == nil {
		return 0, fmt.Errorf("xidle: unable to open X display")
	}
	defer C.XCloseDisplay(display)

	if C.XScreenSaverQueryExtension(display, &eventBase, &errorBase) > 0 {
		C.XScreenSaverQueryInfo(display, C.wrap_DefaultRootWindow(display), &info)
		return uint(info.idle), nil
	}

	return 0, fmt.Errorf("XScreenSaver Extension not present")
}
