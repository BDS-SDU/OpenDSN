#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

total=0
MINER_IP_LOG="$REPO_ROOT/miner_ip.log"

touch "$MINER_IP_LOG"

echo "start listen on port 9999"
echo "miner ip log: $MINER_IP_LOG"

nc -l -k -p 9999 | while read -r addr miner_ip; do
	if [ -z "$addr" ] || [ -z "$miner_ip" ]; then
		echo "invalid message, expected: <wallet_addr> <miner_ip>"
		continue
	fi

	echo "receive address: $addr"
	echo "receive miner ip: $miner_ip"

	echo "$miner_ip" >> "$MINER_IP_LOG"

	echo "./lotus send $addr 10000000"
	./lotus send "$addr" 10000000

	total=$((total+1))
	echo "total send: $total"
done
