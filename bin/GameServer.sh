#!/bin/bash
cd "$(dirname "$0")" || exit 1
mkdir -p "log"
./GoGameServer run game 0
