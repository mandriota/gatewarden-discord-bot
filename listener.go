package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/mandriota/gatewarden-bot/pkg/bytes"
	"github.com/steambap/captcha"
)

func newCommandListener(ctx context.Context) func(acic *events.ApplicationCommandInteractionCreate) {
	cDatas := &sync.Map{}

	return func(acic *events.ApplicationCommandInteractionCreate) {
		err := error(nil)

		switch cname := acic.SlashCommandInteractionData().CommandName(); cname {
		case "captcha":
			err = captchaCommandListener(ctx, cDatas, acic)
		case "submit":
			err = submitCommandListener(ctx, cDatas, acic)
		case "config":
			err = configCommandListener(ctx, acic)
		default:
			err = fmt.Errorf("unknow command: %v", cname)
		}

		if err != nil {
			log.Error(err)
		}
	}
}

func captchaCommandListener(ctx context.Context, cdatas *sync.Map, acic *events.ApplicationCommandInteractionCreate) error {
	data, err := captcha.New(300, 100, func(o *captcha.Options) {
		o.TextLength = 5
		o.CurveNumber = 7
	})
	if err != nil {
		return fmt.Errorf("error while generating captcha: %v", err)
	}

	cdatas.Store([2]uint16{acic.GuildID().Sequence(), acic.User().ID.Sequence()}, data.Text)

	buf := bytes.AcquireBuffer65536()
	if err := data.WriteJPG(buf, nil); err != nil {
		return fmt.Errorf("error while writing jpeg: %v", err)
	}

	return acic.CreateMessage(newDefaultMessageCreateBuilder(ctx, acic).
		SetContent(":drop_of_blood: Use /submit command to submit answer").
		SetFiles(discord.NewFile("captcha.jpg", "captcha", buf, discord.FileFlagsNone)).
		Build(),
	)
}

func submitCommandListener(ctx context.Context, cdatas *sync.Map, acic *events.ApplicationCommandInteractionCreate) error {
	ans, ok := cdatas.LoadAndDelete([2]uint16{acic.GuildID().Sequence(), acic.User().ID.Sequence()})
	if !ok {
		return acic.CreateMessage(newDefaultMessageCreateBuilder(ctx, acic).
			SetContent(":anger: Use /captcha command first").
			Build(),
		)
	}

	if ansv, _ := acic.SlashCommandInteractionData().OptString("answer"); ans.(string) != ansv {
		return acic.CreateMessage(newDefaultMessageCreateBuilder(ctx, acic).
			SetContent(":x: Wrong answer, please use /captcha command again.").
			Build(),
		)
	}

	id := client.HGet(ctx, "guildsBypassRole", acic.GuildID().String()).Val()
	if id == "" {
		return acic.CreateMessage(newDefaultMessageCreateBuilder(ctx, acic).
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
	return acic.CreateMessage(newDefaultMessageCreateBuilder(ctx, acic).
		SetContent(":o: User has been verified.").
		Build(),
	)
}

func configCommandListener(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	if acic.Member().Permissions.Remove(discord.PermissionManageRoles, discord.PermissionAdministrator) == acic.Member().Permissions {
		return acic.CreateMessage(newDefaultMessageCreateBuilder(ctx, acic).
			SetContent(":anger: You do not have anyone of these permissions: **Manage Roles**, **Administrator**.").
			Build(),
		)
	}

	if role, ok := acic.SlashCommandInteractionData().OptRole("bypass"); ok {
		client.HSet(ctx, "guildsBypassRole", acic.GuildID().String(), role.ID.String())
	}

	if v, ok := acic.SlashCommandInteractionData().OptBool("ephemeral"); ok {
		client.HSet(ctx, "guildsEphemerals", acic.GuildID().String(), v)
	}

	return acic.CreateMessage(newDefaultMessageCreateBuilder(ctx, acic).
		SetContent(":gear: Configuration is updated.").
		Build(),
	)
}

func newDefaultMessageCreateBuilder(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) *discord.MessageCreateBuilder {
	v, _ := client.HGet(ctx, "guildsEphemerals", acic.GuildID().String()).Bool()

	return discord.NewMessageCreateBuilder().
		SetEphemeral(v)
}
