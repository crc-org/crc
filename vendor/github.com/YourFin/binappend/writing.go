// Copyright © 2018 Patrick Nuckolls <nuckollsp at gmail>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package binappend

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"
)

type appendedData struct {
	StartFilePtr int64 `json:"start_file_pointer"`
	BlockSize    int64 `json:"block_size"`
	Zipped       bool  `json:"zipped"`
}

const METADATA_VERSION string = "0.2"

type appendedMetadata struct {
	Version string
	Data    map[string]appendedData
}

type Appender struct {
	fileHandle *os.File
	metadata   appendedMetadata
	mux        *sync.Mutex
	filename   string
}

// Procedure:
//  MakeAppender
// Purpose:
//  To create an Appender
// Parameters:
//  The name of the file to append to: filename string
//  A function that wraps an io.Writer: writeWrapper
//    This can be used to pre-process data before insertion
//    Note: this function will be called every time a file/stream is added
// Produces:
//  A pointer to a new Appender: output *Appender
//  Any filesystem errors that occur in opening $filename: err error
// Preconditions:
//  The file at filename exists and can be written to
// Postconditions:
//  An appender is created that will append to filename through writeWrapper
//  The caller of this function closes the created Appender
func MakeAppender(filename string) (*Appender, error) {
	var err error
	output := Appender{}
	output.fileHandle, err = os.OpenFile(filename, os.O_RDWR, 0755)
	if err != nil {
		return nil, errors.Wrap(err, "Opening file (RW)")
	}
	output.mux = &sync.Mutex{}
	output.metadata = appendedMetadata{}
	output.metadata.Data = make(map[string]appendedData)
	output.metadata.Version = METADATA_VERSION
	output.filename = filename
	return &output, nil
}

// Procedure:
//  *Appender.FileName
// Purpose:
//  To retrieve the filename associated with an Appender
// Parameters:
//  None
// Produces:
//  filename, a string
// Preconditions:
//  None.
// Postconditions:
//  Returns the filename private field
func (appender *Appender) FileName() string {
	return appender.filename
}

// Procedure:
//  *Appender.AppendStreamReader
// Purpose:
//  To append the entirety of a stream in an appended file block
// Parameters:
//  The parent *Appender: appender
//  The unique name of the stream: name string
//  The reader to pull data out of: source io.Reader
//  Whether to compress the stream: compressed bool
// Produces:
//  Side effects
//  Any errors in writing to the filesystem: err error
// Preconditions:
//  reader has a finite amount of data to read
//  $appender.Close() has not been called
// Postconditions:
//  All of the data that reader can read has been written to
//    appender's internal writer at the end of its file
//  appender's internal metadata has been updated to reflect the addition
//  Errors will be filesystem related
//
//  if $compressed: bash equivalent is executed:
//    $source | gzip >> $appender.file
//  else
//    $source >> $appender.file
//
//  $appender.file.ByteArray()[$appender.metadata[$name].StartFilePtr[:$appender].metadata[$name].[BlockSize].gunzip() == $source.ByteArray[]
func (appender *Appender) AppendStreamReader(name string, source io.Reader, compressed bool) error {
	appender.mux.Lock()
	defer appender.mux.Unlock()

	startPtr, err := appender.fileHandle.Seek(0, io.SeekEnd)
	if err != nil {
		return errors.Wrap(err, "Seeking to end of file (1)")
	}
	if compressed {
		gzWriter := gzip.NewWriter(appender.fileHandle)
		_, err = io.Copy(gzWriter, source)
		if err != nil {
			return errors.Wrap(err, "Writing data to file")
		}

		gzWriter.Close()
	} else {
		_, err = io.Copy(appender.fileHandle, source)
		if err != nil {
			return errors.Wrap(err, "Writing data to file")
		}
	}

	endPtr, err := appender.fileHandle.Seek(0, io.SeekEnd)
	if err != nil {
		return errors.Wrap(err, "Output corrupted. Seeking to end of file error")
	}

	fileMetadata := appendedData{}
	fileMetadata.StartFilePtr = startPtr
	fileMetadata.BlockSize = endPtr - startPtr
	fileMetadata.Zipped = compressed

	appender.metadata.Data[name] = fileMetadata
	return nil
}

// Procedure:
//  *Appender.AppendFile
// Purpose:
//  To gzip and pack a file onto the end of the Appender's file
// Parameters:
//  The calling Appender: appender Appender
//  The file to append: source string
//  Whether or not to compress the file: compressed bool
// Produces:
//  Side effects:
//    filesystem
//    internal state changes
//  Any errors in writing to the filesystem: err error
// Preconditions:
//  $source exists and is readable in the file system
//  $source has not been appended already nor has $appender.AppendStreamReader(name, _)
//    been called with name == $source
//  $appender.Close() has not been called
// Postconditions:
//  A reader stream from $source will be passed to $appender.AppendStreamReader,
//    with the name parameter as source
func (appender *Appender) AppendFile(source string, compressed bool) error {
	sourceHandle, err := os.Open(source)
	if err != nil {
		return errors.Wrapf(err, "Opening file: %s", source)
	}

	appender.mux.Lock()
	if _, exists := appender.metadata.Data[source]; exists {
		appender.mux.Unlock()
		return errors.Errorf("File %s has already been added to appender", source)
	}
	appender.mux.Unlock()

	err = appender.AppendStreamReader(source, sourceHandle, compressed)
	if err != nil {
		return errors.Wrap(err, "Appending file stream")
	}
	return sourceHandle.Close()
}

// Procedure:
//  *Appender.AppendByteArray
// Purpose:
//  To append a bytearray to a file
// Parameters:
//  The name of the data: name string
//  The data to append: data []byte
//  Whether to compress the data: compressed bool
// Produces:
//  Side effects:
//    filesystem
//    internal state changes
//  Any errors in writing to the filesystem: err error
// Preconditions:
//  No additional, although you should probably make sure that data is not empty
// Postconditions:
//  AppendSteamReader is called with a reader wrapped around data.
func (appender *Appender) AppendByteArray(name string, data []byte, compressed bool) error {
	return appender.AppendStreamReader(name, bytes.NewReader(data), compressed)
}

// Procedure:
//  *Appender.Close()
// Purpose:
//  To finish writing to the file being appended to
//    and free it for use elsewhere
// Parameters:
//   The Appender being acted upon: appender
// Produces:
//   Any filesystem errors: err error
// Preconditions:
//  $appender.Close() has not been called
// Postconditions:
//  The json-encoded metadata about the appended files has been
//    written out to the end of file being appended to
//  The start of said json block is encoded in the final 8 bytes of
//    the file being appended to as a little endian int64
//  The internal file handle for the file being appended to has been closed
func (appender *Appender) Close() error {
	appender.mux.Lock()
	defer appender.mux.Unlock()

	jsonPtr, err := appender.fileHandle.Seek(0, io.SeekEnd)
	if err != nil {
		return errors.Wrap(err, "Seeking to end of file")
	}
	jsonPtrBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(jsonPtrBytes, uint64(jsonPtr))

	jsonBytes, err := json.Marshal(appender.metadata)
	//Should not happen
	if err != nil {
		return errors.Wrap(err, "Marshaling metadata")
	}
	_, err = appender.fileHandle.Write(jsonBytes)
	if err != nil {
		return errors.Wrap(err, "Writing metadata")
	}
	_, err = appender.fileHandle.Write(jsonPtrBytes)
	if err != nil {
		return errors.Wrap(err, "Writing metadata location")
	}
	return appender.fileHandle.Close()
}
