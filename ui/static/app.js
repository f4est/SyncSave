const state = {
  config: null,
  presets: [],
  statuses: {},
  selectedGameId: null,
  logs: [],
};

const I18N = {
  ru: {
    refresh: "Обновить",
    settings: "Настройки",
    games_worlds: "Игры и миры",
    add_btn: "+ Добавить",
    bg_hint: "Фоновый режим: ~15 МБ ОЗУ",
    welcome_title: "Добро пожаловать в SyncSave",
    welcome_desc: "Добавьте игру, подключите общую папку Google Drive и играйте в коопе без выделенного сервера — любой друг сможет хостить мир с последним сохранением.",
    add_game_btn: "+ Добавить игру",
    setup_gdrive_btn: "Настроить Google Drive",
    theme_label: "Тема оформления",
    lang_label: "Язык интерфейса",
    settings_title: "Настройки SyncSave",
    tab_general: "Общие",
    tab_tray: "Трей и Автозапуск",
    tab_cloud: "Облако",
    tab_sync: "Синхронизация",
    player_name_label: "Ваш никнейм (виден друзьям в истории)",
    autostart_label: "Запускать SyncSave автоматически при входе в Windows",
    minimize_tray: "Сворачивать в трей при закрытии окна (крестик)",
    start_minimized: "Запускать свёрнутым в системный трей",
    tray_single_click: "Открывать окно по одиночному клику на иконку в трее",
    tray_notifications: "Показывать уведомления Windows при синхронизации и конфликтах",
    memory_saver: "Режим экономии ОЗУ: сжимать память при сворачивании в трей",
    backend_label: "Режим синхронизации",
    pause_autosync: "Поставить фоновую авто-синхронизацию на паузу",
    autopull_start: "Проверять и синхронизировать миры при запуске программы",
    poll_label: "Интервал проверки файлов (мс)",
    debounce_label: "Задержка перед авто-загрузкой (сек)",
    backups_label: "Максимум локальных бэкапов на игру",
    state_synced: "Синхронизировано",
    state_local_ahead: "Есть локальные изменения",
    state_remote_ahead: "Доступно обновление",
    state_conflict: "Конфликт версий",
    state_not_synced: "Не синхронизировано",
    state_missing_local: "Папка не найдена",
    state_error: "Ошибка",
    host_free: "Хост свободен",
    host_take: "Взять хост",
    host_release: "Освободить хост",
    host_steal: "Перехватить",
    host_by: "Хостит:",
    rollback: "Откатить",
  },
  en: {
    refresh: "Refresh",
    settings: "Settings",
    games_worlds: "Games & Worlds",
    add_btn: "+ Add",
    bg_hint: "Background RAM: ~15 MB",
    welcome_title: "Welcome to SyncSave",
    welcome_desc: "Add a game, connect a shared Google Drive folder, and play co-op without a dedicated server — any friend can host the world with the latest save.",
    add_game_btn: "+ Add Game",
    setup_gdrive_btn: "Setup Google Drive",
    theme_label: "Theme",
    lang_label: "Language",
    settings_title: "SyncSave Settings",
    tab_general: "General",
    tab_tray: "Tray & Autostart",
    tab_cloud: "Cloud",
    tab_sync: "Sync & Backups",
    player_name_label: "Player nickname (visible to friends)",
    autostart_label: "Launch SyncSave automatically on Windows startup",
    minimize_tray: "Minimize to tray on window close",
    start_minimized: "Start minimized to system tray",
    tray_single_click: "Open window on single tray icon click",
    tray_notifications: "Show Windows notifications on sync and conflicts",
    memory_saver: "Memory Saver: trim working set when minimized to tray",
    backend_label: "Sync backend",
    pause_autosync: "Pause background auto-sync",
    autopull_start: "Check and sync worlds on startup",
    poll_label: "File check interval (ms)",
    debounce_label: "Auto-upload debounce delay (sec)",
    backups_label: "Max local backups per game",
    state_synced: "Synced",
    state_local_ahead: "Local changes",
    state_remote_ahead: "Update available",
    state_conflict: "Version conflict",
    state_not_synced: "Not synced",
    state_missing_local: "Folder not found",
    state_error: "Error",
    host_free: "Host free",
    host_take: "Claim Host",
    host_release: "Release Host",
    host_steal: "Take Over",
    host_by: "Host:",
    rollback: "Rollback",
  },
  uk: {
    refresh: "Оновити",
    settings: "Налаштування",
    games_worlds: "Ігри та світи",
    add_btn: "+ Додати",
    bg_hint: "Фоновий режим: ~15 МБ ОЗП",
    welcome_title: "Ласкаво просимо до SyncSave",
    welcome_desc: "Додайте гру, підключіть спільну папку Google Drive та грайте в коопі без виділеного сервера — будь-хто з друзів зможе хостити світ з останнім збереженням.",
    add_game_btn: "+ Додати гру",
    setup_gdrive_btn: "Налаштувати Google Drive",
    theme_label: "Тема оформлення",
    lang_label: "Мова інтерфейсу",
    settings_title: "Налаштування SyncSave",
    tab_general: "Загальні",
    tab_tray: "Трей та Автозапуск",
    tab_cloud: "Хмара",
    tab_sync: "Синхронізація",
    player_name_label: "Ваш нікнейм (видно друзям в історії)",
    autostart_label: "Запускати SyncSave автоматично при вході в Windows",
    minimize_tray: "Згортати в трей при закритті вікна",
    start_minimized: "Запускати згорнутим у системний трей",
    tray_single_click: "Відкривати вікно одним кліком по іконці в треї",
    tray_notifications: "Показувати сповіщення Windows при синхронізації та конфліктах",
    memory_saver: "Режим економії ОЗП: стискати пам'ять при згортанні в трей",
    backend_label: "Режим синхронізації",
    pause_autosync: "Поставити фонову авто-синхронізацію на паузу",
    autopull_start: "Перевіряти та синхронізувати світи при запуску",
    poll_label: "Інтервал перевірки файлів (мс)",
    debounce_label: "Затримка перед авто-завантаженням (сек)",
    backups_label: "Максимум локальних бекапів на гру",
    state_synced: "Синхронізовано",
    state_local_ahead: "Є локальні зміни",
    state_remote_ahead: "Доступне оновлення",
    state_conflict: "Конфлікт версій",
    state_not_synced: "Не синхронізовано",
    state_missing_local: "Папку не знайдено",
    state_error: "Помилка",
    host_free: "Хост вільний",
    host_take: "Взяти хост",
    host_release: "Звільнити хост",
    host_steal: "Перехопити",
    host_by: "Хостить:",
    rollback: "Відкотити",
  },
};

function t(key) {
  const lang = state.config?.language || "ru";
  return (I18N[lang] && I18N[lang][key]) || I18N.ru[key] || key;
}

function applyTheme(theme) {
  document.documentElement.setAttribute("data-theme", theme || "dark");
}

function applyLanguage(lang) {
  const dict = I18N[lang] || I18N.ru;
  document.querySelectorAll("[data-i18n]").forEach((el) => {
    const key = el.getAttribute("data-i18n");
    if (dict[key]) el.textContent = dict[key];
  });
}

function stateLabel(code) {
  return t("state_" + code) || code;
}

async function api(path, method = "GET", body = null) {
  const opts = { method, headers: {} };
  if (body !== null) {
    opts.headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(path, opts);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || `HTTP ${res.status}`);
  }
  return data;
}

function toast(msg, isErr = false) {
  const el = document.getElementById("toast");
  el.textContent = msg;
  el.style.borderColor = isErr ? "var(--danger)" : "var(--accent)";
  el.classList.remove("hidden");
  clearTimeout(window._toastTimer);
  window._toastTimer = setTimeout(() => el.classList.add("hidden"), 4000);
}

async function init() {
  await Promise.all([loadConfig(), loadPresets(), loadLogs()]);
  await refreshAllStatuses();
  setInterval(loadLogs, 3000);
  setInterval(refreshAllStatuses, 15000);
}

async function loadConfig() {
  state.config = await api("/api/config");
  applyTheme(state.config.theme || "dark");
  applyLanguage(state.config.language || "ru");
  document.getElementById("playerNameLabel").textContent = state.config.player_name || "Player";
  if (!state.selectedGameId && state.config.games.length > 0) {
    state.selectedGameId = state.config.games[0].id;
  }
  checkBackendPill();
  renderGameList();
}

async function checkBackendPill() {
  const badge = document.getElementById("backendBadge");
  const text = document.getElementById("backendText");
  try {
    const res = await api("/api/backend/check");
    badge.className = "pill ok";
    text.textContent = res.label;
  } catch (e) {
    badge.className = "pill err";
    text.textContent = state.config?.backend === "local" ? "Папка не настроена" : "Google Drive не подключен";
  }
}

async function loadPresets() {
  state.presets = await api("/api/presets");
}

async function loadLogs() {
  state.logs = await api("/api/logs");
  renderLogs();
}

async function refreshAllStatuses() {
  if (!state.config) return;
  for (const g of state.config.games) {
    try {
      state.statuses[g.id] = await api(`/api/games/${encodeURIComponent(g.id)}/status`);
    } catch (e) {
      state.statuses[g.id] = { game: g, state: "error", error: e.message, versions: [] };
    }
  }
  renderGameList();
  renderDetail();
}

function renderGameList() {
  const list = document.getElementById("gameList");
  list.innerHTML = "";
  const games = state.config?.games || [];
  if (games.length === 0) {
    document.getElementById("emptyState").classList.remove("hidden");
    document.getElementById("gameDetail").classList.add("hidden");
    return;
  }
  document.getElementById("emptyState").classList.add("hidden");
  document.getElementById("gameDetail").classList.remove("hidden");

  for (const g of games) {
    const st = state.statuses[g.id];
    const stCode = st?.state || "not_synced";
    const div = document.createElement("div");
    div.className = "game-item" + (g.id === state.selectedGameId ? " active" : "");
    div.onclick = () => {
      state.selectedGameId = g.id;
      renderGameList();
      renderDetail();
    };
    div.innerHTML = `
      <div>
        <div class="game-item-name">${escapeHtml(g.name)}</div>
        <div class="game-item-sub">${escapeHtml(g.remote_folder)} • v${st?.remote_version ?? g.last_synced_ver ?? 0}</div>
      </div>
      <span class="status-dot st-${stCode}" title="${stateLabel(stCode)}"></span>
    `;
    list.appendChild(div);
  }
}

function renderDetail() {
  const id = state.selectedGameId;
  if (!id || !state.config) return;
  const g = state.config.games.find((x) => x.id === id);
  if (!g) return;
  const st = state.statuses[id] || { game: g, state: "not_synced", versions: [] };

  document.getElementById("detailName").textContent = g.name;
  document.getElementById("detailStateBadge").textContent = stateLabel(st.state);
  document.getElementById("detailLocalPath").textContent = g.local_path;
  document.getElementById("detailRemoteFolder").textContent = g.remote_folder;
  document.getElementById("detailVersion").textContent = `v${st.remote_version || 0}`;
  document.getElementById("detailFiles").textContent = st.local_files || 0;
  document.getElementById("detailSize").textContent = ((st.local_size || 0) / (1024 * 1024)).toFixed(2) + " МБ";
  document.getElementById("chkAutoSync").checked = !!g.auto_sync;

  // Host lock
  const lockInfo = document.getElementById("hostLockInfo");
  const lockBtn = document.getElementById("btnHostLock");
  if (st.host_lock) {
    lockInfo.textContent = `${t("host_by")} ${st.host_lock.holder}`;
    lockBtn.textContent = st.host_lock.holder === state.config.player_name ? t("host_release") : t("host_steal");
  } else {
    lockInfo.textContent = t("host_free");
    lockBtn.textContent = t("host_take");
  }

  // Conflict / error banner
  const banner = document.getElementById("conflictBanner");
  const bannerText = document.getElementById("conflictText");
  if (st.state === "conflict") {
    banner.classList.remove("hidden");
    bannerText.textContent = `Внимание: локальные сохранения изменены И в облаке есть новая версия v${st.remote_version} от ${st.remote_author}.`;
  } else if (st.error) {
    banner.classList.remove("hidden");
    bannerText.textContent = st.error;
  } else {
    banner.classList.add("hidden");
  }

  // Versions table
  const tbody = document.getElementById("versionBody");
  tbody.innerHTML = "";
  const versions = st.versions || [];
  document.getElementById("versionCount").textContent = `${versions.length} версий`;
  for (const v of versions) {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td><strong>v${v.version}</strong></td>
      <td>${escapeHtml(v.author)}</td>
      <td>${escapeHtml(v.timestamp)}</td>
      <td>${(v.zip_size / (1024 * 1024)).toFixed(2)} МБ</td>
      <td>${escapeHtml(v.note || "—")}</td>
      <td><button class="btn btn-secondary btn-xs" onclick="rollbackVersion(${v.version})">${t("rollback")}</button></td>
    `;
    tbody.appendChild(tr);
  }
}

function renderLogs() {
  const el = document.getElementById("activityLog");
  el.innerHTML = "";
  for (const l of state.logs) {
    const div = document.createElement("div");
    div.className = `log-item ${l.level}`;
    div.innerHTML = `<span class="log-time">${l.time}</span>${l.game ? `<strong>[${escapeHtml(l.game)}]</strong> ` : ""}${escapeHtml(l.message)}`;
    el.appendChild(div);
  }
}

async function pushGame(force = false) {
  const id = state.selectedGameId;
  if (!id) return;
  const note = document.getElementById("pushNote").value;
  try {
    toast("Упаковка и загрузка сохранения...");
    const ver = await api(`/api/games/${encodeURIComponent(id)}/push`, "POST", { note, force });
    document.getElementById("pushNote").value = "";
    toast(`Версия v${ver.version} загружена!`);
    await loadConfig();
    await refreshAllStatuses();
  } catch (e) {
    toast(e.message, true);
  }
}

async function pullGame(version = 0, force = false) {
  const id = state.selectedGameId;
  if (!id) return;
  try {
    toast("Скачивание сохранения...");
    const ver = await api(`/api/games/${encodeURIComponent(id)}/pull`, "POST", { version, force });
    toast(`Установлена версия v${ver.version}!`);
    await loadConfig();
    await refreshAllStatuses();
  } catch (e) {
    toast(e.message, true);
  }
}

window.rollbackVersion = async (ver) => {
  if (!confirm(`Откатить локальное сохранение к версии v${ver}? Текущее сохранение будет автоматически забэкаплено.`)) return;
  await pullGame(ver, true);
};

// Modals
window.closeModals = () => {
  document.getElementById("modalAddGame").classList.add("hidden");
  document.getElementById("modalSettings").classList.add("hidden");
};

function openAddGameModal() {
  const grid = document.getElementById("presetGrid");
  grid.innerHTML = "";
  for (const p of state.presets) {
    const btn = document.createElement("button");
    btn.className = "preset-chip";
    btn.innerHTML = `${escapeHtml(p.name)}${p.installed ? `<span class="installed">✓ Найдена на ПК</span>` : ""}`;
    btn.onclick = () => {
      document.querySelectorAll(".preset-chip").forEach((c) => c.classList.remove("selected"));
      btn.classList.add("selected");
      if (p.id !== "custom") {
        document.getElementById("addGameName").value = p.name;
        document.getElementById("addRemoteFolder").value = p.name.replace(/[^a-zA-Z0-9_-]/g, "_") + "_World";
      }
      if (p.detected_path) {
        document.getElementById("addLocalPath").value = p.detected_path;
      }
    };
    grid.appendChild(btn);
  }
  document.getElementById("remoteWorldSelect").classList.add("hidden");
  document.getElementById("modalAddGame").classList.remove("hidden");
}

function openSettingsModal() {
  const c = state.config;
  document.getElementById("setPlayerName").value = c.player_name || "";
  document.getElementById("setBackend").value = c.backend || "gdrive";
  document.getElementById("setClientID").value = c.gdrive?.client_id || "";
  document.getElementById("setClientSecret").value = c.gdrive?.client_secret || "";
  document.getElementById("setFolderID").value = c.gdrive?.folder_id || "";
  document.getElementById("setLocalSyncDir").value = c.local_sync_dir || "";
  document.getElementById("setTheme").value = c.theme || "dark";
  document.getElementById("setLanguage").value = c.language || "ru";
  document.getElementById("setAutostart").checked = !!c.autostart;
  document.getElementById("setMinimizeToTray").checked = !!c.minimize_to_tray;
  document.getElementById("setStartMinimized").checked = !!c.start_minimized;
  document.getElementById("setTraySingleClick").checked = !!c.tray_single_click;
  document.getElementById("setTrayNotifications").checked = !!c.tray_notifications;
  document.getElementById("setMemorySaver").checked = !!c.memory_saver;
  document.getElementById("setAutoSyncPaused").checked = !!c.auto_sync_paused;
  document.getElementById("setAutoPullOnStart").checked = !!c.auto_pull_on_start;
  document.getElementById("setPollMs").value = c.poll_interval_ms || 5000;
  document.getElementById("setDebounceSec").value = c.debounce_seconds || 8;
  document.getElementById("setKeepBackups").value = c.default_keep_backups || 5;
  document.getElementById("oauthEmailBadge").textContent = c.gdrive?.user_email ? `Подключен: ${c.gdrive.user_email}` : "Не авторизован";
  toggleBackendBoxes();
  document.getElementById("modalSettings").classList.remove("hidden");
}

function toggleBackendBoxes() {
  const v = document.getElementById("setBackend").value;
  document.getElementById("gdriveSettingsBox").classList.toggle("hidden", v !== "gdrive");
  document.getElementById("localSettingsBox").classList.toggle("hidden", v !== "local");
}

// Event bindings
document.getElementById("btnAddGame").onclick = openAddGameModal;
document.getElementById("btnEmptyAdd").onclick = openAddGameModal;
document.getElementById("btnSettings").onclick = openSettingsModal;
document.getElementById("btnEmptySettings").onclick = openSettingsModal;
document.getElementById("setBackend").onchange = toggleBackendBoxes;
document.querySelectorAll(".tab-btn").forEach((btn) => {
  btn.onclick = () => {
    document.querySelectorAll(".tab-btn").forEach((b) => b.classList.remove("active"));
    document.querySelectorAll(".tab-pane").forEach((p) => p.classList.add("hidden"));
    btn.classList.add("active");
    document.getElementById(btn.dataset.tab).classList.remove("hidden");
  };
});
document.getElementById("setTheme").onchange = (e) => applyTheme(e.target.value);
document.getElementById("setLanguage").onchange = (e) => {
  if (state.config) state.config.language = e.target.value;
  applyLanguage(e.target.value);
  renderGameList();
  renderDetail();
};
document.getElementById("btnRefreshAll").onclick = refreshAllStatuses;
document.getElementById("btnPush").onclick = () => pushGame(false);
document.getElementById("btnPull").onclick = () => pullGame(0, false);
document.getElementById("btnForcePush").onclick = () => pushGame(true);
document.getElementById("btnForcePull").onclick = () => pullGame(0, true);

document.getElementById("chkAutoSync").onchange = async (e) => {
  const id = state.selectedGameId;
  if (!id) return;
  await api(`/api/games/${encodeURIComponent(id)}`, "PATCH", { auto_sync: e.target.checked });
  await loadConfig();
};

document.getElementById("btnHostLock").onclick = async () => {
  const id = state.selectedGameId;
  if (!id) return;
  const st = state.statuses[id];
  const acquire = !(st?.host_lock && st.host_lock.holder === state.config.player_name);
  try {
    await api(`/api/games/${encodeURIComponent(id)}/lock`, "POST", { acquire });
    await refreshAllStatuses();
  } catch (e) {
    toast(e.message, true);
  }
};

document.getElementById("btnOpenFolder").onclick = async () => {
  const g = state.config.games.find((x) => x.id === state.selectedGameId);
  if (g) await api("/api/open-folder", "POST", { path: g.local_path });
};

document.getElementById("btnRemoveGame").onclick = async () => {
  const id = state.selectedGameId;
  if (!id || !confirm("Удалить игру из списка SyncSave? (Локальные файлы сохранений НЕ удаляются)")) return;
  await api(`/api/games/${encodeURIComponent(id)}`, "DELETE");
  state.selectedGameId = null;
  await loadConfig();
  await refreshAllStatuses();
};

document.getElementById("btnBrowseLocal").onclick = async () => {
  const res = await api("/api/browse-folder", "POST");
  if (res.path) document.getElementById("addLocalPath").value = res.path;
};

document.getElementById("btnBrowseSyncDir").onclick = async () => {
  const res = await api("/api/browse-folder", "POST");
  if (res.path) document.getElementById("setLocalSyncDir").value = res.path;
};

document.getElementById("btnPickRemoteWorld").onclick = async () => {
  try {
    const worlds = await api("/api/remote-worlds");
    const sel = document.getElementById("remoteWorldSelect");
    sel.innerHTML = `<option value="">-- Выберите мир из облака --</option>` +
      worlds.map((w) => `<option value="${escapeHtml(w)}">${escapeHtml(w)}</option>`).join("");
    sel.classList.remove("hidden");
    sel.onchange = () => {
      if (sel.value) document.getElementById("addRemoteFolder").value = sel.value;
    };
  } catch (e) {
    toast(e.message, true);
  }
};

document.getElementById("btnSaveNewGame").onclick = async () => {
  const name = document.getElementById("addGameName").value.trim();
  const local_path = document.getElementById("addLocalPath").value.trim();
  const remote_folder = document.getElementById("addRemoteFolder").value.trim();
  const auto_sync = document.getElementById("addAutoSync").checked;
  if (!name || !local_path || !remote_folder) {
    toast("Заполните название, локальный путь и имя мира", true);
    return;
  }
  try {
    const g = await api("/api/games", "POST", { name, local_path, remote_folder, auto_sync });
    state.selectedGameId = g.id;
    closeModals();
    await loadConfig();
    await refreshAllStatuses();
  } catch (e) {
    toast(e.message, true);
  }
};

document.getElementById("btnSaveSettings").onclick = async () => {
  try {
    await api("/api/config", "POST", {
      player_name: document.getElementById("setPlayerName").value.trim(),
      theme: document.getElementById("setTheme").value,
      language: document.getElementById("setLanguage").value,
      backend: document.getElementById("setBackend").value,
      local_sync_dir: document.getElementById("setLocalSyncDir").value.trim(),
      minimize_to_tray: document.getElementById("setMinimizeToTray").checked,
      start_minimized: document.getElementById("setStartMinimized").checked,
      autostart: document.getElementById("setAutostart").checked,
      tray_single_click: document.getElementById("setTraySingleClick").checked,
      tray_notifications: document.getElementById("setTrayNotifications").checked,
      memory_saver: document.getElementById("setMemorySaver").checked,
      auto_sync_paused: document.getElementById("setAutoSyncPaused").checked,
      auto_pull_on_start: document.getElementById("setAutoPullOnStart").checked,
      poll_interval_ms: parseInt(document.getElementById("setPollMs").value, 10) || 5000,
      debounce_seconds: parseInt(document.getElementById("setDebounceSec").value, 10) || 8,
      default_keep_backups: parseInt(document.getElementById("setKeepBackups").value, 10) || 5,
      gdrive_client_id: document.getElementById("setClientID").value.trim(),
      gdrive_client_secret: document.getElementById("setClientSecret").value.trim(),
      gdrive_folder_id: document.getElementById("setFolderID").value.trim(),
    });
    closeModals();
    await loadConfig();
    await refreshAllStatuses();
    toast("Настройки сохранены");
  } catch (e) {
    toast(e.message, true);
  }
};

document.getElementById("btnGoogleLogin").onclick = async () => {
  const client_id = document.getElementById("setClientID").value.trim();
  const client_secret = document.getElementById("setClientSecret").value.trim();
  if (!client_id || !client_secret) {
    toast("Сначала введите Client ID и Client Secret", true);
    return;
  }
  try {
    toast("Откроется браузер для входа в Google...");
    const res = await api("/api/gdrive/oauth", "POST", { client_id, client_secret });
    document.getElementById("oauthEmailBadge").textContent = res.email ? `Подключен: ${res.email}` : "Подключен";
    toast("Google Drive успешно подключен!");
    await loadConfig();
  } catch (e) {
    toast(e.message, true);
  }
};

document.getElementById("btnTestConn").onclick = async () => {
  await document.getElementById("btnSaveSettings").onclick();
  try {
    const res = await api("/api/backend/check");
    toast("Успешно: " + res.label);
  } catch (e) {
    toast(e.message, true);
  }
};

function escapeHtml(s) {
  return String(s || "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

init();
