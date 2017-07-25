// Copyright (C) 2017 Kazumasa Kohtaka <kkohtaka@gmail.com> All right reserved
// This file is available under the MIT license.

package types

// Config represents a single configuration to template a file with a variables set.
type Config struct {
	Name          string
	Template      string
	Backend       string
	Source        string
	Destination   string
	CheckCommand  string
	ReloadCommand string
}

// ConfigList represents a list of configurations.
type ConfigList []Config
