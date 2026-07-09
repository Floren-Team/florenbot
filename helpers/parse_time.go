package helpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseDuration(timeStr string) (time.Duration, error) {
	timeStr = strings.ToLower(strings.TrimSpace(timeStr))

	// Защита: если строка пустая
	if timeStr == "" {
		return 0, fmt.Errorf("пустая строка")
	}

	// Обработка дней
	if strings.HasSuffix(timeStr, "d") {
		daysStr := strings.TrimSuffix(timeStr, "d")
		// Если пользователь ввел просто "d" без числа
		if daysStr == "" {
			return 0, fmt.Errorf("отсутствует число перед 'd'")
		}
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// Для всего остального (1h, 5m, 30s)
	return time.ParseDuration(timeStr)
}
