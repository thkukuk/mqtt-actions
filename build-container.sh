#!/bin/bash

set -e

UPLOAD=0

if [ "$1" = "--upload" ]; then
	UPLOAD=1
fi


VERSION=$(cat VERSION)
VER=( ${VERSION//./ } )

sudo podman pull registry.opensuse.org/opensuse/busybox:latest
sudo podman build --rm --no-cache --build-arg VERSION="${VERSION}" --build-arg BUILDTIME=$(date +%Y-%m-%dT%TZ) -t mqtt-actions .
sudo podman tag localhost/mqtt-actions thkukuk/mqtt-actions:"${VERSION}"
sudo podman tag localhost/mqtt-actions thkukuk/mqtt-actions:latest
sudo podman tag localhost/mqtt-actions thkukuk/mqtt-actions:"${VER[0]}"
sudo podman tag localhost/mqtt-actions thkukuk/mqtt-actions:"${VER[0]}.${VER[1]}"
if [ $UPLOAD -eq 1 ]; then
	sudo podman login docker.io
	sudo podman push thkukuk/mqtt-actions:"${VERSION}"
	sudo podman push thkukuk/mqtt-actions:latest
	sudo podman push thkukuk/mqtt-actions:"${VER[0]}"
	sudo podman push thkukuk/mqtt-actions:"${VER[0]}.${VER[1]}"
fi
