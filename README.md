# mqtt-actions
**MQTT Actions - listens to MQTT topics and create new ones**


The mqtt-actions daemon listens on MQTT topics and creates new ones if a matching one was seen.

```plaintext
 IoT Sensors -> publish -> MQTT Broker <- subcribed <- MQTT Actions -> publish -> MQTT Broker <- subscribed <- IoT devices
```

## Subscribed and Published Topics

The configuration file for mqtt-actions should contain the data to connect with a MQTT Broker and a list of "actions". For every "action" there will be a topic to which mqtt-actions will subscribe and wait for a specific message. If this message is seen, mqtt-actions will go through a list of topics and publishes a corresponding message to each of them.
As example, if somebody presses a button which will generate e.g. the message "on" for the topic "livingroom/button", mqtt-actions can send the message "toggle" to several Shelly Plug S devices, which will toggle the power switch.

It is possible to have devices which publish every message in an own topic, as JSON struct per topic or a mix of both.
For Shelly Plug S and TRÅDFRI Shortcut button from IKEA this MQTT messages could look like:

```plaintext
shellies/shelly-plug-s1/relay/0/power 20.58
shellies/shelly-plug-s1/relay/0/energy 94736
shellies/shelly-plug-s1/relay/0 on
shellies/shelly-plug-s1/temperature 22.28
shellies/shelly-plug-s1/temperature_f 72.10
shellies/shelly-plug-s1/overtemperature 0
zigbee2mqtt/IKEA-Shortcut-Button1 {"action":"on","battery":50,"linkquality":255,"update":{"installed_version":604241926,"latest_version":604241926,"state":"idle"},"update_available":false}
```

Wildcards for the topics are currently not supported.

## Container

### Public Container Image

To run the public available image:

```bash
podman run --rm -v <path>/config.yaml:/config.yaml registry.opensuse.org/home/kukuk/containerfile/mqtt-actions
```

You can replace `podman` with `docker` without any further changes.

### Build locally

To build the container image with the `mqtt-actions` binary included run:

```bash
sudo podman build --rm --no-cache --build-arg VERSION=$(cat VERSION) --build-arg BUILDTIME=$(date +%Y-%m-%dT%TZ) -t mqtt-actions .
```

You can of cource replace `podman` with `docker`, no other arguments needs to be adjusted.

## Configuration

mqtt-actions will be configured via command line and configuration file.

### Commandline

Available options are:
```plaintext
Usage:
  mqtt-actions [flags]

Flags:
  -c, --config string   configuration file (default "config.yaml")
  -h, --help            help for mqtt-actions
  -q, --quiet           don't print any informative messages
  -v, --verbose         become really verbose in printing messages
      --version         version for mqtt-actions
```

### Configuration File

By default `mqtt-actions` looks for the file `config.yaml` in the local directory. This can be overriden with the `--config` option.

Here is an example configuration file, which can be used to controll two Shelly Plus Plug S with a TRÅDFRI Shortcut button or another process which sends sunrise/sunset messages.

```yaml
# Optional, if set, /healthz liveness probe and /readyz readiness
# probes will be provided
#health_check: ":8080"
verbose: true
mqtt:
  # Required: The MQTT broker to connect to
  broker: mqtt.example.com
  # Optional: Port of the MQTT broker
  port: 8883
  protocol: mqtts
  # Optional: Username and Password for authenticating with the MQTT Server
  #user: <username>
  #password: <password>
  # Optional: Used to specify ClientID. The default is <hostname>-<pid>
  # client_id: somedevice
  # The MQTT QoS level
  qos: 0
  # MQTT retain messages
  retain: false
actions:
  - name: IKEA Button 1
    # topic to subscribe on
    watch: zigbee2mqtt/IKEA-Shortcut-Button1
    # for JSON values, e.g.: {"action":"on",...}
    path: action
    # execute if message is "on"
    trigger: on
    action:
    # Publish message "toogle" for 1. Shelly Plus Plug S
    - topic: shellies/shelly-plus-plug-s1/command/switch:0
      message: toggle
    # Publish message "toogle" for 2. Shelly Plus Plug S
    - topic: shellies/shelly-plus-plug-s2/command/switch:0
      message: toggle
  # Switch on two Shelly Plus Plug S on sunset
  - name: Sunset timer
    # topic to subscribe on
    watch: timer/sun
    # No json, execute if message is "sunset"
    trigger: sunset
    action:
    # Publish message "on" for 1. Shelly Plus Plug S
    - topic: shellies/shelly-plus-plug-s1/command/switch:0
      message: on
    # Publish message "on" for 2. Shelly Plus Plug S
    - topic: shellies/shelly-plus-plug-s2/command/switch:0
      message: on
```

### Explanation

The action sections define, for which MQTT topic the program should look, how to parse the data and how to act.

* **name** is descriptive name of the action, it's only used for debug output.
* **watch** is the topic to subscribe to.
* **path** is the path inside the JSON struct, if the message is in JSON format. The names are separated via "dots". More information about this can be found in the `find` examples of the [gojsonq](https://github.com/thedevsaddam/gojsonq) documentation.
* **trigger** is the content of the message which triggers this action.
* **action** is a list of **topic** and **message** pairs which will be published.

## Environment Variables

Having the login details in the config file runs the risk of publishing them to a version control system. To avoid this, you can supply these parameters via environment variables. mqtt-actions will look for MQTT_USER and MQTT_PASSWORD in the local environment at startup.

### Example usage with container

```bash
  sudo podman run -e MQTT_USER="user" -e MQTT_PASSWORD="password" ...
```

### Example usage with kubernetes or podman kube play

```yaml
...
spec:
  containers:
  - name: mqtt-actions
    image: <image>
    env:
    - name: MQTT_USER
      value: <username>
    - name: MQTT_PASSWORD
      value: <password>
...
```
## Liveness and readiness probes

This liveness and readiness health checks are needed if the service runs in Kubernetes. The livness probe tells kubernetes that the application is alive, if the service does not answer, the service will be restarted. The readiness probe tells kubernetes, when the container is ready to serve traffic.

The endpoints are:

* *IP:Port*/healthz for the liveness probe
* *IP:Port*/readyz for the readiness probe


The **IP:Port** will be defined with the `health_check` option in the configuration file. If this config variable is not set, the health check stay disabled.
