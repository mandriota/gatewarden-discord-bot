package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/log"
	"github.com/redis/go-redis/v9"
)

var config = struct {
	Token string `json:"token"`
	Redis struct {
		Addr     string `json:"address"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
}{}

var client *redis.Client

func init() {
	json.NewDecoder(os.Stdin).Decode(&config)

	client = redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})
}

var commandsCreate = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "captcha",
		Description: "generates new captcha",
	},
	discord.SlashCommandCreate{
		Name:        "submit",
		Description: "submit answer to captcha",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "answer",
				Description: "answer to captcha",
				Required:    true,
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "config",
		Description: "configure",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionRole{
				Name:        "bypass",
				Description: "bypass role",
			},
			discord.ApplicationCommandOptionBool{
				Name:        "ephemeral",
				Description: "bot message invisibility to each other",
			},
		},
	},
}

func main() {
	log.SetLevel(log.LevelDebug)

	ctx := context.Background()

	client, err := disgo.New(config.Token,
		bot.WithDefaultGateway(),
		bot.WithEventListenerFunc(newCommandListener(ctx)),
	)
	if err != nil {
		log.Fatal("error while building disgo instance: ", err)
	}
	defer client.Close(ctx)

	if _, err := client.Rest().SetGlobalCommands(client.ApplicationID(), commandsCreate); err != nil {
		log.Fatal("error while registering commands: ", err)
	}

	if err := client.OpenGateway(ctx); err != nil {
		log.Fatal("error while connecting to gateway: ", err)
	}

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	<-ctx.Done()
}
