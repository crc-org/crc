package errors

import "encoding/json"

func ToSerializableError(err error) *SerializableError {
	if err == nil {
		return nil
	}
	return &SerializableError{err}
}

type SerializableError struct {
	error
}

func (e *SerializableError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error())
}

func (e *SerializableError) Unwrap() error {
	return e.error
}
