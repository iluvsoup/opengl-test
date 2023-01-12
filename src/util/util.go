package util

import (
	"fmt"
	"reflect"
	"strconv"
)

func Sizeofint() int {
	return int(reflect.TypeOf((*int)(nil)).Elem().Size())
}

func Sizeofuint32() int { return 4 }
func Sizeoffloat32() int { return 4 }

func ThrowNotification(notificationMessage string) {
	fmt.Println("\x1b[1;34m" + notificationMessage + "\x1b[0m")
}

func ThrowWarning(warningMessage string) {
	fmt.Println("\x1b[1;33m" + warningMessage + "\x1b[0m")
}

func ThrowError(errorMessage error) {
	fmt.Print("\n")
	panic(
		"\x1b[1;31m" + fmt.Sprint(errorMessage),
	)
}

func ParseFloat(s string, bitSize int) float64 {
	number, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		ThrowError(err)
	}
	return number
}

func ParseInt(s string, base int, bitSize int) int64 {
	number, err := strconv.ParseInt(s, base, bitSize)
	if err != nil {
		ThrowError(err)
	}
	return number
}