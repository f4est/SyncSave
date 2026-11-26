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
