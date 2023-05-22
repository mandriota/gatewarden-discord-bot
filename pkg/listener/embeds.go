package listener

import (
	"bytes"
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (l *Listener) messageCreateBuilderBase(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) *discord.MessageCreateBuilder {
	v, _ := l.dbClient.HGet(ctx, "guildsEphemerals", acic.GuildID().String()).Bool()

	return discord.NewMessageCreateBuilder().
		SetEphemeral(v)
}

func (l *Listener) embedBuilderBase(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) *discord.EmbedBuilder {
	return discord.NewEmbedBuilder().
		SetColor(0xFF6798)
}

func (l *Listener) reconfiguredCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic).
			SetTitle(l.config.Localization.Messages.Reconfigured[acic.Locale()]).
			Build()).
		Build())
}

func (l *Listener) generateCaptchaCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate, captcha []byte) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetFiles(discord.NewFile(
			"captcha.jpg",
			"captcha",
			bytes.NewReader(captcha),
			discord.FileFlagsNone,
		)).
		SetEmbeds(l.embedBuilderBase(ctx, acic).
			SetImage("attachment://captcha.jpg").
			SetFooterText(l.config.Localization.Messages.SubmissionRequired[acic.Locale()]).
			Build()).
		Build(),
	)
}

func (l *Listener) bypassDeniedCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic).
			SetTitle(l.config.Localization.Messages.BypassDenied[acic.Locale()]).
			Build()).
		Build(),
	)
}

func (l *Listener) captchaRequiredCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic).
			SetTitle(l.config.Localization.Messages.CaptchaRequired[acic.Locale()]).
			Build()).
		Build(),
	)
}

func (l *Listener) bypassRoleRequiredCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic).
			SetTitle(l.config.Localization.Messages.BypassRoleRequired[acic.Locale()]).
			Build()).
		Build(),
	)
}

func (l *Listener) bypassedCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic).
			SetTitle(l.config.Localization.Messages.Bypassed[acic.Locale()]).
			Build()).
		Build())
}
