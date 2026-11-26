const state = {
  config: null,
  presets: [],
  statuses: {},
  selectedGameId: null,
  logs: [],
};

const stateLabels = {
  synced: "Синхронизировано",
  local_ahead: "Есть локальные изменения",
  remote_ahead: "Доступно обновление",
  conflict: "Конфликт версий",
  not_synced: "Не синхронизировано",
  missing_local: "Папка не найдена",
  error: "Ошибка",
};

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
      <span class="status-dot st-${stCode}" title="${stateLabels[stCode] || stCode}"></span>
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
  document.getElementById("detailStateBadge").textContent = stateLabels[st.state] || st.state;
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
    lockInfo.textContent = `Хостит: ${st.host_lock.holder}`;
    lockBtn.textContent = st.host_lock.holder === state.config.player_name ? "Освободить хост" : "Перехватить";
  } else {
    lockInfo.textContent = "Хост свободен";
    lockBtn.textContent = "Взять хост";
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
      <td><button class="btn btn-secondary btn-xs" onclick="rollbackVersion(${v.version})">Откатить</button></td>
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
  document.getElementById("setMinimizeToTray").checked = !!c.minimize_to_tray;
  document.getElementById("setPollMs").value = c.poll_interval_ms || 5000;
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
      backend: document.getElementById("setBackend").value,
      local_sync_dir: document.getElementById("setLocalSyncDir").value.trim(),
      minimize_to_tray: document.getElementById("setMinimizeToTray").checked,
      poll_interval_ms: parseInt(document.getElementById("setPollMs").value, 10) || 5000,
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
