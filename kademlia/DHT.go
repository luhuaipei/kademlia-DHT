package kademlia

type DHT struct {
	NodeID   ID
	DHTtable map[ID]([]byte)
}

//Initialize the DHT struct
func (table *DHT) Init(node ID) {
	table.NodeID = node
	table.DHTtable = make(map[ID]([]byte))

}

//Insert key/value pairs into the DHT
func (table *DHT) Put(key ID, value []byte) {
	table.DHTtable[key] = value
}

//Delete pairs
func (table *DHT) Del(key ID) {
	delete(table.DHTtable, key)
}

//Find key
func (table *DHT) FindValue(key ID) ([]byte, bool) {
	i, ok := table.DHTtable[key]
	if !ok {
		return nil, ok
	}
	return i, ok
}
