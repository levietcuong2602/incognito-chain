#!/bin/bash

# Kiểm tra tham số đầu vào
if [ "$#" -lt 1 ]; then
    echo "Usage: $0 <session_name|all> [output_directory] [interval_seconds]"
    echo "Example: $0 all logs 5"
    echo "Example: $0 my_session /var/logs 10"
    exit 1
fi

SESSION="$1"
OUTPUT_DIR="${2:-logs}"  # Thư mục mặc định là "logs" nếu không chỉ định
SLEEP_INTERVAL="${3:-5}"  # Mặc định 5 giây nếu không chỉ định

# Tạo thư mục logs nếu chưa tồn tại
mkdir -p "$OUTPUT_DIR"

# Tạo tên file log với timestamp và session name
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
if [ "$SESSION" = "all" ]; then
    OUTPUT_FILE="$OUTPUT_DIR/tmux_all_sessions_${TIMESTAMP}.log"
    echo "📄 Capturing logs from ALL tmux sessions every ${SLEEP_INTERVAL}s"
else
    OUTPUT_FILE="$OUTPUT_DIR/tmux_session_${SESSION}_${TIMESTAMP}.log"
    echo "📄 Capturing logs from tmux session '${SESSION}' every ${SLEEP_INTERVAL}s"
fi

echo "📄 Log will be saved to: $OUTPUT_FILE"
echo "Press Ctrl+C to stop..."

# Tạo header cho file log
echo "====== TMUX LOG CAPTURE ======" > "$OUTPUT_FILE"
echo "Started: $(date)" >> "$OUTPUT_FILE"
echo "Session: ${SESSION}" >> "$OUTPUT_FILE"
echo "Interval: ${SLEEP_INTERVAL}s" >> "$OUTPUT_FILE"
echo "==============================" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Xử lý khi nhận tín hiệu Ctrl+C
trap 'echo -e "\n✅ Capturing stopped. Log saved to: $OUTPUT_FILE"; exit 0' INT

# Hàm lấy danh sách session và pane
get_tmux_panes() {
    if [ "$SESSION" = "all" ]; then
        tmux list-panes -a -F "#{session_name}:#{window_index}.#{pane_index}"
    else
        tmux list-panes -t "$SESSION" -a -F "#{session_name}:#{window_index}.#{pane_index}" 2>/dev/null
        if [ $? -ne 0 ]; then
            echo "Error: Tmux session '$SESSION' not found."
            exit 1
        fi
    fi
}

# File tạm để lưu danh sách pane đã thấy
SEEN_PANES_FILE="/tmp/tmux_seen_panes_$$"
touch "$SEEN_PANES_FILE"

# Xóa file tạm khi thoát
trap 'echo -e "\n✅ Capturing stopped. Log saved to: $OUTPUT_FILE"; rm -f "$SEEN_PANES_FILE"; exit 0' INT TERM EXIT

while true; do
    # Lấy timestamp hiện tại
    CURRENT_TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S")
    
    # Ghi timestamp vào file
    echo -e "\n\n==============================================" >> "$OUTPUT_FILE"
    echo "TIMESTAMP: $CURRENT_TIMESTAMP" >> "$OUTPUT_FILE"
    echo "==============================================" >> "$OUTPUT_FILE"
    
    # Lặp qua tất cả các pane được chỉ định
    get_tmux_panes | while read -r pane; do
        # Kiểm tra xem pane này đã được thấy chưa
        if ! grep -q "^$pane$" "$SEEN_PANES_FILE"; then
            # Nếu chưa thấy, thêm vào file danh sách đã thấy
            echo "$pane" >> "$SEEN_PANES_FILE"
            SESSION_NAME=$(echo "$pane" | cut -d':' -f1)
            echo -e "\n\n▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓" >> "$OUTPUT_FILE"
            echo "▶ NEW PANE DETECTED: $pane (Session: $SESSION_NAME)" >> "$OUTPUT_FILE" 
            echo "▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓" >> "$OUTPUT_FILE"
        fi
        
        # Capture toàn bộ nội dung pane
        CONTENT=$(tmux capture-pane -p -t "$pane")
        
        # Ghi thông tin pane và nội dung vào file
        echo -e "\n────────────────────────────────────" >> "$OUTPUT_FILE"
        echo "PANE: $pane" >> "$OUTPUT_FILE"
        echo "────────────────────────────────────" >> "$OUTPUT_FILE"
        echo "$CONTENT" >> "$OUTPUT_FILE"
        
        # Hiển thị tiến trình trên terminal
        echo -ne "Capturing pane: $pane at $CURRENT_TIMESTAMP   \r"
    done
    
    # Hiển thị thông báo hoàn thành trên terminal
    echo -e "Completed capture at $CURRENT_TIMESTAMP. Next capture in ${SLEEP_INTERVAL}s...\r"
    
    # Đợi trước khi quét lại
    sleep "$SLEEP_INTERVAL"
done
