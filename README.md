# (Ре)трансляция аудиопотока в Telegram

Запуск (ре)трансляции потокового аудио в Telegram.

## Docker

### Подготовка

1. Создать образ контейнера

```bash
docker compose build
```
2. Создать чат в Telegram.
3. Запустить в чате аудио-видео звонок в режиме стрима.

### Запуск

В файле docker-compose.yml установить переменные окружения.

DEBUG (false) - включение режима отладки
CHECK (true) - включение проверки Icecast
CHECK_URL (http://icecast:8000/status-json.xsl) - URL проверки
CHECK_INTERVAL (60) - интервал проверки
STREAM_URL (https://stream.radio-t.com) - URL потока вещания
TG_SERVER (dc4-1.rtmp.t.me) - адрес сервера Telegram для приема потока. Выдается при старте вещания в Telegram
TG_KEY - секретный ключ вещания. Выдается при старте вещания в Telegram

```bash
docker compose up -d
```

## Без контейнера

1. Установить ffmpeg и nushell
2. Создать чат в Telegram.
3. Запустить в чате аудио-видео звонок в режиме стрима.
4. TG_KEY=111:AAA nu ./entrypoint.nu
