const state = {
  config: null,
  presets: [],
  statuses: {},
  selectedGameId: null,
  viewMode: "home", // "home" or "detail"
  libraryFilter: "all",
  searchQuery: "",
  selectedPreset: null,
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

const GAME_ART = {
  valheim: { g: "linear-gradient(135deg,#2563eb,#0f172a)", icon: "⚔️", glow: "rgba(37,99,235,0.22)" },
  terraria: { g: "linear-gradient(135deg,#10b981,#065f46)", icon: "🌳", glow: "rgba(16,185,129,0.22)" },
  palworld: { g: "linear-gradient(135deg,#06b6d4,#3b82f6)", icon: "🐾", glow: "rgba(6,182,212,0.22)" },
  minecraft: { g: "linear-gradient(135deg,#22c55e,#78350f)", icon: "🧱", glow: "rgba(34,197,94,0.20)" },
  stardew: { g: "linear-gradient(135deg,#f59e0b,#10b981)", icon: "🌾", glow: "rgba(245,158,11,0.20)" },
  factorio: { g: "linear-gradient(135deg,#ea580c,#78350f)", icon: "⚙️", glow: "rgba(234,88,12,0.22)" },
  "7 days": { g: "linear-gradient(135deg,#dc2626,#1f2937)", icon: "🧟", glow: "rgba(220,38,38,0.22)" },
  raft: { g: "linear-gradient(135deg,#0284c7,#14b8a6)", icon: "⛵", glow: "rgba(2,132,199,0.22)" },
  forest: { g: "linear-gradient(135deg,#15803d,#1e293b)", icon: "🌲", glow: "rgba(21,128,61,0.22)" },
  zomboid: { g: "linear-gradient(135deg,#991b1b,#111827)", icon: "🪓", glow: "rgba(153,27,27,0.22)" },
  starve: { g: "linear-gradient(135deg,#7c3aed,#1e1b4b)", icon: "🔥", glow: "rgba(124,58,237,0.22)" },
  satisfactory: { g: "linear-gradient(135deg,#f97316,#1e293b)", icon: "🏭", glow: "rgba(249,115,22,0.22)" },
  rising: { g: "linear-gradient(135deg,#be123c,#31102f)", icon: "🦇", glow: "rgba(190,18,60,0.22)" },
  grounded: { g: "linear-gradient(135deg,#84cc16,#0d9488)", icon: "🐜", glow: "rgba(132,204,22,0.20)" },
  core: { g: "linear-gradient(135deg,#6366f1,#0ea5e9)", icon: "💎", glow: "rgba(99,102,241,0.22)" },
  astroneer: { g: "linear-gradient(135deg,#38bdf8,#6366f1)", icon: "🚀", glow: "rgba(56,189,248,0.22)" },
  deep: { g: "linear-gradient(135deg,#eab308,#713f12)", icon: "⛏️", glow: "rgba(234,179,8,0.20)" },
  vintage: { g: "linear-gradient(135deg,#d97706,#451a03)", icon: "🏺", glow: "rgba(217,119,6,0.20)" },
  enshrouded: { g: "linear-gradient(135deg,#0ea5e9,#1e1b4b)", icon: "🛡️", glow: "rgba(14,165,233,0.22)" },
  baldur: { g: "linear-gradient(135deg,#b45309,#450a0a)", icon: "🐉", glow: "rgba(180,83,9,0.22)" },
  divinity: { g: "linear-gradient(135deg,#7c3aed,#9f1239)", icon: "✨", glow: "rgba(124,58,237,0.22)" },
  subnautica: { g: "linear-gradient(135deg,#06b6d4,#0f172a)", icon: "🌊", glow: "rgba(6,182,212,0.22)" },
};

const PALETTES = [
  { g: "linear-gradient(135deg,#3b82f6,#8b5cf6)", glow: "rgba(59,130,246,0.20)" },
  { g: "linear-gradient(135deg,#10b981,#06b6d4)", glow: "rgba(16,185,129,0.20)" },
  { g: "linear-gradient(135deg,#f43f5e,#fb923c)", glow: "rgba(244,63,94,0.20)" },
  { g: "linear-gradient(135deg,#8b5cf6,#ec4899)", glow: "rgba(139,92,246,0.20)" },
];

function getGameArt(name) {
  const lower = String(name || "").toLowerCase();
  for (const [key, val] of Object.entries(GAME_ART)) {
    if (lower.includes(key)) return val;
  }
  let hash = 0;
  for (let i = 0; i < lower.length; i++) hash = (hash * 31 + lower.charCodeAt(i)) | 0;
  const p = PALETTES[Math.abs(hash) % PALETTES.length];
  return { g: p.g, glow: p.glow, icon: (name || "G").trim().slice(0, 2).toUpperCase() };
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
  checkBackendPill();
  renderGameList();
  renderLibraryGrid();
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
  renderLibraryGrid();
}

window.setLibraryFilter = (f) => {
  state.libraryFilter = f;
  document.querySelectorAll(".filter-pill").forEach((el) => el.classList.remove("active"));
  if (f === "all") document.getElementById("filterAll").classList.add("active");
  if (f === "installed") document.getElementById("filterInstalled").classList.add("active");
  if (f === "added") document.getElementById("filterAdded").classList.add("active");
  renderLibraryGrid();
};

function showHome() {
  state.viewMode = "home";
  state.selectedGameId = null;
  document.getElementById("btnNavHome").classList.add("active");
  document.getElementById("homeView").classList.remove("hidden");
  document.getElementById("gameDetail").classList.add("hidden");
  renderGameList();
  renderLibraryGrid();
}

function renderLibraryGrid() {
  const grid = document.getElementById("libraryGrid");
  if (!grid || !state.presets) return;
  const games = state.config?.games || [];
  const addedByName = new Map(games.map((g) => [g.name.toLowerCase(), g]));
  const q = state.searchQuery.trim().toLowerCase();

  let items = state.presets.filter((p) => p.id !== "custom");
  if (q) items = items.filter((p) => p.name.toLowerCase().includes(q) || p.category.toLowerCase().includes(q));
  if (state.libraryFilter === "installed") items = items.filter((p) => p.installed);
  if (state.libraryFilter === "added") items = items.filter((p) => addedByName.has(p.name.toLowerCase()));

  document.getElementById("librarySubtitle").textContent = `${items.length} игр • Нажмите на карточку, чтобы подключить или открыть мир`;
  grid.innerHTML = "";

  for (const p of items) {
    const added = addedByName.get(p.name.toLowerCase());
    const art = getGameArt(p.name);
    const card = document.createElement("div");
    card.className = "lib-card";
    card.onclick = () => {
      if (added) {
        state.selectedGameId = added.id;
        state.viewMode = "detail";
        document.getElementById("btnNavHome").classList.remove("active");
        renderGameList();
        renderDetail();
      } else {
        openAddGameModal(p);
      }
    };
    const badge = added
      ? `<span class="lib-badge synced">Подключена</span>`
      : p.installed
      ? `<span class="lib-badge">Найдена на ПК</span>`
      : "";
    const coverHtml = p.cover_url
      ? `<img src="${escapeHtml(p.cover_url)}" loading="lazy" alt="${escapeHtml(p.name)}" onerror="this.style.display='none'" />`
      : "";
    card.innerHTML = `
      <div class="lib-cover" style="background:${art.g}">
        <span class="lib-cover-fallback">${art.icon}</span>
        ${coverHtml}
        ${badge}
      </div>
      <div class="lib-card-body">
        <div style="min-width:0">
          <div class="lib-card-title">${escapeHtml(p.name)}</div>
          <div class="lib-card-cat">${escapeHtml(p.category)}</div>
        </div>
      </div>
    `;
    grid.appendChild(card);
  }
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
  const q = state.searchQuery.trim().toLowerCase();

  if (state.viewMode === "home" || !state.selectedGameId) {
    document.getElementById("homeView").classList.remove("hidden");
    document.getElementById("gameDetail").classList.add("hidden");
  } else {
    document.getElementById("homeView").classList.add("hidden");
    document.getElementById("gameDetail").classList.remove("hidden");
  }

  const sortedGames = [...games].sort((a, b) => (b.favorite ? 1 : 0) - (a.favorite ? 1 : 0));
  for (const g of sortedGames) {
    if (q && !g.name.toLowerCase().includes(q) && !g.remote_folder.toLowerCase().includes(q)) continue;
    const st = state.statuses[g.id];
    const stCode = st?.state || "not_synced";
    const div = document.createElement("div");
    div.className = "game-item" + (state.viewMode === "detail" && g.id === state.selectedGameId ? " active" : "");
    div.onclick = () => {
      state.selectedGameId = g.id;
      state.viewMode = "detail";
      document.getElementById("btnNavHome").classList.remove("active");
      renderGameList();
      renderDetail();
    };
    const art = getGameArt(g.name);
    const thumb = g.cover_url
      ? `<img src="${escapeHtml(g.cover_url)}" style="width:100%;height:100%;object-fit:cover;border-radius:7px" onerror="this.remove()" />`
      : art.icon;
    div.innerHTML = `
      <div class="game-item-left">
        <div class="game-art-sm" style="background:${art.g};overflow:hidden">${thumb}</div>
        <div>
          <div class="game-item-name">${g.favorite ? "★ " : ""}${escapeHtml(g.name)}</div>
          <div class="game-item-sub">${escapeHtml(g.remote_folder)} • v${st?.remote_version ?? g.last_synced_ver ?? 0}</div>
        </div>
      </div>
      <span class="status-dot st-${stCode}" title="${stateLabel(stCode)}"></span>
    `;
    list.appendChild(div);
  }
}

function renderDetail() {
  const id = state.selectedGameId;
  if (!id || !state.config || state.viewMode !== "detail") return;
  const g = state.config.games.find((x) => x.id === id);
  if (!g) return;
  document.getElementById("homeView").classList.add("hidden");
  document.getElementById("gameDetail").classList.remove("hidden");
  const st = state.statuses[id] || { game: g, state: "not_synced", versions: [] };

  const art = getGameArt(g.name);
  const iconEl = document.getElementById("detailGameIcon");
  iconEl.style.background = art.g;
  iconEl.innerHTML = g.cover_url
    ? `<img src="${escapeHtml(g.cover_url)}" style="width:100%;height:100%;object-fit:cover;border-radius:9px;display:block" onerror="this.remove()" />`
    : art.icon;
  const heroBg = document.getElementById("worldHeroBg");
  heroBg.style.backgroundImage = g.cover_url ? `url("${g.cover_url}")` : "none";
  document.getElementById("worldCard").style.background = g.cover_url
    ? "var(--panel)"
    : `radial-gradient(circle at top left, ${art.glow}, transparent 55%), var(--panel)`;
  document.getElementById("detailName").textContent = g.name;
  document.getElementById("btnPlayGame").classList.toggle("hidden", !g.steam_app_id);
  document.getElementById("btnFavGame").textContent = g.favorite ? "★ В избранном" : "☆ Избранное";
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
  document.getElementById("modalShare").classList.add("hidden");
  document.getElementById("modalJoin").classList.add("hidden");
  document.getElementById("modalHistory").classList.add("hidden");
};

function openAddGameModal(preselectPreset = null) {
  const grid = document.getElementById("presetGrid");
  grid.innerHTML = "";
  state.selectedPreset = preselectPreset;
  if (!preselectPreset) {
    document.getElementById("addGameName").value = "";
    document.getElementById("addRemoteFolder").value = "";
    document.getElementById("addLocalPath").value = "";
  }
  for (const p of state.presets) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "preset-chip" + (preselectPreset && preselectPreset.id === p.id ? " selected" : "");
    const art = getGameArt(p.name);
    const iconHtml = p.cover_url
      ? `<img src="${escapeHtml(p.cover_url)}" style="width:100%;height:100%;object-fit:cover;display:block" onerror="this.remove()" />`
      : art.icon;
    btn.innerHTML = `
      <div class="game-art-sm" style="background:${art.g};width:44px;height:24px;font-size:12px">${iconHtml}</div>
      <div style="min-width:0">
        <div style="white-space:nowrap;overflow:hidden;text-overflow:ellipsis">${escapeHtml(p.name)}</div>
        ${p.installed ? `<span class="installed">✓ Найдена на ПК</span>` : ""}
      </div>
    `;
    const applyPreset = () => {
      document.querySelectorAll(".preset-chip").forEach((c) => c.classList.remove("selected"));
      btn.classList.add("selected");
      state.selectedPreset = p;
      if (p.id !== "custom") {
        document.getElementById("addGameName").value = p.name;
        document.getElementById("addRemoteFolder").value = p.name.replace(/[^a-zA-Z0-9_-]/g, "_") + "_World";
      }
      if (p.detected_path) {
        document.getElementById("addLocalPath").value = p.detected_path;
      } else {
        document.getElementById("addLocalPath").value = "";
      }
    };
    btn.onclick = applyPreset;
    grid.appendChild(btn);
    if (preselectPreset && preselectPreset.id === p.id) {
      applyPreset();
    }
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
  document.getElementById("setCustomSyncDir").value = c.local_sync_dir || "";
  document.getElementById("setWebdavURL").value = c.webdav?.url || "";
  document.getElementById("setWebdavUser").value = c.webdav?.username || "";
  document.getElementById("setWebdavPass").value = c.webdav?.password || "";
  document.getElementById("setWebdavBasePath").value = c.webdav?.base_path || "SyncSave";
  document.getElementById("setS3Endpoint").value = c.s3?.endpoint || "";
  document.getElementById("setS3Region").value = c.s3?.region || "auto";
  document.getElementById("setS3Bucket").value = c.s3?.bucket || "";
  document.getElementById("setS3Prefix").value = c.s3?.prefix || "";
  document.getElementById("setS3Access").value = c.s3?.access_key || "";
  document.getElementById("setS3Secret").value = c.s3?.secret_key || "";
  document.getElementById("setGitRepo").value = c.git?.repo_url || "";
  document.getElementById("setGitBranch").value = c.git?.branch || "main";
  document.getElementById("setLanHost").checked = !!c.lan?.host_enabled;
  document.getElementById("setLanPort").value = c.lan?.port || 47821;
  document.getElementById("setLanPin").value = c.lan?.room_pin || "";
  document.getElementById("setLanPeer").value = c.lan?.peer_url || "";
  document.getElementById("setCustomPullCmd").value = c.custom_sync?.pull_cmd || "";
  document.getElementById("setCustomPushCmd").value = c.custom_sync?.push_cmd || "";
  refreshLanIps();
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
  document.getElementById("webdavSettingsBox").classList.toggle("hidden", v !== "webdav");
  document.getElementById("s3SettingsBox").classList.toggle("hidden", v !== "s3");
  document.getElementById("gitSettingsBox").classList.toggle("hidden", v !== "git");
  document.getElementById("lanSettingsBox").classList.toggle("hidden", v !== "lan");
  document.getElementById("localSettingsBox").classList.toggle("hidden", v !== "local");
  document.getElementById("customSettingsBox").classList.toggle("hidden", v !== "custom");
}

async function refreshLanIps() {
  try {
    const info = await api("/api/lan/info");
    const el = document.getElementById("lanIpsHint");
    if (info.ips && info.ips.length > 0) {
      el.textContent = "Ваши адреса для друзей: " + info.ips.map((ip) => `http://${ip}:${info.port}`).join(" • ");
    } else {
      el.textContent = `Порт: ${info.port}`;
    }
  } catch (_) {}
}

window.fillWebDAVPreset = (kind) => {
  if (kind === "yandex") {
    document.getElementById("setWebdavURL").value = "https://webdav.yandex.ru";
    document.getElementById("setWebdavBasePath").value = "SyncSave";
  } else if (kind === "nextcloud") {
    document.getElementById("setWebdavURL").value = "https://your-server.com/remote.php/dav/files/USERNAME/";
    document.getElementById("setWebdavBasePath").value = "SyncSave";
  }
};

// Event bindings
document.getElementById("btnNavHome").onclick = showHome;
document.getElementById("sidebarSearch").oninput = (e) => {
  state.searchQuery = e.target.value;
  renderGameList();
  renderLibraryGrid();
};
document.getElementById("btnAddGame").onclick = () => openAddGameModal(null);
document.getElementById("btnEmptyAdd").onclick = () => openAddGameModal(null);
document.getElementById("btnJoinCode").onclick = () => {
  document.getElementById("joinCodeInput").value = "";
  document.getElementById("joinLocalPath").value = "";
  document.getElementById("modalJoin").classList.remove("hidden");
};
document.getElementById("btnBrowseJoinPath").onclick = async () => {
  const res = await api("/api/browse-folder", "POST");
  if (res.path) document.getElementById("joinLocalPath").value = res.path;
};
document.getElementById("btnConfirmJoin").onclick = async () => {
  const code = document.getElementById("joinCodeInput").value.trim();
  const local_path = document.getElementById("joinLocalPath").value.trim();
  if (!code) {
    toast("Вставьте код приглашения", true);
    return;
  }
  try {
    const g = await api("/api/join", "POST", { code, local_path });
    state.selectedGameId = g.id;
    state.viewMode = "detail";
    document.getElementById("btnNavHome").classList.remove("active");
    closeModals();
    await loadConfig();
    await refreshAllStatuses();
    toast(`Вы присоединились к миру "${g.name}"!`);
  } catch (e) {
    toast(e.message, true);
  }
};

async function generateShareCode() {
  const id = state.selectedGameId;
  if (!id) return;
  const include_secrets = document.getElementById("chkShareSecrets").checked;
  const res = await api(`/api/games/${encodeURIComponent(id)}/share`, "POST", { include_secrets });
  document.getElementById("shareCodeText").value = res.code;
}

document.getElementById("btnShareWorld").onclick = async () => {
  try {
    await generateShareCode();
    document.getElementById("modalShare").classList.remove("hidden");
  } catch (e) {
    toast(e.message, true);
  }
};
document.getElementById("chkShareSecrets").onchange = generateShareCode;
document.getElementById("btnPlayGame").onclick = async () => {
  const id = state.selectedGameId;
  if (!id) return;
  try {
    await api(`/api/games/${encodeURIComponent(id)}/launch`, "POST");
    toast("Запуск игры через Steam...");
  } catch (e) {
    toast(e.message, true);
  }
};

document.getElementById("btnFavGame").onclick = async () => {
  const id = state.selectedGameId;
  const g = state.config?.games.find((x) => x.id === id);
  if (!g) return;
  await api(`/api/games/${encodeURIComponent(id)}`, "PATCH", { favorite: !g.favorite });
  await loadConfig();
  renderDetail();
};

document.getElementById("btnOpenHistory").onclick = () => {
  const id = state.selectedGameId;
  const g = state.config?.games.find((x) => x.id === id);
  const st = state.statuses[id];
  if (!g || !st) return;
  document.getElementById("historyModalTitle").textContent = `История — ${g.remote_folder}`;
  const box = document.getElementById("historyModalList");
  const versions = st.versions || [];
  if (versions.length === 0) {
    box.innerHTML = `<div class="muted">Пока нет загруженных версий.</div>`;
  } else {
    box.innerHTML = versions.map((v, idx) => `
      <div class="hist-card">
        <div>
          <div style="font-weight:700">#v${v.version} • ${escapeHtml(v.author)}</div>
          <div class="muted" style="font-size:11.5px">${escapeHtml(v.timestamp)} • ${(v.zip_size / (1024 * 1024)).toFixed(2)} МБ${v.note ? " • " + escapeHtml(v.note) : ""}</div>
        </div>
        ${idx === 0
          ? `<span class="hist-pill-cur">Текущая версия</span>`
          : `<button class="btn btn-secondary btn-xs" onclick="rollbackVersion(${v.version}); closeModals();">Восстановить</button>`}
      </div>
    `).join("");
  }
  document.getElementById("modalHistory").classList.remove("hidden");
};

document.getElementById("btnLanPopover").onclick = async (e) => {
  e.stopPropagation();
  const pop = document.getElementById("lanPopover");
  pop.classList.toggle("hidden");
  if (!pop.classList.contains("hidden")) {
    const info = await api("/api/lan/info");
    const ip = info.ips && info.ips.length > 0 ? `http://${info.ips[0]}:${info.port}` : `127.0.0.1:${info.port}`;
    document.getElementById("lanPopIp").value = ip;
    const btn = document.getElementById("btnToggleLanHost");
    btn.textContent = info.host_enabled ? "Отключить LAN Sync" : "Включить LAN Sync";
    btn.className = info.host_enabled ? "btn btn-danger-ghost btn-sm" : "btn btn-primary btn-sm";
    btn.dataset.enabled = info.host_enabled ? "1" : "0";
  }
};
document.getElementById("lanPopover").onclick = (e) => e.stopPropagation();
document.addEventListener("click", () => document.getElementById("lanPopover")?.classList.add("hidden"));
document.getElementById("btnCopyLanIp").onclick = async () => {
  await navigator.clipboard.writeText(document.getElementById("lanPopIp").value);
  toast("LAN/VPN адрес скопирован!");
};
document.getElementById("btnToggleLanHost").onclick = async () => {
  const btn = document.getElementById("btnToggleLanHost");
  const next = btn.dataset.enabled !== "1";
  const info = await api("/api/lan/toggle", "POST", { enabled: next });
  btn.textContent = info.host_enabled ? "Отключить LAN Sync" : "Включить LAN Sync";
  btn.className = info.host_enabled ? "btn btn-danger-ghost btn-sm" : "btn btn-primary btn-sm";
  btn.dataset.enabled = info.host_enabled ? "1" : "0";
  await loadConfig();
  toast(info.host_enabled ? "LAN Sync сервер запущен!" : "LAN Sync сервер остановлен");
};

document.getElementById("btnCopyShareCode").onclick = async () => {
  const txt = document.getElementById("shareCodeText").value;
  await navigator.clipboard.writeText(txt);
  toast("Код приглашения скопирован! Отправьте его другу.");
};
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

document.getElementById("btnBrowseCustomDir").onclick = async () => {
  const res = await api("/api/browse-folder", "POST");
  if (res.path) document.getElementById("setCustomSyncDir").value = res.path;
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
    const g = await api("/api/games", "POST", {
      name,
      local_path,
      remote_folder,
      auto_sync,
      preset_id: state.selectedPreset?.id || "",
      steam_app_id: state.selectedPreset?.steam_app_id || "",
      cover_url: state.selectedPreset?.cover_url || "",
    });
    state.selectedGameId = g.id;
    state.viewMode = "detail";
    document.getElementById("btnNavHome").classList.remove("active");
    closeModals();
    await loadConfig();
    await refreshAllStatuses();
  } catch (e) {
    toast(e.message, true);
  }
};

document.getElementById("btnSaveSettings").onclick = async () => {
  try {
    const backendVal = document.getElementById("setBackend").value;
    const syncDirVal = backendVal === "custom"
      ? document.getElementById("setCustomSyncDir").value.trim()
      : document.getElementById("setLocalSyncDir").value.trim();
    await api("/api/config", "POST", {
      player_name: document.getElementById("setPlayerName").value.trim(),
      theme: document.getElementById("setTheme").value,
      language: document.getElementById("setLanguage").value,
      backend: backendVal,
      local_sync_dir: syncDirVal,
      webdav_url: document.getElementById("setWebdavURL").value.trim(),
      webdav_username: document.getElementById("setWebdavUser").value.trim(),
      webdav_password: document.getElementById("setWebdavPass").value,
      webdav_base_path: document.getElementById("setWebdavBasePath").value.trim(),
      s3_endpoint: document.getElementById("setS3Endpoint").value.trim(),
      s3_region: document.getElementById("setS3Region").value.trim(),
      s3_bucket: document.getElementById("setS3Bucket").value.trim(),
      s3_prefix: document.getElementById("setS3Prefix").value.trim(),
      s3_access_key: document.getElementById("setS3Access").value.trim(),
      s3_secret_key: document.getElementById("setS3Secret").value,
      git_repo_url: document.getElementById("setGitRepo").value.trim(),
      git_branch: document.getElementById("setGitBranch").value.trim(),
      lan_host_enabled: document.getElementById("setLanHost").checked,
      lan_port: parseInt(document.getElementById("setLanPort").value, 10) || 47821,
      lan_room_pin: document.getElementById("setLanPin").value.trim(),
      lan_peer_url: document.getElementById("setLanPeer").value.trim(),
      custom_pull_cmd: document.getElementById("setCustomPullCmd").value.trim(),
      custom_push_cmd: document.getElementById("setCustomPushCmd").value.trim(),
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
