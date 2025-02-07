package photon

import (
	lru "github.com/hashicorp/golang-lru"
)

// FragmentBuffer Provides an LRU backed buffer which will assemble ReliableFragments
// into a single PhotonCommand with type ReliableMessage
type FragmentBuffer struct {
	cache *lru.Cache
}

// Offer Offers a message to the buffer. Returns nil when no new commands could be assembled from the
// buffer's contents.
func (buf *FragmentBuffer) Offer(msg ReliableFragment) *Command {
	var entry fragmentBufferEntry

	if buf.cache.Contains(msg.SequenceNumber) {
		obj, _ := buf.cache.Get(msg.SequenceNumber)
		entry = obj.(fragmentBufferEntry)
		entry.Fragments[int(msg.FragmentNumber)] = msg.Data

	} else {
		entry.SequenceNumber = msg.SequenceNumber
		entry.FragmentsNeeded = int(msg.FragmentCount)
		entry.Fragments = make(map[int][]byte)
		entry.Fragments[int(msg.FragmentNumber)] = msg.Data
	}

	if entry.Finished() {
		command := entry.Make()
		buf.cache.Remove(msg.SequenceNumber)
		return &command
	} else {
		buf.cache.Add(msg.SequenceNumber, entry)
		return nil
	}
}

type fragmentBufferEntry struct {
	SequenceNumber  uint32
	FragmentsNeeded int
	Fragments       map[int][]byte
}

func (buf fragmentBufferEntry) Finished() bool {
	return len(buf.Fragments) == buf.FragmentsNeeded
}

func (buf fragmentBufferEntry) Make() Command {
	var data []byte

	for i := 0; i < buf.FragmentsNeeded; i++ {
		data = append(data, buf.Fragments[i]...)
	}

	return Command{
		Type:                   SendReliableType,
		Data:                   data,
		ReliableSequenceNumber: buf.SequenceNumber,
	}
}

// NewFragmentBuffer Makes a new instance of a FragmentBuffer
func NewFragmentBuffer() *FragmentBuffer {
	var f FragmentBuffer
	f.cache, _ = lru.New(128)
	return &f
}
