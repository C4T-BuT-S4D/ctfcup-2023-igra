package input

import gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"

func ToProtoEvent(e *Input) *gameserverpb.ClientEvent {
	return &gameserverpb.ClientEvent{
		KeysPressed: &gameserverpb.ClientEvent_KeysPressed{
			WPressed: e.WPressed,
			APressed: e.APressed,
			SPressed: e.SPressed,
			DPressed: e.DPressed,
		},
	}
}

func FromProtoEvent(e *gameserverpb.ClientEvent) *Input {
	return &Input{
		WPressed: e.KeysPressed.WPressed,
		APressed: e.KeysPressed.APressed,
		SPressed: e.KeysPressed.SPressed,
		DPressed: e.KeysPressed.DPressed,
	}
}
