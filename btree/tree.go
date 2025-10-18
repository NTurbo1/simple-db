package btree

import (
	"encoding/binary"
)

type BTree struct {
	root uint64 // pointer (a nonzero page number)

	// callbacks for managing on-disk pages
	get func(uint64) []byte // deference a pointer
	new func([]byte) uint64 // allocate a new page
	del func(uint64) // deallocate a page
}

type BNode []byte

func (node BNode) nodeType() uint16 {
	return binary.LittleEndian.Uint16(node[0:2])
}

func (node BNode) numKeys() uint16 {
	return binary.LittleEndian.Uint16(node[2:4])
}

func (node BNode) setHeader(nodeType, nKeys uint16) {
	binary.LittleEndian.PutUint16(node[0:2], nodeType)
	binary.LittleEndian.PutUint16(node[2:4], nKeys)
}

func (node BNode) getPtr(index uint16) (uint64, error) {
	if err := validateIndex(index, node); err != nil {
		return 0, err
	}
	ptrPos := BTREE_NODE_POINTERS_OFFSET + (index * 8)
	return binary.LittleEndian.Uint64(node[ptrPos:]), nil
}

func (node BNode) setPtr(index uint16, val uint64) error {
	if err := validateIndex(index, node); err != nil {
		return err
	}
	ptrPos := BTREE_NODE_POINTERS_OFFSET + (index * 8)
	binary.LittleEndian.PutUint64(node[ptrPos:], val)

	return nil
}

func (node BNode) getOffset(index uint16) (uint16, error) {
	if index == 0 {
		return 0, nil
	}

	pos, err := offsetPos(node, index)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(node[pos:]), nil
}

func (node BNode) setOffset(index uint16, offset uint16) error {
	var pos uint16 = 0
	var err error
	if index > 0 {
		pos, err = offsetPos(node, index)
		if err != nil {
			return err
		}
	}
	binary.LittleEndian.PutUint16(node[pos:], offset)

	return nil
}

func (node BNode) kvPos(index uint16) (uint16, error) {
	kvOffset, err := node.getOffset(index)
	if err != nil {
		return 0, err
	}

	return BTREE_NODE_POINTERS_OFFSET + (8 * node.numKeys()) + (2 * node.numKeys()) + kvOffset, nil
}

func (node BNode) getKey(index uint16) ([]byte, error) {
	pos, err := node.kvPos(index)
	if err != nil {
		return nil, err
	}
	klen := binary.LittleEndian.Uint16(node[pos:])

	pos += 4 // skip the klen and vlen bytes
	return node[pos:pos + klen], nil
}

func (node BNode) getVal(index uint16) ([]byte, error) {
	pos, err := node.kvPos(index)
	if err != nil {
		return nil, err
	}

	klen := binary.LittleEndian.Uint16(node[pos:])
	vlen := binary.LittleEndian.Uint16(node[pos + 2:])

	pos += (4 + klen) // skip the klen, vlen, and key bytes
	return node[pos:pos + vlen], nil
}

// Size of the node in bytes
func (node BNode) size() (uint16, error) {
	lastKVPos, err := node.kvPos(node.numKeys() - 1)
	if err != nil {
		return 0, err
	}
	klen := binary.LittleEndian.Uint16(node[lastKVPos:])
	vlen := binary.LittleEndian.Uint16(node[lastKVPos + 2:])
	kvlen := klen + vlen // !!! RISK OF INTEGER OVERFLOW !!!

	return lastKVPos + 4 + kvlen, nil
}
