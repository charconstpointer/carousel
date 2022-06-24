package carousel_test

import (
	"context"
	"github.com/charconstpointer/carousel"
	"sync"
	"testing"
	"time"
)

func TestScheduler_HandleSchedule(t *testing.T) {
	exec := carousel.NewEveryCoordinator(time.Nanosecond)
	scheduler := carousel.NewCarousel(exec, nil, nil)
	members := []*carousel.Rider{
		carousel.NewRider("member1"),
		carousel.NewRider("member2"),
		carousel.NewRider("member3"),
	}
	for _, member := range members {
		if err := scheduler.AddRider(member); err != nil {
			t.Errorf("cannot add member: %v", err)
		}
	}
	var orderedResult []*carousel.Rider
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var handler carousel.HandlerFunc = func(ctx context.Context, member *carousel.Rider) error {
		orderedResult = append(orderedResult, member)
		if len(orderedResult) == len(members) {
			cancel()
		}
		return nil
	}

	if err := scheduler.HandleRide(ctx, handler); err != nil {
		t.Errorf("could not handle schedule: %v", err)
	}

	if len(orderedResult) != len(members) {
		t.Errorf("expected %d members, got %d", len(members), len(orderedResult))
	}
	for i := 0; i < len(members); i++ {
		if orderedResult[i] != members[i] {
			t.Errorf("expected %v, got %v", members[i], orderedResult[i])
		}
	}
}

type EveryOtherReadinessChecker struct {
	counter int
	mu      sync.Mutex
}

func (e *EveryOtherReadinessChecker) IsReady(context.Context, *carousel.Rider) bool {
	e.mu.Lock()
	defer func() {
		e.counter++
		e.mu.Unlock()
	}()
	if e.counter%2 == 0 {
		return true
	}
	return false
}

func TestScheduler_HandleSchedule_WithReadiness(t *testing.T) {
	everyOtherChecker := &EveryOtherReadinessChecker{}
	exec := carousel.NewEveryCoordinator(time.Nanosecond)
	scheduler := carousel.NewCarousel(exec, everyOtherChecker, nil)
	members := []*carousel.Rider{
		carousel.NewRider("member1"),
		carousel.NewRider("member2"),
		carousel.NewRider("member3"),
	}
	for _, member := range members {
		if err := scheduler.AddRider(member); err != nil {
			t.Errorf("cannot add member: %v", err)
		}
	}
	expectedOrder := []*carousel.Rider{
		members[0],
		members[2],
		members[1],
	}
	var orderedResult []*carousel.Rider
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var handler carousel.HandlerFunc = func(ctx context.Context, member *carousel.Rider) error {
		orderedResult = append(orderedResult, member)
		if len(orderedResult) == len(members) {
			cancel()
		}
		return nil
	}

	if err := scheduler.HandleRide(ctx, handler); err != nil {
		t.Errorf("could not handle schedule: %v", err)
	}

	for i := 0; i < len(expectedOrder); i++ {
		if orderedResult[i] != expectedOrder[i] {
			t.Errorf("expected %v, got %v", expectedOrder[i], orderedResult[i])
		}
	}
}
