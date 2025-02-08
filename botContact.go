package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	url          string
	conn         *websocket.Conn
	sendMx       *sync.Mutex
	sendChan     chan []byte
	charcterId   uuid.UUID
	charcterName string
	allianceName string
	guildName    string
}

func NewWebSocketClient(url string) *WebSocketClient {
	return &WebSocketClient{
		url:      url,
		sendMx:   new(sync.Mutex),
		sendChan: make(chan []byte, 100),
	}
}

func (c *WebSocketClient) Connect(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.url, nil)
		if err != nil {
			log.Printf("Failed to connect, retrying in 5 seconds... %v\n", err)
			<-ticker.C
			continue
		}

		log.Println("Connected to WebSocket server")

		c.conn = conn

		if err := c.SendInitialize(); err != nil {
			return err
		}

		go c.readLoop(ctx)
		go c.writeLoop(ctx)

		return nil
	}
}

func (c *WebSocketClient) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Read error, reconnecting... %v\n", err)
			c.reconnect(ctx)
			return
		}
		log.Println("Received:", string(message))
	}
}

func (c *WebSocketClient) writeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.sendChan:
			c.sendMx.Lock()
			var message map[string]interface{}

			if err := json.Unmarshal(msg, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v\n", err)
				c.sendMx.Unlock()
				continue
			}

			if c.conn == nil {
				c.sendChan <- msg
				c.sendMx.Unlock()
				continue
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("Write error, reconnecting... %v\n", err)
				c.sendMx.Unlock()
				c.reconnect(ctx)
				return
			}
			c.sendMx.Unlock()
		}
	}
}

func (c *WebSocketClient) reconnect(ctx context.Context) {
	c.conn.Close()
	c.conn = nil
	go c.Connect(ctx)
}

func (c *WebSocketClient) Send(ctx context.Context, message map[string]interface{}) error {
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case c.sendChan <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *WebSocketClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func (c *WebSocketClient) Initialize(id uuid.UUID, name string, guildName string, allianceName string) error {
	c.charcterId = id
	c.charcterName = name
	c.guildName = guildName
	c.allianceName = allianceName

	if c.conn == nil {
		if err := c.Connect(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

func (c *WebSocketClient) SendInitialize() error {
	msg := map[string]interface{}{
		"action":   "initialize",
		"id":       c.charcterId,
		"name":     c.charcterName,
		"guild":    c.guildName,
		"alliance": c.allianceName,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send initialize message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) CreateNewChar(uid uuid.UUID, name string, guildName string, allianceName string) error {
	msg := map[string]interface{}{
		"action":   "new_character",
		"id":       uid,
		"name":     name,
		"guild":    guildName,
		"alliance": allianceName,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send create_new_char message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) UpdateCharacterStats(name string, guild string, alliance string) interface{} {
	msg := map[string]interface{}{
		"action":   "update_character_stats",
		"name":     name,
		"guild":    guild,
		"alliance": alliance,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send update_character_stats message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) JoinParty(leader uuid.UUID, playersUuid []uuid.UUID) error {
	msg := map[string]interface{}{
		"action":  "join_party",
		"leader":  leader,
		"players": playersUuid,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send create_party_or_update message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) AddPartyPlayer(uid uuid.UUID) error {
	msg := map[string]interface{}{
		"action": "add_member",
		"id":     uid,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send add_party_player message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) RemovePartyPlayer(uid uuid.UUID) error {
	msg := map[string]interface{}{
		"action": "remove_member",
		"id":     uid,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send remove_party_player message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) DisbandParty() error {
	msg := map[string]interface{}{
		"action": "disband_party",
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send disband_party message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) UpdatePartyLeader(leader uuid.UUID) error {
	msg := map[string]interface{}{
		"action": "update_leader",
		"leader": leader,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send update_party_leader message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) AttachItemContainer(id int, uuid uuid.UUID, items []int, slots int) error {
	msg := map[string]interface{}{
		"action": "attach_item_container",
		"id":     id,
		"uuid":   uuid,
		"items":  items,
		"slots":  slots,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send attach_item_container message: %v\n", err)
	}

	return nil
}
func (c *WebSocketClient) MoveItems(fromSlot int, fromUUID uuid.UUID, toSlot int, toUUID uuid.UUID) error {
	msg := map[string]interface{}{
		"action":   "move_items",
		"fromSlot": fromSlot,
		"fromUUID": fromUUID,
		"toSlot":   toSlot,
		"toUUID":   toUUID,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send move_items message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) PutItems(id int, containerId uuid.UUID, slotId int) error {
	msg := map[string]interface{}{
		"action":      "put_items",
		"id":          id,
		"containerId": containerId,
		"slotId":      slotId,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send put_items message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) CreateNewLootChest(id int, owner string) error {
	msg := map[string]interface{}{
		"action": "create_new_loot_chest",
		"id":     id,
		"owner":  owner,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send create_new_loot_chest message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) CreateNewLoot(id int, owner string) error {
	msg := map[string]interface{}{
		"action": "create_new_loot",
		"id":     id,
		"owner":  owner,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send create_new_loot message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) UpdateLootChest(id int) error {
	msg := map[string]interface{}{
		"action": "update_loot_chest",
		"id":     id,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send update_loot_chest message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) OtherGrabLoot(lootedFrom string, lootedBy string, silver bool, index int, quantity int) error {
	msg := map[string]interface{}{
		"action":     "other_grab_loot",
		"lootedFrom": lootedFrom,
		"lootedBy":   lootedBy,
		"silver":     silver,
		"index":      index,
		"quantity":   quantity,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send other_grab_loot message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) DetachItemContainer(containerUUID uuid.UUID) error {
	msg := map[string]interface{}{
		"action":        "detach_item_container",
		"containerUUID": containerUUID,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send detach_item_container message: %v\n", err)
	}

	return nil
}

func (c *WebSocketClient) NewSimpleItem(id int, index int, quantity int) interface{} {
	msg := map[string]interface{}{
		"action":   "new_simple_item",
		"id":       id,
		"index":    index,
		"quantity": quantity,
	}

	if err := c.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("Failed to send new_simple_item message: %v\n", err)
	}

	return nil
}
