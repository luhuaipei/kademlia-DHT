package kademlia
import(
	"sync"
)

type VDOT struct {
	VDOID   ID
	VDOtable map[ID] VanashingDataObject
	VDOLock sync.Mutex
}

// Initialize
func (myVDO *VDOT) Init(node ID) {
	myVDO.VDOLock = sync.Mutex{}
	myVDO.VDOID = node
	myVDO.VDOtable = make(map[ID](VanashingDataObject))
}

//Insert key/value pairs into the DHT
func (myVDO *VDOT) Put(VDOID ID, VDOobj VanashingDataObject) {
	myVDO.VDOLock.Lock()
	myVDO.VDOtable[VDOID] = VDOobj
	myVDO.VDOLock.Unlock()
}

//Find key
func (myVDO *VDOT) Get(VDOID ID) (*VanashingDataObject, bool) {
	myVDO.VDOLock.Lock()
	i, ok := myVDO.VDOtable[VDOID]
	if !ok {
		myVDO.VDOLock.Unlock()
		return nil, ok
	}
	myVDO.VDOLock.Unlock()
	return &i, ok
}
