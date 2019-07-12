#!/usr/bin/env bash

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

attempts=1
max_attempts=10
wait=5

function compose() {
	docker-compose --file ${__dir}/../docker-compose.yml ${*:1}
}

while ! compose logs | grep Ready > /dev/null; do
	((attempts++))
	if [ "$attempts" -gt "$max_attempts" ]; then
		echo "Giving up! See the logs below.";
		echo "---------------------------------------"
		compose logs
		echo "---------------------------------------"
		exit 1;
	fi;
	echo "Not ready yet..."
	sleep $wait
done
