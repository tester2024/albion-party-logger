package main

import (
	"M00DSWINGS/protocol/enums"
	"M00DSWINGS/protocol/packets"
	"M00DSWINGS/utils"
	"flag"
	"github.com/google/gopacket/pcap"
	"log"
)

var (
	err           error
	interfaceName string
	serverAddr    string
	interfaceObj  pcap.Interface
)

func init() {
	flag.StringVar(&interfaceName, "interface", "", "Network interface to use")
	flag.StringVar(&serverAddr, "server", "ws://85.192.42.52:3000", "Server address")

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

	// Party events
	l.RegisterEvent(enums.EventTypePartyPlayerJoined, packets.EvPartySinglePlayerJoined{})
	l.RegisterEvent(enums.EventTypePartyJoined, packets.EvPartyJoined{})
	l.RegisterEvent(enums.EventTypePartyPlayerLeft, packets.EvPartyLeft{})
	l.RegisterEvent(enums.EventTypeNewCharacter, packets.EvNewCharacter{})
	l.RegisterEvent(enums.EventTypeCharacterStats, packets.EvCharacterStats{})
	l.RegisterEvent(enums.EventTypePartyLeaderChanged, packets.EvPartyLeaderChanged{})
	l.RegisterEvent(enums.EventTypePartyDisbanded, packets.EvPartyDisbanded{})

	// Log Party events
	l.RegisterEvent(enums.EventTypePartyReadyCheckUpdate, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyPlayerUpdated, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyInvitationAnswer, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyJoinRequestAnswer, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyLootItems, packets.Logger{})
	l.RegisterEvent(enums.EventTypePartyLootItemsRemoved, packets.Logger{})

	// Loot events
	l.RegisterOperation(enums.OpTypeInventoryMoveItem, packets.OpInventoryMoveItems{})
	l.RegisterEvent(enums.EventTypeNewSimpleItem, packets.EvNewSimpleItem{})
	l.RegisterEvent(enums.EventTypeNewLootChest, packets.EvNewLootChest{})
	l.RegisterEvent(enums.EventTypeNewLoot, packets.EvNewLoot{})
	l.RegisterEvent(enums.EventTypeAttachItemContainer, packets.EvAttachItemContainer{})
	l.RegisterEvent(enums.EventTypeDetachItemContainer, packets.EvDetachItemContainer{})
	l.RegisterEvent(enums.EventTypeUpdateLootChest, packets.EvUpdateLootChest{})
	l.RegisterEvent(enums.EventTypeOtherGrabbedLoot, packets.EvOtherGrabbedLoot{})
	l.RegisterEvent(enums.EventTypeInventoryPutItem, packets.EvInventoryPutItems{})

	ws := NewWebSocketClient(serverAddr)
	defer ws.Close()

	l.RegisterListeners(func(data interface{}) {
		switch d := data.(type) {
		case *packets.OpJoinGame:
			log.Printf("Joined game with Character ID: %s, Name: %s, Guild: %s, Alliance: %s", d.CharacterID, d.CharacterName, d.GuildName, d.AllianceName)
			if err := ws.Initialize(d.CharacterID, d.CharacterName, d.GuildName, d.AllianceName); err != nil {
				log.Fatal(err)
			}

		case *packets.EvNewCharacter:
			if err := ws.CreateNewChar(d.PlayerUID, d.PlayerName, d.GuildName, d.AllianceName); err != nil {
				log.Println(err)
			}

		case *packets.EvCharacterStats:
			if err := ws.UpdateCharacterStats(d.PlayerName, d.GuildName, d.AllianceName); err != nil {
				log.Println(err)
			}

		case *packets.EvPartySinglePlayerJoined:
			if err := ws.CreateNewChar(d.PlayerUID, d.PlayerName, "", ""); err != nil {
				log.Println(err)
			}

			if err := ws.AddPartyPlayer(d.PlayerUID); err != nil {
				log.Println(err)
			}

		case *packets.EvPartyJoined:
			for i, playerUsername := range d.PlayerUsernames {
				if err := ws.CreateNewChar(d.PlayersUuid[i], playerUsername, "", ""); err != nil {
					log.Println(err)
				}
			}

			if err := ws.JoinParty(d.PartyLeader, d.PlayersUuid); err != nil {
				log.Println(err)
			}

		case *packets.EvPartyLeft:
			if err := ws.RemovePartyPlayer(d.PlayerUID); err != nil {
				log.Println(err)
			}

		case *packets.EvPartyDisbanded:
			if err := ws.DisbandParty(); err != nil {
				log.Println(err)
			}

		case *packets.EvPartyLeaderChanged:
			if err := ws.UpdatePartyLeader(d.NewPartyLeader); err != nil {
				log.Println(err)
			}

		case *packets.OpInventoryMoveItems:
			if err := ws.MoveItems(d.FromSlot, d.FromUUID, d.ToSlot, d.ToUUID); err != nil {
				log.Println(err)
			}

		case *packets.EvInventoryPutItems:
			if err := ws.PutItems(d.ObjectId, d.ContainerId, d.SlotId); err != nil {
				log.Println(err)
			}

		case *packets.EvNewLootChest:
			if err := ws.CreateNewLootChest(d.Id, d.Owner); err != nil {
				log.Println(err)
			}

		case *packets.EvNewLoot:
			if err := ws.CreateNewLoot(d.Id, d.Owner); err != nil {
				log.Println(err)
			}

		case *packets.EvUpdateLootChest:
			if err := ws.UpdateLootChest(d.Id); err != nil {
				log.Println(err)
			}

		case *packets.EvOtherGrabbedLoot:
			if err := ws.OtherGrabLoot(d.LootedFromName, d.LooterByName, d.IsSilver, d.ItemIndex, d.Quantity); err != nil {
				log.Println(err)
			}

		case *packets.EvNewSimpleItem:
			if err := ws.NewSimpleItem(d.Id, d.ItemIndex, d.Quantity); err != nil {
				log.Println(err)
			}

		case *packets.EvAttachItemContainer:
			if err := ws.AttachItemContainer(d.Id, d.ContainerUUID, d.Items, d.Slots); err != nil {
				log.Println(err)
			}

		case *packets.EvDetachItemContainer:
			if err := ws.DetachItemContainer(d.ContainerUUID); err != nil {
				log.Println(err)
			}

		default:
			log.Println(data)
		}
	})

	l.RegisterDisconnect(func() {
		log.Printf("Disconnceted!")
	})

	l.ListenAndServe()
}
