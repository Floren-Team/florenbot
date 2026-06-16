package helpers

import (
	"crypto/sha256"
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
	"hash"
	"io"
	"log"
	"os"
	"strings"
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

	if actualHash != expectedHash {
		panic("Критическая ошибка: хеш-сумма не совпадает! Программа была изменена. " +
			"Безопасный источник: https://github.com/Floren-Team/florenbot/releases")
	}

	log.Println("Хеш-сумма успешно проверена.")
}

func ReadHashFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}