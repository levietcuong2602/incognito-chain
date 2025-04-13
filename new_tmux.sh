#!/bin/bash

# Create a new tmux session named 'nodes' and detach immediately
tmux new-session -d -s nodes

# Create new windows for each node (adjust as necessary)
tmux new-session -d -s 'shard00'
tmux new-session -d -s 'shard10'
tmux new-session -d -s 'shard01'
tmux new-session -d -s 'shard11'
tmux new-session -d -s 'beacon0'
tmux new-session -d -s 'beacon1'
tmux new-session -d -s 'beacon2'
tmux new-session -d -s 'beacon3'
tmux new-session -d -s 'shard02'
tmux new-session -d -s 'shard03'
tmux new-session -d -s 'shard12'
tmux new-session -d -s 'shard13'
tmux new-session -d -s 'fullnode'

# Send commands to each window
tmux send-keys -t shard00 C-C 'ENTER' './start_node.sh shard0-0' 'ENTER'
sleep 1
tmux send-keys -t shard10 C-C 'ENTER' './start_node.sh shard1-0' 'ENTER'
sleep 1
tmux send-keys -t shard01 C-C 'ENTER' './start_node.sh shard0-1' 'ENTER'
sleep 1
tmux send-keys -t shard11 C-C 'ENTER' './start_node.sh shard1-1' 'ENTER'
sleep 1

tmux send-keys -t beacon0 C-C 'ENTER' './start_node.sh beacon-0' 'ENTER'
sleep 1
tmux send-keys -t beacon1 C-C 'ENTER' './start_node.sh beacon-1' 'ENTER'
sleep 1
tmux send-keys -t beacon2 C-C 'ENTER' './start_node.sh beacon-2' 'ENTER'
sleep 1
tmux send-keys -t beacon3 C-C 'ENTER' './start_node.sh beacon-3' 'ENTER'
sleep 1

tmux send-keys -t shard02 C-C 'ENTER' './start_node.sh shard0-2' 'ENTER'
sleep 1
tmux send-keys -t shard03 C-C 'ENTER' './start_node.sh shard0-3' 'ENTER'
sleep 1
tmux send-keys -t shard12 C-C 'ENTER' './start_node.sh shard1-2' 'ENTER'
sleep 1
tmux send-keys -t shard13 C-C 'ENTER' './start_node.sh shard1-3' 'ENTER'
sleep 1
tmux send-keys -t fullnode C-C 'ENTER' './start_node.sh full_node' 'ENTER'

# Attach to the tmux session
tmux attach-session -t nodes
