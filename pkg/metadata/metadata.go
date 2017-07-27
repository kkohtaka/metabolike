// Copyright (C) 2017 Kazumasa Kohtaka <kkohtaka@gmail.com> All right reserved
// This file is available under the MIT license.

package metadata

// UpdateHandler represents a function which will be called when a metada is updated.
type UpdateHandler func(key string, data []byte) error

// Monitor monitors update on a metadata specified by a key.
type Monitor interface {
	MonitorUpdate(key string, handler UpdateHandler) error
}
