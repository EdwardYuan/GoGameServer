#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")" || exit 1
mkdir -p log

BIN="./GoGameServer"
CONFIG_FILE="config/config.toml"
STARTED_PIDS=()

if [[ ! -x "$BIN" ]]; then
  echo "missing executable: $BIN"
  echo "run ../build.sh first"
  exit 1
fi

if [[ ! -f "$CONFIG_FILE" ]]; then
  echo "missing config file: $CONFIG_FILE"
  exit 1
fi

config_value() {
  local section="$1"
  local key="$2"
  awk -F '=' -v section="[$section]" -v key="$key" '
    $0 == section { in_section = 1; next }
    /^\[/ { in_section = 0 }
    in_section && $1 == key {
      gsub(/[[:space:]\"]/, "", $2)
      print $2
      exit
    }
  ' "$CONFIG_FILE"
}

host_from_addr() {
  local addr="$1"
  echo "${addr%%:*}"
}

port_from_addr() {
  local addr="$1"
  echo "${addr##*:}"
}

is_port_open() {
  local host="$1"
  local port="$2"
  bash -c ":</dev/tcp/$host/$port" >/dev/null 2>&1
}

wait_for_port() {
  local name="$1"
  local host="$2"
  local port="$3"
  local attempts="${4:-30}"

  for ((i = 1; i <= attempts; i++)); do
    if is_port_open "$host" "$port"; then
      echo "$name is ready on $host:$port"
      return 0
    fi
    sleep 1
  done

  echo "timeout waiting for $name on $host:$port"
  return 1
}

start_service() {
  local service="$1"
  local idx="$2"
  local host="$3"
  local port="$4"

  if is_port_open "$host" "$port"; then
    echo "$service already running on $host:$port"
    return 0
  fi

  echo "starting $service..."
  "$BIN" run "$service" "$idx" >"log/${service}_${idx}.bootstrap.log" 2>&1 &
  local pid=$!
  STARTED_PIDS+=("$pid")

  if ! wait_for_port "$service" "$host" "$port"; then
    echo "last log lines from log/${service}_${idx}.bootstrap.log:"
    tail -n 40 "log/${service}_${idx}.bootstrap.log" || true
    exit 1
  fi
}

cleanup() {
  if [[ ${#STARTED_PIDS[@]} -eq 0 ]]; then
    return
  fi
  echo "stopping background services..."
  for pid in "${STARTED_PIDS[@]}"; do
    kill "$pid" >/dev/null 2>&1 || true
  done
}

trap cleanup EXIT INT TERM

gate_host="$(config_value gamegate addr)"
gate_port="$(config_value gamegate port)"
db_host="$(config_value dbserver addr)"
db_port="$(config_value dbserver port)"
proxy_addr="$(config_value proxy addr)"
proxy_host="$(host_from_addr "$proxy_addr")"
proxy_port="$(port_from_addr "$proxy_addr")"
etcd_addr="$(config_value etcd endpoints)"
etcd_host="$(host_from_addr "$etcd_addr")"
etcd_port="$(port_from_addr "$etcd_addr")"

if ! is_port_open "$etcd_host" "$etcd_port"; then
  echo "warning: etcd is not reachable on $etcd_host:$etcd_port; proxy may fail to start"
fi

start_service gate 0 "$gate_host" "$gate_port"
start_service dbserver 0 "$db_host" "$db_port"
start_service proxy 0 "$proxy_host" "$proxy_port"

echo "starting game..."
"$BIN" run game 0
