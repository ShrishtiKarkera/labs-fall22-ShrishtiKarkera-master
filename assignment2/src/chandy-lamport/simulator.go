package chandy_lamport

import (
	"fmt"
	"log"
	"math/rand"
)

// Max random delay added to packet delivery
const maxDelay = 5

// Simulator is the entry point to the distributed snapshot application.
//
// It is a discrete time simulator, i.e. events that happen at time t + 1 come
// strictly after events that happen at time t. At each time step, the simulator
// examines messages queued up across all the links in the system and decides
// which ones to deliver to the destination.
//
// The simulator is responsible for starting the snapshot process, inducing servers
// to pass tokens to each other, and collecting the snapshot state after the process
// has terminated.
type Simulator struct {
	time           int
	nextSnapshotId int
	servers        map[string]*Server // key = server ID
	logger         *Logger
	// TODO: ADD MORE FIELDS HERE
	//started bool
	snapshotid_rec *SyncMap // to keep track whether a snapshot for snapshotids has been started or not (0 - started), 1- completed
	server_rec     *SyncMap // snapshotids and their corresponding completed serverids (array)
	blocker        []int
	b              map[int]chan bool //make map of channels
	counting       *SyncMap
	snapshot_rec   *SyncMap
}

func NewSimulator() *Simulator {
	return &Simulator{
		0,
		0,
		make(map[string]*Server),
		NewLogger(),
		//false,
		NewSyncMap(),
		NewSyncMap(),
		make([]int, 20),
		make(map[int]chan bool),
		NewSyncMap(),
		NewSyncMap(),
	}
}

// Return the receive time of a message after adding a random delay.
// Note: since we only deliver one message to a given server at each time step,
// the message may be received *after* the time step returned in this function.
func (sim *Simulator) GetReceiveTime() int {
	return sim.time + 1 + rand.Intn(5)
}

// Add a server to this simulator with the specified number of starting tokens
func (sim *Simulator) AddServer(id string, tokens int) {
	server := NewServer(id, tokens, sim)
	sim.servers[id] = server
}

// Add a unidirectional link between two servers
func (sim *Simulator) AddForwardLink(src string, dest string) {
	server1, ok1 := sim.servers[src]
	server2, ok2 := sim.servers[dest]
	if !ok1 {
		log.Fatalf("Server %v does not exist\n", src)
	}
	if !ok2 {
		log.Fatalf("Server %v does not exist\n", dest)
	}
	server1.AddOutboundLink(server2)
}

// Run an event in the system
func (sim *Simulator) InjectEvent(event interface{}) {
	switch event := event.(type) {
	case PassTokenEvent:
		src := sim.servers[event.src]
		src.SendTokens(event.tokens, event.dest)
	case SnapshotEvent:
		sim.StartSnapshot(event.serverId)
	default:
		log.Fatal("Error unknown event: ", event)
	}
}

// Advance the simulator time forward by one step, handling all send message events
// that expire at the new time step, if any.
func (sim *Simulator) Tick() {
	sim.time++
	sim.logger.NewEpoch()
	// Note: to ensure deterministic ordering of packet delivery across the servers,
	// we must also iterate through the servers and the links in a deterministic way
	for _, serverId := range getSortedKeys(sim.servers) {
		server := sim.servers[serverId]
		for _, dest := range getSortedKeys(server.outboundLinks) {
			link := server.outboundLinks[dest]
			// Deliver at most one packet per server at each time step to
			// establish total ordering of packet delivery to each server
			if !link.events.Empty() {
				e := link.events.Peek().(SendMessageEvent)
				if e.receiveTime <= sim.time {
					link.events.Pop()
					sim.logger.RecordEvent(
						sim.servers[e.dest],
						ReceivedMessageEvent{e.src, e.dest, e.message})
					sim.servers[e.dest].HandlePacket(e.src, e.message)
					break
				}
			}
		}
	}
}

// Start a new snapshot process at the specified server
func (sim *Simulator) StartSnapshot(serverId string) {
	snapshotId := sim.nextSnapshotId
	sim.nextSnapshotId++
	// TODO: IMPLEMENT ME
	//Get the snapshotID from simulator struct to initiate snapshot of snapshotId on
	//	serverId that this function receives as argument.
	//	Before starting a snapshot, a check is required whether if the snapshot process for
	//	snapshotId has started or not.
	//	Only if there is no snapshot, initiate one by calling StartSnapshot on the server.

	//flag := 0
	//if len(sim.snapshotid_rec)
	//sim.snapshotid_rec.Range(func(k interface{}, v interface{}) bool{
	//	key := k.(int)
	//	value := v.(int)
	//	if key == snapshotId && value == 0{
	//		flag = 1
	//	}
	//	return true
	//})

	sim.servers[serverId].StartSnapshot(snapshotId)
	//sim.compl
	//m := make(map[int]bool, 0)
	//	m[snapshotId] = false
	//sim.b <- m
	sim.b[snapshotId] = make(chan bool)
	sim.snapshot_rec.Store(snapshotId, make([]string, 0))
}

// Callback for servers to notify the simulator that the snapshot process has
// completed on a particular server
func (sim *Simulator) NotifySnapshotComplete(serverId string, snapshotId int) {
	sim.logger.RecordEvent(sim.servers[serverId], EndSnapshot{serverId, snapshotId})
	// TODO: IMPLEMENT ME
	//This is a function that the server calls to notify simulator that snapshot on that
	//server is completed.
	//• Whenever a server notifies you need to keep track of the serverID in some data
	//structure.
	//b := make(chan int)
	//sim.mu.Lock()
	fmt.Println("notify started")
	// code comment1
	//a, ok := sim.server_rec.Load(serverId)
	//if ok == true {
	//	fmt.Println("has value in the map ,, cooll")
	//	q := a.([]int)
	//	q = append(q, snapshotId)
	//	sim.server_rec.Store(serverId, q)
	//} else {
	//	fmt.Println("notify started: does not have value in map")
	//	lis := make([]int, 0)
	//	lis = append(lis, snapshotId)
	//	sim.server_rec.Store(serverId, lis)
	//}
	// code comment 1 end

	a, ok := sim.snapshot_rec.Load(snapshotId)
	if ok == true {
		fmt.Println("has value in the map ,, cooll")
		q := a.([]string)
		q = append(q, serverId)
		sim.snapshot_rec.Store(snapshotId, q)
	} else {
		fmt.Println("notify started: does not have value in map")
		lis := make([]string, 0)
		lis = append(lis, serverId)
		sim.snapshot_rec.Store(snapshotId, lis)
		fmt.Println("after storing the value")
	}
	//count := 0
	// code comment 2
	//sim.server_rec.Range(func(k interface{}, v interface{}) bool {
	//	snapShotIdsOfServer := v.([]int)
	//	for i := range snapShotIdsOfServer {
	//		if i == snapshotId {
	//			fmt.Println("matched the server in server_recording")
	//			new_count, ok := sim.counting.Load(serverId)
	//			if ok == false {
	//				fmt.Println("inside false of counter")
	//				sim.counting.Store(serverId, 1)
	//				val, _ := sim.counting.Load(serverId)
	//				fmt.Println("Stored value in counter for ok = false")
	//				fmt.Printf("value of counter %v in ok false", val)
	//			}
	//			fmt.Println("just outside ok = true")
	//			if ok == true {
	//				//valfmt.Printf("inside ok = true of counter %v", )
	//				count_new := new_count.(int)
	//				count_new++
	//				sim.counting.Store(serverId, count_new)
	//				fmt.Printf("incremented the counter inside ok = true %v", count_new)
	//			}
	//			//fmt.Printf("value of count in notify %v\n", count)
	//		}
	//	}
	//	return true
	//})
	//code comment 2 ends

	sarray, _ := sim.snapshot_rec.Load(snapshotId)
	srr := sarray.([]string)
	fmt.Println("snap id matched")
	fmt.Printf("srr %v length", len(srr))
	if len(srr) == len(sim.servers) {
		fmt.Println("insert into channel")
		sim.b[snapshotId] <- true
	}

	//c, _ := sim.counting.Load(serverId)
	//countingOfServers := c.(int)
	//if len(sim.servers) == countingOfServers {
	//	fmt.Println("insert into channel")
	//	//mp := make(map[int]bool, 0)
	//	//mp[snapshotId] = true
	//	sim.b[snapshotId] <- true
	//	//mp <- sim.b
	//	//mp.
	//	fmt.Println("notify completed and marker process completed")
	//}
	//sim.mu.Unlock()

}

// Collect and merge snapshot state from all the servers.
// This function blocks until the snapshot process has completed on all servers.
func (sim *Simulator) CollectSnapshot(snapshotId int) *SnapshotState {
	//one more checking to be done for the """""snapshotID"""""
	// TODO: IMPLEMENT ME

	fmt.Println("collect started")
	snap := SnapshotState{snapshotId, make(map[string]int), make([]*SnapshotMessage, 0)}
	<-sim.b[snapshotId]
	//fmt.Printf("\n collect started with channel %v", ch[snapshotId])
	//if ch[snapshotId] {
	fmt.Println("true for channel id")
	for _, server := range sim.servers {
		server.snapshotid_rec.Range(func(key interface{}, value interface{}) bool {
			k := key.(int)
			v := value.(SnapshotState)
			fmt.Printf("val for snapshot in collect %v\n", v)
			if k == snapshotId { // snapshotstate - server_id, tokens, nil
				snap.tokens[server.Id] = v.tokens[server.Id]
				snap.messages = append(snap.messages, v.messages...)
			}
			fmt.Printf("val for final snap in collect %v\n", snap)
			return true
		})
	}

	//}
	return &snap
}
