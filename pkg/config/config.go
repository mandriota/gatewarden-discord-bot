package config

import (
	"encoding/json"
	"os"

	"github.com/disgoorg/disgo/discord"
)

type Command[Opts any, Responses any] struct {
	Name                     string
	Description              string
	DescriptionLocalizations map[discord.Locale]string
	Options                  Opts
	Responses                Responses
}

type Option struct {
	Name                     string
	Description              string
	DescriptionLocalizations map[discord.Locale]string
}

type Response struct {
	TitleLocalizations       map[discord.Locale]string
	DescriptionLocalizations map[discord.Locale]string
	FooterLocalizations      map[discord.Locale]string
}

type Config struct {
	Token string
	Redis struct {
		Addr     string
		Password string
		DB       int
	}
	Commands struct {
		Config Command[
			struct {
				Bypass    Option
				Ephemeral Option
			},
			struct{ Success Response },
		]
		Captcha Command[
			struct{},
			struct{ Success Response },
		]
		Submit Command[
			struct{ Answer Option },
			struct {
				Success                         Response
				ErrWrongAnswer                  Response
				ErrUndefinedCaptcha             Response
				ErrUndefinedBypassRole          Response
				ErrMissingManageRolesPermission Response
			},
		]
	}
}

func FromStdin(cfg *Config) *Config {
	json.NewDecoder(os.Stdin).Decode(&cfg)
	return cfg
}
