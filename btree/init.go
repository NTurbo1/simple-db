package btree

import (
	"github.com/nturbo1/simple-db/log"
)
func init() {
	node1max := BTREE_NODE_HEADER_SIZE + 8 + 2 + 4 + BTREE_MAX_KEY_SIZE + BTREE_MAX_VAL_SIZE
	if node1max > BTREE_PAGE_SIZE {
		log.Error("The max node size %d exceeds the BTree page size %d", node1max, BTREE_PAGE_SIZE)
		panic("Max node size is too big! I haven't signed for THAT! I'M OUT!")
	}
}
