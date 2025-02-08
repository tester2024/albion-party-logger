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

type EvNewCharacter struct {
	PlayerUID    uuid.UUID `albion:"7"`
	PlayerName   string    `albion:"1"`
	GuildName    string    `albion:"8"`
	AllianceName string    `albion:"51"`
}

type EvCharacterStats struct {
	PlayerName   string `albion:"1"`
	GuildName    string `albion:"2"`
	AllianceName string `albion:"4"`
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
