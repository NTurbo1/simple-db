package btree

// NODE BYTE FORMAT
// | type | nkeys |  pointers  |  offsets   | key-values | unused |
// |  2B  |  2B   | nkeys * 8B | nkeys * 2B |    ...     |        |

// KV pair byte format
// | klen | vlen | key | val |
// |  2B  |  2B  | ... | ... |

const (
	BTREE_NODE_HEADER_SIZE = 4
	BTREE_PAGE_SIZE = 4096
	BTREE_MAX_KEY_SIZE = 1000
	BTREE_MAX_VAL_SIZE = 3000

	BTREE_NODE_POINTERS_OFFSET = 4

	BNODE_LEAF_TYPE = uint16(1)
	BNODE_INTERNAL_TYPE = uint16(0)
)
