package config

import (
	"encoding/json"
	"os"

	"github.com/disgoorg/disgo/discord"
)

type command[Options any] struct {
	Description map[discord.Locale]string
	Options     Options
}

type Config struct {
	Token string
	Redis struct {
		Addr     string
		Password string
		DB       int
	}
	Localization struct {
		Commands struct {
			Config command[struct {
				Bypass    map[discord.Locale]string
				Ephemeral map[discord.Locale]string
			}]
			Captcha command[struct{}]
			Submit  command[struct {
				Answer map[discord.Locale]string
			}]
		}
		Messages struct {
			BypassRoleRequired map[discord.Locale]string
			SubmissionRequired map[discord.Locale]string
			CaptchaRequired    map[discord.Locale]string
			Reconfigured       map[discord.Locale]string
			BypassDenied       map[discord.Locale]string
			Bypassed           map[discord.Locale]string
		}
	}
}

func FromStdin(cfg *Config) *Config {
	json.NewDecoder(os.Stdin).Decode(&cfg)
	return cfg
}
