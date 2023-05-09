package main

import (
	"context"
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
	switch cname := acic.SlashCommandInteractionData().CommandName(); cname {
	case "captcha":
		captchaCommandListener(acic)
	case "submit":
		submitCommandListener(acic)
	case "config":
		configCommandListener(acic)
	default:
		log.Error("unknow command: ", cname)
	}
}

func captchaCommandListener(acic *events.ApplicationCommandInteractionCreate) {
	data, err := captcha.New(300, 100, func(o *captcha.Options) {
		o.TextLength = 5
		o.CurveNumber = 7
	})
	if err != nil {
		log.Error("error while generating captcha: ", err)
	}

	cDatas.Store([2]uint16{acic.GuildID().Sequence(), acic.User().ID.Sequence()}, data.Text)

	buf := bytes.AcquireBuffer65536()
	if err := data.WriteJPG(buf, nil); err != nil {
		log.Error("error while writing jpeg: ", err)
	}

	if err := acic.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(":drop_of_blood: Use /submit command to submit answer").
		SetFiles(discord.NewFile("captcha.jpg", "captcha", buf, discord.FileFlagsNone)).
		Build(),
	); err != nil {
		log.Error("error while creating message: ", err)
	}
}

func submitCommandListener(acic *events.ApplicationCommandInteractionCreate) {
	ans, ok := cDatas.LoadAndDelete([2]uint16{acic.GuildID().Sequence(), acic.User().ID.Sequence()})
	if !ok {
		acic.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(":anger: Use /captcha command first").
			Build(),
		)
	}

	if ansv, _ := acic.SlashCommandInteractionData().OptString("answer"); ans.(string) == ansv {
		acic.Client().Rest().AddMemberRole(*acic.GuildID(),
			acic.User().ID,
			snowflake.MustParse(client.HGet(context.TODO(), "guildsBypassRole", acic.GuildID().String()).Val()),
		)
		acic.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(":o: You are verified.").
			Build(),
		)
	} else {
		acic.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(":x: Wrong answer, please use /captcha command again.").
			Build(),
		)
	}
}

func configCommandListener(acic *events.ApplicationCommandInteractionCreate) {
	if role, ok := acic.SlashCommandInteractionData().OptRole("bypass"); ok {
		client.HSetNX(context.TODO(), "guildsBypassRole", acic.GuildID().String(), role.ID.String())
		acic.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(":gear: Configuration is updated.").
			Build(),
		)
	}
}
