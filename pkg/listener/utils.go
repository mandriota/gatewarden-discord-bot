package listener

import (
	"unsafe"

	"github.com/disgoorg/disgo/events"
)

func generateCaptchaID(acic *events.ApplicationCommandInteractionCreate) string {
	return unsafe.String((*byte)(unsafe.Pointer(&[2]uint16{
		acic.GuildID().Sequence(),
		acic.User().ID.Sequence(),
	})), 4)
}
