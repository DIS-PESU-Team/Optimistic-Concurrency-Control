#!/bin/bash

# kill -9 $(lsof -i tcp:4040 -t)
# kill -9 $(lsof -i tcp:4041 -t)
# kill -9 $(lsof -i tcp:4042 -t)

tmux -u new -d -s 'OCC'
tmux -u split-window -h -p 50
tmux -u selectp -t 0
tmux -u split-window -v -p 50

tmux -u send-keys -t 0 "cd client" Enter
tmux -u send-keys -t 1 "cd client" Enter
tmux -u send-keys -t 2 "cd server" Enter

tmux -u send-keys -t 2 "go run server.go 10 | tee logs.txt" Enter
sleep 5
tmux -u send-keys -t 0 "go run client.go 2 5 1 5 6" Enter
tmux -u send-keys -t 1 "go run client.go 5 2 1 5 6" Enter

tmux -u attach -t 'OCC'
tmux -u send-keys -t 2 C-c Enter
tmux -u kill-session -t 'OCC'
