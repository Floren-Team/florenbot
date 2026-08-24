Как сделать билд проекта на Linux?


1. Скачайте архив Go
 ```bash
    wget https://go.dev/dl/go1.26.4.linux-amd64.tar.gz
   ```

2. Распакуйте архив в /usr/local/bin:
```bash
   sudo rm -rf /usr/local/go (Если была инсталяция ранее)
   sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz 
```
3. Настройте PATH:

```bash
  echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
  source ~/.bashrc
```

Проверьте что у вас есть MariaDB (MySQL), Redis и настройте .env

   ```env
    REDIS_ADDR=localhost:6379
    REDIS_PASSWORD=
    REDIS_DB=0
    
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=
    DB_PASSWORD=
    DB_NAME=
    DB_ROOT_PASSWORD=
   REDIS_ADDR=localhost:6379
```

Далее скомпилируйте исходный код

```bash
go build
```

## Использование Docker 

Важно! На компьютере должен быть установлен Docker если нет установите его

Для сборки проекта просто пропишите команду


Но очень важно скопировать файл .env.docker.example как .env.example

1. Скопируй файл
   
```bash
  cp .env.docker.example .env.docker
 ```

2. Заполни его
     ```env
    REDIS_ADDR=redis:6379
    REDIS_PASSWORD=
    REDIS_DB=0
    
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=
    DB_PASSWORD=
    DB_NAME=
    DB_ROOT_PASSWORD=
   REDIS_ADDR=localhost:6379
   ```
     
После заполнения файла запустите команду 

```bash
 ./build.sh
```

или

```bash
  bash build.sh
```
Все готово!





