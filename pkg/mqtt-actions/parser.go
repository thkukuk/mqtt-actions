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

package mqttActions

import (
	"fmt"

	log "github.com/thkukuk/mqtt-actions/pkg/logger"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/thedevsaddam/gojsonq/v2"
)

func getMsgValue(action ActionType, msg mqtt.Message) (string, error) {

	payload := string(msg.Payload())

	if len(action.Path) > 0 {
		// gojsonq.Find is of form "a.b.c.d"
		entry := gojsonq.New().FromString(payload).Find(action.Path)
		if entry == nil {
			if Verbose {
				log.Warnf("WARNING: %q not found in '%s'!", action.Path, payload)
			}
			return "", nil
		}
		payload = fmt.Sprintf("%v", entry)
	}

	return payload, nil
}
