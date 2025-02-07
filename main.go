package main

import (
	"M00DSWINGS/protocol/enums"
	"M00DSWINGS/protocol/packets"
	"M00DSWINGS/utils"
	"context"
	"flag"
	"github.com/google/gopacket/pcap"
	"github.com/google/uuid"
	"log"
	"syscall"
)

var (
	err            error
	interfaceName  string
	interfaceObj   pcap.Interface
	partyLoggerUrl = getEnv("PARTY_WEBHOOK")
	lootLoggerUrl  = getEnv("LOOT_WEBHOOK")
)

func getEnv(key string) string {
	value, exists := syscall.Getenv(key)
	if !exists {
		log.Printf("Environment variable %s not set", key)
		return ""
	}
	return value
}

func init() {
	flag.StringVar(&interfaceName, "interface", "", "Network interface to use")
	flag.Parse()

	if !utils.CheckPcapInstalled() {
		log.Fatal("pcap is not installed")
	}

	switch {
	case interfaceName == "":
		interfaceObj, err = utils.GetDefaultDevice()
	default:
		interfaceObj, err = utils.FindDevice(interfaceName)
	}
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Using interface %s", interfaceObj.Description)

}

func main() {
	l := NewLogger(interfaceObj)

	l.RegisterOperation(enums.OpTypeJoin, packets.OpJoinGame{})
	l.RegisterOperation(enums.OpTypePartyLeave, packets.Logger{})
	l.RegisterOperation(enums.OpTypePartyKickPlayer, packets.Logger{})

	l.RegisterEvent(enums.EventTypePartyPlayerJoined, packets.EvPartySinglePlayerJoined{})
	l.RegisterEvent(enums.EventTypePartyJoined, packets.EvPartyJoined{})
	l.RegisterEvent(enums.EventTypePartyPlayerLeft, packets.EvPartyLeft{})
	l.RegisterEvent(enums.EventTypeNewCharacter, packets.EvNewCharacter{})
	l.RegisterEvent(enums.EventTypePartyLeaderChanged, packets.EvPartyLeaderChanged{})

	l.RegisterEvent(enums.EventTypePartyDisbanded, packets.EvPartyDisbanded{})

	l.RegisterEvent(enums.EventTypePartyReadyCheckUpdate, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyPlayerLeaveScheduled, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyPlayerUpdated, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyInvitationAnswer, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyJoinRequestAnswer, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyOnClusterPartyJoined, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyLootItems, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyLootItemsRemoved, packets.Logger{})
	l.RegisterEvent(enums.EventTypeOtherGrabbedLoot, packets.EvGrabbedLoot{})
	l.RegisterEvent(enums.EventTypeNewLoot, packets.Logger{})

	gm := NewGameDataManager()
	defer gm.Close(context.TODO())

	logger := NewDiscordLogger(gm, partyLoggerUrl, lootLoggerUrl)
	defer logger.Close(context.TODO())

	l.RegisterListeners(func(data interface{}) {
		switch d := data.(type) {
		case *packets.OpJoinGame:
			log.Printf("Joined as %s in the game with name %s and guild %s", d.CharacterID, d.CharacterName, d.GuildName)
			gm.Initialize(d.CharacterID, d.CharacterName)

		case *packets.EvNewCharacter:
			gm.CreateNewChar(d.PlayerUID, d.PlayerName)

		case *packets.EvPartySinglePlayerJoined:
			if gm.CurrentParty == nil {
				panic("currentParty is nil")
			}

			gm.CreateNewChar(d.PlayerUID, d.PlayerName)
			gm.CurrentParty.AddPlayer(d.PlayerUID)

			logger.LogPartyAdd(d.PlayerUID)
		case *packets.EvPartyJoined:
			for i, playerUsername := range d.PlayerUsernames {
				gm.CreateNewChar(d.PlayersUuid[i], playerUsername)
			}

			var newParty bool

			var removedPlayers, addedPlayers []uuid.UUID
			gm.CurrentParty, newParty, removedPlayers, addedPlayers = gm.CreatePartyOrUpdate(d.PartyLeader, d.PlayersUuid)

			if newParty {
				logger.LogPartyCreate(gm.CurrentParty)
			} else {
				logger.LogPartyUpdate(removedPlayers, addedPlayers)
			}

		case *packets.EvPartyLeft:
			if gm.CurrentParty == nil {
				panic("currentParty is nil")
			}

			if d.PlayerUID == gm.CurrentUser {
				gm.CurrentParty.RemoveSelf(d.PlayerUID)
				logger.LogSelfLeave(d.PlayerUID)
			} else {
				gm.CurrentParty.RemovePlayer(d.PlayerUID)
				logger.LogPartyLeft(d.PlayerUID)
			}

		case *packets.EvPartyDisbanded:
			if gm.CurrentParty == nil {
				panic("currentParty is nil")
			}

			gm.DisbandParty(gm.CurrentParty)
			logger.LogPartyDisband()
			gm.CurrentParty = nil

		case *packets.EvPartyLeaderChanged:
			if gm.CurrentParty == nil {
				panic("currentParty is nil")
			}

			currentParty := gm.CurrentParty

			gm.Parties.Delete(currentParty.PartyOwner)
			currentParty.PartyOwner = d.NewPartyLeader
			gm.Parties.Store(d.NewPartyLeader, currentParty)

		default:
			log.Println(data)
		}
	})

	l.RegisterDisconnect(func() {
		log.Printf("Disconnceted!")
	})

	l.ListenAndServe()
}
