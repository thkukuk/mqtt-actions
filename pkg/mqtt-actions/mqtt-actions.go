// Copyright 2023, 2024 Thorsten Kukuk
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

package mqttActions

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"net/http"

	log "github.com/thkukuk/mqtt-actions/pkg/logger"
	"github.com/thkukuk/mqtt-actions/pkg/health"
	"github.com/eclipse/paho.mqtt.golang"
)

const (
	defMQTTPort = "1883"
	defMQTTSPort = "8883"
	defMQTTProtocol = "mqtt"
	defMQTTSProtocol = "mqtts"
)

type ConfigType struct {
	HealthCheckListener *string         `yaml:"health_check,omitempty"`
	Verbose             *bool           `yaml:"verbose,omitempty"`
	MQTT                *MQTTConfig     `yaml:"mqtt"`
	Actions             []ActionType    `yaml:"actions"`
}

type MQTTConfig struct {
	Broker                 string `yaml:"broker"`
	Port                   string `yaml:"port"`
	Protocol               string `yaml:"protocol"`
	User                   string `yaml:"user"`
	Password               string `yaml:"password"`
	ClientID               string `yaml:"client_id"`
	QoS                    byte   `yaml:"qos"`
	Retain                 bool   `yaml:"retain"`
}

type ActionType struct {
	Name      string `yaml:"name"`
	Watch     string `yaml:"watch"`
	Path      string `yaml:"path"`
	Trigger   string `yaml:"trigger"`
	Action    []struct {
		  Topic   string `yaml:"topic"`
		  Message string `yaml:"message"`
	} `yaml:"action"`
	Enabled   *bool  `yaml:"enabled,omitempty"`
	Ignore2nd *bool  `yaml:"ignore_second,omitempty"`
	Counter   byte
}

var (
	Version = "unreleased"
	Quiet   = false
	Verbose = false
	Config ConfigType
	healthstate = health.NewHealthState()
)

func createMQTTClientID() string {
	host, err := os.Hostname()
        if err != nil {
		// XXX we should implement correct error handling
                panic(fmt.Sprintf("failed to get hostname: %v", err))
        }
        pid := os.Getpid()
        return fmt.Sprintf("%s-%d", host, pid)
}

func msgHandler(client mqtt.Client, msg mqtt.Message) {
	if Verbose {
		log.Debugf("Received message: topic: %s - %s\n", msg.Topic(), msg.Payload())
	}
	for i := range Config.Actions {
		if Config.Actions[i].Watch == msg.Topic() {
			if Verbose {
				log.Debugf("Found topic: %s\n", Config.Actions[i].Name)
			}
			message, _ := getMsgValue(Config.Actions[i], msg)
			if len(message) > 0 && message == Config.Actions[i].Trigger &&
				(Config.Actions[i].Enabled == nil || *Config.Actions[i].Enabled == true) {
				if Verbose {
					log.Debugf("Found message match: %s\n", message)
				} else {
					log.Infof("Execute '%s'\n", Config.Actions[i].Name)
				}
				if Config.Actions[i].Ignore2nd != nil && *Config.Actions[i].Ignore2nd == true {
				   	if Config.Actions[i].Counter > 0 {
					   	Config.Actions[i].Counter = 0;
						if Verbose {
							log.Debugf("Ignore 2nd call of %s\n", Config.Actions[i].Name)
						}
						return
					} else {
						Config.Actions[i].Counter++
					}
				}
				for j := range Config.Actions[i].Action {
					topic := Config.Actions[i].Action[j].Topic
					cmd := Config.Actions[i].Action[j].Message
					if Verbose {
						log.Debugf("Publish: %s = %s\n", topic, cmd)
					}
					client.Publish(topic, byte(Config.MQTT.QoS), Config.MQTT.Retain, cmd)
				}
			}
		}
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Info("Connection to MQTT Broker established")

	// Establish the subscription - doing this here means that it
	// will happen every time a connection is established

	// the connection handler is called in a goroutine so blocking
	// here would hot cause an issue. However as blocking in other
	// handlers does cause problems its best to just assume we should
	// not block
	for i := range Config.Actions {
		topic := Config.Actions[i].Watch
		token := client.Subscribe(topic, Config.MQTT.QoS, msgHandler)

		go func() {
			//_ = token.Wait()
			<-token.Done()

			if token.Error() != nil {
				log.Errorf("Error subscribing: %s", token.Error())
			} else {
				if !Quiet {
					log.Infof("Subscribed to topic: %s", topic)
				}
			}
		}()
	}
	healthstate.IsReady()
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	healthstate.NotReady()
	log.Errorf("Connection to MQTT Broker lost: %v", err)
}

func RunServer() {
	if !Quiet {
		log.Infof("MQTT Actions (mqtt-actions) %s is starting...\n", Version)
	}

	var mqtt_client mqtt.Client

	healthstate.DebugMode(Verbose)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	go func() {
		<-quit
		log.Info("Terminated via Signal. Shutting down...")
		if mqtt_client != nil && mqtt_client.IsConnectionOpen() {
			mqtt_client.Disconnect(250)
		}
		os.Exit(0)
	}()

	if Config.HealthCheckListener != nil &&
		len(*Config.HealthCheckListener) > 0 {
		// Start the state server
		stateServer := &http.Server{
			Addr:    *Config.HealthCheckListener,
			Handler: healthstate,
		}
		go stateServer.ListenAndServe()
	}

	opts := mqtt.NewClientOptions()

	if len(Config.MQTT.Protocol) == 0 {
		if Config.MQTT.Port == defMQTTSPort {
			Config.MQTT.Protocol = defMQTTSProtocol
		} else {
			Config.MQTT.Protocol = defMQTTProtocol
		}
	}

	if len(Config.MQTT.Port) == 0 {
		if Config.MQTT.Protocol == defMQTTSProtocol {
			Config.MQTT.Port = defMQTTSPort
		} else {
			Config.MQTT.Port = defMQTTPort
		}
	}

	brokerUrl := fmt.Sprintf("%s://%s:%s",
		Config.MQTT.Protocol, Config.MQTT.Broker,
		Config.MQTT.Port)
	if !Quiet {
		log.Infof("Broker: %s", brokerUrl)
	}

	opts.AddBroker(brokerUrl)
	opts.SetAutoReconnect(true)
	if len(Config.MQTT.ClientID) > 0 {
		opts.SetClientID(Config.MQTT.ClientID)
	} else {
		opts.SetClientID(createMQTTClientID())
	}
	if len(Config.MQTT.User) > 0 {
		opts.SetUsername(Config.MQTT.User)
	}
	if len(Config.MQTT.Password) > 0 {
		opts.SetPassword(Config.MQTT.Password)
	}
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

        errorChan := make(chan error, 1)

        for {
		mqtt_client = mqtt.NewClient(opts)
		if token := mqtt_client.Connect(); token.Wait() && token.Error() != nil {
			log.Warnf("Could not connect to mqtt broker, sleep 10 second: %v", token.Error())
			time.Sleep(10 * time.Second)
		} else {
                        break
                }
        }

	// loop forever and print error messages if they arrive
	// app is quit with above signal handler "quit".
	for {
                select {
                case err := <-errorChan:
                        log.Errorf("Error while processing message: %v", err)
                }
        }
}
