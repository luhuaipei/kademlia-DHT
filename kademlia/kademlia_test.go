package kademlia

import (
	"net"
	"strconv"
	"testing"
    "time"
	"strings"
	"math"
)
var N=50 //number of node in the network   
var CONCURRENCY = 5 //number of doping rpc
var TimeStrap = time.Millisecond*100     //takes about 20+ second for testing a function
func StringToIpPort(laddr string) (ip net.IP, port uint16, err error) {
	hostString, portString, err := net.SplitHostPort(laddr)
	if err != nil {
		return
	}
	ipStr, err := net.LookupHost(hostString)
	if err != nil {
		return
	}
	for i := 0; i < len(ipStr); i++ {
		ip = net.ParseIP(ipStr[i])
		if ip.To4() != nil {
			break
		}
	}
	portInt, err := strconv.Atoi(portString)
	port = uint16(portInt)
	return
}

func TestPing(t *testing.T) {
	instance1 := NewKademlia("localhost:7890")
	instance2 := NewKademlia("localhost:7891")
	host2, port2, _ := StringToIpPort("localhost:7891")
	//    t.Log(host2,port2)
	instance1.DoPing(host2, port2)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	contact1, err := instance2.FindContact(instance1.NodeID)
	if err != nil {
		t.Error("Instance 1's contact not found in Instance 2's contact list")
		return
	}
	if contact1.NodeID != instance1.NodeID {
		t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
	}
	if contact2.NodeID != instance2.NodeID {
		t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
	}
	return
}

func TestStore(t *testing.T) {
	instance1 := NewKademlia("localhost:7892")
	instance2 := NewKademlia("localhost:7893")
	host2, port2, _ := StringToIpPort("localhost:7892")
	instance1.DoPing(host2, port2)
	key, err := IDFromString("aaaa")
	if err != nil {
		t.Error([]byte("ERR: IDFromString"))
		return
	}
	res1 := instance1.DoStore(&instance2.SelfContact, key, []byte("key:aaaa value:testDoStore"))
	if res1 != "OK: Store" {
		t.Error([]byte(res1))
		return
	}
	res2 := instance2.LocalFindValue(key)
	if res2 != "key:aaaa value:testDoStore" {
		t.Error([]byte(res2))
		return
	}
	return
}

func TestFindContact(t *testing.T) {
	instance1 := NewKademlia("localhost:7886")
	instance2 := NewKademlia("localhost:7887")
	host2, port2, _ := StringToIpPort("localhost:7887")
	//    t.Log(host2,port2)
	instance1.DoPing(host2, port2)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	contact1, err := instance2.FindContact(instance1.NodeID)
	if err != nil {
		t.Error("Instance 1's contact not found in Instance 2's contact list")
		return
	}
	if contact1.NodeID != instance1.NodeID {
		t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
	}
	if contact2.NodeID != instance2.NodeID {
		t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
	}
	return
}

func TestLocalFindValue(t *testing.T) {
	instance1 := NewKademlia("localhost:7888")
	instance2 := NewKademlia("localhost:7889")
	host2, port2, _ := StringToIpPort("localhost:7888")
	instance1.DoPing(host2, port2)
	key, err := IDFromString("aaaa")
	if err != nil {
		t.Error([]byte("ERR: IDFromString"))
		return
	}
	res1 := instance1.DoStore(&instance2.SelfContact, key, []byte("Testing"))
	if res1 != "OK: Store" {
		t.Error([]byte(res1))
		return
	}
	res2 := instance2.LocalFindValue(key)
	if res2 != "Testing" {
		t.Error([]byte(res2))
		return
	}
	return
}
func pingWithChan(node *Kademlia, host net.IP, port uint16, chnl chan int) {
	node.DoPing(host, port)
	chnl <- 1
	return
}
func TestFindNode(t *testing.T) {
	num_of_node := N
	var node []*Kademlia
	initPort := 7894
	for i := 0; i < num_of_node; i++ {
		port := strconv.Itoa(initPort + i)
		str_port := "localhost:" + port
		node = append(node, NewKademlia(str_port))
	}
	chnl := make(chan int, CONCURRENCY)
	for c := 0; c < CONCURRENCY; c++ {
		chnl <- 1
	}

	//TODO: use a more realistic network layout
	port:=""
	for j := 0; j < num_of_node; j++ {
		for k := 0; k <= 20; k += 1 {
			port_temp := 2*k
			if int(port_temp)+j >= num_of_node {
				port = strconv.Itoa(initPort+(j+int(port_temp))%(num_of_node))
			} else {
				port = strconv.Itoa(initPort +j+ int(port_temp))
			}
			targetHost, targetPort, _:= StringToIpPort("localhost:" + port)
			select {
			case <-chnl:
	            time.Sleep(time.Millisecond*30)
				go pingWithChan(node[j], targetHost, targetPort, chnl)
			}
		}
	}
	
	res1, _ := node[0].DoFindNode(&node[num_of_node-2].SelfContact, node[num_of_node-1].NodeID)
	if res1 != "OK : FindNode" {
		t.Error([]byte(res1))
		return	
	}
}
func TestFindValue(t *testing.T) {
	num_of_node := N
	var node []*Kademlia
	initPort := 8500
	for i := 0; i < num_of_node; i++ {
		port := strconv.Itoa(initPort + i)
		str_port := "localhost:" + port
		node = append(node, NewKademlia(str_port))
	}
	//TODO: use a more realistic network layout
	chnl := make(chan int, CONCURRENCY)
	for c := 0; c < CONCURRENCY; c++ {
		chnl <- 1
	}

	//TODO: use a more realistic network layout
	port:=""
	for j := 0; j < num_of_node; j++ {
		for k := 0; k <= int(math.Floor(math.Log2(float64(num_of_node)))); k += 1 {
			port_temp := math.Pow(2, float64(k))
			if int(port_temp)+j >= num_of_node {
				port = strconv.Itoa(initPort+(j+int(port_temp))%(num_of_node))
			} else {
				port = strconv.Itoa(initPort +j+ int(port_temp))
			}
			targetHost, targetPort, _:= StringToIpPort("localhost:" + port)
			select {
			case <-chnl:
	            time.Sleep(TimeStrap)
				go pingWithChan(node[j], targetHost, targetPort, chnl)
			}
		}
	}
	//Store a data in a node
	nodeStoreData := 1
	key, err := IDFromString("1111111111111111111111111111111111111111")
	if err != nil {
		t.Error([]byte("ERR: IDFromString"))
		return
	}
	node[0].DoStore(&node[nodeStoreData].SelfContact, key, []byte("TESTING DATA "))
	//FindValue instance 1
	_, res2 := node[0].DoFindValue(&node[nodeStoreData].SelfContact, key)
	if string(res2.Value) == "E" {
		t.Error([]byte("Err: FindNode, instance1, can't find value"))
		return
	}

	//  t.Log("Found Value", string(res2.Value),"in", res2.SenderContact.NodeID.AsString(),"   PASS")
	_, res2 = node[0].DoFindValue(&node[0].SelfContact, key)
	if string(res2.Value) != "E" {
		t.Error([]byte("Err: FindNode, instance1, can't find value"))
		return
	}
	// t.Log("-----------------DoFindValue Result--Case 2:Not Found----------")
	//for i := 0; i < len(res2.Nodes); i++ {
	//   t.Log("     ",i,":",res2.Nodes[i].NodeID.AsString())
	// }
	return
}
func TestIterativeFindNode(t *testing.T) {
	//Initialization
	num_of_node := N
	var node []*Kademlia
	initPort := 9500
	for i := 0; i < num_of_node; i++ {
		port := strconv.Itoa(initPort + i)
		str_port := "localhost:" + port
		node = append(node, NewKademlia(str_port))
	}
	//TODO: use a more realistic network layout
	chnl := make(chan int, CONCURRENCY)
	for c := 0; c < CONCURRENCY; c++ {
		chnl <- 1
	}

	//TODO: use a more realistic network layout
	port:=""
	for j := 0; j < num_of_node; j++ {
		for k := 0; k <= int(math.Floor(math.Log2(float64(num_of_node)))); k += 1 {
			port_temp := math.Pow(2, float64(k))
			if int(port_temp)+j >= num_of_node {
				port = strconv.Itoa(initPort+(j+int(port_temp))%(num_of_node))
			} else {
				port = strconv.Itoa(initPort +j+ int(port_temp))
			}
			targetHost, targetPort, _:= StringToIpPort("localhost:" + port)
			select {
			case <-chnl:
	            time.Sleep(TimeStrap)
				go pingWithChan(node[j], targetHost, targetPort, chnl)
			}
		}
	}
	_, err := node[0].DoIterativeFindNode(node[0].NodeID)
	if err != nil {
		t.Error(err)
	}

	return
}

func TestIterativeStore(t *testing.T) {
	//Initialization
	num_of_node := N
	var node []*Kademlia
	initPort := 10000
	for i := 0; i < num_of_node; i++ {
		port := strconv.Itoa(initPort + i)
		str_port := "localhost:" + port
		node = append(node, NewKademlia(str_port))
	}
	//TODO: use a more realistic network layout
	chnl := make(chan int, CONCURRENCY)
	for c := 0; c < CONCURRENCY; c++ {
		chnl <- 1
	}

	//TODO: use a more realistic network layout
	port:=""
	for j := 0; j < num_of_node; j++ {
		for k := 0; k <= int(math.Floor(math.Log2(float64(num_of_node)))); k += 1 {
			port_temp := math.Pow(2, float64(k))
			if int(port_temp)+j >= num_of_node {
				port = strconv.Itoa(initPort+(j+int(port_temp))%(num_of_node))
			} else {
				port = strconv.Itoa(initPort +j+ int(port_temp))
			}
			targetHost, targetPort, _:= StringToIpPort("localhost:" + port)
			select {
			case <-chnl:
	            time.Sleep(TimeStrap)
				go pingWithChan(node[j], targetHost, targetPort, chnl)
			}
		}
	}

	key := NewRandomID()
	res, err := node[0].DoIterativeStore(key, []byte("TESTING"))
	if err != nil {
		t.Error(err)
	}
	res = strings.Replace(res, "\n", "", -1)
	num_of_contact := len(res) / 40
	for i := 0; i < num_of_contact; i++ {
		NodeID, _ := IDFromString(res[i*40 : (i+1)*40])

		contact, err2 :=
			node[0].FindContact(NodeID)
		if err2 != nil {
			t.Error()
		}
		_, res2 := node[0].DoFindValue(contact, key)

		if string(res2.Value) == "E" {
			t.Error([]byte("Err: FindNode, instance1, can't find value"))
			return
		}
	}

	return
}

func TestIterativeFindValue(t *testing.T) {
	//Initialization
	num_of_node := N
	var node []*Kademlia
	initPort := 11000
	for i := 0; i < num_of_node; i++ {
		port := strconv.Itoa(initPort + i)
		str_port := "localhost:" + port
		node = append(node, NewKademlia(str_port))
	}
	//TODO: use a more realistic network layout
	chnl := make(chan int, CONCURRENCY)
	for c := 0; c < CONCURRENCY; c++ {
		chnl <- 1
	}

	//TODO: use a more realistic network layout
	port:=""
	for j := 0; j < num_of_node; j++ {
		for k := 0; k <= int(math.Floor(math.Log2(float64(num_of_node)))); k += 1 {
			port_temp := math.Pow(2, float64(k))
			if int(port_temp)+j >= num_of_node {
				port = strconv.Itoa(initPort+(j+int(port_temp))%(num_of_node))
			} else {
				port = strconv.Itoa(initPort +j+ int(port_temp))
			}
			targetHost, targetPort, _:= StringToIpPort("localhost:" + port)
			select {
			case <-chnl:
	            time.Sleep(TimeStrap)
				go pingWithChan(node[j], targetHost, targetPort, chnl)
			}
		}
	}
	key := NewRandomID()
	node[1].DoStore(&node[15].SelfContact, key, []byte("Testing"))
	res2, _ := node[0].DoIterativeFindValue(key)
	if res2 == "ERR" {
		t.Error([]byte("Can't Find the value"))
	}
	key = NewRandomID()
	res2, _ = node[0].DoIterativeFindValue(key)
	if res2 != "ERR" {
		t.Error([]byte("should output ERR"))
	}

	return
}

func TestVanish(t *testing.T){
	
}
