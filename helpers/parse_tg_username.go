package helpers

import (
	"regexp"
	"errors"
	"strings"
)

// parseTelegramUsername парсит юзернейм Telegram
func ParseTelegramUsername(input string) (string, error) {
	// регулярное выражение для юзернейма
	re := regexp.MustCompile(`^@?([a-zA-Z0-9_]{5,32})$`)
	
	input = strings.TrimSpace(input)
	matches := re.FindStringSubmatch(input)
	
	if len(matches) < 2 {
		return "", errors.New("Некорректный юзернейм Telegram")
	}
	
	//Возвращаем юзернейм
	return matches[1], nil
}