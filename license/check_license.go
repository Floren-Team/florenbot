package license

import (
	"bufio"
	consts "florenbot/consts"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func LoadLicense() bool {
	// Прочитаем файл лицензии

	if _, err := os.Stat(consts.LICENSE_FILE); os.IsNotExist(err) {
		log.Println("[LicenseManager] Лицензия не найдена пожалуйста запросите лицензию")
		return false
	} else {
		log.Println("[LicenseManager] ✅ Лицензия успешно загружена")
		return true
	}
}

func CheckLicense() bool {
	// Используем GPG для проверки лицензии

	cmd := exec.Command("gpg", "--verify", consts.LICENSE_FILE_ASC, consts.LICENSE_FILE)
	err := cmd.Run()
	if err != nil {
		log.Println("[LicenseManager] ❌ Лицензия не прошла проверку пожалуйста запросите новую лицензию")
		return false
	} else {
		log.Println("[LicenseManager] ✅ Лицензия прошла проверку. Идентифицировано")
		return true
	}
}

func GetExpireLicense() (*bool, error) {
	if _, err := os.Stat(consts.LICENSE_FILE); os.IsNotExist(err) {
		log.Println("[LicenseManager] Лицензия не найдена")
		return nil, nil // Файл відсутній
	}

	file, err := os.Open(consts.LICENSE_FILE)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ExpirationDate:") {
			dateStr := strings.TrimSpace(strings.TrimPrefix(line, "ExpirationDate:"))

			expirationDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return nil, fmt.Errorf("Ошибка парсинга даты: %w", err)
			}

			isActive := time.Now().Before(expirationDate)
			if isActive {
				log.Printf("[LicenseManager] ✅ Лицензия активна, дата окончания: %s\n", expirationDate.Format("2006-01-02"))
			} else {
				log.Println("[LicenseManager] ❌ Лицензия истекла")
			}
			return &isActive, nil
		}
	}
	return nil, fmt.Errorf("поле ExpirationDate не найдено в лицензии")
}
