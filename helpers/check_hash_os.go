package helpers

import (
	"errors"
	"fmt"
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
		// Проверка поддерживаемых архитектур для Windows
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
		// Проверка поддерживаемой архитектуры для Linux
		// if runtime.GOARCH != "amd64" {
		// 	return fmt.Errorf("архитектура %s не поддерживается на Linux (разрешено только amd64)", runtime.GOARCH)
		// }

		ascFilePath := exePath + ".asc"
		if !fileExists(ascFilePath) {
			return fmt.Errorf("файл .asc не найден")
		}
		return VerifyASC(ascFilePath, exePath)

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