package kademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"time"
	"sort"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID      ID
	SelfContact Contact
	SelfTable   RoutingTable
	SelfDHT     DHT
	SelfVDOT	VDOT
}

func NewKademlia(laddr string) *Kademlia {
	// TODO: Initialize other state here as you add functionality.
	k := new(Kademlia)
	k.NodeID = NewRandomID()
	kRoutingLock := sync.Mutex{}
	kDHTLock     := sync.Mutex{}
	//initialize routing table
	k.SelfTable.NewRoutingTable(k.NodeID,&kRoutingLock)
	k.SelfDHT.Init(k.NodeID)
	// Set up RPC server
	// NOTE: KademliaCore is just a wrapingpper around Kademlia. This type includes
	// the RPC functions.
	s := rpc.NewServer()
	s.Register(&KademliaCore{k,&kDHTLock})
	_, port,_:=net.SplitHostPort(laddr)
	s.HandleHTTP(rpc.DefaultRPCPath+port, rpc.DefaultDebugPath+port)
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}
	// Run RPC server forever.
	go http.Serve(l, nil)
	// Add self contact
	hostname, port, _ := net.SplitHostPort(l.Addr().String())
	port_int, _ := strconv.Atoi(port)
	ipAddrStrings, err := net.LookupHost(hostname)
	var host net.IP
	for i := 0; i < len(ipAddrStrings); i++ {
		host = net.ParseIP(ipAddrStrings[i])
		if host.To4() != nil {
			break
		}
	}
	k.SelfContact = Contact{k.NodeID, host, uint16(port_int)}
	return k
}

type NotFoundError struct {
	id  ID
	msg string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}
// Given ID, return contact if existing in the routing table
// if ID is self, we return the selfContact
func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
	if nodeId == k.NodeID {
		return &k.SelfContact, nil
	}
	prefix_len := nodeId.Xor(k.NodeID).PrefixLen()
	bucket := k.SelfTable.Buckets[prefix_len]
	exist := false
	var temp *list.Element
	for e := bucket.Front(); e != nil; e = e.Next() {
		if ((*e).Value.(Contact).NodeID).Equals(nodeId) { //* or no *
			exist = true
			temp = e
			break
		}
	}
	if exist {
		var res Contact
		res = (*temp).Value.(Contact)
		return &res, nil
	}
	return nil, &NotFoundError{nodeId, "Not found"}
}

// This is the function to perform the RPC
func (k *Kademlia) DoPing(host net.IP, port uint16) string {
	// TODO: Implement
//	dstHostPort := HostPortGenerator(host, port)
	port_str := strconv.Itoa(int(port))
	client, err := rpc.DialHTTPPath("tcp", host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	defer client.Close()
	ping := new(PingMessage)
	pong := new(PongMessage)
	ping.MsgID = NewRandomID()
	ping.Sender = k.SelfContact
	err = client.Call("KademliaCore.Ping", ping, &pong)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERR:Doping call err"
	}
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	if pong.MsgID == ping.MsgID {
		k.SelfTable.Update(pong.Sender)
		return "OK: PONG"
	} else {
		return "ERR: PONG.MsgID != ping.MsgID"
	}

}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	port_str := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	defer client.Close()
	storeRequest := new(StoreRequest)
	storeResult := new(StoreResult)
	storeRequest.MsgID = NewRandomID()
	storeRequest.Sender = k.SelfContact
	storeRequest.Key = key
	storeRequest.Value = value
	err = client.Call("KademliaCore.Store", storeRequest, &storeResult)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERR: Dostore call err"
	}
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	if storeRequest.MsgID == storeResult.MsgID {
		k.SelfTable.Update(*contact)
		fmt.Println("\"",string(value),"\" stored in \"",contact.NodeID.AsString(),"\"")
		return "OK: Store"
	} else {
		return "ERR: storeRequest.MsgID != storeResult.MsgID"
	}
	return "ERR: Not implemented"
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) (string,FindNodeResult) {
	// TODO: Implement
	// Warining:The recipient must return k triples if at all possible.
	//          It may only return fewer than k if it is returning all of the contacts that it has knowledge of.  -----FIXME
	port_str := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	defer client.Close()
	request := new(FindNodeRequest)
	result := new(FindNodeResult)
	request.MsgID = NewRandomID()
	request.Sender = k.SelfContact
	request.NodeID = searchKey
	err = client.Call("KademliaCore.FindNode", request, &result)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERR:DoFindNode call err",*result
	}
	if request.MsgID == result.MsgID {
		k.SelfTable.Update(*contact)
		//print the contacts
		fmt.Println("--------Closest Nodes to",searchKey.AsString(),"---------")
		for i := 0; i < len(result.Nodes); i++ {
			fmt.Println(result.Nodes[i].NodeID.AsString())
		}
		return "OK : FindNode",*result
	}
	return "ERR: FindNOde->req.MsgID!=res.MsgID",*result
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
}

func (k *Kademlia) DoFindValue(contact *Contact, searchKey ID) (string,FindValueResult){
	// TODO: Implement
	// Warining:The recipient must return k triples if at all possible.
	//          It may only return fewer than k if it is returning all of the contacts that it has knowledge of.  -----FIXME
	port_str := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	defer client.Close()
	request := new(FindValueRequest)
	result := new(FindValueResult)
	request.MsgID = NewRandomID()
	request.Sender = k.SelfContact
	request.Key = searchKey
	err = client.Call("KademliaCore.FindValue", request, &result)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERR:DoFindValue call err",*result
	}
	if request.MsgID == result.MsgID {
		k.SelfTable.Update(*contact)
		if string(result.Value) == "E" {
			fmt.Println("--------Closest Nodes to",searchKey.AsString(),"---------")

			for i := 0; i < len(result.Nodes); i++ {
				fmt.Println(result.Nodes[i].NodeID.AsString())
			}
		} else {		
			fmt.Println("Found Value!",string(result.Value),"AT",contact.NodeID.AsString())
		}
		return "OK: Findvalue",*result
	} else {
		return "ERR: FindValue: req.MsgID != res.MsgID",*result
	}
}

func (k *Kademlia) LocalFindValue(searchKey ID) string {
	// TODO: Implement
	i, ok := k.SelfDHT.FindValue(searchKey)
	if ok {
		return string(i)
	} else {
		return "ERR:Can't find data in DHT"
	}
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	return "ERR: In LocalFindValue "
}

//-----------------------------------------------------------------//
func (k *Kademlia) DoFindNodeIter(contact Contact, searchKey ID, chnl chan FindNodeResult) string {
	// TODO: Implement
	// Warining:The recipient must return k triples if at all possible.
	//          It may only return fewer than k if it is returning all of the contacts that it has knowledge of.  -----FIXME
	port_str := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)
	if err != nil {
		return "DialHTTP:err"
	}
	defer client.Close()
	request := new(FindNodeRequest)
	result := new(FindNodeResult)
	request.MsgID = NewRandomID()
	request.Sender = k.SelfContact
	request.NodeID = searchKey
	err = client.Call("KademliaCore.FindNode", request, &result)
	if err != nil {
		return "ERR:DoFindValue call err"
	}
	if request.MsgID == result.MsgID {
		/* put returned contact into channel */
		result.SenderContact = contact
		chnl <- *result
		return "OK, Found Nodes"
	}
	return "ERR: FindNOde->req.MsgID!=res.MsgID"
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
}
type closest_element struct {
	contact Contact
	flag  int   //0 for default, 1 for requested, 2 for active
	prefix_len int 
}

func (k *Kademlia) DoFindValueIter(contact Contact, searchKey ID, chnl chan FindValueResult) string {
	// 	The recipient must return k triples if at all possible.
	//  It may only return fewer than k if it is returning all of the contacts that it has knowledge of. 
	port_str := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	defer client.Close()
	request := new(FindValueRequest)
	result := new(FindValueResult)
	request.MsgID = NewRandomID()
	request.Sender = k.SelfContact
	request.Key = searchKey
	err = client.Call("KademliaCore.FindValue", request, &result)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERR:DoFindValue call err"
	}
	if request.MsgID == result.MsgID {
		result.SenderContact = contact
		k.SelfTable.Update(contact)
		chnl <- *result 
		return "OK: Findvalue"
	} else {
		return "ERR: FindValue: req.MsgID != res.MsgID"
	}
}


func closest_all_active(closest_list []closest_element) (bool) {
	var ret bool
	ret = true
	if len(closest_list)==0{
		ret = false
	}
	for _, v := range closest_list{
		if v.flag != 2{
			ret = false
			break
		}
	}
	return ret
}

type ByPrefix []closest_element

func (a ByPrefix) Len() int           { return len(a) }

func (a ByPrefix) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (a ByPrefix) Less(i, j int) bool { return a[i].prefix_len < a[j].prefix_len }

func (k *Kademlia) DoIterativeFindNode(id ID) (string,error) {
	/* a dictionary, a short_list, a closest_node, a channel */
	/* Initialization */
	short_list := make(map[ID] bool)
	var closest_list []closest_element
	closest_node_prefix_old := -1
	closest_node_prefix_new := 0
	chnl := make(chan FindNodeResult,70)
	/* loop up local table */
	k.DoFindNodeIter(k.SelfContact, id, chnl)
	/* wait for 300ms */
	time.Sleep(time.Millisecond * 300)
	/* break loop until one of the condition is matched */
	for !(closest_node_prefix_old >= closest_node_prefix_new || closest_all_active(closest_list)){
		closest_node_prefix_old = closest_node_prefix_new
		/* Recieve all responds from channel */
		chnl_len := len(chnl)
		for i := 0; i < chnl_len; i++ {
			v := <- chnl
			/* Updates local table */
			k.SelfTable.Update(v.SenderContact)
			/* Label repsond node */
			for i,u := range closest_list{
				if u.contact.NodeID==v.SenderContact.NodeID{
					closest_list[i].flag = 2
					break
				}
			}
			for _, j := range v.Nodes{
				/* Updates closest_list */
				if !short_list[j.NodeID]{
					this_prefix_len := id.Xor(j.NodeID).PrefixLen()
					this_closest_element := closest_element{j, 0, this_prefix_len}
					closest_list = append(closest_list,this_closest_element)	
				}				
				/* Updates short_list */
				short_list[j.NodeID] = true
			}
		}
		/* Remove all the inactive(flag=1) contact in closest_list*/
//		var temp_closest_list []closest_element 
//		for _,val := range closest_list{
//			if val.flag != 1 {
//				temp_closest_list=append(temp_closest_list,val)
//			}
//		}
//		closest_list = temp_closest_list
		/* Cut len(closest_list) to 20 and closest_node */
		if len(closest_list)!=0{
			sort.Sort(ByPrefix(closest_list))
			/* closest_node */
			closest_node_prefix_new = closest_list[0].prefix_len
		}
//		if len(closest_list)<20{
//			closest_list=closest_list[:]
//		}else{	
//			closest_list=closest_list[:20]
//		}
		/* Pick 3 contact which is "default". Pack up 3 RPC requests */ 
		counter:=0
		for i,v := range closest_list{
			if v.flag==0{
				if counter==3{
					break
				}
				closest_list[i].flag = 1
				go k.DoFindNodeIter(v.contact, id, chnl)
				counter++
			}
		}
		/* wait for 300ms */
		time.Sleep(time.Millisecond * 300)
	}
	/* clean up channel and return result */
	var result_array string
	result_array=""
	for _,val:= range closest_list{
		result_array += (val.contact.NodeID.AsString()+"\n")
	}
	fmt.Println(result_array)
	return result_array,nil
}
func (k *Kademlia) DoIterativeStore(key ID, value []byte) (string,error) {	
	//This function first calls iterativeFindNode() and receives a set of k triples. A STORE rpc is sent to each of these Contacts.
	// It returns a string containing the ID of the node that received the final STORE operation

	short_list := make(map[ID] bool)
	var closest_list []closest_element
	closest_node_prefix_old := -1
	closest_node_prefix_new := 0
	chnl := make(chan FindNodeResult,70)
	/* loop up local table */
	go k.DoFindNodeIter(k.SelfContact, key, chnl)
	/* wait for 300ms */
	time.Sleep(time.Millisecond * 300)
	/* break loop until one of the condition is matched */
	for !(closest_node_prefix_old >= closest_node_prefix_new || closest_all_active(closest_list)){
		closest_node_prefix_old = closest_node_prefix_new
		/* Recieve all responds from channel */
		chnl_len := len(chnl)
		for i := 0; i < chnl_len; i++ {
			v := <- chnl
			/* Updates local table */
			k.SelfTable.Update(v.SenderContact)
			/* Label repsond node */
			for i,u := range closest_list{
				if u.contact.NodeID==v.SenderContact.NodeID{
					closest_list[i].flag = 2
					break
				}
			}

			for _, j := range v.Nodes{
				/* Updates closest_list */
				if !short_list[j.NodeID]{
					this_prefix_len := key.Xor(j.NodeID).PrefixLen()
					this_closest_element := closest_element{j, 0, this_prefix_len}
					closest_list = append(closest_list,this_closest_element)	
				}				
				/* Updates short_list */
				short_list[j.NodeID] = true
			}
		}
		/* Remove all the inactive(flag=1) contact in closest_list*/
//		var temp_closest_list []closest_element 
//		for _,val := range closest_list{
//			if val.flag != 1 {
//				temp_closest_list=append(temp_closest_list,val)
//			}
//		}
//		closest_list = temp_closest_list
		/* Cut len(closest_list) to 20 and closest_node */
		if len(closest_list)!=0{
		sort.Sort(ByPrefix(closest_list))
		/* closest_node */
		closest_node_prefix_new = closest_list[0].prefix_len
		}
//		if len(closest_list)<20{
//			closest_list=closest_list[:]
//		}else{	
//			closest_list=closest_list[:20]
//		}
		
		/* Pick 3 contact which is "default". Pack up 3 RPC requests */ 
		counter:=0
		for i,v := range closest_list{
			if v.flag==0{
				if counter==3{
					break
				}
				closest_list[i].flag = 1
				go k.DoFindNodeIter(v.contact, key, chnl)
				counter++
			}
		}
		/* wait for 300ms */
		time.Sleep(time.Millisecond * 300)
	}
	var result_array string
	result_array=""
	//for _,val := range closest_list{
	//k.DoStore(&val.contact,key,value)
	k.DoStore(&closest_list[0].contact, key, value)
	result_array += (closest_list[0].contact.NodeID.AsString()+"\n")		
	//}
	return result_array,nil
}
func (k *Kademlia) DoIterativeFindValue(key ID) (string,error) {
	/* a dictionary, a short_list, a closest_node, a channel */
	/* Initialization */
	short_list := make(map[ID] bool)
	var closest_list []closest_element
	chnl := make(chan FindValueResult,70)
	/* loop up local table */
	go k.DoFindValueIter(k.SelfContact, key, chnl)
	/* wait for 300ms */
	time.Sleep(time.Millisecond * 300)
	haveFind:=false
	var res FindValueResult
	/* break loop until one of the condition is matched */
	for !(closest_all_active(closest_list)|| haveFind){   
		/* Recieve all responds from channel */
		chnl_len := len(chnl)
		for i := 0; i < chnl_len; i++ {
			v := <- chnl
			if string(v.Value) != "E" {
				haveFind = true
				res = v
			}
			/* Updates local table */
			k.SelfTable.Update(v.SenderContact)
			/* Label repsond node */

			for i,u := range closest_list{
				if u.contact.NodeID==v.SenderContact.NodeID{
					closest_list[i].flag = 2
					break
				}
			}
			for _, j := range v.Nodes{
				/* Updates closest_list */
				if !short_list[j.NodeID]{
					this_prefix_len := key.Xor(j.NodeID).PrefixLen()
					this_closest_element := closest_element{j, 0, this_prefix_len}
					closest_list = append(closest_list,this_closest_element)	
				}				
				/* Updates short_list */
				short_list[j.NodeID] = true
			}
		}
		/* Remove all the inactive(flag=1) contact in closest_list*/
//		var temp_closest_list []closest_element 
//		for _,val := range closest_list{
//			if val.flag != 1 {
//				temp_closest_list=append(temp_closest_list,val)
//			}
//		}
//		closest_list = temp_closest_list
		/* Cut len(closest_list) to 20 and closest_node */
		if len(closest_list)!=0{
		sort.Sort(ByPrefix(closest_list))
		}
	//	if len(closest_list)<20{
	//		closest_list=closest_list[:]
	//	}else{	
	//		closest_list=closest_list[:20]
	//	}
		/* closest_node */
		/* Pick 3 contact which is "default". Pack up 3 RPC requests */ 
		counter:=0
		for i,v := range closest_list{
			if v.flag==0{
				if counter==3{
					break
				}
				closest_list[i].flag = 1
				go k.DoFindValueIter(v.contact, key, chnl)
				counter++
			}
		}
		/* wait for 300ms */
		time.Sleep(time.Millisecond * 300)
	}
	if haveFind{
		for _,val := range closest_list{
			if val.flag==2 && val.contact.NodeID!=res.SenderContact.NodeID{
				k.DoStore(&val.contact,key,res.Value)
				fmt.Println("store in node:",val.contact.NodeID.AsString())
				//return res.SenderContact.NodeID.AsString()+" "+ string(res.Value),nil
				return string(res.Value), nil
			}
		}
	}
	/* return result */
	var result_array string
	result_array=""
	for _,val:= range closest_list{
		result_array += (val.contact.NodeID.AsString()+"\n")
	}
	return "ERR",nil
}

func HostPortGenerator(Host net.IP, Port uint16) string {

	hostTemp := Host.String()

	portTemp := strconv.Itoa(int(Port))

	return net.JoinHostPort(hostTemp, portTemp)
}



func (k *Kademlia) DoGetVDO(contact Contact, vdoId ID) (string,error){
	port_str := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	defer client.Close()
	req := new(GetVDORequest)
	res := new(GetVDOResult)
	req.Sender = k.SelfContact
	req.MsgID = NewRandomID()
	req.VdoID = vdoId
	err = client.Call("KademliaCore.GetVDO", req, &res)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERR:Doping call err", nil 
	}
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	if req.MsgID == res.MsgID {
		k.SelfTable.Update(res.SenderContact)
		secret := UnvanishData(k, res.VDO)
		return string(secret), nil
	} else {
		return "ERR: req.MsgID != res.MsgID", nil
	}


}
