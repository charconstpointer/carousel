package carousel_test

import (
	"container/list"
	"context"
	"testing"
	"time"

	"github.com/charconstpointer/carousel"
)

func TestScheduler_HandleSchedule(t *testing.T) {
	members := list.New()
	ms := []*carousel.Rider{
		carousel.NewRider("member1"),
		carousel.NewRider("member2"),
		carousel.NewRider("member3"),
	}
	for _, m := range ms {
		members.PushBack(m)
	}
	orderedChooser := carousel.NewOrderedChooser[*carousel.Rider](members)
	readinessChecker := carousel.NewNoopReadinessChecker[*carousel.Rider]()
	readinessChooser := carousel.NewReadinessChooser[*carousel.Rider](orderedChooser, readinessChecker)
	exec := carousel.NewEveryExecutor(time.Nanosecond)
	scheduler := carousel.NewCarousel[*carousel.Rider](exec, readinessChooser, nil)

	var orderedResult []*carousel.Rider
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var handler carousel.HandlerFunc[*carousel.Rider] = func(_ context.Context, member *carousel.Rider) error {
		orderedResult = append(orderedResult, member)
		if len(orderedResult) == len(ms) {
			cancel()
		}
		return nil
	}

	if err := scheduler.HandleRide(ctx, handler); err != nil {
		t.Errorf("could not handle schedule: %v", err)
	}

	if len(orderedResult) != len(ms) {
		t.Errorf("expected %d members, got %d", len(ms), len(orderedResult))
	}
	for i := 0; i < len(ms); i++ {
		if orderedResult[i] != ms[i] {
			t.Errorf("expected %v, got %v", ms[i], orderedResult[i])
		}
	}
}
