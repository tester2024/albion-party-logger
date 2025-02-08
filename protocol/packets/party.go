package packets

import "github.com/google/uuid"

type EvPartyJoined struct {
	PartyLeader     uuid.UUID   `albion:"3"`
	PlayersUuid     []uuid.UUID `albion:"4"`
	PlayerUsernames []string    `albion:"5"`
}

type EvPartySinglePlayerJoined struct {
	PlayerUID  uuid.UUID `albion:"1"`
	PlayerName string    `albion:"2"`
}

type EvPartyLeft struct {
	PlayerUID uuid.UUID `albion:"1"`
}

type EvPartyLeaderChanged struct {
	NewPartyLeader uuid.UUID `albion:"1"`
}

type EvPartyDisbanded struct {
}
