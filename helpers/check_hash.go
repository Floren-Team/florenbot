package helpers

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
	"log"
	"os"
	"strings"
	"os/exec"
)

func GetFileHashAndVerify(filePath string, algo string, expectedHash string) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var hasher hash.Hash
	if algo == "sha256" {
		hasher = sha256.New()
	} else {
		hasher = sha512.New()
	}

	if _, err := io.Copy(hasher, f); err != nil {
		log.Fatal(err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))

	// Порівняння хешів
	if actualHash != expectedHash {
		panic("Ошибка при проверки подписи! Программа не может быть дальше работать. " +
			"Рекомендую установить из наших источников: https://github.com/Floren-Team/florenbot/releases")
	}

	log.Println("Программа успешно проверена.")
}

func VerifyASC(signatureFile string, targetFile string) error {
	cmd := exec.Command("gpg", "--verify", signatureFile, targetFile)
	
	// Виводимо вивід команди (stdout/stderr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic("Ошибка при проверки подписи! Программа не может быть дальше работать. Рекомендую установить из наших источников: https://github.com/Floren-Team/florenbot/releases")
	}
	
	log.Println("Подпись успешно проверена:")
	log.Println(string(output))
	return nil
}

func ReadHashFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// Очищаємо від переносів рядків та зайвих пробілів
	return strings.TrimSpace(string(content)), nil
}