package kademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	"net"
	"sync"
	"container/list"
)

type KademliaCore struct {
	kademlia   *Kademlia
	serverDHTLock *sync.Mutex
}

// Host identification.
type Contact struct {
	NodeID ID
	Host   net.IP
	Port   uint16
}

///////////////////////////////////////////////////////////////////////////////
// PING
///////////////////////////////////////////////////////////////////////////////
type PingMessage struct {
	Sender Contact
	MsgID  ID
}

type PongMessage struct {
	MsgID  ID
	Sender Contact
}

func (kc *KademliaCore) Ping(ping PingMessage, pong *PongMessage) error {
	// lock
//	kc.serverLock.Lock()
	// TODO: Finish implementation
	pong.MsgID = CopyID(ping.MsgID)
	// Specify the sender
	pong.Sender = kc.kademlia.SelfContact
	// Update contact, etc
	if ping.Sender.NodeID != pong.Sender.NodeID {
		kc.kademlia.SelfTable.Update(ping.Sender)
	}
	// Unlock
//	kc.serverLock.Unlock()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STORE
///////////////////////////////////////////////////////////////////////////////
type StoreRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
	Value  []byte
}

type StoreResult struct {
	MsgID ID
	Err   error
}

func (kc *KademliaCore) Store(req StoreRequest, res *StoreResult) error {
	// lock
//	kc.serverLock.Lock()
	// TODO: Implement.
	res.MsgID = req.MsgID
	key := req.Key
	value := req.Value
    kc.serverDHTLock.Lock()
	kc.kademlia.SelfDHT.Put(key, value)
	kc.serverDHTLock.Unlock()
	kc.kademlia.SelfTable.Update(req.Sender)
	// Unlock
//	kc.serverLock.Unlock()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_NODE
///////////////////////////////////////////////////////////////////////////////
type FindNodeRequest struct {
	Sender Contact
	MsgID  ID
	NodeID ID
}

type FindNodeResult struct {
	SenderContact Contact
	MsgID ID
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	// lock
//	kc.serverLock.Lock()
	res.MsgID = req.MsgID
	// update local rountig table
	kc.kademlia.SelfTable.Update(req.Sender)

	// delcare counter, res is returned when coutner reaches 20
	counter := 0
	bucketDist := 1
	prefix_len := req.NodeID.Xor(kc.kademlia.SelfTable.NodeID).PrefixLen()
	var bucket *list.List

	if prefix_len != 160 {
		bucket = kc.kademlia.SelfTable.Buckets[prefix_len]
	} else {
		bucket = kc.kademlia.SelfTable.Buckets[159]
	}

	// iterate bucket, add element to Nodes[]
	for e := bucket.Front(); e != nil; e = e.Next() {
		res.Nodes = append(res.Nodes, e.Value.(Contact))
		counter += 1
	}
	if counter == 20 {
		// Unlock
//		kc.serverLock.Unlock()
		return nil
	}

	// find nearby buckets until find 20 nodes or finish all 
	flag1 := true
	flag2 := true
	for flag1 || flag2 {
		if prefix_len+bucketDist >= 160 {
			flag1 = false
		}
		// search upward buckets
		if flag1 == true {
			bucket1 := kc.kademlia.SelfTable.Buckets[prefix_len+bucketDist]
			for e := bucket1.Front(); e != nil; e = e.Next() {
				res.Nodes = append(res.Nodes, e.Value.(Contact))
				counter += 1
				if counter == 20 {
					// Unlock
//					kc.serverLock.Unlock()
					return nil
				}
			}

		}
		if prefix_len-bucketDist < 0 {
			flag2 = false
		}
		// search downward buckets
		if flag2 == true {
			bucket2 := kc.kademlia.SelfTable.Buckets[prefix_len-bucketDist]
			for e := bucket2.Front(); e != nil; e = e.Next() {
				res.Nodes = append(res.Nodes, e.Value.(Contact))
				counter += 1
				if counter == 20 {
					// Unlock
	//				kc.serverLock.Unlock()
					return nil
				}
			}
		}
		bucketDist += 1
	}
	// Unlock
//	kc.serverLock.Unlock()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_VALUE
///////////////////////////////////////////////////////////////////////////////
type FindValueRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
}

// If Value is nil, it should be ignored, and Nodes means the same as in a
// FindNodeResult.
type FindValueResult struct {
	SenderContact Contact
	MsgID ID
	Value []byte
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindValue(req FindValueRequest, res *FindValueResult) error {
	// lock
	//kc.serverLock.Lock()
	// TODO: Implement.
	res.MsgID = req.MsgID
	kc.kademlia.SelfTable.Update(req.Sender)
	key := req.Key
	res.SenderContact=kc.kademlia.SelfContact
	ok := false
	// find value
    kc.serverDHTLock.Lock()
	res.Value, ok = kc.kademlia.SelfDHT.FindValue(key)
	kc.serverDHTLock.Unlock()

	// if value is not find, find nearest nodes
	if !ok {
		//Find closest node,return k triples
		res.Value = []byte("E")
		counter := 0
		bucketDist := 1
		prefix_len := req.Key.Xor(kc.kademlia.SelfTable.NodeID).PrefixLen()
		var bucket *list.List

		if prefix_len != 160 {
			bucket = kc.kademlia.SelfTable.Buckets[prefix_len]
		} else {
			bucket = kc.kademlia.SelfTable.Buckets[159]
		}

		//iterate bucket,add element to Nodes[]
		for e := bucket.Front(); e != nil; e = e.Next() {
			res.Nodes = append(res.Nodes, e.Value.(Contact))
			counter += 1
		}
		if counter == 20 {
			// Unlock
	//		kc.serverLock.Unlock()
			return nil
		}
		flag1 := true
		flag2 := true
		for flag1 || flag2 {
			if prefix_len+bucketDist >= 160 {
				flag1 = false
			}
			if flag1 == true {
				bucket1 := kc.kademlia.SelfTable.Buckets[prefix_len+bucketDist]
				for e := bucket1.Front(); e != nil; e = e.Next() {
					res.Nodes = append(res.Nodes, e.Value.(Contact))
					counter += 1
					if counter == 20 {
						// Unlock
	//					kc.serverLock.Unlock()
						return nil
					}
				}

			}
			if prefix_len-bucketDist < 0 {
				flag2 = false
			}
			if flag2 == true {
				bucket2 := kc.kademlia.SelfTable.Buckets[prefix_len-bucketDist]
				for e := bucket2.Front(); e != nil; e = e.Next() {
					res.Nodes = append(res.Nodes, e.Value.(Contact))
					counter += 1
					if counter == 20 {
						// Unlock
	//					kc.serverLock.Unlock()
						return nil
					}
				}
			}
			bucketDist += 1
		}

	}
	// Unlock
//	kc.serverLock.Unlock()
	return nil
}
type GetVDORequest struct {
    Sender Contact
    MsgID ID
    VdoID ID
}

type GetVDOResult struct{
    SenderContact Contact
    MsgID ID
    VDO VanashingDataObject
}

func (kc *KademliaCore) GetVDO(req GetVDORequest, res *GetVDOResult) error {
    // fill in
    kc.kademlia.SelfTable.Update(req.Sender)
	res.MsgID = req.MsgID
	res.SenderContact = kc.kademlia.SelfContact
    i,_ := kc.kademlia.SelfVDOT.Get(req.VdoID)
    if i != nil{
    	res.VDO = *i
    } 
    return nil
}