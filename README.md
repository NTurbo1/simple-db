## B+ Tree Details
### Node
* A Node includes:
    1. A fixed-size header, which contains:
        * The type of the node (leaf or internal).
        * The number of keys.
    2. A list of pointers to child nodes for internal nodes.
    1. A list of KVs pairs.
    1. A list of offsets to KV pairs, which can be used to binary search KVs.
* A Node byte format:
```
| type | nkeys |  pointers  |  offsets   | key-values | unused |
|  2B  |  2B   | nkeys * 8B | nkeys * 2B |    ...     |        |
```
* A KV pair byte format:
```
| klen | vlen | key | val |
|  2B  |  2B  | ... | ... |
```
