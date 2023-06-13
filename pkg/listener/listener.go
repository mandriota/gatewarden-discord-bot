package listener

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/mandriota/gatewarden-bot/pkg/config"
	captcha "github.com/mandriota/gatewarden-captcha"
	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, cfg *config.Config) func(acic *events.ApplicationCommandInteractionCreate) {
	const defaultShowLine = captcha.OptionShowHollowLine | captcha.OptionShowSineLine | captcha.OptionShowSlimeLine

	l := Listener{
		config: cfg,
		dbClient: redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
		drivers: map[string]captcha.Driver{
			"mathematical":   captcha.NewDriverMath(200, 600, 10, defaultShowLine, nil, captcha.DefaultEmbeddedFonts, []string{"Comismsh.ttf"}),
			"alphanumerical": captcha.NewDriverString(200, 600, 10, defaultShowLine, 5, captcha.TxtAlphabet+captcha.TxtNumbers, nil, captcha.DefaultEmbeddedFonts, []string{"Comismsh.ttf"}),
			"simplified":     captcha.NewDriverString(200, 600, 10, defaultShowLine, 5, captcha.TxtSimpleCharaters, nil, captcha.DefaultEmbeddedFonts, []string{"Comismsh.ttf"}),
			"alphabetical":   captcha.NewDriverString(200, 600, 10, defaultShowLine, 5, captcha.TxtAlphabet, nil, captcha.DefaultEmbeddedFonts, []string{"Comismsh.ttf"}),
			"numerical":      captcha.NewDriverString(200, 600, 10, defaultShowLine, 5, captcha.TxtNumbers, nil, captcha.DefaultEmbeddedFonts, []string{"Comismsh.ttf"}),
		},
		memStore: captcha.NewMemoryStore(1000, time.Minute),
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
	config   *config.Config
	dbClient *redis.Client

	drivers  map[string]captcha.Driver
	memStore captcha.Store
}

func (l *Listener) commandConfigListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	if enc, ok := acic.SlashCommandInteractionData().OptString("driver"); ok {
		l.dbClient.HSet(ctx, acic.GuildID().String(), "driver", enc)
	}

	if role, ok := acic.SlashCommandInteractionData().OptRole("bypass"); ok {
		l.dbClient.HSet(ctx, acic.GuildID().String(), "bypass_role", role.ID.String())
	}

	if v, ok := acic.SlashCommandInteractionData().OptBool("ephemeral"); ok {
		l.dbClient.HSet(ctx, acic.GuildID().String(), "ephemeral", v)
	}

	return l.reconfiguredCreateMessage(ctx, acic)
}

func (l *Listener) commandCaptchaListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	driver, ok := l.drivers[l.dbClient.HGet(ctx, acic.GuildID().String(), "driver").Val()]
	if !ok {
		driver = l.drivers["alphabetical"]
	}

	_, query, answer := driver.GenerateIdQuestionAnswer()
	capt, err := driver.DrawCaptcha(query)
	if err != nil {
		return fmt.Errorf("error while drawing captcha: %v", err)
	}
	if err := l.memStore.Set(generateCaptchaID(acic), strings.ToLower(answer)); err != nil {
		return fmt.Errorf("error while storing captcha: %v", err)
	}

	return l.generateCaptchaCreateMessage(ctx, acic, capt.EncodeBinary())
}

func (l *Listener) commandSubmitListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	if l.memStore.Get(generateCaptchaID(acic), false) == "" {
		return l.captchaRequiredCreateMessage(ctx, acic)
	}

	answer, _ := acic.SlashCommandInteractionData().OptString("answer")

	if !l.memStore.Verify(generateCaptchaID(acic), strings.ToLower(answer), true) {
		return l.bypassDeniedCreateMessage(ctx, acic)
	}

	id := l.dbClient.HGet(ctx, acic.GuildID().String(), "bypass_role").Val()
	if id == "" {
		return l.bypassRoleRequiredCreateMessage(ctx, acic)
	}

	sfid, err := snowflake.Parse(id)
	if err != nil {
		return fmt.Errorf("error while writing jpeg: %v", err)
	}

	if err := acic.Client().Rest().AddMemberRole(*acic.GuildID(),
		acic.User().ID,
		sfid,
	); err != nil {
		return l.manageRolesPermissionRequiredCreateMessage(ctx, acic)
	}
	return l.bypassedCreateMessage(ctx, acic)
}
