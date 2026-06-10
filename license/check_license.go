package license


import (
	"os"
	"log"
	consts "florenbot/consts"
	"os/exec"
)

func LoadLicense() bool {
	// Прочитаем файл лицензии

	if _, err := os.Stat(consts.***REMOVED***_FILE); os.IsNotExist(err) {
		log.Println("[LicenseManager] Лицензия не найдена пожалуйста запросите лицензию")
		return false
	} else {
		log.Println("[LicenseManager] ✅ Лицензия успешно загружена")
		return true
	}
}

func CheckLicense() bool {
	// Используем GPG для проверки лицензии


	cmd := exec.Command("gpg", "--verify", consts.***REMOVED***_FILE)
	err := cmd.Run()
	if err != nil {
		log.Println("[LicenseManager] ❌ Лицензия не прошла проверку пожалуйста запросите новую лицензию")
		return false
	} else {
		log.Println("[LicenseManager] ✅ Лицензия прошла проверку. Идентифицировано")
		return true
	}
}