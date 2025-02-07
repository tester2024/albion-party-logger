package packets

import (
	"M00DSWINGS/protocol"
	"M00DSWINGS/protocol/enums"
	"M00DSWINGS/protocol/photon"
	"github.com/google/uuid"
	"log"
)

type EvLeave struct {
	Id int64 `albion:"0"`
}

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

type EvNewCharacter struct {
	PlayerUID  uuid.UUID `albion:"7"`
	PlayerName string    `albion:"1"`
}

type EvLootEvent struct {
	ObjectID int    `albion:"0"`
	LootBody string `albion:"3"`
}

type EvGrabbedLoot struct {
	LootedFromName string `albion:"1"`
	LooterByName   string `albion:"2"`
	IsSilver       bool   `albion:"3"`
	ItemIndex      int    `albion:"4"`
	Quantity       int    `albion:"5"`
}

type EvPartyDisbanded struct {
}

type Logger struct {
}

func (e *Logger) Decode(params photon.ReliableMessageParamaters) {
	if val, ok := params[252]; ok {
		var event = enums.EventType(protocol.DecodeInteger(val))
		log.Printf("[Logger] Type %s, Params: %v", event, params)

	} else if val, ok := params[253]; ok {
		var event = enums.OperationType(protocol.DecodeInteger(val))
		log.Printf("[Logger] Type %s, Params: %v", event, params)

	} else {
		log.Printf("[Logger] Type Unknown, Params: %v", params)
	}
}
