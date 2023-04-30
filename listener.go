package main

import (
	"bytes"
	"runtime"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/steambap/captcha"
)

var (
	cDatas = sync.Map{}
	cBuffs = sync.Pool{
		New: func() any {
			return &bytes.Buffer{}
		},
	}
)

func getBuffer() *bytes.Buffer {
	buf := cBuffs.Get().(*bytes.Buffer)
	runtime.SetFinalizer(buf, cBuffs.Put)
	buf.Reset()
	return buf
}

func commandListener(acic *events.ApplicationCommandInteractionCreate) {
	switch cname := acic.SlashCommandInteractionData().CommandName(); cname {
	case "captcha":
		captchaCommandListener(acic)
	case "submit":
		submitCommandListener(acic)
	default:
		log.Error("unknow command: ", cname)
	}
}

func captchaCommandListener(acic *events.ApplicationCommandInteractionCreate) {
	data, err := captcha.New(320, 128, func(o *captcha.Options) {
		o.TextLength = 5
		o.CurveNumber = 7
		o.Noise = 3
	})
	if err != nil {
		log.Error("error while generating captcha: ", err)
	}

	cDatas.Store([2]uint16{acic.GuildID().Sequence(), acic.User().ID.Sequence()}, data.Text)

	buf := getBuffer()
	data.WriteJPG(buf, nil)

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
			snowflake.MustParse(config.GuildsBypassRole[acic.GuildID().String()]),
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
