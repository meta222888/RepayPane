package ui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
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
	app  *App
	mu   sync.Mutex
	jobs []transferJob

	active      bool
	progress    float64
	speedText   string
	queueRemain int
	bytesDone   int64
	lastTick    time.Time
	lastBytes   int64
}

func NewTransferQueue(app *App) *TransferQueue {
	return &TransferQueue{app: app, speedText: i18n.T(i18n.KeyTransferIdle)}
}

func (q *TransferQueue) EnqueueUpload(client *remote.Client, localPath, remotePath string, onDone func(error)) {
	info, err := os.Stat(localPath)
	if err != nil {
		onDone(err)
		return
	}
	q.enqueue(transferJob{
		kind:       transferUpload,
		client:     client,
		localPath:  localPath,
		remotePath: remotePath,
		totalBytes: info.Size(),
		onDone:     onDone,
	})
}

func (q *TransferQueue) EnqueueDownload(client *remote.Client, remotePath, localPath string, onDone func(error)) {
	st, err := client.Stat(remotePath)
	if err != nil {
		onDone(err)
		return
	}
	q.enqueue(transferJob{
		kind:       transferDownload,
		client:     client,
		localPath:  localPath,
		remotePath: remotePath,
		totalBytes: st.Size,
		onDone:     onDone,
	})
}

func (q *TransferQueue) enqueue(job transferJob) {
	q.mu.Lock()
	q.jobs = append(q.jobs, job)
	q.queueRemain = len(q.jobs)
	q.mu.Unlock()
	q.app.statusBar.RefreshTransfer()
	go q.pump()
}

func (q *TransferQueue) pump() {
	q.mu.Lock()
	if q.active {
		q.mu.Unlock()
		return
	}
	if len(q.jobs) == 0 {
		q.mu.Unlock()
		return
	}
	q.active = true
	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	q.queueRemain = len(q.jobs)
	q.bytesDone = 0
	q.progress = 0
	q.lastTick = time.Now()
	q.lastBytes = 0
	q.mu.Unlock()

	q.app.statusBar.RefreshTransfer()

	var err error
	progressFn := func(n int64) {
		q.mu.Lock()
		q.bytesDone = n
		if job.totalBytes > 0 {
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
		q.mu.Unlock()
		fyne.Do(func() { q.app.statusBar.RefreshTransfer() })
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
	q.active = false
	if len(q.jobs) == 0 {
		q.progress = 0
		q.speedText = i18n.T(i18n.KeyTransferIdle)
	}
	q.mu.Unlock()
	fyne.Do(func() { q.app.statusBar.RefreshTransfer() })

	q.pump()
}

func (q *TransferQueue) Snapshot() (active bool, progress float64, speed string, queue int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	queue = q.queueRemain
	if q.active {
		queue++
	}
	return q.active, q.progress, q.speedText, queue
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
