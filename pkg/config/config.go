package config

import (
	"encoding/json"
	"os"

	"github.com/disgoorg/disgo/discord"
)

type command[T any] struct {
	Description map[discord.Locale]string `json:"description"`
	Options     T                         `json:"options"`
}

type Config struct {
	Token string `json:"token"`
	Redis struct {
		Addr     string `json:"address"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
	Localization struct {
		Commands struct {
			Config command[struct {
				Bypass    map[discord.Locale]string `json:"bypass"`
				Ephemeral map[discord.Locale]string `json:"ephemeral"`
			}] `json:"config"`
			Captcha command[struct{}] `json:"captcha"`
			Submit  command[struct {
				Answer map[discord.Locale]string `json:"answer"`
			}] `json:"submit"`
		} `json:"commands"`
		Messages struct {
			PermissionsMissed     map[discord.Locale]string `json:"permissions-missed"`
			SubmissionRequired    map[discord.Locale]string `json:"submission-required"`
			ConfigurationUpdated  map[discord.Locale]string `json:"configuration-updated"`
			CaptchaRequired       map[discord.Locale]string `json:"captcha-required"`
			VerificationFailed    map[discord.Locale]string `json:"verification-failed"`
			BypassMissed          map[discord.Locale]string `json:"bypass-missed"`
			VerificationSuccessed map[discord.Locale]string `json:"verification-successed"`
		}
	} `json:"localization"`
}

func FromStdin(cfg *Config) *Config {
	json.NewDecoder(os.Stdin).Decode(&cfg)
	return cfg
}
