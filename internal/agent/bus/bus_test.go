package bus

import (
	"sync"
	"testing"
	"time"
)

func TestPublishDeliversToDomainSubscribers(t *testing.T) {
	b := New()
	ch1, ch2 := make(chan Record, 4), make(chan Record, 4)
	b.Subscribe("cpu", Func(func(rec Record) { ch1 <- rec }))
	b.Subscribe("cpu", Func(func(rec Record) { ch2 <- rec }))

	rec := Record{Domain: "cpu", At: time.Unix(0, 0), Payload: "p"}
	b.Publish(rec)

	for _, ch := range []chan Record{ch1, ch2} {
		select {
		case got := <-ch:
			if got != rec {
				t.Fatalf("got %+v, want %+v", got, rec)
			}
		case <-time.After(time.Second):
			t.Fatal("record not delivered")
		}
	}
}

func TestPublishSkipsOtherDomains(t *testing.T) {
	b := New()
	ch := make(chan Record, 4)
	b.Subscribe("cpu", Func(func(rec Record) { ch <- rec }))

	b.Publish(Record{Domain: "locks", Payload: "p"})

	select {
	case got := <-ch:
		t.Fatalf("got unexpected record %+v", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestPublishWithoutSubscriberDrops(t *testing.T) {
	New().Publish(Record{Domain: "cpu", Payload: "p"})
}

func TestConcurrentPublishAndSubscribe(t *testing.T) {
	b := New()
	var mu sync.Mutex
	got := 0
	b.Subscribe("cpu", Func(func(Record) {
		mu.Lock()
		got++
		mu.Unlock()
	}))

	const publishers, perPublisher = 8, 100
	var wg sync.WaitGroup
	for i := 0; i < publishers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perPublisher; j++ {
				b.Publish(Record{Domain: "cpu"})
			}
		}()
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if got != publishers*perPublisher {
		t.Fatalf("delivered %d records, want %d", got, publishers*perPublisher)
	}
}
