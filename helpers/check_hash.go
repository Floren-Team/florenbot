package helpers

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

//go:embed assets/publickey.asc
var publicKeyContent []byte

//go:embed assets/florenbot_linux_amd64.asc
var signatureContent []byte

// VerifyBotSignature теперь использует встроенные данные, а не системный GPG

func VerifyBotSignature(binaryPath string) error {
    // 1. Читаємо ключ (це у вас вже працює)
    keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(publicKeyContent))
    if err != nil {
        return fmt.Errorf("помилка ключа: %w", err)
    }

    // 2. Відкриваємо бінарник
    binaryFile, err := os.Open(binaryPath)
    if err != nil {
        return fmt.Errorf("помилка файлу: %w", err)
    }
    defer binaryFile.Close()

    // 3. РОЗПАШОВУЄМО ARMOR підпису (перетворюємо текст на бінарний потік)
    block, err := armor.Decode(bytes.NewReader(signatureContent))
    if err != nil {
        return fmt.Errorf("помилка декодування armor: %w", err)
    }

    // 4. Тепер передаємо цей бінарний потік у перевірку
    _, err = openpgp.CheckDetachedSignature(keyring, binaryFile, block.Body, nil)
    if err != nil {
        return fmt.Errorf("помилка верифікації: %w", err)
    }

    log.Println("Підпис успішно перевірено.")
    return nil
}


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
