//go:build windows
// +build windows

package main

// #cgo LDFLAGS: -Wl,--allow-multiple-definition
import "C"
import _ "github.com/go-gl/gl/v3.2-core/gl"
