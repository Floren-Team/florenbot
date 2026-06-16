package helpers

import (
	"errors"
	"os"
	"runtime"
)

func CheckHashAndGpg() error {

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		switch runtime.GOARCH {
		case "amd64", "386":
			// Проверяем SHA-512
			hash512Path := exePath + ".sha512"
			expectedHash512, err := ReadHashFromFile(hash512Path)
			if err == nil {
				GetFileHashAndVerify(exePath, "sha512", expectedHash512)
			}

			// Проверяем SHA-256
			hash256Path := exePath + ".sha256"
			expectedHash256, err := ReadHashFromFile(hash256Path)
			if err == nil {
				GetFileHashAndVerify(exePath, "sha256", expectedHash256)
			}

			// Если оба файла отсутствуют, возвращаем ошибку
			if _, err := os.Stat(hash512Path); os.IsNotExist(err) {
				if _, err := os.Stat(hash256Path); os.IsNotExist(err) {
					return errors.New("Файлы хешей (.sha512 или .sha256) не найдены")
				}
			}
			return nil
		default:
			return errors.New("Данная архитектура не поддерживается")
		}

	case "linux":
		switch runtime.GOARCH {
		case "amd64", "386", "arm64":
			// Проверка цифровой подписи GPG для Linux
			ascFilePath := exePath + ".asc"
			// Проверяем, существует ли файл .asc перед проверкой
            if _, err := os.Stat(ascFilePath); err == nil {
                // Файл есть, выполняем проверку
                if err := VerifyASC(ascFilePath, exePath); err != nil {
                    return err // Если файл есть, но подпись неверна — возвращаем ошибку
                }
            } else if os.IsNotExist(err) {
                return errors.New("Файл подписи (.asc) не найден! Программа не может быть дальше работать. " +
                    "Рекомендую установить из наших источников: https://github.com/Floren-Team/florenbot/releases")
            } else {
                // Ошибка доступа к файлу (например, нет прав)
                return err
            }
			return nil
		default:
			return errors.New("Данная архитектура не поддерживается")
		}

	default:
		return errors.New("Данная ОС не поддерживается")
	}
}