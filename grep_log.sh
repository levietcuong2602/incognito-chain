#!/bin/bash

# Kiểm tra xem có truyền pattern grep hay không
if [ -z "$1" ]; then
    echo "Usage: $0 <grep_pattern> [sleep_interval]"
    echo "Example: $0 'error' 2"
    exit 1
fi

GREP_PATTERN="$1"
SLEEP_INTERVAL="${2:-1}"  # Mặc định 1 giây nếu không chỉ định

echo "🔍 Monitoring all tmux panes for pattern: '$GREP_PATTERN' (Interval: ${SLEEP_INTERVAL}s)"
echo "Press Ctrl+C to stop..."

while true; do
    # Lặp qua tất cả các pane trong tmux
    tmux list-panes -a -F "#{session_name}:#{window_index}.#{pane_index}" | while read -r pane; do
        # Capture nội dung pane và grep
        CONTENT=$(tmux capture-pane -p -t "$pane")
        
        # Kiểm tra pattern và hiển thị kết quả (nếu có)
        if echo "$CONTENT" | grep -q "$GREP_PATTERN"; then
            echo "=== 🟢 Found in $pane ==="
            echo "$CONTENT" | grep --color=always "$GREP_PATTERN"
            echo ""
        fi
    done
    
    sleep "$SLEEP_INTERVAL"  # Đợi trước khi quét lại
done
