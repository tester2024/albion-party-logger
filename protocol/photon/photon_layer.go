package photon

import (
	"bytes"
	"encoding/binary"

	"github.com/google/gopacket"
)

const (
	CommandHeaderLength = 12
)

var LayerType = gopacket.RegisterLayerType(5056,
	gopacket.LayerTypeMetadata{
		Name:    "PhotonLayerType",
		Decoder: gopacket.DecodeFunc(DecodePhotonPacket)})

type Layer struct {
	// Header
	PeerID       uint16
	CrcEnabled   uint8
	CommandCount uint8
	Timestamp    uint32
	Challenge    int32

	// Commands
	Commands []Command

	// Interface stuff
	contents []byte
	payload  []byte
}

func (p Layer) LayerType() gopacket.LayerType { return LayerType }
func (p Layer) LayerContents() []byte         { return p.contents }
func (p Layer) LayerPayload() []byte          { return p.payload }

func DecodePhotonPacket(data []byte, p gopacket.PacketBuilder) error {
	layer := Layer{}
	buf := bytes.NewBuffer(data)

	// Read the header
	if err := binary.Read(buf, binary.BigEndian, &layer.PeerID); err != nil {
		return err
	}

	if err := binary.Read(buf, binary.BigEndian, &layer.CrcEnabled); err != nil {
		return err
	}

	if err := binary.Read(buf, binary.BigEndian, &layer.CommandCount); err != nil {
		return err
	}

	if err := binary.Read(buf, binary.BigEndian, &layer.Timestamp); err != nil {
		return err
	}

	if err := binary.Read(buf, binary.BigEndian, &layer.Challenge); err != nil {
		return err
	}

	var commands []Command

	// Read each command
	for i := 0; i < int(layer.CommandCount); i++ {
		var command Command

		// Command header
		if err := binary.Read(buf, binary.BigEndian, &command.Type); err != nil {
			return err
		}

		if err := binary.Read(buf, binary.BigEndian, &command.ChannelID); err != nil {
			return err
		}

		if err := binary.Read(buf, binary.BigEndian, &command.Flags); err != nil {
			return err
		}

		if err := binary.Read(buf, binary.BigEndian, &command.ReservedByte); err != nil {
			return err
		}

		if err := binary.Read(buf, binary.BigEndian, &command.Length); err != nil {
			return err
		}

		if err := binary.Read(buf, binary.BigEndian, &command.ReliableSequenceNumber); err != nil {
			return err
		}

		// Command data
		dataLength := int(command.Length) - CommandHeaderLength

		// Ensure we don't try to read more than we have
		if dataLength > buf.Len() {
			panic("Data is malformed")
		}

		command.Data = make([]byte, dataLength)
		if _, err := buf.Read(command.Data); err != nil {
			return err
		}

		commands = append(commands, command)
	}

	layer.Commands = commands

	// Split and store the read and unread data
	dataUsed := len(data) - buf.Len()
	layer.contents = data[0:dataUsed]
	layer.payload = buf.Bytes()

	p.AddLayer(layer)
	return p.NextDecoder(gopacket.LayerTypePayload)
}
