package grummi

import "fmt"

var CurrentLanguage = "fr"

var translations = map[string]map[string]string{
	"en": {
		"menu_file":             "File",
		"menu_new_game":         "New Game",
		"menu_save":             "Save",
		"menu_quit":             "Quit",
		"menu_display":          "Display",
		"menu_language":         "Language",
		"status_lang_changed":   "Language changed to %s",
		"menu_theme_dark":       "Dark Theme",
		"menu_theme_light":      "Light Theme",
		"ai_thinking":           "🤖 %s is thinking...",
		"ai_merge":              "🤖 Merging two combinations on the table.",
		"ai_replace_joker":      "🤖 %s replaces a Joker on the table with %s.",
		"status_welcome":        "Welcome",
		"dialog_new_game_title": "New Game",
		"label_your_name":       "Your Name:",
		"label_ai_opponents":    "AI Opponents:",
		"btn_start":             "Start",
		"btn_cancel":            "Cancel",
	},
	"fr": {
		"menu_file":             "Fichier",
		"menu_new_game":         "Nouvelle Partie",
		"menu_save":             "Sauvegarder",
		"menu_quit":             "Quitter",
		"menu_display":          "Affichage",
		"menu_language":         "Langue",
		"status_lang_changed":   "Langue changée en %s",
		"menu_theme_dark":       "Thème Sombre",
		"menu_theme_light":      "Thème Clair",
		"ai_thinking":           "🤖 %s réfléchit...",
		"ai_merge":              "🤖 Fusion de deux combinaisons sur la table.",
		"ai_replace_joker":      "🤖 %s remplace un Joker sur la table par %s.",
		"status_welcome":        "Bienvenue",
		"dialog_new_game_title": "Nouvelle Partie",
		"label_your_name":       "Votre Nom :",
		"label_ai_opponents":    "Adversaires AI :",
		"btn_start":             "Démarrer",
		"btn_cancel":            "Annuler",
	},
}

// T returns the translated string for the given key.
// If the key is not found, it returns the key itself.
func T(key string, args ...interface{}) string {
	langMap, ok := translations[CurrentLanguage]
	if !ok {
		langMap = translations["en"]
	}

	val, ok := langMap[key]
	if !ok {
		// Fallback to English if key missing in current language
		val, ok = translations["en"][key]
		if !ok {
			return key
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

// SetLanguage updates the global language setting.
func SetLanguage(lang string) {
	if _, ok := translations[lang]; ok {
		CurrentLanguage = lang
	}
}
