// Copyright 2022 Chainguard, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Loads an image configuration given a configuration file path.
func (ic *ImageConfiguration) Load(imageConfigPath string) error {
	data, err := os.ReadFile(imageConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read image configuration file: %w", err)
	}

	if err := yaml.Unmarshal(data, ic); err != nil {
		return fmt.Errorf("failed to parse image configuration: %w", err)
	}

	return nil
}

// Do preflight checks and mutations on an image configuration.
func (ic *ImageConfiguration) Validate() error {
	if ic.Entrypoint.Type == "service-bundle" {
		if err := ic.ValidateServiceBundle(); err != nil {
			return err
		}
	}

	for _, u := range ic.Accounts.Users {
		if u.UserName == "" {
			return fmt.Errorf("configured user %v has no configured user name", u)
		}

		if u.UID == 0 {
			return fmt.Errorf("configured user %v has UID 0", u)
		}
	}

	for _, g := range ic.Accounts.Groups {
		if g.GroupName == "" {
			return fmt.Errorf("configured group %v has no configured group name", g)
		}

		if g.GID == 0 {
			return fmt.Errorf("configured group %v has GID 0", g)
		}
	}

	return nil
}

// Do preflight checks and mutations on an image configured to manage
// a service bundle.
func (ic *ImageConfiguration) ValidateServiceBundle() error {
	ic.Entrypoint.Command = "/bin/s6-svscan /sv"

	// It's harmless to have a duplicate entry in /etc/apk/world,
	// apk will fix it up when the fixate op happens.
	ic.Contents.Packages = append(ic.Contents.Packages, "s6")

	return nil
}

func (ic *ImageConfiguration) Summarize() {
	log.Printf("image configuration:")
	log.Printf("  contents:")
	log.Printf("    repositories: %v", ic.Contents.Repositories)
	log.Printf("    keyring:      %v", ic.Contents.Keyring)
	log.Printf("    packages:     %v", ic.Contents.Packages)
	if ic.Entrypoint.Type != "" || ic.Entrypoint.Command != "" || len(ic.Entrypoint.Services) != 0 {
		log.Printf("  entrypoint:")
		log.Printf("    type:    %s", ic.Entrypoint.Type)
		log.Printf("    cmd:     %s", ic.Entrypoint.Command)
		log.Printf("    service: %v", ic.Entrypoint.Services)
	}
	if ic.Accounts.RunAs != "" || len(ic.Accounts.Users) != 0 || len(ic.Accounts.Groups) != 0 {
		log.Printf("  accounts:")
		log.Printf("    runas:  %s", ic.Accounts.RunAs)
		log.Printf("    users:  %v", ic.Accounts.Users)
		log.Printf("    groups: %v", ic.Accounts.Groups)
	}
}
