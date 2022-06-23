package carousel_test

import (
	"context"
	"github.com/charconstpointer/carousel"
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
}
