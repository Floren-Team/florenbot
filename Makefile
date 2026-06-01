# Змінні
BINARY_NAME=myapp
GO_FILES=./...

.PHONY: all build clean test lint vet check-deprecated

all: clean build test lint

# Збірка проекту
build:
	go build -o bin/${BINARY_NAME} ***REMOVED***

# Очищення бінарних файлів
clean:
	go clean
	rm -rf bin/

# Запуск тестів
test:
	go test -v ${GO_FILES}

# Запуск стандартного аналізатора (vet)
vet:
	-go vet ${GO_FILES}

# Запуск GolangCI-Lint (потребує встановлення golangci-lint)
lint:
	golangci-lint run ${GO_FILES}

# Запуск Staticcheck для пошуку deprecated коду
check-deprecated:
	staticcheck ${GO_FILES}

# Загальна перевірка
check: vet lint check-deprecated