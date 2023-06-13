package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/json"
	"github.com/disgoorg/log"
	"github.com/mandriota/gatewarden-bot/pkg/config"
	"github.com/mandriota/gatewarden-bot/pkg/listener"
)

func newCommandCreate(cfg *config.Config) []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        cfg.Commands.Config.Name,
			Description: cfg.Commands.Config.Description,
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "driver",
					Description: "set captcha driver",
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{
							Name: "Default",
						},
						{
							Name:  "Mathematical",
							Value: "mathematical",
						},
						{
							Name:  "Alphanumerical",
							Value: "alphanumerical",
						},
						{
							Name:  "Simplified Alphanumerical",
							Value: "simplified",
						},
						{
							Name:  "Alphabetical",
							Value: "alphabetical",
						},
						{
							Name:  "Numerical",
							Value: "numerical",
						},
					},
				},
				discord.ApplicationCommandOptionRole{
					Name:                     cfg.Commands.Config.Options.Bypass.Name,
					Description:              cfg.Commands.Config.Options.Bypass.Description,
					DescriptionLocalizations: cfg.Commands.Config.Options.Bypass.DescriptionLocalizations,
				},
				discord.ApplicationCommandOptionBool{
					Name:                     cfg.Commands.Config.Options.Ephemeral.Name,
					Description:              cfg.Commands.Config.Options.Ephemeral.Description,
					DescriptionLocalizations: cfg.Commands.Config.Options.Ephemeral.DescriptionLocalizations,
				},
			},
			DescriptionLocalizations: cfg.Commands.Config.DescriptionLocalizations,
			DefaultMemberPermissions: json.NewNullablePtr(discord.PermissionAdministrator),
		},
		discord.SlashCommandCreate{
			Name:                     cfg.Commands.Captcha.Name,
			Description:              cfg.Commands.Captcha.Description,
			DescriptionLocalizations: cfg.Commands.Captcha.DescriptionLocalizations,
		},
		discord.SlashCommandCreate{
			Name:        cfg.Commands.Submit.Name,
			Description: cfg.Commands.Submit.Description,
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:                     cfg.Commands.Submit.Options.Answer.Name,
					Description:              cfg.Commands.Submit.Options.Answer.Description,
					Required:                 true,
					DescriptionLocalizations: cfg.Commands.Submit.Options.Answer.DescriptionLocalizations,
				},
			},
			DescriptionLocalizations: cfg.Commands.Submit.DescriptionLocalizations,
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
