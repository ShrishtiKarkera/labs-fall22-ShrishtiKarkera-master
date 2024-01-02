package chandy_lamport

import (
	"fmt"
	"log"
)

// The main participant of the distributed snapshot protocol.
// Servers exchange token messages and marker messages among each other.
// Token messages represent the transfer of tokens from one server to another.
// Marker messages represent the progress of the snapshot process. The bulk of
// the distributed protocol is implemented in `HandlePacket` and `StartSnapshot`.
type Server struct {
	Id            string
	Tokens        int
	sim           *Simulator
	outboundLinks map[string]*Link // key = link.dest
	inboundLinks  map[string]*Link // key = link.src
	// TODO: ADD MORE FIELDS HERE
	marker_yn      *SyncMap //snapshotid - 0/1    whether marker is received or not and whether the snapshot is completed or not (int to bool)
	snapshotid_rec *SyncMap //snapshotid - snapshotstate   snapshot id and the number of tokens
	marker_link    *SyncMap // (not used yet)   stores the snapshotid and the links from which it has received the id from(int for now)
	//snapshot_link  *SyncMap //     to store the messages received from a particular link
	link_tracker *SyncMap //[snapshotid] - [all links - active/inactive(bool)]
	//syncmap of a syncmap for storing the active links
}

// A unidirectional communication channel between two servers
// Each link contains an event queue (as opposed to a packet queue)
type Link struct {
	src    string
	dest   string
	events *Queue
}

func NewServer(id string, tokens int, sim *Simulator) *Server {
	return &Server{
		id,
		tokens,
		sim,
		make(map[string]*Link),
		make(map[string]*Link),
		NewSyncMap(),
		NewSyncMap(),
		NewSyncMap(),
		//NewSyncMap(),
		NewSyncMap(),
	}
}

// Add a unidirectional link to the destination server
func (server *Server) AddOutboundLink(dest *Server) {
	if server == dest {
		return
	}
	l := Link{server.Id, dest.Id, NewQueue()}
	server.outboundLinks[dest.Id] = &l
	dest.inboundLinks[server.Id] = &l
}

// Send a message on all of the server's outbound links
func (server *Server) SendToNeighbors(message interface{}) {
	for _, serverId := range getSortedKeys(server.outboundLinks) {
		link := server.outboundLinks[serverId]
		server.sim.logger.RecordEvent(
			server,
			SentMessageEvent{server.Id, link.dest, message})
		link.events.Push(SendMessageEvent{
			server.Id,
			link.dest,
			message,
			server.sim.GetReceiveTime()})
	}
}

// Send a number of tokens to a neighbor attached to this server
func (server *Server) SendTokens(numTokens int, dest string) {
	if server.Tokens < numTokens {
		log.Fatalf("Server %v attempted to send %v tokens when it only has %v\n",
			server.Id, numTokens, server.Tokens)
	}
	message := TokenMessage{numTokens}
	server.sim.logger.RecordEvent(server, SentMessageEvent{server.Id, dest, message})
	// Update local state before sending the tokens
	server.Tokens -= numTokens
	link, ok := server.outboundLinks[dest]
	if !ok {
		log.Fatalf("Unknown dest ID %v from server %v\n", dest, server.Id)
	}
	link.events.Push(SendMessageEvent{
		server.Id,
		dest,
		message,
		server.sim.GetReceiveTime()})
}

// Callback for when a message is received on this server.
// When the snapshot algorithm completes on this server, this function
// should notify the simulator by calling `sim.NotifySnapshotComplete`.
func (server *Server) HandlePacket(src string, message interface{}) {
	// TODO: IMPLEMENT ME
	switch message.(type) {
	case MarkerMessage:
		//flag := false
		m := message.(MarkerMessage)
		fmt.Println("marker msg")
		_, ok := server.marker_yn.Load(m.snapshotId) // started = false , completed = true
		if ok == true {                              // not take snapshot again // listening to channel
			fmt.Println("marker msg listenng again")
			link, _ := server.link_tracker.Load(m.snapshotId)
			link_new := link.(*SyncMap)
			link_new.Store(src, false)
			server.link_tracker.Store(m.snapshotId, link_new)
			//server.link_tracker.Range(func(k interface{}, v interface{}) bool {
			//	key := k.(int) //getting snapshotid and value has the src and whether they are active or not
			//	if key == m.snapshotId {
			//		a, _ := server.link_tracker.Load(m.snapshotId) //
			//
			//		a_new := a.(*SyncMap)
			//		a_new.Store(src, false)
			//		server.link_tracker.Store(m.snapshotId, a_new)
			//		fmt.Println("storing into link tracker")
			//	}
			//	return true
			//})
			//has sserver recieved marker from evrry other server
			fmt.Println("has sserver recieved marker from evrry other server")
			checking, ok := server.link_tracker.Load(m.snapshotId)
			if ok == true {
				c_new := checking.(*SyncMap)
				counter := 0
				c_new.Range(func(k interface{}, v interface{}) bool {
					//key := k.(int)
					value := v.(bool)
					if value == false {
						counter += 1
					}
					return true
				})
				if counter == len(server.inboundLinks) {
					fmt.Printf("Yay all count matched for server %v\n", server.Id)
					server.marker_yn.Store(m.snapshotId, true)
					server.sim.NotifySnapshotComplete(server.Id, m.snapshotId)
					// setting true in marker_yn means snapshot completed for server

				}
			}
		} else {
			fmt.Printf(" seeing marker first time %v", server.Id)
			server.StartSnapshot(m.snapshotId)
			//making inactive for that link
			//make a new entry
			link, _ := server.link_tracker.Load(m.snapshotId)
			link_new := link.(*SyncMap)
			//matchingSrc,_:=link_new.Load(src)
			//matchingSrc()
			link_new.Store(src, false)

			fmt.Printf(" updating link tracker with  %v", link_new)
			server.link_tracker.Store(m.snapshotId, link_new)

			// check for link = false
			checking, ok := server.link_tracker.Load(m.snapshotId)
			if ok == true {
				c_new := checking.(*SyncMap)
				counter := 0
				c_new.Range(func(k interface{}, v interface{}) bool {
					//key := k.(int)
					value := v.(bool)
					if value == false {
						counter += 1
					}
					return true
				})
				fmt.Printf("check count for server %v\n", server.Id)
				if counter == len(server.inboundLinks) {
					fmt.Printf("Yay count got match %v\n", server.Id)
					server.sim.NotifySnapshotComplete(server.Id, m.snapshotId)
					// setting true in marker_yn means snapshot completed for server
					server.marker_yn.Store(m.snapshotId, true)
				}
			} // checking for link = false ends here
		}

	case TokenMessage:
		m := message.(TokenMessage)

		server.marker_yn.Range(func(k interface{}, v interface{}) bool {
			key := k.(int)
			value := v.(bool)
			if value == false {
				ongoing, _ := server.link_tracker.Load(key)
				ongoing_new := ongoing.(*SyncMap)
				ongoing_new.Range(func(i interface{}, j interface{}) bool {
					ikey := i.(string) //link
					jvalue := j.(bool) //active/inactive
					if ikey == src {
						if jvalue == true {
							//take snapshot of the channel because it is inactive - read common.go
							//but should I take the snapshot if I'm not really listening
							shot := SnapshotMessage{src, server.Id, TokenMessage{m.numTokens}}
							//server.snapshot_link.Store(ikey, shot) // but how will the snap be retrieved by the system???
							s, _ := server.snapshotid_rec.Load(key)
							s_new := s.(SnapshotState)
							s_new.messages = append(s_new.messages, &shot)
							server.snapshotid_rec.Store(key, s_new)
						}
					}
					return true
				})
			}
			return true
		})
		server.Tokens += m.numTokens
		//here the code for checking the channels will come and then updating the channel snapshot
	}

	//server.sim.NotifySnapshotComplete()
}

// Start the chandy-lamport snapshot algorithm on this server.
// This should be called only once per server.
func (server *Server) StartSnapshot(snapshotId int) {
	// TODO: IMPLEMENT ME
	//Before starting a snapshot, a check is required that snapshot with snapshotID has
	//already been started or was completed.
	// If not record the server snapshot state and send Marker messages to all neighbors of
	//current server.
	// In the whole process the data structures that are keeping track of the ongoing
	//snapshots/completed snapshots/marker messages should be updated.
	fmt.Printf("snapshot started for %v", server.Id)
	shot := make(map[string]int)
	shot[server.Id] = server.Tokens
	s := SnapshotState{snapshotId, shot, make([]*SnapshotMessage, 0)}
	server.snapshotid_rec.Store(snapshotId, s)
	server.marker_yn.Store(snapshotId, false) // snapshot started for snapshotid but not completed
	fmt.Println("snapshot started for snapshotid but not completed")
	newSyncMap := NewSyncMap()
	for _, link := range server.inboundLinks {
		newSyncMap.Store(link.src, true)
	}
	//server.inboundLink.forr
	//newSyncMap.Store()
	server.link_tracker.Store(snapshotId, newSyncMap)
	server.SendToNeighbors(MarkerMessage{snapshotId: snapshotId})
	fmt.Printf("snapshot completed for %v", server.Id)
	//shift above
	//server.link_tracker
	//add data into marker link
	//listen for all channels if from simulator
}

// make a tokens array for the tokens that it is receiving from that link
