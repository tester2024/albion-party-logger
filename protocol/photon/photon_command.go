package photon

import (
	"M00DSWINGS/protocol"
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	// AcknowledgeType Command types
	AcknowledgeType          = 1
	ConnectType              = 2
	VerifyConnectType        = 3
	DisconnectType           = 4
	PingType                 = 5
	SendReliableType         = 6
	SendUnreliableType       = 7
	SendReliableFragmentType = 8
	// OperationRequest Message types
	OperationRequest       = 2
	otherOperationResponse = 3
	EventDataType          = 4
	OperationResponse      = 7
)

type Command struct {
	// Header
	Type                   uint8
	ChannelID              uint8
	Flags                  uint8
	ReservedByte           uint8
	Length                 uint32
	ReliableSequenceNumber uint32

	// Body
	Data []byte
}

type ReliableMessage struct {
	// Header
	Signature uint8
	Type      uint8

	// OperationRequest
	OperationCode uint8

	// EventData
	EventCode uint8

	// OperationResponse
	OperationResponseCode uint16
	OperationDebugString  string

	ParamaterCount uint16
	Data           []byte
}

type ReliableFragment struct {
	SequenceNumber uint32
	FragmentCount  int32
	FragmentNumber int32
	TotalLength    int32
	FragmentOffset int32

	Data []byte
}

func (c Command) ReliableMessage() (msg ReliableMessage, err error) {
	if c.Type != SendReliableType {
		return msg, fmt.Errorf("command can't be converted")
	}

	buf := bytes.NewBuffer(c.Data)

	if err := binary.Read(buf, binary.BigEndian, &msg.Signature); err != nil {
		return ReliableMessage{}, err
	}

	if err := binary.Read(buf, binary.BigEndian, &msg.Type); err != nil {
		return ReliableMessage{}, err
	}

	if msg.Type > 128 {
		return msg, protocol.EncryptionNotSupported
	}

	if msg.Type == otherOperationResponse {
		msg.Type = OperationResponse
	}

	switch msg.Type {
	case OperationRequest:
		if err := binary.Read(buf, binary.BigEndian, &msg.OperationCode); err != nil {
			return ReliableMessage{}, err
		}
	case EventDataType:
		if err := binary.Read(buf, binary.BigEndian, &msg.EventCode); err != nil {
			return ReliableMessage{}, err
		}
	case OperationResponse, otherOperationResponse:
		if err := binary.Read(buf, binary.BigEndian, &msg.OperationCode); err != nil {
			return ReliableMessage{}, err
		}
		if err := binary.Read(buf, binary.BigEndian, &msg.OperationResponseCode); err != nil {
			return ReliableMessage{}, err
		}
		var paramType uint8
		if err := binary.Read(buf, binary.BigEndian, &paramType); err != nil {
			return ReliableMessage{}, err
		}
		paramValue, err := decodeType(buf, paramType)
		if err != nil {
			return msg, err
		}

		if paramValue == nil {
			paramValue = ""
		}

		msg.OperationDebugString = paramValue.(string)
	}

	if err := binary.Read(buf, binary.BigEndian, &msg.ParamaterCount); err != nil {
		return ReliableMessage{}, err
	}

	msg.Data = buf.Bytes()

	return
}

func (c Command) ReliableFragment() (msg ReliableFragment, err error) {
	if c.Type != SendReliableFragmentType {
		return msg, fmt.Errorf("command can't be converted")
	}

	buf := bytes.NewBuffer(c.Data)

	if err := binary.Read(buf, binary.BigEndian, &msg.SequenceNumber); err != nil {
		return ReliableFragment{}, err
	}

	if err := binary.Read(buf, binary.BigEndian, &msg.FragmentCount); err != nil {
		return ReliableFragment{}, err
	}

	if err := binary.Read(buf, binary.BigEndian, &msg.FragmentNumber); err != nil {
		return ReliableFragment{}, err
	}

	if err := binary.Read(buf, binary.BigEndian, &msg.TotalLength); err != nil {
		return ReliableFragment{}, err
	}

	if err := binary.Read(buf, binary.BigEndian, &msg.FragmentOffset); err != nil {
		return ReliableFragment{}, err
	}

	msg.Data = buf.Bytes()

	return
}
