package kademlia

import (
	"log"
	"net"
	"container/list"
	"net/rpc"
	"strconv"
	"sync"
)

const BUCKETSIZE = 20

type RoutingTable struct {
	NodeID  ID
	Buckets [160]*list.List
	RoutingLock *sync.Mutex
}

func (table *RoutingTable) NewRoutingTable(node ID, RoutingLock *sync.Mutex) {
	table.NodeID      = node
	table.RoutingLock = RoutingLock
	for i := 0; i < 160; i++ {
		table.Buckets[i] = list.New()
	}
}

func (table *RoutingTable) Update(contact Contact) {
	//LOCK
	table.RoutingLock.Lock()
	// updates table if contact is not itself
	if contact.NodeID != table.NodeID {
		prefix_len := (contact.NodeID.Xor(table.NodeID)).PrefixLen()
		var bucket *list.List
		bucket = table.Buckets[prefix_len]
		exist := false
		var temp *list.Element

		// check if contact exists in the table
		for e := (*bucket).Front(); e != nil; e = e.Next() {
			if (*e).Value.(Contact).NodeID.Equals(contact.NodeID) {
				exist = true
				temp = e
				break
			}
		}

		// if contact not in the table, check different conditions
		if exist == false {
			// push contact if bucket if not full
			// O.W. ping first contact
			if (bucket).Len() < BUCKETSIZE { 
				(bucket).PushFront(contact)
			} else {
				//if bucket has been full,ping->remove or discard
				nodeHostTemp := string((bucket).Back().Value.(Contact).Host)
				nodePortTemp := strconv.Itoa(int((bucket).Back().Value.(Contact).Port))
				nodeHostPortTemp := net.JoinHostPort(nodeHostTemp, nodePortTemp)
				//-------------------------------------------------
				client, err := rpc.DialHTTP("tcp", nodeHostPortTemp)
				if err != nil {
					log.Fatal("DialHTTP: ", err)
				}

				ping := new(PingMessage)
				ping.MsgID = NewRandomID()
				var pong PongMessage
				err = client.Call("KademliaCore.Ping", ping, &pong)
				if err != nil {
					log.Fatal("Call: ", err)
				}
				var backTemp *list.Element
				if pong.MsgID != ping.MsgID {
					backTemp := (bucket).Back()
					(bucket).MoveToFront(temp)
					(bucket).Remove(backTemp)
				} else {
					(bucket).MoveToFront(backTemp)
				}

				//-----------------------------------------------------------------------
			}
		} else {
			(*bucket).MoveToFront(temp)

		}

	}
	table.RoutingLock.Unlock()

}
