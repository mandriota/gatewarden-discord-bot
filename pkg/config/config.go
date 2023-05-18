package config

import (
	"encoding/json"
	"os"

	"github.com/disgoorg/disgo/discord"
)

type Config struct {
	Token string `json:"token"`
	Redis struct {
		Addr     string `json:"address"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
	Localizations struct {
		Config struct {
			Description map[discord.Locale]string `json:"description"`
			Options     struct {
				Bypass    map[discord.Locale]string `json:"bypass"`
				Ephemeral map[discord.Locale]string `json:"ephemeral"`
			} `json:"options"`
		} `json:"config"`
		Captcha struct {
			Description map[discord.Locale]string `json:"description"`
		} `json:"captcha"`
		Submit struct {
			Description map[discord.Locale]string `json:"description"`
			Options     struct {
				Answer map[discord.Locale]string `json:"answer"`
			} `json:"options"`
		} `json:"submit"`
	} `json:"localizations"`
}

func FromStdin(cfg *Config) *Config {
	json.NewDecoder(os.Stdin).Decode(&cfg)
	return cfg
}
