package main

import (
	"M00DSWINGS/protocol"
	"M00DSWINGS/protocol/enums"
	"M00DSWINGS/protocol/photon"
	"encoding/hex"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/uuid"
	"log"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

type Logger struct {
	device     pcap.Interface
	events     map[enums.EventType]reflect.Type
	listeners  []func(interface{})
	disconnect []func()
	operations map[enums.OperationType]reflect.Type
	mx         *sync.Mutex
	fragments  *photon.FragmentBuffer
}

func NewLogger(device pcap.Interface) *Logger {
	return &Logger{
		device:     device,
		disconnect: make([]func(), 0),
		listeners:  make([]func(interface{}), 0),
		operations: make(map[enums.OperationType]reflect.Type),
		events:     make(map[enums.EventType]reflect.Type),
		mx:         new(sync.Mutex),
		fragments:  photon.NewFragmentBuffer(),
	}
}

func (e *Logger) ListenAndServe() {
	handle, err := pcap.OpenLive(e.device.Name, 65535, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	if err := handle.SetBPFFilter("udp and (dst port 5056 or src port 5056)"); err != nil {
		log.Fatal(err)
	}

	for _, port := range []int{5055, 5056} {
		layers.RegisterUDPPortLayerType(layers.UDPPort(port), photon.LayerType)
		layers.RegisterTCPPortLayerType(layers.TCPPort(port), photon.LayerType)
	}

	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.NoCopy = true // more performance

	for packet := range packetSource.Packets() {
		if p, ok := packet.Layer(photon.LayerType).(photon.Layer); ok {
			for _, command := range p.Commands {
				e.handleCommand(command)
			}
		}
	}
}

func (e *Logger) handleReliableCommand(cmd *photon.Command) {
	msg, err := cmd.ReliableMessage()
	if err != nil {
		if errors.Is(err, protocol.EncryptionNotSupported) {
			return
		}

		log.Printf("Could not decode reliable message: %v - %v", err, hex.EncodeToString(cmd.Data))
		return
	}

	params, err := photon.DecodeReliableMessage(msg)
	if err != nil {
		log.Printf("Could not decode reliable message: %v - %v", err, hex.EncodeToString(cmd.Data))
		return
	}

	switch msg.Type {
	case photon.OperationRequest, photon.OperationResponse:
		if val, ok := params[253]; ok {
			var opType = enums.OperationType(val.(int16))

			e.handleOperation(opType, params)
		} else {
			log.Printf("ERROR: Could not decode operation: [%d] (%d) (%d) %v", msg.Type,
				msg.ParamaterCount, len(msg.Data),
				hex.EncodeToString(msg.Data))
		}
	case photon.EventDataType:
		if msg.EventCode == 3 {
			params[252] = int16(enums.EventTypeMove)
		}

		if val, ok := params[252]; ok {
			var eventType = enums.EventType(val.(int16))
			e.handleEvent(eventType, params)
		}
	default:
		return
	}
}

func (e *Logger) RegisterListeners(f func(event interface{})) {
	e.mx.Lock()
	defer e.mx.Unlock()

	e.listeners = append(e.listeners, f)
}

func (e *Logger) RegisterOperation(optype enums.OperationType, op interface{}) {
	e.mx.Lock()
	defer e.mx.Unlock()

	e.operations[optype] = reflect.TypeOf(op)
}

func (e *Logger) RegisterEvent(evtype enums.EventType, op interface{}) {
	e.mx.Lock()
	defer e.mx.Unlock()

	e.events[evtype] = reflect.TypeOf(op)
}

func (e *Logger) handleOperation(opType enums.OperationType, params photon.ReliableMessageParamaters) {
	if _, ok := e.operations[opType]; !ok {
		return
	}

	operation := e.operations[opType]
	value := reflect.New(operation).Interface()

	e.updateData(params, value)

	for _, listener := range e.listeners {
		listener(value)
	}
}

func (e *Logger) handleEvent(eventType enums.EventType, params photon.ReliableMessageParamaters) {
	if _, ok := e.events[eventType]; !ok {
		return
	}

	event := e.events[eventType]
	value := reflect.New(event).Interface()

	e.updateData(params, value)

	for _, listener := range e.listeners {
		listener(value)
	}
}

func (e *Logger) updateData(params photon.ReliableMessageParamaters, input any) {
	if params == nil || input == nil {
		return
	}

	valueOf := reflect.ValueOf(input)
	typeOf := reflect.TypeOf(input)

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			log.Println(input, params)
			log.Printf("Input %s data update failed: %v", typeOf.String(), r)
		}
	}()

	val := valueOf.Elem()
	typ := val.Type()

	if !val.IsValid() {
		return
	}

	if typeOf.Kind() == reflect.Func {
		log.Println("Used func instead of object to register event/operation", params)
		return
	}

	if m := valueOf.MethodByName("Decode"); m.IsValid() {
		m.Call([]reflect.Value{reflect.ValueOf(params)})
		return
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		target := typ.Field(i)

		if _, ok := target.Tag.Lookup("albion"); !ok {
			continue
		}

		if !field.CanSet() {
			log.Printf("Cannot set field %s", target.Name)
			continue
		}

		tags := strings.Split(target.Tag.Get("albion"), ",")

		for _, tag := range tags {
			num, err := strconv.Atoi(strings.TrimSpace(tag))
			if err != nil {
				log.Printf("Invalid tag value %s for field %s: %v", tag, target.Name, err)
				continue
			}

			if v, ok := params[uint8(num)]; ok {
				if field.Kind() == reflect.Int64 {
					v = protocol.DecodeInt64(v)
				}
				if field.Kind() == reflect.Int {
					v = protocol.DecodeInteger(v)
				}

				if field.Type() == reflect.TypeOf([]int8{}) {
					v = protocol.DecodeIntegers(v)
				}

				if field.Type() == reflect.TypeOf(uuid.Nil) {
					v = protocol.DecodeCharacterID(v.([]int8))
				}

				if field.Type() == reflect.TypeOf([]uuid.UUID{}) {
					var uui []uuid.UUID

					for _, id := range v.([][]int8) {
						uui = append(uui, protocol.DecodeCharacterID(id))
					}

					v = uui
				}

				if field.Kind() == reflect.String {
					if data, ok := target.Tag.Lookup("not-contains"); ok {
						if strings.Contains(v.(string), data) {
							continue
						}
					}
				}

				value := reflect.ValueOf(v)

				if field.Type() == value.Type() {
					field.Set(value)
					break
				} else {
					log.Printf("Type mismatch for field %s.%s: expected %s, got %s", typ.Name(), target.Name, field.Type(), value.Type())
				}

			}
		}
	}
}

func (e *Logger) handleCommand(command photon.Command) {
	switch command.Type {
	case photon.SendReliableType:
		e.handleReliableCommand(&command)
	case photon.SendUnreliableType:
		var s = make([]byte, len(command.Data)-4)
		copy(s, command.Data[4:])

		command.Data = s
		command.Length -= 4
		command.Type = 6

		e.handleReliableCommand(&command)
	case photon.DisconnectType:
		e.handleDisconnect()
	case photon.SendReliableFragmentType:
		msg, err := command.ReliableFragment()
		if err != nil {
			return
		}

		result := e.fragments.Offer(msg)
		if result != nil {
			e.handleReliableCommand(result)
		}
	}
}

func (e *Logger) RegisterDisconnect(f func()) {
	e.mx.Lock()
	defer e.mx.Unlock()

	e.disconnect = append(e.disconnect, f)
}

func (e *Logger) handleDisconnect() {
	e.mx.Lock()
	defer e.mx.Unlock()

	for _, f := range e.disconnect {
		f()
	}
}
