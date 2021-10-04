package scheduler

const (
	headroom   = 5
	numBuckets = 10
	bufferPercentage = 0.2
)

type bucket struct {
	count    float64
	duration float64
}

func (b *bucket) reset() {
	b.count = 0
	b.duration = 0
}

func (b *bucket) collect(duration float64) {
	b.duration += duration
	b.count++
}

type bucketManager struct {
	buckets    []bucket
	currentIdx int
}

func NewBucketManager() *bucketManager {
	return &bucketManager{
		buckets:    make([]bucket, numBuckets),
		currentIdx: 0,
	}
}

func (bm *bucketManager) ResetCurrent() {
	bm.buckets[bm.currentIdx].reset()
}

func (bm *bucketManager) Total() float64 {
	var totalCount float64
	var totalDuration float64
	for _, bucket := range bm.buckets {
		totalCount += bucket.count
		totalDuration += bucket.duration
	}
	return totalDuration / totalCount
}

func (bm *bucketManager) Last() float64 {
	bucket := bm.buckets[bm.currentIdx]
	return bucket.duration / bucket.count
}

func (bm *bucketManager) Next() {
	newIdx := (bm.currentIdx + 1) % len(bm.buckets)
	bm.buckets[newIdx].reset()
	bm.currentIdx = newIdx
}

func (bm *bucketManager) Collect(duration float64) {
	bm.buckets[bm.currentIdx].collect(duration)
}

func CalculateConcurrency(bm *bucketManager, oldConcurrency float64) float64 {
	minRTT := bm.Total()
	buffer := minRTT * bufferPercentage
	gradient := (minRTT + buffer) / bm.Last()
	limit := gradient*oldConcurrency + headroom
	return limit
}

func SetConcurrency(q *Queue, bm *bucketManager) {
	newConcurrency := CalculateConcurrency(bm,float64(q.config.MaxConcurrency))
	q.config.MaxConcurrency = int(newConcurrency)
	q.reconfigure()
	bm.Next()
}

