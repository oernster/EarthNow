//go:build windows

package oslocale

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/oernster/EarthNow/internal/domain/region"
)

// GetUserDefaultGeoName answers Settings' "Country or region" as an ISO alpha-2
// code; a UN M.49 code such as 001 where no country is set (Windows 10 1709
// and later). The reference machine answers GB (measured).
var procGetUserDefaultGeoName = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetUserDefaultGeoName")

// geoNameCapacity holds either kind of code with its terminator and room to spare.
const geoNameCapacity = 16

func read() (string, func(string) (string, bool), error) {
	if err := procGetUserDefaultGeoName.Find(); err != nil {
		return "", region.Code, err
	}
	buf := make([]uint16, geoNameCapacity)
	n, _, err := procGetUserDefaultGeoName.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return "", region.Code, err
	}
	return windows.UTF16ToString(buf), region.Code, nil
}
