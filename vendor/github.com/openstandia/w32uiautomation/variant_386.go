// +build 386
package w32uiautomation

import "github.com/go-ole/go-ole"

// type VARIANT struct {
// 	VT         uint16 //  2
// 	wReserved1 uint16 //  4
// 	wReserved2 uint16 //  6
// 	wReserved3 uint16 //  8
// 	Val        int64  // 16
// }

func VariantToUintptrArray(v ole.VARIANT) []uintptr {
	// Size of uintptr on 32bit system is 4
	return []uintptr{
		uintptr(v.VT), // uintptr(v.wReserved1)<<16 | uintptr(v.VT),
		uintptr(0),    // uintptr(v.wReserved3)<<16 | uintptr(v.wReserved2),
		uintptr(v.Val & 0xffffffff),
		uintptr(v.Val >> 32),
	}
}
