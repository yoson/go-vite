package net

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vitelabs/go-vite/common/types"

	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/vite/net/message"
)

type SyncState uint

const (
	SyncNotStart SyncState = iota
	Syncing
	Syncdone
	Syncerr
	SyncCancel
	SyncDownloaded
)

var syncStatus = [...]string{
	SyncNotStart:   "Sync Not Start",
	Syncing:        "Synchronising",
	Syncdone:       "Sync done",
	Syncerr:        "Sync error",
	SyncCancel:     "Sync canceled",
	SyncDownloaded: "Sync downloaded",
}

func (s SyncState) String() string {
	if s > SyncDownloaded {
		return "unknown sync state"
	}
	return syncStatus[s]
}

// the minimal height difference between snapshot chain of ours and bestPeer
// if the difference is little than this value, then we deem no need sync
const minHeightDifference = 3600
const waitEnoughPeers = 10 * time.Second
const enoughPeers = 3
const chainGrowInterval = time.Second

func shouldSync(from, to uint64) bool {
	if to >= from+minHeightDifference {
		return true
	}

	return false
}

type fileRecord struct {
	File
	add bool
}

type syncer struct {
	from, to uint64
	current  uint64 // current height
	aCount   uint64 // count of snapshot blocks have downloaded
	sCount   uint64 // count of account blocks have download

	state SyncState

	peers *peerSet

	pending   int // pending count of FileList msg
	responsed int // number of FileList msg received
	mu        sync.Mutex
	fileMap   map[filename]*fileRecord

	// query current block and height
	chain Chain

	// get peer add/delete event
	eventChan chan peerEvent

	// handle blocks
	verifier Verifier
	notifier blockNotifier

	// for sync tasks
	fc   *fileClient
	pool *chunkPool
	exec syncTaskExecutor

	// for subscribe
	curSubId int
	subs     map[int]SyncStateCallback

	running int32
	term    chan struct{}
	log     log15.Logger
}

func (s *syncer) receiveAccountBlock(block *ledger.AccountBlock) error {
	err := s.verifier.VerifyNetAb(block)
	if err != nil {
		return err
	}

	s.notifier.notifyAccountBlock(block, types.RemoteSync)
	atomic.AddUint64(&s.aCount, 1)
	return nil
}

func (s *syncer) receiveSnapshotBlock(block *ledger.SnapshotBlock) error {
	err := s.verifier.VerifyNetSb(block)
	if err != nil {
		return err
	}

	s.notifier.notifySnapshotBlock(block, types.RemoteSync)
	atomic.AddUint64(&s.sCount, 1)
	return nil
}

func newSyncer(chain Chain, peers *peerSet, verifier Verifier, gid MsgIder, notifier blockNotifier) *syncer {
	s := &syncer{
		state:     SyncNotStart,
		chain:     chain,
		peers:     peers,
		fileMap:   make(map[filename]*fileRecord),
		eventChan: make(chan peerEvent, 1),
		verifier:  verifier,
		notifier:  notifier,
		subs:      make(map[int]SyncStateCallback),
		log:       log15.New("module", "net/syncer"),
	}

	pool := newChunkPool(peers, gid, s)
	fc := newFileClient(chain, s, peers)
	s.exec = newExecutor(s)

	s.pool = pool
	s.fc = fc

	return s
}

func (s *syncer) SubscribeSyncStatus(fn SyncStateCallback) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.curSubId++
	s.subs[s.curSubId] = fn
	return s.curSubId
}

func (s *syncer) UnsubscribeSyncStatus(subId int) {
	if subId <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subs, subId)
}

func (s *syncer) SyncState() SyncState {
	return s.state
}

func (s *syncer) setState(st SyncState) {
	s.state = st
	for _, sub := range s.subs {
		sub(st)
	}
}

func (s *syncer) Stop() {
	if atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		if s.term == nil {
			return
		}

		select {
		case <-s.term:
		default:
			close(s.term)
			s.exec.terminate()
			s.pool.stop()
			s.fc.stop()
			s.clear()
			s.peers.unSub(s.eventChan)
		}
	}
}

func (s *syncer) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.aCount = 0
	s.sCount = 0
	s.pending = 0
	s.fileMap = make(map[filename]*fileRecord)
}

func (s *syncer) Start() {
	// is running
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		return
	}
	s.term = make(chan struct{})

	s.peers.sub(s.eventChan)

	defer s.Stop()

	start := time.NewTimer(waitEnoughPeers)

wait:
	for {
		select {
		case e := <-s.eventChan:
			if e.count >= enoughPeers {
				break wait
			}
		case <-start.C:
			break wait
		case <-s.term:
			s.log.Warn("sync cancel")
			s.setState(SyncCancel)
			start.Stop()
			return
		}
	}

	start.Stop()

	// for now syncState is SyncNotStart
	syncPeer := s.peers.syncPeer()
	if syncPeer == nil {
		s.setState(Syncerr)
		s.log.Error("sync error: no peers")
		return
	}

	syncPeerHeight := syncPeer.height()

	// compare snapshot chain height
	current := s.chain.GetLatestSnapshotBlock()
	// p is not all enough, no need to sync
	if current.Height+minHeightDifference > syncPeerHeight {
		if current.Height < syncPeerHeight {
			err := syncPeer.send(GetSnapshotBlocksCode, 0, &message.GetSnapshotBlocks{
				From:    ledger.HashHeight{Height: syncPeerHeight},
				Count:   1,
				Forward: true,
			})

			if err != nil {
				s.log.Error(fmt.Sprintf("Failed to send GetSnapshotBlocks to %s", syncPeer.Address()))
				return
			}
		}

		s.log.Info(fmt.Sprintf("sync done: syncPeer %s at %d, our height: %d", syncPeer.Address(), syncPeerHeight, current.Height))
		s.setState(Syncdone)
		return
	}

	s.current = current.Height
	s.from = current.Height + 1
	s.to = syncPeerHeight
	s.current = current.Height
	s.setState(Syncing)
	// todo

	// check chain height
	checkChainTicker := time.NewTicker(chainGrowInterval)
	defer checkChainTicker.Stop()
	var lastCheckTime = time.Now()

	for {
		select {
		case e := <-s.eventChan:
			if e.code == delPeer {
				// a taller peer is disconnected, maybe is the peer we syncing to
				// because peer`s height is growing
				if e.peer.height() >= s.to {
					if syncPeer = s.peers.syncPeer(); syncPeer != nil {
						syncPeerHeight = syncPeer.height()
						if shouldSync(current.Height, syncPeerHeight) {
							s.setTarget(syncPeerHeight)
						} else {
							// no need sync
							s.log.Info(fmt.Sprintf("no need sync to bestPeer %s at %d, our height: %d", syncPeer, syncPeerHeight, current.Height))
							s.setState(Syncdone)
							return
						}
					} else {
						// have no peers
						s.log.Error("sync error: no peers")
						s.setState(Syncerr)
						// no peers, then quit
						return
					}
				}
			} else if shouldSync(current.Height, e.peer.height()) {
				// todo
			}

		case now := <-checkChainTicker.C:
			current = s.chain.GetLatestSnapshotBlock()

			if current.Height >= s.to {
				s.log.Info(fmt.Sprintf("sync done, current height: %d", current.Height))
				s.setState(Syncdone)
				return
			}

			s.log.Info(fmt.Sprintf("sync current: %d, chain speed %d", current.Height, current.Height-s.current))

			if current.Height == s.current && now.Sub(lastCheckTime) > 10*time.Minute {
				s.setState(Syncerr)
			} else if s.state == Syncing {
				s.current = current.Height
				lastCheckTime = now
				s.exec.runTo(s.current + 3600)
			}

		case <-s.term:
			s.log.Warn("sync cancel")
			s.setState(SyncCancel)
			return
		}
	}
}

// this method will be called when our target Height changed, (eg. the best peer disconnected)
func (s *syncer) setTarget(to uint64) {
	atomic.StoreUint64(&s.to, to)
}

func (s *syncer) taskDone(t *syncTask, err error) {
	if err != nil {
		s.log.Error(fmt.Sprintf("sync task %s error", t.String()))

		if s.state != Syncing || atomic.LoadInt32(&s.running) == 0 {
			return
		}

		_, to := t.bound()
		target := atomic.LoadUint64(&s.to)

		if to <= target {
			s.setState(Syncerr)
			return
		}
	}
}

func (s *syncer) allTaskDone(last *syncTask) {
	_, to := last.bound()
	target := atomic.LoadUint64(&s.to)

	if to >= target {
		s.setState(SyncDownloaded)
		return
	}

	// use chunk
	cks := splitChunk(to+1, target, 3600)
	for _, ck := range cks {
		s.exec.add(&syncTask{
			task: &chunkTask{
				from:       ck[0],
				to:         ck[1],
				downloader: s.pool,
			},
			typ: syncChunkTask,
		})
	}

	s.pool.start()
}

type SyncStatus struct {
	From     uint64
	To       uint64
	Current  uint64
	Received uint64
	State    SyncState
}

func (s *syncer) Status() SyncStatus {
	current := s.chain.GetLatestSnapshotBlock()

	return SyncStatus{
		From:     s.from,
		To:       s.to,
		Current:  current.Height,
		Received: s.sCount,
		State:    s.state,
	}
}

type SyncDetail struct {
	SyncStatus
	ExecutorStatus
	FileClientStatus
	ChunkPoolStatus
}

func (s *syncer) Detail() SyncDetail {
	return SyncDetail{
		SyncStatus:       s.Status(),
		ExecutorStatus:   s.exec.status(),
		FileClientStatus: s.fc.status(),
		ChunkPoolStatus:  s.pool.status(),
	}
}
