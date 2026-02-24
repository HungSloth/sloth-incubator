FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    xvfb \
    x11vnc \
    novnc \
    websockify \
    openbox \
    xterm \
    ca-certificates \
    bash \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
COPY .incubator/preview/entrypoint.sh /usr/local/bin/preview-entrypoint.sh
RUN chmod +x /usr/local/bin/preview-entrypoint.sh

EXPOSE 5900 6080

ENTRYPOINT ["/usr/local/bin/preview-entrypoint.sh"]
