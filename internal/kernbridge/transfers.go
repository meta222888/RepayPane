package kernbridge

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"
)

type transferKind int

const (
	transferUpload transferKind = iota
	transferDownload
)

type transferJob struct {
	kind       transferKind
	client     *remote.Client
	localPath  string
	remotePath string
	totalBytes int64
	onDone     func(err error)
}

type TransferQueue struct {
	app *App
	mu  sync.Mutex

	jobs []transferJob

	active      bool
	activeKind  transferKind
	activeName  string
	progress    float64
	speedText   string
	queueRemain int
	bytesDone   int64
	batchTotal  int64
	batchDone   int64
	lastTick    time.Time
	lastBytes   int64
}

func NewTransferQueue(app *App) *TransferQueue {
	return &TransferQueue{app: app, speedText: i18n.T(i18n.KeyTransferIdle)}
}

func (q *TransferQueue) EnqueueUpload(client *remote.Client, localPath, remotePath string, onDone func(error)) {
	info, err := os.Stat(localPath)
	if err != nil {
		if onDone != nil {
			onDone(err)
		}
		return
	}
	q.enqueue(transferJob{
		kind: transferUpload, client: client, localPath: localPath,
		remotePath: remotePath, totalBytes: info.Size(), onDone: onDone,
	})
}

func (q *TransferQueue) EnqueueDownload(client *remote.Client, remotePath, localPath string, onDone func(error)) {
	st, err := client.Stat(remotePath)
	if err != nil {
		if onDone != nil {
			onDone(err)
		}
		return
	}
	q.enqueue(transferJob{
		kind: transferDownload, client: client, localPath: localPath,
		remotePath: remotePath, totalBytes: st.Size, onDone: onDone,
	})
}

func (q *TransferQueue) enqueue(job transferJob) {
	q.mu.Lock()
	if !q.active && len(q.jobs) == 0 {
		q.batchTotal = 0
		q.batchDone = 0
	}
	q.batchTotal += job.totalBytes
	q.jobs = append(q.jobs, job)
	q.queueRemain = len(q.jobs)
	q.mu.Unlock()
	q.app.host.refreshTransferUI()
	go q.pump()
}

func (q *TransferQueue) pump() {
	q.mu.Lock()
	if q.active || len(q.jobs) == 0 {
		q.mu.Unlock()
		return
	}
	q.active = true
	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	q.activeKind = job.kind
	q.activeName = filepath.Base(job.localPath)
	if job.kind == transferDownload {
		q.activeName = filepath.Base(job.remotePath)
	}
	q.queueRemain = len(q.jobs)
	q.bytesDone = 0
	if q.batchTotal > 0 {
		q.progress = float64(q.batchDone) / float64(q.batchTotal) * 100
	}
	q.lastTick = time.Now()
	q.lastBytes = 0
	q.mu.Unlock()
	q.app.host.refreshTransferUI()

	var err error
	var lastUIRefresh time.Time
	progressFn := func(n int64) {
		q.mu.Lock()
		q.bytesDone = n
		if q.batchTotal > 0 {
			q.progress = float64(q.batchDone+n) / float64(q.batchTotal) * 100
		} else if job.totalBytes > 0 {
			q.progress = float64(n) / float64(job.totalBytes) * 100
		}
		now := time.Now()
		elapsed := now.Sub(q.lastTick).Seconds()
		if elapsed >= 0.25 {
			delta := n - q.lastBytes
			if delta > 0 && elapsed > 0 {
				q.speedText = formatSpeed(float64(delta) / elapsed)
			}
			q.lastTick = now
			q.lastBytes = n
		}
		refreshUI := now.Sub(lastUIRefresh) >= 100*time.Millisecond
		if refreshUI {
			lastUIRefresh = now
		}
		q.mu.Unlock()
		if refreshUI {
			q.app.host.refreshTransferUI()
		}
	}

	switch job.kind {
	case transferUpload:
		err = job.client.UploadWithProgress(job.localPath, job.remotePath, progressFn)
	case transferDownload:
		err = job.client.DownloadWithProgress(job.remotePath, job.localPath, progressFn)
	}
	if job.onDone != nil {
		job.onDone(err)
	}
	q.mu.Lock()
	if err == nil {
		q.batchDone += job.totalBytes
	} else {
		q.batchDone += q.bytesDone
	}
	q.active = false
	q.activeName = ""
	if len(q.jobs) == 0 {
		q.progress = 0
		q.speedText = i18n.T(i18n.KeyTransferIdle)
		q.batchTotal = 0
		q.batchDone = 0
	}
	q.mu.Unlock()
	q.app.host.refreshTransferUI()
	q.pump()
}

func (q *TransferQueue) Snapshot() (active bool, progress float64, speed string, queue int, fileName string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	queue = q.queueRemain
	if q.active {
		queue++
	}
	return q.active, q.progress, q.speedText, queue, q.activeName
}

func formatSpeed(bps float64) string {
	if bps < 1024 {
		return fmt.Sprintf("%.0f B/s", bps)
	}
	if bps < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", bps/1024)
	}
	return fmt.Sprintf("%.1f MB/s", bps/(1024*1024))
}
