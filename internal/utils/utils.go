package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 检查目录是否可写
func IsWritableDir(dir string) bool {
	// 尝试在目录中创建临时文件
	testFile := filepath.Join(dir, ".write_test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}

// 检查字符串切片是否包含指定字符串
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func RemoveAllControlCharacters(data *[]byte) {
	var cleanedData []byte
	original := *data
	for len(original) > 0 {
		r, size := utf8.DecodeRune(original)
		if r != utf8.RuneError && (r >= 32 && r <= 126) || r == '\n' || r == '\t' || r == '\r' || unicode.Is(unicode.Han, r) {
			cleanedData = append(cleanedData, original[:size]...)
		}
		original = original[size:]
	}
	*data = cleanedData
}
func IsDebug() bool {
	debug := os.Getenv("NODEHUB_DEBUG")
	return strings.ToLower(debug) == "true"
}

// IPToUint32 将IP地址转换为uint32
func IPToUint32(ip string) uint32 {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return 0
	}

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return 0
	}

	var result uint32
	for i, part := range parts {
		partInt, err := strconv.Atoi(part)
		if err != nil || partInt < 0 || partInt > 255 {
			return 0
		}
		result |= uint32(partInt) << ((3 - i) * 8)
	}
	return result
}

// Uint32ToIP 将uint32转换为IP地址
func Uint32ToIP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(ip>>24)&0xFF,
		(ip>>16)&0xFF,
		(ip>>8)&0xFF,
		ip&0xFF)
}
