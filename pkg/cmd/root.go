// Copyright (C) 2017 Kazumasa Kohtaka <kkohtaka@gmail.com> All right reserved
// This file is available under the MIT license.

package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sync"
	"text/template"

	"github.com/kkohtaka/metabolike/pkg/metadata"
	"github.com/kkohtaka/metabolike/pkg/types"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var configFilePath string

// RootCmd represents a root command of Metabolike
var RootCmd = &cobra.Command{
	Use:   "metabolike",
	Short: "Metabolike is a tool to template configuration files with metadata API",
}

func init() {
	RootCmd.PersistentFlags().StringVar(&configFilePath, "config", "config.yml", "a path to a configuration file")

	RootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		data, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			log.
				WithError(err).
				WithField("path", configFilePath).
				Error("Failed to read configuration from a path")
			return err
		}
		list := types.ConfigList{}
		err = yaml.Unmarshal(data, &list)
		if err != nil {
			log.
				WithError(err).
				WithField("data", string(data)).
				Error("Failed to unmarshal data as a YAML")
			return err
		}
		log.
			WithField("path", configFilePath).
			Info("Configuration was loaded from a path")

		wg := sync.WaitGroup{}
		for _, config := range list {
			log.
				WithField("config", config.Name).
				Info("Processing configuration")

			var monitor metadata.Monitor
			switch config.Backend {
			case "gce":
				monitor = metadata.NewGCEMonitor()
			case "file":
				monitor = metadata.NewFileMonitor()
			default:
				log.
					WithField("backend", config.Backend).
					Warn("Invalid backend was specified")
				continue
			}

			handler, err := generateMetadataUpdateHandler(config)
			if err != nil {
				log.
					WithError(err).
					Warn("Failed to generate a metadata update handler")
				continue
			}

			wg.Add(1)
			go func(config types.Config) {
				defer wg.Done()
				monitor.MonitorUpdate(config.Source, handler)
			}(config)
		}
		wg.Wait()

		return nil
	}
}

func generateMetadataUpdateHandler(config types.Config) (metadata.UpdateHandler, error) {
	tmplData, err := ioutil.ReadFile(config.Template)
	if err != nil {
		log.
			WithError(err).
			WithField("name", config.Name).
			WithField("path", config.Template).
			Error("Failed to read a template from a path")
		return nil, err
	}

	tmpl, err := template.New(config.Name).Parse(string(tmplData))
	if err != nil {
		log.
			WithError(err).
			WithField("name", config.Name).
			WithField("path", config.Template).
			WithField("text", string(tmplData)).
			Error("Failed to parse a text as a template")
		return nil, err
	}

	return func(key string, data []byte) error {
		variables := struct{}{}
		err = yaml.Unmarshal(data, &variables)
		if err != nil {
			log.
				WithError(err).
				WithField("data", string(data)).
				Error("Failed to unmarshal data as a YAML")
			return err
		}

		tmpFile, err := ioutil.TempFile("", "metabolike")
		if err != nil {
			log.
				WithError(err).
				Error("Failed to crate a temporary file")
			return err
		}
		defer os.Remove(tmpFile.Name())

		w := bufio.NewWriter(tmpFile)
		err = tmpl.Execute(w, variables)
		if err != nil {
			log.
				WithError(err).
				WithField("variables", variables).
				Error("Failed to execute a template")
			return err
		}
		w.Flush()

		if len(config.CheckCommand) > 0 {
			command := fmt.Sprintf("%s %s", config.CheckCommand, tmpFile.Name())
			var result []byte
			result, err = exec.Command("sh", "-c", command).Output()
			if err != nil {
				log.
					WithError(err).
					WithField("command", command).
					Error("Failed to execute a check command")
				return err
			}
			log.
				WithField("command", command).
				WithField("result", string(result)).
				Info("A check command succeeded")
		}

		dir := path.Dir(config.Destination)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.
				WithError(err).
				WithField("directory", dir).
				Error("Failed to create a directory")
			return err
		}

		err = os.Remove(config.Destination)
		if err != nil {
			log.
				WithError(err).
				WithField("path", config.Destination).
				Error("Failed to remove a file at a path")
			return err
		}

		err = os.Link(tmpFile.Name(), config.Destination)
		if err != nil {
			log.
				WithError(err).
				WithField("destination", config.Destination).
				Error("Failed to export to a destination")
			return err
		}

		if len(config.ReloadCommand) > 0 {
			command := config.ReloadCommand
			var result []byte
			result, err = exec.Command("sh", "-c", command).Output()
			if err != nil {
				log.
					WithError(err).
					WithField("command", command).
					Error("Failed to execute a reload command")
				return err
			}
			log.
				WithField("command", command).
				WithField("result", string(result)).
				Info("A reload command succeeded")
		}

		return nil
	}, nil
}
