#!/bin/bash

REQ="hello"

RES=$(docker run --rm --network tp0_testing_net busybox sh -c "printf '$MESSAGE' | nc server 12345")

if [ "$REQ" = "$RES" ]; then
  echo "action: test_echo_server | result: success"
else
  echo "action: test_echo_server | result: fail"
fi
