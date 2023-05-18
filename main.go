package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/log"
	"github.com/mandriota/gatewarden-bot/pkg/config"
	"github.com/mandriota/gatewarden-bot/pkg/listener"
)

func newCommandCreate(cfg *config.Config) []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "config",
			Description: "configure bot",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionRole{
					Name:                     "bypass",
					Description:              "set bypass role",
					DescriptionLocalizations: cfg.Localization.Commands.Config.Options.Bypass,
				},
				discord.ApplicationCommandOptionBool{
					Name:                     "ephemeral",
					Description:              "set bot messages to be ephemeral",
					DescriptionLocalizations: cfg.Localization.Commands.Config.Options.Ephemeral,
				},
			},
			DescriptionLocalizations: cfg.Localization.Commands.Config.Description,
		},
		discord.SlashCommandCreate{
			Name:                     "captcha",
			Description:              "generates new captcha",
			DescriptionLocalizations: cfg.Localization.Commands.Captcha.Description,
		},
		discord.SlashCommandCreate{
			Name:        "submit",
			Description: "submit captcha solution",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:                     "answer",
					Description:              "captcha solution",
					Required:                 true,
					DescriptionLocalizations: cfg.Localization.Commands.Submit.Options.Answer,
				},
			},
			DescriptionLocalizations: cfg.Localization.Commands.Submit.Description,
		},
	}
}

func main() {
	log.SetLevel(log.LevelDebug)

	ctx := context.Background()

	cfg := config.FromStdin(&config.Config{})

	client, err := disgo.New(cfg.Token,
		bot.WithDefaultGateway(),
		bot.WithEventListenerFunc(listener.New(ctx, cfg)),
	)
	if err != nil {
		log.Fatal("error while building disgo instance: ", err)
	}
	defer client.Close(ctx)

	if _, err := client.Rest().SetGlobalCommands(client.ApplicationID(), newCommandCreate(cfg)); err != nil {
		log.Fatal("error while registering commands: ", err)
	}

	if err := client.OpenGateway(ctx); err != nil {
		log.Fatal("error while connecting to gateway: ", err)
	}

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	<-ctx.Done()
}
