#!/bin/bash

tmux new -d -s 'OCC'
tmux split-window -h
tmux split-window -h
tmux select-layout even-horizontal
tmux selectp -t 0
tmux split-window -v -p 50
tmux selectp -t 2
tmux split-window -v -p 50
tmux selectp -t 4
tmux split-window -v -p 50

tmux send-keys -t 0 "cd server" Enter
tmux send-keys -t 0 "go run server.go 10" Enter

sleep 3

tmux send-keys -t 1 "cd client" Enter
tmux send-keys -t 1 "go run client.go 5 2 1 5 4" Enter

tmux send-keys -t 2 "cd client" Enter
tmux send-keys -t 2 "go run client.go 2 5 1 5 4" Enter

tmux send-keys -t 3 "cd client" Enter
tmux send-keys -t 3 "go run client.go 3 1 1 5 4" Enter

tmux send-keys -t 4 "cd client" Enter
tmux send-keys -t 4 "go run client.go 5 2 4 3 4" Enter

tmux send-keys -t 5 "cd client" Enter
tmux send-keys -t 5 "go run client.go 5 2 1 8 4" Enter

tmux attach -t 'OCC'
tmux kill-session -t 'OCC'
