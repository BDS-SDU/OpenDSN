#!/bin/bash

collect_pids() {
	local pattern
	local pid
	local -A seen=()

	for pattern in "$@"; do
		while read -r pid; do
			if [ -n "$pid" ]; then
				seen["$pid"]=1
			fi
		done < <(pgrep -f "$pattern" || true)
	done

	for pid in "${!seen[@]}"; do
		echo "$pid"
	done
}

stop_group() {
	local name="$1"
	shift
	local pid
	local pids=()

	while read -r pid; do
		if [ -n "$pid" ]; then
			pids+=("$pid")
		fi
	done < <(collect_pids "$@")

	if [ "${#pids[@]}" -eq 0 ]; then
		echo "$name is not running."
		return
	fi

	echo "Stopping $name: ${pids[*]}"
	kill "${pids[@]}" 2>/dev/null || true

	sleep 1

	local remaining=()
	for pid in "${pids[@]}"; do
		if kill -0 "$pid" 2>/dev/null; then
			remaining+=("$pid")
		fi
	done

	if [ "${#remaining[@]}" -gt 0 ]; then
		echo "Force stopping $name: ${remaining[*]}"
		kill -9 "${remaining[@]}" 2>/dev/null || true
	fi
}

collect_listen_pids_by_port() {
	local port="$1"
	local pid
	local -A seen=()

	if command -v lsof >/dev/null 2>&1; then
		while read -r pid; do
			if [ -n "$pid" ]; then
				seen["$pid"]=1
			fi
		done < <(lsof -t -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)
	fi

	if [ "${#seen[@]}" -eq 0 ] && command -v ss >/dev/null 2>&1; then
		while read -r pid; do
			if [ -n "$pid" ]; then
				seen["$pid"]=1
			fi
		done < <(
			ss -ltnp "( sport = :$port )" 2>/dev/null \
				| grep -o 'pid=[0-9]\+' \
				| cut -d= -f2 \
				| sort -u
		)
	fi

	for pid in "${!seen[@]}"; do
		echo "$pid"
	done
}

stop_listen_port() {
	local port="$1"
	local name="$2"
	local pid
	local pids=()

	while read -r pid; do
		if [ -n "$pid" ]; then
			pids+=("$pid")
		fi
	done < <(collect_listen_pids_by_port "$port")

	if [ "${#pids[@]}" -eq 0 ]; then
		echo "$name on port $port is not running."
		return
	fi

	echo "Stopping $name on port $port: ${pids[*]}"
	kill "${pids[@]}" 2>/dev/null || true

	sleep 1

	local remaining=()
	for pid in "${pids[@]}"; do
		if kill -0 "$pid" 2>/dev/null; then
			remaining+=("$pid")
		fi
	done

	if [ "${#remaining[@]}" -gt 0 ]; then
		echo "Force stopping $name on port $port: ${remaining[*]}"
		kill -9 "${remaining[@]}" 2>/dev/null || true
	fi
}

rm -rf devgen.car
rm -rf localnet.json
rm -rf ~/.genesis-sectors
rm -rf ~/.lotus
rm -rf ~/.lotusminer
rm -rf DB
rm -rf SST
rm -rf client_upload
rm -rf client_download
rm -rf acc
rm -rf leaf
rm -rf paths
rm -rf zk_output
rm lotus.log
rm miner.log
rm fund.log
rm api.log
rm demo-web.log
rm aggregate.log
rm filenames.log


./lotus fetch-params 8MiB
./lotus-seed pre-seal --sector-size 8MiB --num-sectors 2
./lotus-seed genesis new localnet.json
./lotus-seed genesis add-miner localnet.json ~/.genesis-sectors/pre-seal-t01000.json
#tmux new-session -s "lotus-daemon" -d "./lotus daemon --lotus-make-genesis=devgen.car --genesis-template=localnet.json --bootstrap=false"
nohup ./lotus daemon --lotus-make-genesis=devgen.car --genesis-template=localnet.json --bootstrap=false > lotus.log 2>&1 &

ps -ef | grep lotus

echo "sleep 30s" && sleep 30s
./lotus wallet import --as-default ~/.genesis-sectors/pre-seal-t01000.key
./lotus-miner init --genesis-miner --actor=t01000 --sector-size=8MiB --pre-sealed-sectors=~/.genesis-sectors --pre-sealed-metadata=~/.genesis-sectors/pre-seal-t01000.json --nosync
#tmux new-session -s "lotus-miner" -d "./lotus-miner run --nosync"
nohup ./lotus-miner run --nosync > miner.log 2>&1 &

sleep 30

./lotus net listen
sleep 3
./lotus-miner net listen
sleep 3
./lotus-miner info

stop_group "listen_and_send.sh" \
	"bash scripts/listen_and_send.sh" \
	"scripts/listen_and_send.sh"
stop_listen_port 9999 "listen_and_send nc listener"

nohup bash scripts/listen_and_send.sh > fund.log 2>&1 &

sleep 3

nohup go run ./cmd/opendsn-api > api.log 2>&1 &
nohup bash -lc 'cd demo-web && npm run dev' > demo-web.log 2>&1 &

#nohup bash scripts_FileDES/server_listen_aggregate.sh > aggregate.log 2>&1 &


#./main bls.txt
