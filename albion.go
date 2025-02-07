package main

import (
	"context"
	"github.com/google/uuid"
	"sync"
)

type GameDataManager struct {
	Parties    *sync.Map
	Characters *sync.Map

	CurrentUser  uuid.UUID
	CurrentParty *Party
}

func NewGameDataManager() *GameDataManager {
	return &GameDataManager{
		Parties:     new(sync.Map),
		Characters:  new(sync.Map),
		CurrentUser: uuid.Nil,
	}
}

func (m *GameDataManager) CreateNewChar(playerUuid uuid.UUID, playerName string) {
	if playerName == "" || playerUuid == uuid.Nil {
		panic("Player information is incorrect")
	}

	_, ok := m.Characters.Load(playerUuid)
	if ok {
		return
	}

	m.Characters.Store(playerUuid, playerName)
}

func (m *GameDataManager) Initialize(playerUuid uuid.UUID, playerName string) {
	if playerName == "" || playerUuid == uuid.Nil {
		panic("Game information is incorrect")
	}

	m.CreateNewChar(playerUuid, playerName)

	m.CurrentUser = playerUuid
}

func (m *GameDataManager) CreatePartyOrUpdate(leader uuid.UUID, members []uuid.UUID) (*Party, bool, []uuid.UUID, []uuid.UUID) {
	value, ok := m.Parties.Load(leader)
	if ok {
		party := value.(*Party)
		removedPlayers, addedPlayers := party.SetMembers(members)
		return value.(*Party), false, removedPlayers, addedPlayers
	}

	value = NewParty(leader, members)
	m.Parties.Store(leader, value)

	return value.(*Party), true, make([]uuid.UUID, 0), members
}

func (m *GameDataManager) DisbandParty(party *Party) {
	m.Parties.Delete(party.PartyOwner)
}

func (m *GameDataManager) GetUsername(userId uuid.UUID) string {
	value, ok := m.Characters.Load(userId)
	if !ok {
		return ""
	}

	return value.(string)
}

func (m *GameDataManager) GetSelfUsername() string {
	return m.GetUsername(m.CurrentUser)
}

func (m *GameDataManager) Close(todo context.Context) {

}
