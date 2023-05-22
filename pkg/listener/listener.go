package listener

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/mandriota/base64Captcha"
	"github.com/mandriota/gatewarden-bot/pkg/config"
	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, cfg *config.Config) func(acic *events.ApplicationCommandInteractionCreate) {
	l := Listener{
		config: cfg,
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
		driver: base64Captcha.NewDriverString(
			200, 600, 10,
			base64Captcha.OptionShowHollowLine|
				base64Captcha.OptionShowSineLine|
				base64Captcha.OptionShowSlimeLine,
			5, base64Captcha.TxtAlphabet, nil,
			base64Captcha.DefaultEmbeddedFonts,
			[]string{"Comismsh.ttf"},
		),
		memStore: base64Captcha.NewMemoryStore(1000, time.Minute),
	}

	return func(acic *events.ApplicationCommandInteractionCreate) {
		err := error(nil)

		switch cname := acic.SlashCommandInteractionData().CommandName(); cname {
		case "captcha":
			err = l.commandCaptchaListener(ctx, acic)
		case "submit":
			err = l.commandSubmitListener(ctx, acic)
		case "config":
			err = l.commandConfigListener(ctx, acic)
		default:
			err = fmt.Errorf("unknow command: %v", cname)
		}

		if err != nil {
			log.Error(err)
		}
	}
}

type Listener struct {
	config *config.Config
	client *redis.Client

	driver   base64Captcha.Driver
	memStore base64Captcha.Store
}

func (l *Listener) commandConfigListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	if role, ok := acic.SlashCommandInteractionData().OptRole("bypass"); ok {
		l.client.HSet(ctx, "guildsBypassRole", acic.GuildID().String(), role.ID.String())
	}

	if v, ok := acic.SlashCommandInteractionData().OptBool("ephemeral"); ok {
		l.client.HSet(ctx, "guildsEphemerals", acic.GuildID().String(), v)
	}

	return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
		SetEmbeds(l.newDefaultEmbedBuilder(ctx, acic).
			SetTitle(l.config.Localization.Messages.Reconfigured[acic.Locale()]).
			Build()).
		Build(),
	)
}

func (l *Listener) commandCaptchaListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	_, query, answer := l.driver.GenerateIdQuestionAnswer()
	i, _ := l.driver.DrawCaptcha(query)
	l.memStore.Set(generateCaptchaID(acic), strings.ToLower(answer))

	return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
		SetFiles(discord.NewFile(
			"captcha.jpg",
			"captcha",
			bytes.NewReader(i.EncodeBinary()),
			discord.FileFlagsNone,
		)).
		SetEmbeds(l.newDefaultEmbedBuilder(ctx, acic).
			SetImage("attachment://captcha.jpg").
			SetFooterText(l.config.Localization.Messages.SubmissionRequired[acic.Locale()]).
			Build()).
		Build(),
	)
}

func (l *Listener) commandSubmitListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	if l.memStore.Get(generateCaptchaID(acic), false) == "" {
		return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
			SetEmbeds(l.newDefaultEmbedBuilder(ctx, acic).
				SetTitle(l.config.Localization.Messages.CaptchaRequired[acic.Locale()]).
				Build()).
			Build(),
		)
	}

	ansv, _ := acic.SlashCommandInteractionData().OptString("answer")

	if !l.memStore.Verify(generateCaptchaID(acic), strings.ToLower(ansv), true) {
		return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
			SetEmbeds(l.newDefaultEmbedBuilder(ctx, acic).
				SetTitle(l.config.Localization.Messages.BypassDenied[acic.Locale()]).
				Build()).
			Build(),
		)
	}

	id := l.client.HGet(ctx, "guildsBypassRole", acic.GuildID().String()).Val()
	if id == "" {
		log.Info("bypass role required")
		return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
			SetEmbeds(l.newDefaultEmbedBuilder(ctx, acic).
				SetTitle(l.config.Localization.Messages.BypassRoleRequired[acic.Locale()]).
				Build()).
			Build(),
		)
	}

	sfid, err := snowflake.Parse(id)
	if err != nil {
		return fmt.Errorf("error while writing jpeg: %v", err)
	}

	if err := acic.Client().Rest().AddMemberRole(*acic.GuildID(),
		acic.User().ID,
		sfid,
	); err != nil {
		return fmt.Errorf("error while giving role: %v", err)
	}
	return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
		SetEmbeds(l.newDefaultEmbedBuilder(ctx, acic).
			SetTitle(l.config.Localization.Messages.Bypassed[acic.Locale()]).
			Build()).
		Build(),
	)
}

func (l *Listener) newDefaultMessageCreateBuilder(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) *discord.MessageCreateBuilder {
	v, _ := l.client.HGet(ctx, "guildsEphemerals", acic.GuildID().String()).Bool()

	return discord.NewMessageCreateBuilder().
		SetEphemeral(v)
}

func (l *Listener) newDefaultEmbedBuilder(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) *discord.EmbedBuilder {
	return discord.NewEmbedBuilder().
		SetColor(0xFF0000)
}

func generateCaptchaID(acic *events.ApplicationCommandInteractionCreate) string {
	return unsafe.String((*byte)(unsafe.Pointer(&[2]uint16{
		acic.GuildID().Sequence(),
		acic.User().ID.Sequence(),
	})), 4)
}
