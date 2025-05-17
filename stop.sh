#!/usr/bin/env bash

# Define colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# List of tmux sessions to kill
SESSIONS=(
    "beacon0"
    "beacon1"
    "beacon2"
    "beacon3"
    "fullnode"
    "shard00"
    "shard01"
    "shard02"
    "shard03"
    "shard10"
    "shard11"
    "shard12"
    "shard13"
)

# Function to display timestamped messages
log_message() {
    echo -e "${2}[$(date '+%Y-%m-%d %H:%M:%S')] $1${NC}"
}

# Function to check if a tmux session exists
session_exists() {
    tmux has-session -t "$1" 2>/dev/null
}

# Display banner
echo -e "${BLUE}"
echo "====================================================="
echo "  Kill Tmux Sessions Script for Incognito Chain"
echo "====================================================="
echo -e "${NC}"

# Display list of sessions before killing
log_message "Current tmux sessions:" "${YELLOW}"
tmux list-sessions 2>/dev/null || echo -e "${RED}No running sessions${NC}"

# Count killed sessions
count_killed=0

# Kill each session
log_message "Starting the process of killing tmux sessions..." "${BLUE}"

for session in "${SESSIONS[@]}"; do
    if session_exists "$session"; then
        log_message "Stopping application in session $session..." "${YELLOW}"
        # Send Ctrl+C signal to gracefully stop the application
        tmux send-keys -t "$session" C-c
        sleep 2
        
        log_message "Killing session $session..." "${RED}"
        # Kill the session
        tmux kill-session -t "$session"
        
        if ! session_exists "$session"; then
            log_message "Session $session has been killed successfully" "${GREEN}"
            ((count_killed++))
        else
            log_message "Could not kill session $session, trying force kill..." "${RED}"
            tmux kill-session -t "$session" 2>/dev/null || true
            sleep 1
            
            if ! session_exists "$session"; then
                log_message "Session $session has been force killed successfully" "${GREEN}"
                ((count_killed++))
            else
                log_message "WARNING: Could not kill session $session!" "${RED}"
            fi
        fi
    else
        log_message "Session $session does not exist, skipping" "${BLUE}"
    fi
    
    # Brief pause between sessions to avoid overload
    sleep 0.5
done

# Check remaining sessions after killing
remaining_sessions=$(tmux list-sessions 2>/dev/null | wc -l)

# Display results
echo -e "\n${BLUE}====================================================${NC}"
if [ "$count_killed" -gt 0 ]; then
    log_message "Successfully killed $count_killed tmux sessions" "${GREEN}"
else
    log_message "No sessions were killed" "${YELLOW}"
fi

if [ "$remaining_sessions" -gt 0 ]; then
    log_message "WARNING: There are still $remaining_sessions running sessions" "${RED}"
    echo -e "${YELLOW}Remaining sessions:${NC}"
    tmux list-sessions
else
    log_message "All tmux sessions have been successfully killed" "${GREEN}"
fi

echo -e "${BLUE}====================================================${NC}"

rm -rf data LOG LOCK MANIFEST* CURRENT *.log

echo -e "${GREEN}All data directories have been removed${NC}"
echo -e "${BLUE}====================================================${NC}"
echo -e "${GREEN}Script completed successfully!${NC}"
echo -e "${BLUE}====================================================${NC}"
echo -e "${GREEN}Thank you for using the Kill Tmux Sessions Script!${NC}"
echo -e "${BLUE}====================================================${NC}"
echo -e "${GREEN}Have a great day!${NC}"
echo -e "${BLUE}====================================================${NC}"
echo -e "${GREEN}Goodbye!${NC}"
echo -e "${BLUE}====================================================${NC}"
echo -e "${GREEN}Exiting script...${NC}"
echo -e "${BLUE}====================================================${NC}"
