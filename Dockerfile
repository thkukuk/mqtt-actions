FROM registry.opensuse.org/opensuse/tumbleweed:latest AS build-stage
RUN zypper install --no-recommends --auto-agree-with-product-licenses -y git go make
#RUN git clone https://github.com/thkukuk/mqtt-actions
COPY . mqtt-actions
RUN cd mqtt-actions && make update && make tidy && make

FROM registry.opensuse.org/opensuse/busybox:latest
LABEL maintainer="Thorsten Kukuk <kukuk@thkukuk.de>"

ARG BUILDTIME=
ARG VERSION=unreleased
LABEL org.opencontainers.image.title="MQTT-Actions"
LABEL org.opencontainers.image.description="Listens to MQTT topics and creates new ones if a matching one was seen."
LABEL org.opencontainers.image.created=$BUILDTIME
LABEL org.opencontainers.image.version=$VERSION

COPY --from=build-stage /mqtt-actions/bin/mqtt-actions /usr/local/bin
COPY entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/local/bin/mqtt-actions"]
