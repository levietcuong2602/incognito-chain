package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Tạo logger để ghi vào file
var TxLogger *log.Logger
var txLogFile *os.File
var logEnabled bool = false

// InitTxLogger khởi tạo logger để ghi log transaction vào file
func InitTxLogger(logDir string, logFile string) error {
	// Tạo thư mục log nếu chưa tồn tại
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("cannot create log directory: %v", err)
	}

	// Tạo tên file log với timestamp
	timestamp := time.Now().Format("2006-01-02")
	logFilePath := filepath.Join(logDir, fmt.Sprintf("%s_%s", timestamp, logFile))

	// Mở file để ghi log, tạo nếu chưa tồn tại, append nếu đã tồn tại
	var err error
	txLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open log file: %v", err)
	}

	// Tạo logger mới ghi vào file
	TxLogger = log.New(txLogFile, "", log.LstdFlags)
	logEnabled = true

	return nil
}

// CloseTxLogger đóng file log khi ứng dụng kết thúc
func CloseTxLogger() {
	if txLogFile != nil {
		txLogFile.Close()
		logEnabled = false
	}
}

// IsEnabled kiểm tra xem logger đã được khởi tạo chưa
func IsEnabled() bool {
	return logEnabled && TxLogger != nil
}

// LogPrintf ghi log với định dạng (nếu logger đã được kích hoạt)
func LogPrintf(format string, v ...interface{}) {
	if IsEnabled() {
		TxLogger.Printf(format, v...)
	}
}

// LogPrintln ghi log dạng dòng (nếu logger đã được kích hoạt)
func LogPrintln(v ...interface{}) {
	if IsEnabled() {
		TxLogger.Println(v...)
	}
}
