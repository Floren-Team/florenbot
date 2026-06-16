package helpers


import (
	"os"
	"runtime"
	"errors"
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
        hashFilePath := exePath + ".asc"
        if err := VerifyASC(hashFilePath, exePath); err != nil {
            return errors.New("Ошибка при проверки подписи! Рекомендую: https://github.com/Floren-Team/florenbot/releases")
        }
        return nil

    default:
        return errors.New("Данная ОС не поддерживается")
    }
}