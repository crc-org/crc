package libhosty

import (
	"errors"
	"fmt"
)

//ErrNotAnAddressLine used when operating on a non-address line for operation
// related to address lines, such as comment/uncomment
var ErrNotAnAddressLine = errors.New("this line is not of type ADDRESS")

//ErrUncommentableLine used when try to comment a line that cannot be commented
var ErrUncommentableLine = errors.New("this line cannot be commented")

//ErrAlredyCommentedLine used when try to comment an alredy commented line
var ErrAlredyCommentedLine = errors.New("this line is alredy commented")

//ErrAlredyUncommentedLine used when try to uncomment an alredy uncommented line
var ErrAlredyUncommentedLine = errors.New("this line is alredy uncommented")

//ErrAddressNotFound used when provided address is not found
var ErrAddressNotFound = errors.New("cannot find a line with given address")

//ErrHostnameNotFound used when provided hostname is not found
var ErrHostnameNotFound = errors.New("cannot find a line with given hostname")

//ErrUnknown used when we don't know what's happened
var ErrUnknown = errors.New("unknown error")

//ErrCannotParseIPAddress used when unable to parse given ip address
func ErrCannotParseIPAddress(ip string) error {
	return fmt.Errorf("cannot parse IP Address: %s", ip)
}

//ErrUnrecognizedOS used when unable to recognize OS
func ErrUnrecognizedOS(os string) error {
	return fmt.Errorf("unrecognized OS: %s", os)
}
