package listener

import (
	"bytes"
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/mandriota/gatewarden-bot/pkg/config"
)

func (l *Listener) messageCreateBuilderBase(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) *discord.MessageCreateBuilder {
	v, _ := l.dbClient.HGet(ctx, acic.GuildID().String(), "ephemeral").Bool()

	return discord.NewMessageCreateBuilder().
		SetEphemeral(v)
}

func (l *Listener) embedBuilderBase(ctx context.Context, acic *events.ApplicationCommandInteractionCreate, resp config.Response) *discord.EmbedBuilder {
	return discord.NewEmbedBuilder().
		SetColor(0xFF6798).
		SetTitle(resp.TitleLocalizations[acic.Locale()]).
		SetDescription(resp.DescriptionLocalizations[acic.Locale()]).
		SetFooterText(resp.FooterLocalizations[acic.Locale()])
}

func (l *Listener) reconfiguredCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic, l.config.Commands.Config.Responses.Success).
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
		SetEmbeds(l.embedBuilderBase(ctx, acic, l.config.Commands.Captcha.Responses.Success).
			SetImage("attachment://captcha.jpg").
			Build()).
		Build(),
	)
}

func (l *Listener) bypassDeniedCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic, l.config.Commands.Submit.Responses.ErrWrongAnswer).
			Build()).
		Build(),
	)
}

func (l *Listener) captchaRequiredCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic, l.config.Commands.Submit.Responses.ErrUndefinedCaptcha).
			Build()).
		Build(),
	)
}

func (l *Listener) bypassRoleRequiredCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic, l.config.Commands.Submit.Responses.ErrUndefinedBypassRole).
			Build()).
		Build(),
	)
}

func (l *Listener) bypassedCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic, l.config.Commands.Submit.Responses.Success).
			Build()).
		Build())
}

func (l *Listener) manageRolesPermissionRequiredCreateMessage(ctx context.Context, acic *events.ApplicationCommandInteractionCreate) error {
	return acic.CreateMessage(l.messageCreateBuilderBase(ctx, acic).
		SetEmbeds(l.embedBuilderBase(ctx, acic, l.config.Commands.Submit.Responses.ErrMissingManageRolesPermission).
			Build()).
		Build())
}
