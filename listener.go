package main

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/mandriota/gatewarden-bot/pkg/bytes"
	"github.com/steambap/captcha"
)

var cDatas sync.Map

func commandListener(acic *events.ApplicationCommandInteractionCreate) {
	err := error(nil)

	switch cname := acic.SlashCommandInteractionData().CommandName(); cname {
	case "captcha":
		err = captchaCommandListener(acic)
	case "submit":
		err = submitCommandListener(acic)
	case "config":
		err = configCommandListener(acic)
	default:
		err = fmt.Errorf("unknow command: %v", cname)
	}

	if err != nil {
		log.Error(errors.Join(err, acic.CreateMessage(discord.NewMessageCreateBuilder().
			SetContentf(":x: %v", err).
			Build(),
		)))
	}
}

func captchaCommandListener(acic *events.ApplicationCommandInteractionCreate) error {
	data, err := captcha.New(300, 100, func(o *captcha.Options) {
		o.TextLength = 5
		o.CurveNumber = 7
	})
	if err != nil {
		return fmt.Errorf("error while generating captcha: %v", err)
	}

	cDatas.Store([2]uint16{acic.GuildID().Sequence(), acic.User().ID.Sequence()}, data.Text)

	buf := bytes.AcquireBuffer65536()
	if err := data.WriteJPG(buf, nil); err != nil {
		log.Error("error while writing jpeg: ", err)
	}

	return acic.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(":drop_of_blood: Use /submit command to submit answer").
		SetFiles(discord.NewFile("captcha.jpg", "captcha", buf, discord.FileFlagsNone)).
		Build(),
	)
}

func submitCommandListener(acic *events.ApplicationCommandInteractionCreate) error {
	ans, ok := cDatas.LoadAndDelete([2]uint16{acic.GuildID().Sequence(), acic.User().ID.Sequence()})
	if !ok {
		return acic.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(":anger: Use /captcha command first").
			Build(),
		)
	}

	if ansv, _ := acic.SlashCommandInteractionData().OptString("answer"); ans.(string) != ansv {
		return acic.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(":x: Wrong answer, please use /captcha command again.").
			Build(),
		)
	}

	id := client.HGet(context.TODO(), "guildsBypassRole", acic.GuildID().String()).Val()
	if id == "" {
		return acic.CreateMessage(discord.NewMessageCreateBuilder().
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
	return acic.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(":o: You are verified.").
		Build(),
	)
}

func configCommandListener(acic *events.ApplicationCommandInteractionCreate) error {
	if acic.Member().Permissions.Remove(discord.PermissionManageRoles, discord.PermissionAdministrator) == acic.Member().Permissions {
		return acic.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(":anger: You do not have anyone of these permissions: **Manage Roles**, **Administrator**.").
			Build(),
		)
	}

	if role, ok := acic.SlashCommandInteractionData().OptRole("bypass"); ok {
		client.HSetNX(context.TODO(), "guildsBypassRole", acic.GuildID().String(), role.ID.String())
		return acic.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(":gear: Configuration is updated.").
			Build(),
		)
	}

	return acic.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(":anger: At least 1 option is required.").
		Build(),
	)
}
