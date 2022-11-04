#!/bin/bash

tmux -u new -d -s 'OCC'
tmux -u split-window -h
tmux -u split-window -h
tmux -u select-layout even-horizontal
tmux -u selectp -t 0
tmux -u split-window -v -p 50
tmux -u selectp -t 2
tmux -u split-window -v -p 50
tmux -u selectp -t 4
tmux -u split-window -v -p 50

tmux -u send-keys -t 0 "cd server" Enter
tmux -u send-keys -t 1 "cd client" Enter
tmux -u send-keys -t 2 "cd client" Enter
tmux -u send-keys -t 3 "cd client" Enter
tmux -u send-keys -t 4 "cd client" Enter
tmux -u send-keys -t 5 "cd client" Enter

tmux -u send-keys -t 0 "go run server.go 10 | tee logs.txt" Enter
sleep 2
tmux -u send-keys -t 1 "go run client.go 5 0 1 3 4" Enter
tmux -u send-keys -t 2 "go run client.go 6 1 1 5 4" Enter
tmux -u send-keys -t 3 "go run client.go 2 0 2 6 4" Enter
tmux -u send-keys -t 4 "go run client.go 3 1 1 5 4" Enter
tmux -u send-keys -t 5 "go run client.go 5 0 1 5 4" Enter

tmux -u attach -t 'OCC'
tmux -u send-keys -t 0 C-c Enter
tmux -u kill-session -t 'OCC'
