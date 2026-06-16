package helpers

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func CheckHashAndGpg() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("ошибка определения пути к файлу: %w", err)
	}

	if runtime.GOOS == "windows" && filepath.Ext(exePath) != ".exe" {
		exePath += ".exe"
	}

	fmt.Printf("[DEBUG] --- СТАРТ ПРОВЕРКИ ---\n")
	fmt.Printf("[DEBUG] ОС: %s, Архитектура: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("[DEBUG] Путь к бинарному файлу: %s\n", exePath)

	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH != "amd64" && runtime.GOARCH != "386" {
			return fmt.Errorf("архитектура %s не поддерживается на Windows", runtime.GOARCH)
		}

		hash512Path := exePath + ".sha512"
		hash256Path := exePath + ".sha256"

		exists512 := fileExists(hash512Path)
		exists256 := fileExists(hash256Path)

		if !exists512 && !exists256 {
			return errors.New("Критическая ошибка: файлы хешей не найдены")
		}

		if exists512 {
			rawHash, _ := ReadHashFromFile(hash512Path)
			parts := strings.Fields(rawHash)
			if len(parts) == 0 {
				return errors.New("SHA512 файл пуст")
			}
			GetFileHashAndVerify(exePath, "sha512", parts[0])
		} else if exists256 {
			rawHash, _ := ReadHashFromFile(hash256Path)
			parts := strings.Fields(rawHash)
			if len(parts) == 0 {
				return errors.New("SHA256 файл пуст")
			}
			GetFileHashAndVerify(exePath, "sha256", parts[0])
		}
		return nil

	case "linux":
		// Теперь функция VerifyBotSignature требует только путь к бинарнику,
		// так как публичный ключ и подпись уже вшиты (embedded) в код.
		log.Printf("Запуск верификации подписи для: %s", exePath)

		err := VerifyBotSignature(exePath)
		if err != nil {
			return fmt.Errorf("ошибка проверки подписи: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("ОС %s не поддерживается", runtime.GOOS)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
