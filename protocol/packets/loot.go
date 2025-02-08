package packets

import "github.com/google/uuid"

type EvNewLootChest struct {
	Id    int    `albion:"0"`
	Owner string `albion:"3"`
}

type EvNewLoot struct {
	Id    int    `albion:"0"`
	Owner string `albion:"3"`
}

type EvNewSimpleItem struct {
	Id        int `albion:"0"`
	ItemIndex int `albion:"1"`
	Quantity  int `albion:"2"`
}

type EvAttachItemContainer struct {
	Id            int       `albion:"0"`
	ContainerUUID uuid.UUID `albion:"1"`
	Items         []int     `albion:"3"`
	Slots         int       `albion:"4"`
}

type EvDetachItemContainer struct {
	ContainerUUID uuid.UUID `albion:"0"`
}

type EvUpdateLootChest struct {
	Id int `albion:"0"`
}

type EvOtherGrabbedLoot struct {
	LootedFromName string `albion:"1"`
	LooterByName   string `albion:"2"`
	IsSilver       bool   `albion:"3"`
	ItemIndex      int    `albion:"4"`
	Quantity       int    `albion:"5"`
}

type EvInventoryPutItems struct {
	ObjectId    int       `albion:"0"`
	SlotId      int       `albion:"1"`
	ContainerId uuid.UUID `albion:"2"`
}

type OpInventoryMoveItems struct {
	FromSlot int       `albion:"0"`
	FromUUID uuid.UUID `albion:"1"`
	ToSlot   int       `albion:"3"`
	ToUUID   uuid.UUID `albion:"4"`
}
