#!/bin/sh
set -eu

interval="${DEV_WATCH_INTERVAL:-1}"
pid=""
binary="/tmp/invitely-api-dev"

current_hash() {
	find . -type f \
		\( -name "*.go" -o -name "*.yaml" -o -name "go.mod" -o -name "go.sum" \) \
		-not -path "./.git/*" \
		-not -path "./vendor/*" \
		-print | sort | xargs sha256sum 2>/dev/null | sha256sum
}

stop_app() {
	if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
		kill "$pid" 2>/dev/null || true
		wait "$pid" 2>/dev/null || true
	fi
	pid=""
}

start_app() {
	echo "[dev] building API"
	if ! go build -o "$binary" ./cmd/api; then
		echo "[dev] build failed, waiting for changes"
		return
	fi

	echo "[dev] starting API"
	"$binary" &
	pid="$!"
}

cleanup() {
	stop_app
	exit 0
}

trap cleanup INT TERM

last_hash="$(current_hash)"
start_app

while true; do
	sleep "$interval"
	next_hash="$(current_hash)"

	if [ "$next_hash" != "$last_hash" ]; then
		echo "[dev] change detected, restarting API"
		last_hash="$next_hash"
		stop_app
		start_app
		continue
	fi

	if [ -n "$pid" ] && ! kill -0 "$pid" 2>/dev/null; then
		wait "$pid" 2>/dev/null || true
		echo "[dev] API stopped, waiting for next successful run"
		start_app
	fi
done
