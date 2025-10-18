package btree

import (
	"encoding/binary"
	"bytes"
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



// Returns the first kid node whose range intersects the key. (kid[i] <= key)
//
// TODO: binary search
//
// NOTE: The function is called nodeLookupLE because it uses the Less-than-or-Equal operator. For
// point queries, we should use the equal operator instead, which is a step we can add later.
func nodeLookupLE(node BNode, key []byte) (uint16, error) {
	numKeys := node.numKeys()
	var found uint16 = 0

	// The first key is a copy from the parent node,
	// thus it's always less than or equal to the key
	for i := uint16(1); i < numKeys; i++ {
		foundKey, err := node.getKey(i)
		if err != nil {
			return 0, err
		}
		cmpRes := bytes.Compare(foundKey, key)
		if cmpRes <= 0 {
			found = i
		}
		if cmpRes >= 0 {
			break
		}
	}

	return found, nil
}

// Adds a new key to a leaf node
func leafInsert(newNode, oldNode BNode, index uint16, key, val []byte) error {
	newNode.setHeader(BNODE_LEAF_TYPE, oldNode.numKeys() + 1)
	err := nodeAppendKVRange(newNode, oldNode, 0, 0, index)
	if err != nil {
		return err
	}
	err = nodeAppendKV(newNode, index, 0, key, val)
	if err != nil {
		return err
	}
	err =nodeAppendKVRange(newNode, oldNode, index + 1, index, oldNode.numKeys() - index)
	if err != nil {
		return err
	}

	return nil
}

// Copies a KV into a position.
func nodeAppendKV(newNode BNode, index uint16, ptr uint64, key, val []byte) error {
	if err := newNode.setPtr(index, ptr); err != nil {
		return err
	}
	kvPos, err := newNode.kvPos(index)
	if err != nil {
		return err
	}

	klen, vlen := uint16(len(key)), uint16(len(val)) // RISK OF INTEGER OVERFLOW AND SILENT TRUNCATION !!!
	binary.LittleEndian.PutUint16(newNode[kvPos:], klen)
	binary.LittleEndian.PutUint16(newNode[kvPos + 2:], vlen)
	copy(newNode[kvPos + 4:], key)
	copy(newNode[kvPos + 4 + klen:], val)
	offsetAtIdx, err := newNode.getOffset(index)
	if err != nil {
		return err
	}
	nextKVOffset := offsetAtIdx + 4 + uint16(klen + vlen) 	// RISK OF OVERFLOW!!! 
															// JUST SAYING, BUD...
	
	// Setting the offset for the next KV
	// TODO: IDK WHY WE'RE DOING IT RIGHT NOW BUT MAKE SURE TO CHECK IT LATER !!!
	// YOU MAY RUN INTO A BUG BECAUSE OF THIS LATER !!!
	if err := newNode.setOffset(index + 1, nextKVOffset); err != nil {
		return err
	}

	return nil
}

// Copies multiple KVs into a position from an old node.
func nodeAppendKVRange(newNode, oldNode BNode, dstNew, srcOld, n uint16) error {
	for i := uint16(0); i < n; i++ {
		oldIndex := uint16(srcOld + i)
		ptr, err := oldNode.getPtr(oldIndex)
		if err != nil {
			return err
		}
		key, err := oldNode.getKey(oldIndex)
		if err != nil {
			return err
		}
		val, err := oldNode.getVal(oldIndex)
		if err != nil {
			return err
		}

		return nodeAppendKV(newNode, dstNew + i, ptr, key, val)
	}

	return nil
}

// Replaces a link with one or multiple links
func nodeReplaceKidN(bTree *BTree, newNode, oldNode BNode, index uint16, kids ...BNode) error {
	inc := uint16(len(kids)) // RISK OF INTEGER OVERFLOW AND SILENT TRUNCATION !!!
	newNode.setHeader(BNODE_INTERNAL_TYPE, oldNode.numKeys() + inc - 1)
	err := nodeAppendKVRange(newNode, oldNode, 0, 0, index)
	if err != nil {
		return err
	}
	for i, node := range kids {
		firstKey, err := node.getKey(0)
		if err != nil {
			return err
		}
		err = nodeAppendKV(newNode, index + uint16(i), bTree.new(node), firstKey, nil)
		if err != nil {
			return err
		}
	}
	err = nodeAppendKVRange(newNode, oldNode, index + inc, index + 1, oldNode.numKeys() - (index + 1))
	if err != nil {
		return err
	}

	return nil
}
