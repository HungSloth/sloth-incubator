#!/usr/bin/env bash
set -euo pipefail

DISPLAY_NUM="${DISPLAY_NUM:-:99}"
SCREEN_GEOMETRY="${SCREEN_GEOMETRY:-1280x800x24}"
VNC_PORT="${VNC_PORT:-5900}"
NOVNC_PORT="${NOVNC_PORT:-6080}"
APP_COMMAND="${PREVIEW_APP_COMMAND:-}"

Xvfb "${DISPLAY_NUM}" -screen 0 "${SCREEN_GEOMETRY}" &
XVFB_PID=$!

export DISPLAY="${DISPLAY_NUM}"
openbox >/tmp/openbox.log 2>&1 &

x11vnc -display "${DISPLAY}" -rfbport "${VNC_PORT}" -forever -shared -nopw -listen 0.0.0.0 >/tmp/x11vnc.log 2>&1 &
xterm >/tmp/xterm.log 2>&1 &

if [[ -n "${APP_COMMAND}" ]]; then
  bash -lc "${APP_COMMAND}" >/tmp/app.log 2>&1 &
fi

NOVNC_DIR="/usr/share/novnc"
if [[ ! -d "${NOVNC_DIR}" ]]; then
  NOVNC_DIR="/usr/share/novnc/utils/novnc_proxy"
fi

if [[ -x /usr/share/novnc/utils/novnc_proxy ]]; then
  /usr/share/novnc/utils/novnc_proxy --listen "${NOVNC_PORT}" --vnc "127.0.0.1:${VNC_PORT}" >/tmp/novnc.log 2>&1 &
else
  websockify --web "${NOVNC_DIR}" "${NOVNC_PORT}" "127.0.0.1:${VNC_PORT}" >/tmp/novnc.log 2>&1 &
fi

wait "${XVFB_PID}"
