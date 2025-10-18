package btree

import (
	"fmt"
)

func validateIndex(index uint16, node BNode) error {
	numKeys := node.numKeys()
	if (index >= numKeys) {
		return fmt.Errorf(
			"Index %d is out of range for a number of keys %d", index, numKeys,
		)
	}

	return nil
}

func offsetPos(node BNode, index uint16) (uint16, error) {
	if err := validateIndex(index, node); err != nil {
		return 0, err
	}
	if index < 1 {
		return 0, fmt.Errorf(
			"You probably forgot to check that the index >= 1 " + 
			"before passing it to the offsetPos function.",
		)
	}
	
	return BTREE_NODE_POINTERS_OFFSET + (8 * node.numKeys()) + 2 * (index - 1), nil
}
