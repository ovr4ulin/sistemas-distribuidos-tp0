#!/bin/bash

timeout=10
text="Hello world!"
server_host="server"
server_port=12345
network_name="tp0_testing_net"

output=$(docker run --rm --network "$network_name" alpine sh -c "echo $text | nc -w $timeout $server_host $server_port")

if [ "$output" = "$text" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
