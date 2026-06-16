package helpers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func CheckHashAndGpg() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("помилка визначення шляху до файлу: %w", err)
	}

	// Нормалізація для Windows
	if runtime.GOOS == "windows" && filepath.Ext(exePath) != ".exe" {
		exePath += ".exe"
	}

	fmt.Printf("[DEBUG] ОС: %s, Архітектура: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("[DEBUG] Робочий шлях до бінарника: %s\n", exePath)

	switch runtime.GOOS {
	case "windows":
		hash512Path := exePath + ".sha512"
		hash256Path := exePath + ".sha256"

		exists512 := fileExists(hash512Path)
		exists256 := fileExists(hash256Path)

		if !exists512 && !exists256 {
			return errors.New("Критична помилка: файли хешів (.sha512 або .sha256) не знайдені поруч з .exe")
		}

		// Перевірка SHA512
		if exists512 {
			expectedHash, err := ReadHashFromFile(hash512Path)
			if err != nil {
				return fmt.Errorf("помилка зчитування SHA512: %w", err)
			}
			// ВИКЛИК БЕЗ ПЕРЕВІРКИ ПОМИЛКИ, оскільки функція сама робить panic
			GetFileHashAndVerify(exePath, "sha512", expectedHash)
		}

		// Перевірка SHA256
		if exists256 {
			expectedHash, err := ReadHashFromFile(hash256Path)
			if err != nil {
				return fmt.Errorf("помилка зчитування SHA256: %w", err)
			}
			// ВИКЛИК БЕЗ ПЕРЕВІРКИ ПОМИЛКИ
			GetFileHashAndVerify(exePath, "sha256", expectedHash)
		}
		return nil

	case "linux":
		ascFilePath := exePath + ".asc"
		if !fileExists(ascFilePath) {
			return fmt.Errorf("файл підпису (.asc) не знайдено: %s", ascFilePath)
		}
		return VerifyASC(ascFilePath, exePath)

	default:
		return fmt.Errorf("ОС %s не підтримується", runtime.GOOS)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}