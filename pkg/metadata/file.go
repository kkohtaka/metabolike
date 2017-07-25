// Copyright (C) 2017 Kazumasa Kohtaka <kkohtaka@gmail.com> All right reserved
// This file is available under the MIT license.

package metadata

// FileMonitor monitors update on a metadata on filesystem.
type FileMonitor struct{}

// NewFileMonitor returns a reference to a new instance of FileMonitor.
func NewFileMonitor() *FileMonitor {
	return &FileMonitor{}
}

// MonitorUpdate registers a handler which will be fired on metadata update.
func (m FileMonitor) MonitorUpdate(key string, handler UpdateHandler) {
	// TODO: Implement this function.
}
