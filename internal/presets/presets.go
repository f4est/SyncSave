package presets

import (
	"os"
	"path/filepath"
	"strings"
)

// GamePreset describes a known co-op game and where its saves live on Windows.
type GamePreset struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	SteamAppID   string   `json:"steam_app_id,omitempty"`
	CoverURL     string   `json:"cover_url,omitempty"`
	PathPatterns []string `json:"path_patterns"`
	DetectedPath string   `json:"detected_path,omitempty"`
	Installed    bool     `json:"installed"`
}

// expandEnvVars expands %APPDATA%, %LOCALAPPDATA%, %USERPROFILE%, %DOCUMENTS%.
func expandEnvVars(pattern string) string {
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")
	userProfile := os.Getenv("USERPROFILE")
	localLow := filepath.Join(userProfile, "AppData", "LocalLow")
	docs := filepath.Join(userProfile, "Documents")

	r := strings.NewReplacer(
		"%APPDATA%", appData,
		"%LOCALAPPDATA%", localAppData,
		"%LOCALLOW%", localLow,
		"%USERPROFILE%", userProfile,
		"%DOCUMENTS%", docs,
	)
	return filepath.Clean(r.Replace(pattern))
}

func resolvePattern(pattern string) string {
	expanded := expandEnvVars(pattern)
	if !strings.ContainsAny(expanded, "*?") {
		if info, err := os.Stat(expanded); err == nil && info.IsDir() {
			return expanded
		}
		return ""
	}
	matches, err := filepath.Glob(expanded)
	if err != nil || len(matches) == 0 {
		return ""
	}
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && info.IsDir() {
			return m
		}
	}
	return ""
}

// BuiltinPresets returns the built-in SaveSync-style library of co-op games
// with auto-detection on the current machine.
func BuiltinPresets() []GamePreset {
	raw := []GamePreset{
		{
			ID:           "valheim",
			Name:         "Valheim",
			Category:     "Survival / Co-op",
			SteamAppID:   "892970",
			PathPatterns: []string{`%LOCALLOW%\IronGate\Valheim\worlds_local`, `%LOCALLOW%\IronGate\Valheim\worlds`},
		},
		{
			ID:           "terraria",
			Name:         "Terraria",
			Category:     "Sandbox / Co-op",
			SteamAppID:   "105600",
			PathPatterns: []string{`%DOCUMENTS%\My Games\Terraria\Worlds`},
		},
		{
			ID:           "palworld",
			Name:         "Palworld",
			Category:     "Survival / Co-op",
			SteamAppID:   "1623730",
			PathPatterns: []string{`%LOCALAPPDATA%\Pal\Saved\SaveGames\*`},
		},
		{
			ID:           "minecraft",
			Name:         "Minecraft (Java Edition)",
			Category:     "Sandbox / Co-op",
			CoverURL:     `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 460 215"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0%25" stop-color="%231e3a5f"/><stop offset="55%25" stop-color="%2322543d"/><stop offset="100%25" stop-color="%233b2f1e"/></linearGradient></defs><rect width="460" height="215" fill="url(%23g)"/><rect y="145" width="460" height="22" fill="%234ade80" opacity="0.85"/><rect y="167" width="460" height="48" fill="%236b4226" opacity="0.9"/><g transform="translate(36,42)"><rect width="68" height="68" rx="6" fill="%2322c55e"/><rect x="12" y="16" width="14" height="14" fill="%23111"/><rect x="42" y="16" width="14" height="14" fill="%23111"/><rect x="26" y="30" width="16" height="12" fill="%23111"/><rect x="18" y="38" width="10" height="16" fill="%23111"/><rect x="40" y="38" width="10" height="16" fill="%23111"/></g><text x="125" y="88" fill="%23ffffff" font-family="Segoe UI,sans-serif" font-weight="900" font-size="34" letter-spacing="1">MINECRAFT</text><text x="127" y="114" fill="%23bbf7d0" font-family="Segoe UI,sans-serif" font-weight="700" font-size="14" letter-spacing="2">JAVA EDITION</text></svg>`,
			PathPatterns: []string{`%APPDATA%\.minecraft\saves`},
		},
		{
			ID:           "stardew_valley",
			Name:         "Stardew Valley",
			Category:     "Farming / Co-op",
			SteamAppID:   "413150",
			PathPatterns: []string{`%APPDATA%\StardewValley\Saves`},
		},
		{
			ID:           "factorio",
			Name:         "Factorio",
			Category:     "Automation / Co-op",
			SteamAppID:   "427520",
			PathPatterns: []string{`%APPDATA%\Factorio\saves`},
		},
		{
			ID:           "7_days_to_die",
			Name:         "7 Days to Die",
			Category:     "Survival / Co-op",
			SteamAppID:   "251570",
			PathPatterns: []string{`%APPDATA%\7DaysToDie\Saves`},
		},
		{
			ID:           "raft",
			Name:         "Raft",
			Category:     "Survival / Co-op",
			SteamAppID:   "648800",
			PathPatterns: []string{`%LOCALLOW%\Redbeet Interactive\Raft\User\*\World`},
		},
		{
			ID:           "sons_of_the_forest",
			Name:         "Sons of the Forest",
			Category:     "Survival / Co-op",
			SteamAppID:   "1326470",
			PathPatterns: []string{`%LOCALLOW%\Endnight\SonsOfTheForest\Saves\*\Multiplayer`},
		},
		{
			ID:           "the_forest",
			Name:         "The Forest",
			Category:     "Survival / Co-op",
			SteamAppID:   "242760",
			PathPatterns: []string{`%LOCALLOW%\SKS\TheForest\*\Multiplayer`},
		},
		{
			ID:           "project_zomboid",
			Name:         "Project Zomboid",
			Category:     "Survival / Co-op",
			SteamAppID:   "108600",
			PathPatterns: []string{`%USERPROFILE%\Zomboid\Saves\Multiplayer`},
		},
		{
			ID:           "dont_starve_together",
			Name:         "Don't Starve Together",
			Category:     "Survival / Co-op",
			SteamAppID:   "322330",
			PathPatterns: []string{`%DOCUMENTS%\Klei\DoNotStarveTogether\*`},
		},
		{
			ID:           "satisfactory",
			Name:         "Satisfactory",
			Category:     "Automation / Co-op",
			SteamAppID:   "526870",
			PathPatterns: []string{`%LOCALAPPDATA%\FactoryGame\Saved\SaveGames\*`},
		},
		{
			ID:           "v_rising",
			Name:         "V Rising",
			Category:     "Survival / Co-op",
			SteamAppID:   "1604030",
			PathPatterns: []string{`%LOCALLOW%\Stunlock Studios\VRising\CloudSaves\*`, `%LOCALLOW%\Stunlock Studios\VRising\Saves\v3`},
		},
		{
			ID:           "grounded",
			Name:         "Grounded",
			Category:     "Survival / Co-op",
			SteamAppID:   "962130",
			PathPatterns: []string{`%USERPROFILE%\Saved Games\Grounded`},
		},
		{
			ID:           "core_keeper",
			Name:         "Core Keeper",
			Category:     "Sandbox / Co-op",
			SteamAppID:   "1621690",
			PathPatterns: []string{`%LOCALLOW%\Pugstorm\Core Keeper\Steam\*\worlds`},
		},
		{
			ID:           "astroneer",
			Name:         "Astroneer",
			Category:     "Sandbox / Co-op",
			SteamAppID:   "361420",
			PathPatterns: []string{`%LOCALAPPDATA%\Astro\Saved\SaveGames`},
		},
		{
			ID:           "deep_rock_galactic",
			Name:         "Deep Rock Galactic",
			Category:     "Co-op FPS",
			SteamAppID:   "548430",
			PathPatterns: []string{`C:\Program Files (x86)\Steam\steamapps\common\Deep Rock Galactic\FSD\Saved\SaveGames`},
		},
		{
			ID:           "vintage_story",
			Name:         "Vintage Story",
			Category:     "Sandbox / Co-op",
			CoverURL:     `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 460 215"><defs><linearGradient id="vg" x1="0" y1="0" x2="1" y2="1"><stop offset="0%25" stop-color="%2329190d"/><stop offset="55%25" stop-color="%235c3317"/><stop offset="100%25" stop-color="%2392400e"/></linearGradient></defs><rect width="460" height="215" fill="url(%23vg)"/><circle cx="385" cy="70" r="65" fill="%23f59e0b" opacity="0.18"/><circle cx="385" cy="70" r="34" fill="%23fbbf24" opacity="0.28"/><g transform="translate(40,56)"><rect x="14" y="10" width="36" height="52" rx="6" fill="none" stroke="%23fbbf24" stroke-width="5"/><circle cx="32" cy="36" r="10" fill="%23f59e0b"/></g><text x="115" y="100" fill="%23fef3c7" font-family="Georgia,serif" font-weight="900" font-size="34">VINTAGE STORY</text><text x="117" y="128" fill="%23fcd34d" font-family="Segoe UI,sans-serif" font-weight="600" font-size="13" letter-spacing="2">SURVIVAL SANDBOX</text></svg>`,
			PathPatterns: []string{`%APPDATA%\VintagestoryData\Saves`},
		},
		{
			ID:           "enshrouded",
			Name:         "Enshrouded",
			Category:     "Survival / Co-op",
			SteamAppID:   "1203620",
			PathPatterns: []string{`%USERPROFILE%\Saved Games\Enshrouded`, `C:\Program Files (x86)\Steam\userdata\*\1203620\remote`},
		},
		{
			ID:           "baldurs_gate_3",
			Name:         "Baldur's Gate 3",
			Category:     "RPG / Co-op",
			SteamAppID:   "1086940",
			PathPatterns: []string{`%LOCALAPPDATA%\Larian Studios\Baldur's Gate 3\PlayerProfiles\Public\Savegames\Story`},
		},
		{
			ID:           "divinity_os2",
			Name:         "Divinity: Original Sin 2",
			Category:     "RPG / Co-op",
			SteamAppID:   "435150",
			PathPatterns: []string{`%DOCUMENTS%\Larian Studios\Divinity Original Sin 2 Definitive Edition\PlayerProfiles\*\Savegames\Story`},
		},
		{
			ID:           "subnautica",
			Name:         "Subnautica (Nitrox)",
			Category:     "Survival / Co-op",
			SteamAppID:   "264710",
			PathPatterns: []string{`%APPDATA%\Nitrox\world`},
		},
		{
			ID:           "schedule_1",
			Name:         "Schedule I",
			Category:     "Simulation / Co-op",
			SteamAppID:   "3164500",
			PathPatterns: []string{`%LOCALLOW%\TVGS\Schedule I\Saves\*`},
		},
		{
			ID:           "abiotic_factor",
			Name:         "Abiotic Factor",
			Category:     "Survival / Co-op",
			SteamAppID:   "427410",
			PathPatterns: []string{`%LOCALAPPDATA%\AbioticFactor\Saved\SaveGames\*\Worlds`},
		},
		{
			ID:           "icarus",
			Name:         "Icarus",
			Category:     "Survival / Co-op",
			SteamAppID:   "1149460",
			PathPatterns: []string{`%LOCALAPPDATA%\Icarus\Saved\PlayerData\*\Prospects`},
		},
		{
			ID:           "custom",
			Name:         "Другая игра (вручную)",
			Category:     "Custom",
			PathPatterns: []string{},
		},
	}

	for i := range raw {
		if raw[i].SteamAppID != "" && raw[i].CoverURL == "" {
			raw[i].CoverURL = "https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/" + raw[i].SteamAppID + "/header.jpg"
		}
		for _, pat := range raw[i].PathPatterns {
			if found := resolvePattern(pat); found != "" {
				raw[i].DetectedPath = found
				raw[i].Installed = true
				break
			}
		}
	}
	return raw
}
