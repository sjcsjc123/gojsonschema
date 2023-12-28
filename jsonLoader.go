// Copyright 2015 xeipuuv ( https://github.com/xeipuuv )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author           xeipuuv
// author-github    https://github.com/xeipuuv
// author-mail      xeipuuv@gmail.com
//
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
//
// description		Different strategies to load JSON files.
// 					Includes References (file and HTTP), JSON strings and Go types.
//
// created          01-02-2015

package gojsonschema

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/xeipuuv/gojsonreference"
)

// JSONLoader defines the JSON loader interface
type JSONLoader interface {
	JsonSource() interface{}
	LoadJSON() (interface{}, error)
	JsonReference() (gojsonreference.JsonReference, error)
	LoaderFactory() JSONLoaderFactory
}

// JSONLoaderFactory defines the JSON loader factory interface
type JSONLoaderFactory interface {
	// New creates a new JSON loader for the given source
	New(source string) JSONLoader
}

// DefaultJSONLoaderFactory is the default JSON loader factory
type DefaultJSONLoaderFactory struct {
}

// FileSystemJSONLoaderFactory is a JSON loader factory that uses http.FileSystem
type FileSystemJSONLoaderFactory struct {
}

// New creates a new JSON loader for the given source
func (d DefaultJSONLoaderFactory) New(source string) JSONLoader {
	return &jsonReferenceLoader{
		source: source,
	}
}

// New creates a new JSON loader for the given source
func (f FileSystemJSONLoaderFactory) New(source string) JSONLoader {
	return &jsonReferenceLoader{
		source: source,
	}
}

// JSON Reference loader
// references are used to load JSONs from files and HTTP

type jsonReferenceLoader struct {
	source string
}

func (l *jsonReferenceLoader) JsonSource() interface{} {
	return l.source
}

func (l *jsonReferenceLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference(l.JsonSource().(string))
}

func (l *jsonReferenceLoader) LoaderFactory() JSONLoaderFactory {
	return &FileSystemJSONLoaderFactory{}
}

// NewReferenceLoader returns a JSON reference loader using the given source and the local OS file system.
func NewReferenceLoader(source string) JSONLoader {
	return &jsonReferenceLoader{
		source: source,
	}
}

func (l *jsonReferenceLoader) LoadJSON() (interface{}, error) {
	return nil, nil
}

// JSON string loader

type jsonStringLoader struct {
	source string
}

func (l *jsonStringLoader) JsonSource() interface{} {
	return l.source
}

func (l *jsonStringLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}

func (l *jsonStringLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

// NewStringLoader creates a new JSONLoader, taking a string as source
func NewStringLoader(source string) JSONLoader {
	return &jsonStringLoader{source: source}
}

func (l *jsonStringLoader) LoadJSON() (interface{}, error) {

	return decodeJSONUsingNumber(strings.NewReader(l.JsonSource().(string)))

}

// JSON bytes loader

type jsonBytesLoader struct {
	source []byte
}

func (l *jsonBytesLoader) JsonSource() interface{} {
	return l.source
}

func (l *jsonBytesLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}

func (l *jsonBytesLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

// NewBytesLoader creates a new JSONLoader, taking a `[]byte` as source
func NewBytesLoader(source []byte) JSONLoader {
	return &jsonBytesLoader{source: source}
}

func (l *jsonBytesLoader) LoadJSON() (interface{}, error) {
	return decodeJSONUsingNumber(bytes.NewReader(l.JsonSource().([]byte)))
}

// JSON Go (types) loader
// used to load JSONs from the code as maps, interface{}, structs ...

type jsonGoLoader struct {
	source interface{}
}

func (l *jsonGoLoader) JsonSource() interface{} {
	return l.source
}

func (l *jsonGoLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}

func (l *jsonGoLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

// NewGoLoader creates a new JSONLoader from a given Go struct
func NewGoLoader(source interface{}) JSONLoader {
	return &jsonGoLoader{source: source}
}

func (l *jsonGoLoader) LoadJSON() (interface{}, error) {

	// convert it to a compliant JSON first to avoid types "mismatches"

	jsonBytes, err := json.Marshal(l.JsonSource())
	if err != nil {
		return nil, err
	}

	return decodeJSONUsingNumber(bytes.NewReader(jsonBytes))

}

type jsonIOLoader struct {
	buf *bytes.Buffer
}

// NewReaderLoader creates a new JSON loader using the provided io.Reader
func NewReaderLoader(source io.Reader) (JSONLoader, io.Reader) {
	buf := &bytes.Buffer{}
	return &jsonIOLoader{buf: buf}, io.TeeReader(source, buf)
}

// NewWriterLoader creates a new JSON loader using the provided io.Writer
func NewWriterLoader(source io.Writer) (JSONLoader, io.Writer) {
	buf := &bytes.Buffer{}
	return &jsonIOLoader{buf: buf}, io.MultiWriter(source, buf)
}

func (l *jsonIOLoader) JsonSource() interface{} {
	return l.buf.String()
}

func (l *jsonIOLoader) LoadJSON() (interface{}, error) {
	return decodeJSONUsingNumber(l.buf)
}

func (l *jsonIOLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}

func (l *jsonIOLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

// JSON raw loader
// In case the JSON is already marshalled to interface{} use this loader
// This is used for testing as otherwise there is no guarantee the JSON is marshalled
// "properly" by using https://golang.org/pkg/encoding/json/#Decoder.UseNumber
type jsonRawLoader struct {
	source interface{}
}

// NewRawLoader creates a new JSON raw loader for the given source
func NewRawLoader(source interface{}) JSONLoader {
	return &jsonRawLoader{source: source}
}
func (l *jsonRawLoader) JsonSource() interface{} {
	return l.source
}
func (l *jsonRawLoader) LoadJSON() (interface{}, error) {
	return l.source, nil
}
func (l *jsonRawLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}
func (l *jsonRawLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

func decodeJSONUsingNumber(r io.Reader) (interface{}, error) {

	var document interface{}

	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	err := decoder.Decode(&document)
	if err != nil {
		return nil, err
	}

	return document, nil

}
