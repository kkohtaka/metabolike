// Copyright (C) 2017 Kazumasa Kohtaka <kkohtaka@gmail.com> All right reserved
// This file is available under the MIT license.

package metadata

import (
	gcemetadata "github.com/GoogleCloudPlatform/google-cloud-go/compute/metadata"
	log "github.com/sirupsen/logrus"
)

// GCEMonitor monitors update on a metadata on filesystem.
type GCEMonitor struct{}

// NewGCEMonitor returns a reference to a new instance of GCEMonitor.
func NewGCEMonitor() *GCEMonitor {
	return &GCEMonitor{}
}

// MonitorUpdate registers a handler which will be fired on metadata update.
func (m GCEMonitor) MonitorUpdate(key string, handler UpdateHandler) error {
	return gcemetadata.Subscribe(key, func(v string, ok bool) error {
		if ok {
			err := handler(key, []byte(v))
			if err != nil {
				log.
					WithError(err).
					Error("Failed to handle a metadata update")
			}
		} else {
			log.Info("Checking a metadata updated")
		}
		return nil
	})
}
