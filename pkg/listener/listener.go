package listener

import (
	"context"
	"fmt"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/mandriota/gatewarden-bot/pkg/bytes"
	"github.com/mandriota/gatewarden-bot/pkg/config"
	"github.com/redis/go-redis/v9"
	"github.com/steambap/captcha"
)

func New(ctx context.Context, cfg *config.Config) func(acic *events.ApplicationCommandInteractionCreate) {
	l := Listener{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
	}

	return func(acic *events.ApplicationCommandInteractionCreate) {
		err := error(nil)

		switch cname := acic.SlashCommandInteractionData().CommandName(); cname {
		case "captcha":
			err = l.captchaCommandListener(ctx, acic)
		case "submit":
			err = l.submitCommandListener(ctx, acic)
		case "config":
			err = l.configCommandListener(ctx, acic)
		default:
			err = fmt.Errorf("unknow command: %v", cname)
		}

		if err != nil {
			log.Error(err)
		}
	}
}

type Listener struct {
	client *redis.Client

	captchas sync.Map
}

func (l *Listener) configCommandListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	if acic.Member().Permissions.Remove(discord.PermissionManageRoles, discord.PermissionAdministrator) == acic.Member().Permissions {
		return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
			SetContent(":anger: You do not have anyone of these permissions: **Manage Roles**, **Administrator**.").
			Build(),
		)
	}

	if role, ok := acic.SlashCommandInteractionData().OptRole("bypass"); ok {
		l.client.HSet(ctx, "guildsBypassRole", acic.GuildID().String(), role.ID.String())
	}

	if v, ok := acic.SlashCommandInteractionData().OptBool("ephemeral"); ok {
		l.client.HSet(ctx, "guildsEphemerals", acic.GuildID().String(), v)
	}

	return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
		SetContent(":gear: Configuration is updated.").
		Build(),
	)
}

func (l *Listener) captchaCommandListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	data, err := captcha.New(300, 100, func(o *captcha.Options) {
		o.TextLength = 5
		o.CurveNumber = 7
	})
	if err != nil {
		return fmt.Errorf("error while generating captcha: %v", err)
	}

	l.captchas.Store([2]uint16{acic.GuildID().Sequence(), acic.User().ID.Sequence()}, data.Text)

	buf := bytes.AcquireBuffer65536()
	if err := data.WriteJPG(buf, nil); err != nil {
		return fmt.Errorf("error while writing jpeg: %v", err)
	}

	return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
		SetContent(":drop_of_blood: Use /submit command to submit answer").
		SetFiles(discord.NewFile("captcha.jpg", "captcha", buf, discord.FileFlagsNone)).
		Build(),
	)
}

func (l *Listener) submitCommandListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	ans, ok := l.captchas.LoadAndDelete([2]uint16{acic.GuildID().Sequence(), acic.User().ID.Sequence()})
	if !ok {
		return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
			SetContent(":anger: Use /captcha command first").
			Build(),
		)
	}

	if ansv, _ := acic.SlashCommandInteractionData().OptString("answer"); ans.(string) != ansv {
		return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
			SetContent(":x: Wrong answer, please use /captcha command again.").
			Build(),
		)
	}

	id := l.client.HGet(ctx, "guildsBypassRole", acic.GuildID().String()).Val()
	if id == "" {
		return acic.CreateMessage(l.newDefaultMessageCreateBuilder(ctx, acic).
			SetContent(":anger: bypass role is not configured: use /config command first.").
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
		SetContent(":o: You are verified.").
		Build(),
	)
}

func (l *Listener) newDefaultMessageCreateBuilder(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) *discord.MessageCreateBuilder {
	v, _ := l.client.HGet(ctx, "guildsEphemerals", acic.GuildID().String()).Bool()

	return discord.NewMessageCreateBuilder().
		SetEphemeral(v)
}
