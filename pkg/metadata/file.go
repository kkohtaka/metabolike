// Copyright (C) 2017 Kazumasa Kohtaka <kkohtaka@gmail.com> All right reserved
// This file is available under the MIT license.

package metadata

import (
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

// FileMonitor monitors update on a metadata on filesystem.
type FileMonitor struct{}

// NewFileMonitor returns a reference to a new instance of FileMonitor.
func NewFileMonitor() *FileMonitor {
	return &FileMonitor{}
}

// MonitorUpdate registers a handler which will be fired on metadata update.
func (m FileMonitor) MonitorUpdate(key string, handler UpdateHandler) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.
			WithError(err).
			Error("Failed to instantiate a filesystem watcher")
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					data, err := ioutil.ReadFile(key)
					if err != nil {
						log.
							WithError(err).
							WithField("path", key).
							Error("Failed to read data from a path")
					}
					err = handler(key, data)
					if err != nil {
						log.
							WithError(err).
							Error("Failed to handle a metadata update")
					}
				}
			case err = <-watcher.Errors:
				log.Info("Checking a metadata updated")
			}
		}
	}()

	err = watcher.Add(key)
	if err != nil {
		log.
			WithError(err).
			WithField("path", key).
			Error("Failed to watch a path to filesystem")
		return err
	}
	<-done

	return nil
}
