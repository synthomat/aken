package internal

import "time"

func stripDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

type Bucket struct {
	From time.Time
	To   time.Time

	Secrets []PasswordCredentials
}

func NewBucket(from, to time.Time) *Bucket {
	return &Bucket{from, to, make([]PasswordCredentials, 0)}
}

func (b *Bucket) Match(pc PasswordCredentials) bool {
	return pc.EndDateTime.After(b.From) && pc.EndDateTime.Before(b.To)
}

type Buckets []*Bucket

func NewBuckets(intervals []int) *Buckets {
	now := stripDate(time.Now())

	var buckets Buckets
	// Pre fill buckets with configured intervals
	for i := 1; i < len(intervals); i++ {
		startDay := now.AddDate(0, 0, intervals[i-1])
		endDay := startDay.AddDate(0, 0, intervals[i])

		buckets = append(buckets, NewBucket(startDay, endDay))
	}

	return &buckets
}

func (buckets *Buckets) Put(pc PasswordCredentials) bool {
	for _, bucket := range *buckets {
		if bucket.Match(pc) {
			bucket.Secrets = append(bucket.Secrets, pc)
			return true
		}
	}

	return false
}
