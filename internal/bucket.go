package internal

import "time"

type Bucket struct {
	From time.Time
	To   time.Time

	Secrets []PasswordCredentials
}

func (b *Bucket) Match(pc PasswordCredentials) bool {
	return pc.EndDateTime.After(b.From) && pc.EndDateTime.Before(b.To)
}

type Buckets []*Bucket

func NewBuckets(intervals []int) *Buckets {
	now := time.Now()

	var buckets Buckets
	for i := 1; i < len(intervals); i++ {
		startDay := stripDate(now).AddDate(0, 0, intervals[i-1])
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

func stripDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func NewBucket(from, to time.Time) *Bucket {
	return &Bucket{
		stripDate(from), stripDate(to), make([]PasswordCredentials, 0),
	}
}
