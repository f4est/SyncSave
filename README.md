# SyncSave

Лёгкий, быстрый и бесплатный аналог **SaveSync** для кооперативной синхронизации сохранений игр через **Google Drive** или **локальную/облачную папку**.

> Один включённый ПК за раз: вы загружаете сохранение после игры, а друзья скачивают его позже, когда заходят поиграть. Никаких выделенных серверов и подписок.

---

## Возможности

- **Онлайн-синхронизация через Google Drive API v3** — чистый HTTP-клиент без тяжёлых SDK, OAuth2-вход в один клик и автообновление токенов.
- **Режим локальной / синхронизируемой папки** — можно использовать с Google Drive Desktop, OneDrive, Яндекс.Диском или сетевой папкой без настройки Google Cloud Console.
- **Кооп-механики**:
  - **История версий и откат** — каждый Push создаёт версионированный `.zip`-архив с автором, временем, комментарием и SHA-256 проверкой.
  - **Авто-бэкап перед заменой** — перед любым Pull или откатом старое локальное сохранение автоматически сохраняется в `%APPDATA%\SyncSave\backups`.
  - **Защита от конфликтов** — не даст случайно перезаписать более новый прогресс друга.
  - **Блокировка хоста («Взять / Освободить хост»)** — друзья видят, кто сейчас хостит мир.
- **Библиотека из 24+ игр с автопоиском папок** — Valheim, Terraria, Palworld, Minecraft, Stardew Valley, Factorio, 7 Days to Die, Raft, Sons of the Forest, Project Zomboid, Enshrouded, Baldur's Gate 3 и др. + добавление любых игр вручную.
- **Гибкость и персонализация**:
  - **5 тем оформления**: *SaveSync Dark*, *Midnight Cyber*, *Emerald Forest*, *Crimson Gaming*, *Clean Light*.
  - **3 языка**: Русский, English, Українська.
  - **Сверхнизкое потребление ресурсов**: одиночный `.exe` (~7.5 МБ), фоновый режим в трее без окна выгружает WebView2 (~10–15 МБ ОЗУ и ~0% CPU в простое).

---

## Быстрый старт (Google Drive)

1. Скачайте `SyncSave.exe` из раздела **Releases** (или соберите из исходников).
2. Создайте общую папку на Google Диске и выдайте друзьям доступ **«Редактор»**. Скопируйте **Folder ID** из ссылки (`drive.google.com/drive/folders/<ID>`).
3. В [Google Cloud Console](https://console.cloud.google.com/):
   - **APIs & Services** → включите **Google Drive API**.
   - **Google Auth Platform** → настройте экран согласия (**External**), добавьте свой Gmail и Gmail друзей в **Audience → Test users**.
   - **Clients → Create Client** → выберите **Desktop app** → скопируйте `Client ID` и `Client Secret`.
4. Откройте `SyncSave.exe` → **Настройки** → вставьте ключи и Folder ID → нажмите **«Войти через Google»**.

---

## Сборка из исходников (Windows)

Требуется Go 1.24+:

```powershell
go test -v ./...
$env:CGO_ENABLED="0"
go build -trimpath -ldflags="-H windowsgui -s -w" -o SyncSave.exe .
```

---

## CI/CD и публикация релизов

В репозитории настроен GitHub Actions ([.github/workflows/release.yml](.github/workflows/release.yml)):
- На каждый `push` / `pull_request` в `main` запускаются `go vet`, `go test` и тестовая сборка `SyncSave.exe`.
- При публикации тега вида `v1.0.0` (`git tag v1.0.0 && git push origin v1.0.0`) автоматически создаётся **GitHub Release** с готовым `SyncSave.exe` и `SyncSave.exe.sha256`.
