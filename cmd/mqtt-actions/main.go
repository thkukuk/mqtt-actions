// Copyright 2023 Thorsten Kukuk
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

package main

import (
	"fmt"
        "io/ioutil"
	"os"

	log "github.com/thkukuk/mqtt-actions/pkg/logger"
	"gopkg.in/yaml.v3"
	"github.com/spf13/cobra"
	"github.com/thkukuk/mqtt-actions/pkg/mqtt-actions"
)

var (
	configFile = "config.yaml"
)

func read_yaml_config(conffile string) (mqttActions.ConfigType, error) {

        var config mqttActions.ConfigType

        file, err := ioutil.ReadFile(conffile)
        if err != nil {
                return config, fmt.Errorf("Cannot read %q: %v", conffile, err)
        }
        err = yaml.Unmarshal(file, &config)
        if err != nil {
                return config, fmt.Errorf("Unmarshal error: %v", err)
        }

        return config, nil
}


func main() {
// mqttActionsCmd represents the mqtt-actions command
	mqttActionsCmd := &cobra.Command{
		Use:   "mqtt-actions",
		Short: "Starts a MQTT Actions",
		Long: `Starts a MQTT Actions.
This daemon listens to MQTT topics and publishes new messages depending on them.
`,
		Run: runMqttActionsCmd,
		Args:  cobra.ExactArgs(0),
	}

        mqttActionsCmd.Version = mqttActions.Version

	mqttActionsCmd.Flags().StringVarP(&configFile, "config", "c", configFile, "configuration file")

	mqttActionsCmd.Flags().BoolVarP(&mqttActions.Quiet, "quiet", "q", mqttActions.Quiet, "don't print any informative messages")
	mqttActionsCmd.Flags().BoolVarP(&mqttActions.Verbose, "verbose", "v", mqttActions.Verbose, "become really verbose in printing messages")

	if err := mqttActionsCmd.Execute(); err != nil {
                os.Exit(1)
        }
}

func runMqttActionsCmd(cmd *cobra.Command, args []string) {
	var err error

	if !mqttActions.Quiet {
		log.Infof("Read yaml config %q\n", configFile)
	}
	mqttActions.Config, err = read_yaml_config(configFile)
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	if mqttActions.Config.Verbose != nil {
		mqttActions.Verbose = *mqttActions.Config.Verbose
	}

        mqtt_user := os.Getenv("MQTT_USER")
        if mqtt_user != "" {
                mqttActions.Config.MQTT.User = mqtt_user
        }

	mqtt_password := os.Getenv("MQTT_PASSWORD")
        if mqtt_password != "" {
                mqttActions.Config.MQTT.Password = mqtt_password
        }

	mqttActions.RunServer()
}
