#!/bin/bash
if [! -d "log"]; then
    mkdir "log"
fi
./GoGameServer run game 0
