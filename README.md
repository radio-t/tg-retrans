# (Ре)трансляция аудиопотока в Telegram

Запуск (ре)трансляции потокового аудио в Telegram.

## Docker

### Подготовка

1. Скопировать файл docker-compose.yml.dist в docker-compose.yml.
2. Загрузить образ контейнера с hub.docker.com или собрать его

```bash
docker compose pull
```

или

```bash
docker compose build
```
3. Создать чат в Telegram.
4. Запустить в чате аудио-видео звонок в режиме стрима.

### Запуск

В файле docker-compose.yml установить переменные окружения.

Обязательные:

* `TG_KEY` - секретный ключ вещания. Выдается при старте вещания в Telegram. Между перезапусками вещания сохраняется, но при желании может быть изменён.

Необязательные:

* `DEBUG` (false) - включение режима отладки
* `SKIP_CHECK` (false) - отключение проверки статуса Icecast
* `CHECK_URL` (http://icecast:8000/status-json.xsl) - URL проверки
* `CHECK_INTERVAL` (60s) - интервал проверки
* `CHECK_TIMEOUT` (5s) - таймаут проверки
* `STREAM_URL` (https://stream.radio-t.com) - URL потока вещания
* `FFMPEG_PATH` (/usr/bin/ffmpeg) - путь до ffmpeg
* `TG_SERVER` (dc4-1.rtmp.t.me) - адрес сервера Telegram для приема потока. Выдается при старте вещания в Telegram

```bash
docker compose up -d
```

## Без контейнера

1. Установить `ffmpeg`
2. Создать чат в Telegram.
3. Запустить в чате аудио-видео звонок в режиме стрима.
4. Собрать исполняемый файл. `go build`
5.a Запустить `TG_KEY=111:AAA ./tg-retrans`
5.b Или запустить `TG_KEY=111:AAA nu ./entrypoint.nu`
