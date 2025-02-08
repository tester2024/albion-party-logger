package main

import (
	"M00DSWINGS/utils"
	"github.com/google/uuid"
	"time"
)

//go:generate golang.org/x/tools/cmd/stringer -type=PartyAction
type PartyAction int

const (
	PartyActionJoin PartyAction = iota
	PartyActionLeave
	PartySelfLeave
)

type PartyHistoryEntry struct {
	Action    PartyAction
	Timestamp time.Time
	User      uuid.UUID
}

type Party struct {
	PartyOwner uuid.UUID
	Members    *utils.HashSet[uuid.UUID]
	History    []PartyHistoryEntry
}

func NewPartyHistory(id uuid.UUID, action PartyAction) PartyHistoryEntry {
	return PartyHistoryEntry{
		Action:    action,
		Timestamp: time.Now().UTC(),
		User:      id,
	}
}

func NewParty(partyOwner uuid.UUID, members []uuid.UUID) *Party {
	set := utils.NewHashSet[uuid.UUID](members...)

	return &Party{
		PartyOwner: partyOwner,
		Members:    set,
		History:    make([]PartyHistoryEntry, 0),
	}
}

func (p *Party) AddPlayer(userId uuid.UUID) {
	if p.Members.Contains(userId) {
		return
	}

	p.addHistory(userId, PartyActionJoin)
	p.Members.Add(userId)
}

func (p *Party) RemoveSelf(userId uuid.UUID) {
	if !p.Members.Contains(userId) {
		return
	}

	p.addHistory(userId, PartySelfLeave)
	p.Members.Add(userId)
}

func (p *Party) RemovePlayer(userId uuid.UUID) {
	if !p.Members.Contains(userId) {
		return
	}

	p.addHistory(userId, PartyActionLeave)
	p.Members.Add(userId)
}

func (p *Party) SetMembers(members []uuid.UUID) (removedPlayers []uuid.UUID, addedPlayers []uuid.UUID) {
	removedPlayers = p.determineRemovedPlayers(members)
	addedPlayers = p.determineAddedPlayers(members)

	for _, id := range removedPlayers {
		p.RemovePlayer(id)
	}

	for _, id := range addedPlayers {
		p.AddPlayer(id)
	}

	return removedPlayers, addedPlayers
}

func (p *Party) determineRemovedPlayers(members []uuid.UUID) []uuid.UUID {
	removedPlayers := make([]uuid.UUID, 0)

	for _, u := range p.Members.Values() {
		if !p.contains(u, members) {
			removedPlayers = append(removedPlayers, u)
		}
	}

	return removedPlayers
}

func (p *Party) determineAddedPlayers(members []uuid.UUID) []uuid.UUID {
	addedPlayers := make([]uuid.UUID, 0)

	for _, member := range members {
		if !p.contains(member, p.Members.Values()) {
			addedPlayers = append(addedPlayers, member)
		}
	}

	return addedPlayers
}

func (p *Party) contains(id uuid.UUID, slice []uuid.UUID) bool {
	for _, u := range slice {
		if u == id {
			return true
		}
	}

	return false
}

func (p *Party) addHistory(id uuid.UUID, partyAction PartyAction) {
	p.History = append(p.History, NewPartyHistory(id, partyAction))
}
