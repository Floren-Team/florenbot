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
			hashFilePath := exePath + ".sha512"
			expectedHash, err := ReadHashFromFile(hashFilePath)
			if err != nil {
				return err
			}
			GetFileHashAndVerify(exePath, "sha512", expectedHash)
			return nil
		default:
			return errors.New("Данная архитектура не поддерживается")
		}

	case "linux":
		switch runtime.GOARCH {
		case "amd64", "386", "arm64":
			ascFilePath := exePath + ".asc"
			if err := VerifyASC(ascFilePath, exePath); err != nil {
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