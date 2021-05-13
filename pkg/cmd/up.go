/*
Copyright © 2021 Microshift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "openshift cluster up",
	Long:  `openshift cluster up`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return startCluster(args)
	},
}

func startCluster(args []string) error {
	logrus.Info("starting controllers")
	if err := startControllerOnly(); err != nil {
		return err
	}
	logrus.Info("starting node")
	if err := startNodeOnly(); err != nil {
		return err
	}
	select {}
}
