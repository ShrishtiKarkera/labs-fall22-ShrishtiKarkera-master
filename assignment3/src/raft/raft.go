package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server. - this server is the own raft server ?? (shrishti)
//

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"

	//"flag"
	//"fmt"
	"labrpc"
	"sync"
	"time"
)

// import "bytes"
// import "encoding/gob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//

type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

//
// A Go object implementing a single Raft peer.
//

type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int // index into peers[]

	applyCh           chan ApplyMsg
	voteCount         int
	indicateLeader    chan bool // indicate the peer which stood for election that it has become the leader now
	heartbreatCh      chan bool // to pass or get heartbeats
	informtimeoutCh   chan bool // informs the raft peer that leader has crossed the timeout threshold and not yet sent a heartbeat
	heartbeatfreq     int       // frequancy aat which the leader should send hearbeats (T) - check whether int or some other format
	hbtimeoutInterval int       // [T,2T] heartbeat timeout interval - between which the server, if not received hearbeat, must start an elecion check whether int or some other format
	electionTimeout   int       // the timeout for which a reElection must begin
	hbreceivedlast    int64     // last heartbeat received from the leader
	hbsentlast        int64     // this is for the leader - last time the heartbeat was sent to all the peers

	//votereqest RequestVoteArgs
	//votereply  RequestVoteReply

	current_term int // latest term server has seen (initialized to 0 on first boot, increases monotonically)
	votedFor     int // candidateId that received vote in current term (or null if none)

	//doubtful about this
	log []map[int][]int //log entries; each entry contains command for state machine, and term when entry	was received by leader (first index is 1)
	//its definitely not this - but int->string : index -> client req

	//volatile state on all servers
	commitIndex    int //index of highest log entry known to be committed (initialized to 0, increases monotonically)
	isleader       bool
	current_leader int
	isfollower     bool
	iscandidate    bool
	lastApplied    int //index of highest log entry applied to state machine (initialized to 0, increases monotonically)

	//volatile state on leaders - only for leaders
	//reinitialize these after every election
	// don't know whether can be ignored or not
	nextIndex  []int //for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
	matchIndex []int //for each server, index of highest log entry known to be replicated on server (initialized to 0, increases monotonically)

	// Your data here.
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	//current_term, votedFor, log[], commitIndex, lastApplied, nextIndex, matchIndex
	//also include isLeader bool
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here.
	//rf.mu.Lock()
	//fmt.Println("The leader is", rf.current_leader, rf.me)
	//fmt.Println("Am i the leader", rf.isleader)
	//defer rf.mu.Unlock()
	term = rf.current_term
	isleader = rf.isleader
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here.
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)

	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(rf.current_term)
	e.Encode(rf.votedFor)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)

	// current_term, votedfor and log[] should be stored in persisted storage
	//log entry could be ignored??
	//votedfor - in the current term and null if not voted but the paper says
	//candidateId that received vote in current term (or null if none)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	// Your code here.
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)

	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)
	d.Decode(&rf.current_term)
	d.Decode(&rf.votedFor)
}

func (rf *Raft) checkforHBtime() { // this will check if its time to produce the next heartbeat, if it is, then it signals the leader by sending true in the heartbeat channel
	for {
		if _, isLeader := rf.GetState(); isLeader == true {
			rf.mu.Lock()
			timeElapsed := time.Now().UnixNano() - rf.hbsentlast
			if int(timeElapsed/int64(time.Millisecond)) >= rf.heartbeatfreq {
				rf.heartbreatCh <- true
			}
			rf.mu.Unlock()
			time.Sleep(time.Millisecond * 10)
		} else {
			return
		}
	}
}

// func (rf *Raft) sendHeartbeat() {
// 	if _, isLeader := rf.GetState(); isLeader == true {
// 		rf.mu.Lock()
// 		rf.hbsentlast = time.Now().UnixNano()
// 		//write a function to send heartbeats to other raft peers, in the log or anywhere

// 		for i := range rf.peers {
// 			if i != rf.me {
// 				ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
// 				return ok
// 			}
// 		}

// 		rf.mu.Unlock()
// 	} else {
// 		return
// 	}
// }

func (rf *Raft) checkwhetherTimeout() {
	for {
		if _, isLeader := rf.GetState(); isLeader == false {
			rf.mu.Lock()
			timeElapsed := time.Now().UnixNano() - rf.hbreceivedlast
			if int(timeElapsed/int64(time.Millisecond)) >= rf.electionTimeout {
				rf.informtimeoutCh <- true
			}
			rf.mu.Unlock()

			//check this below once
			time.Sleep(time.Millisecond * time.Duration(rf.heartbeatfreq)) // make it sleep until the
		}
	}
}

// example RequestVote RPC arguments structure.
type RequestVoteArgs struct {
	// Your data here.
	//new_term - the new term for which the requester is starting elections (..)
	//latest_term - this is not required actually
	//log [] - to check the log entry - again not sure whether required or not

	//capitalizing all field names as said below
	Term         int //candidate’s term
	CandidateId  int //candidate requesting vote
	LastLogIndex int //index of candidate’s last log entry (§5.4)
	LastLogTerm  int //term of candidate’s last log entry (§5.4)

}

// example RequestVote RPC reply structure.
type RequestVoteReply struct {
	// Your data here.
	//capitalizing all field names as said below
	Term        int  //currentTerm, for candidate to update itself
	VoteGranted bool //true means candidate received vote
	//latest_term - the latest term of the target server which should be one less than the new_term
	// or equal to the latest_term in RequestVoteArgs
	//voteCount int //number of votes received by the particular raft server
}

// if the target server has not voted before, then it votes for this server (shrishti)
// is this the only condition(..)
// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here.

	rf.mu.Lock()
	//	rf.mu.Unlock()
	term_new := args.Term
	candidate := args.CandidateId
	//latest_term := args.LastLogTerm
	reply.Term = rf.current_term

	if rf.current_term < term_new || rf.votedFor != candidate {
		rf.current_term = term_new
		rf.votedFor = candidate
		rf.isfollower = true
		rf.isleader = false
		rf.iscandidate = false
		//rf.persist()

		//reply.Voted_term = rf.current_term
		reply.VoteGranted = true
		rf.mu.Unlock()
		return
	} else if rf.current_term == term_new && rf.votedFor != candidate {
		reply.VoteGranted = false
		rf.mu.Unlock()
		return
	} else {
		reply.VoteGranted = false
		rf.mu.Unlock()
		return
	}

	//for i := range rf.peers {
	//	sent := rf.sendRequestVote(i, args, reply)
	//	if sent {
	//		if rf.me == i {
	//
	//		}
	//	}
	//}

}

func (rf *Raft) startElection() { // this will call sendRequestVote
	args := &RequestVoteArgs{}
	args.CandidateId = rf.me
	args.Term = rf.current_term // make sure its incremented before
	//args.LastLogTerm = 1
	//args.LastLogIndex = 1
	peers := len(rf.peers)
	listenVotereqs := make(chan int, peers-1)
	for i := 0; i < peers; i++ {
		if i == rf.me {
			continue
		} else {
			listenVotereqs <- i
			go func() {
				reply := &RequestVoteReply{}
				reply.Term = rf.current_term // again make sure to increment the term
				server := <-listenVotereqs   // initialization error (:= <-)
				rf.sendRequestVote(server, *args, reply)
			}()
		}
	}
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.     --  [voted or not and the term last voted for (shrishti)]
// fills in *reply with RPC reply, so caller should
// pass &reply.      -- doubt below (..)
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// returns true if labrpc says the RPC was delivered.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
// doubt - will it be &reply in the example below
// will be called by startElection
// also check the condition where the term of peer election < reply term, then make it follower - its past stage was behing
func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	//trying
	//rf.mu.Lock() -- putting inside condition
	//defer rf.mu.Unlock()  -- putting inside condition  - doesn't matter though - matters
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	if ok {
		rf.mu.Lock()
		rterm := reply.Term
		status := reply.VoteGranted
		//rf.mu.Lock()
		if rterm > rf.current_term {

			rf.current_term = rterm
			rf.isleader = false
			rf.isfollower = true
			rf.iscandidate = false
		} else if status == true && rf.iscandidate == true { // check if the votes crossed the threshold
			//rf.mu.Lock()
			majority := (len(rf.peers) / 2) + 1
			rf.voteCount += 1
			if rf.voteCount >= majority {
				rf.indicateLeader <- true
				rf.iscandidate = false
				rf.isleader = true
				rf.current_leader = rf.me
				fmt.Println("I just became the leader", rf.me, rf.isleader)
				rf.isfollower = false
			}
		}
		rf.mu.Unlock()
	}
	return ok
	//rf.mu.Unlock()
}

// Code for hearbeat
type SendHeartBeat struct {

	//capitalizing all field names as said below
	Term     int //candidate’s term
	Leaderid int // leader peer number

}
type HeartBeatReply struct {

	//capitalizing all field names as said below
	Term     int  // latest term of the peer
	Received bool // heartbeat received  - update the latest hb received time in raft peer struct

}

func (rf *Raft) AfterReceivingHB(args SendHeartBeat, reply *HeartBeatReply) {
	// Your code here.

	// rf.mu.Lock()
	fmt.Println("Now sending HBs", rf.me, rf.current_term)
	fmt.Println("Am i standing for election", rf.iscandidate)
	rf.heartbreatCh <- true
	leader_term := args.Term
	leader := args.Leaderid

	if leader_term >= rf.current_term {
		//rf.mu.Lock()
		fmt.Println("Tester", rf.me)
		rf.current_term = leader_term
		rf.current_leader = leader
		reply.Term = rf.current_term
		reply.Received = true
		if rf.iscandidate {
			rf.isleader = false
			rf.isfollower = true
			rf.iscandidate = false
		}
		//rf.mu.Unlock()
	}
	//if rf.iscandidate == true {
	//	fmt.Println("entered the candidate to leader term", rf.me)
	//	rf.isfollower = true
	//	rf.iscandidate = false
	//	rf.isleader = false
	//}
	if leader_term == rf.current_term && leader == rf.current_leader {
		reply.Term = rf.current_term
		reply.Received = true
	} else if rf.iscandidate || rf.isleader {
		rf.iscandidate = false
		rf.isfollower = true
		rf.isleader = false
		//rf.current_leader = leader
		//fmt.Println("became follower back again")
	} else {
		//start reelection
	}
	// rf.mu.Unlock()
}

func (rf *Raft) ConsistentHBs() { // this will call sendRequestVote
	fmt.Println("Leader entered consistentHBs to send hbs", rf.me)
	args := &SendHeartBeat{}
	args.Leaderid = rf.me
	args.Term = rf.current_term // make sure its not incremented before
	peers := len(rf.peers)
	listenHBs := make(chan int, peers-1)
	for i := 0; i < peers; i++ {
		if i == rf.me {
			continue
		} else {
			listenHBs <- i
			go func() {
				reply := &HeartBeatReply{}
				reply.Term = rf.current_term // again make sure to increment the term
				reply.Received = true
				server := <-listenHBs // initialization error (:= <-)
				rf.sendHeartBeats(server, *args, reply)
			}()
		}
	}
}
func (rf *Raft) sendHeartBeats(server int, args SendHeartBeat, reply *HeartBeatReply) bool {
	//trying
	//rf.mu.Lock() -- putting inside condition
	//defer rf.mu.Unlock()  -- putting inside condition  - doesn't matter though - matters
	fmt.Println("Leader also entered sendhbs", rf.me)
	ok := rf.peers[server].Call("Raft.AfterReceivingHB", args, reply)
	if ok {
		rf.mu.Lock()
		rterm := reply.Term
		rec := reply.Received
		//fmt.Println("Reply term", rterm)
		//fmt.Println("Leader term", args.Term)
		//fmt.Println("The reply is", rec)
		if rterm == args.Term && rec {
			fmt.Println("Entered the not rabbit")
			//change heart beat send time for the leader
			//change some of the variables as well
		} else {
			fmt.Println("Entered the rabbit hole")
			rf.isfollower = true
			rf.iscandidate = false
			rf.isleader = false
			//fmt.Println("became follower in sendhbs")
			//print something here if you like
		}
		rf.mu.Unlock()
	}
	return ok
	//rf.mu.Unlock()
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true
	return index, term, isLeader
}

// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.

func (rf *Raft) Kill() {
	// Your code here, if desired.
}

func (rf *Raft) goForever() {
	//rf.electionTimeout = 0
	for {
		if rf.isleader { // ensure to send heartbeats
			//fmt.Println("Leader entered the loop", rf.me)
			time.Sleep(50 * time.Millisecond)
			go rf.ConsistentHBs()
		} else if rf.iscandidate { // ensure to start election

			rf.votedFor = rf.me
			rf.voteCount = 1
			rf.current_term += 1
			//go rountine to send heartbeats
			fmt.Println("Starting election for", rf.me)
			go rf.startElection()
			afterCh := time.After(500 * time.Millisecond) // let this be constant for now for all peers
			select {
			case lc := <-rf.indicateLeader:
				//case <-rf.indicateLeader:
				fmt.Println("printing leader", rf.me, lc)
				rf.iscandidate = false
				rf.isleader = true
				go rf.ConsistentHBs()
				//send to all the raft peers that this peer is the leader.
			case ac := <-afterCh:
				//case <-afterCh:
				fmt.Println("election failed for", rf.me, ac)
				rf.isleader = false
				rf.iscandidate = false
				rf.isfollower = true
			}
			// ensure to receive heartbeats - if crossed timeout, start election
		} else if rf.isfollower {
			fmt.Println("Entered is follower", rf.me)
			rf.mu.Lock()
			fmt.Println("Entered the goforever isfollower loop")
			rf.electionTimeout = rand.Intn(300 - 150)
			//fmt.Println(rf.electionTimeout)
			afterCh := time.After(time.Millisecond * time.Duration(rf.electionTimeout))
			select {
			case hb := <-rf.heartbreatCh:
				//case <-rf.heartbreatCh:
				fmt.Println("received hb", rf.me, hb)
				//rf.electionTimeout = 0
			case ac := <-afterCh:
				//case <-afterCh:
				fmt.Println("timed out", rf.me, ac)
				//start elecion for the next term
				rf.isleader = false
				rf.isfollower = false
				rf.iscandidate = true // this will be detected in the next term
				//rf.current_term += 1
				//afterCh.Reset(time.Millisecond * time.Duration(rf.electionTimeout))
				rf.electionTimeout = 0
			}
			rf.mu.Unlock()
		}

	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	rf.applyCh = applyCh // chan applyCh of type ApplyMsg
	rf.heartbreatCh = make(chan bool)
	rf.voteCount = 0
	rf.indicateLeader = make(chan bool)
	rf.informtimeoutCh = make(chan bool)
	//rf.heartbeatfreq = 150 // will convert when using it - 2000 * time.Millisecond
	//rf.electionTimeout = rand.Intn(300-150)
	rf.hbsentlast = 0
	rf.hbreceivedlast = 0
	//rf.hbtimeoutInterval = rand.Seed()
	// Your initialization code here.

	rf.current_term = 0
	rf.votedFor = -1

	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.isleader = false
	rf.isfollower = true
	rf.iscandidate = false
	rf.current_leader = -1

	//check this initialization once
	//volatile state on leaders - only for leaders
	//for each server, index of highest log entry known to be replicated on server
	rf.matchIndex = make([]int, len(rf.peers)) // server number - last committed index on that server
	//for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
	rf.nextIndex = make([]int, len(rf.peers)) // server number - next index to be send to that server

	// initializing RequestVoteArgs and RequestVoteReply
	//rf.votereqest.Voting_term = 0
	//rf.votereqest.Latest_term = 0
	//// 	rf.votereqest.log   	// optional
	//rf.votereply.Voted_term = 0
	//rf.votereply.Voted = false

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	//start go routines before returning
	go rf.goForever()

	//go func()
	//go func()
	return rf
}

// a function to reply to the requestVote function has to be made
//check if this server has voted for the term = election term and then cast vote
//reset the timeout when it cast vote and also increment the current_term value
//func (rf *Raft) ProcessVoteReq() bool{  // this function has already been created - RequestVote
//
//}
