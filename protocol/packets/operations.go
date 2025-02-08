package packets

import "github.com/google/uuid"

type OpJoinGame struct {
	CharacterID   uuid.UUID `albion:"1"`
	CharacterName string    `albion:"2"`
	GuildID       uuid.UUID `albion:"53"`
	GuildName     string    `albion:"57"`
	AllianceName  string    `albion:"77"`
}

type OpClusterChange struct {
	Cluster string `albion:"0"`
}
