#!/bin/bash
export DISPLAY=:99
Xvfb :99 -screen 0 1280x720x24 -nolisten tcp &
sleep 1
exec python main.py --server "$@"
