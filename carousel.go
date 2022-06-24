package carousel

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mileusna/crontab"
)

type Rider struct {
	lastPresented *time.Time
	name          string
}

func NewRider(name string) *Rider {
	return &Rider{
		name: name,
	}
}

type Executor interface {
	Execute(ctx context.Context, handler func()) error
}

type Carousel[T any] struct {
	mu               sync.Mutex
	runner           Executor
	readinessChecker ReadinessChecker[T]
	timeout          *time.Duration
	chooser          Chooser[T]
}

func NewCarousel[T any](exec Executor, chooser Chooser[T], timeout *time.Duration) *Carousel[T] {
	return &Carousel[T]{
		runner:  exec,
		timeout: timeout,
		chooser: chooser,
	}
}

type HandlerFunc[T any] func(context.Context, T) error

type Chooser[T any] interface {
	Choose(context.Context) (T, error)
}

func NewOrderedChooser[T any](members *list.List) *OrderedChooser[T] {
	return &OrderedChooser[T]{
		members: members,
	}
}

type OrderedChooser[T any] struct {
	members *list.List
	mu      sync.Mutex
}

func (c *OrderedChooser[T]) Choose(_ context.Context) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	chosen, ok := c.members.Front().Value.(T)
	if !ok {
		return chosen, fmt.Errorf("cannot cast list.Element to T")
	}
	c.members.MoveToBack(c.members.Front())
	return chosen, nil
}

type ReadinessChooser[T any] struct {
	base             Chooser[T]
	readinessChecker ReadinessChecker[T]
}

func NewReadinessChooser[T any](base Chooser[T], readinessChecker ReadinessChecker[T]) *ReadinessChooser[T] {
	return &ReadinessChooser[T]{
		base:             base,
		readinessChecker: readinessChecker,
	}
}

func (r *ReadinessChooser[T]) Choose(ctx context.Context) (T, error) {
	for {
		var chosen T
		chosen, err := r.base.Choose(ctx)
		if err != nil {
			return chosen, err
		}

		ok, err := r.readinessChecker.IsReady(ctx, chosen)
		if err != nil {
			return chosen, err
		}
		if !ok {
			continue
		}
		return chosen, nil
	}
}

type NoopReadinessChecker[T any] struct {
}

func NewNoopReadinessChecker[T any]() *NoopReadinessChecker[T] {
	return &NoopReadinessChecker[T]{}
}

func (c *NoopReadinessChecker[T]) IsReady(_ context.Context, _ T) (bool, error) {
	return true, nil
}

// HandleRide calls runner to execute a closure that will choose a rider and then apply handler function on the rider
func (s *Carousel[T]) HandleRide(ctx context.Context, handler HandlerFunc[T]) error {
	defer ctx.Done()
	errs := make(chan error)
	if err := s.runner.Execute(ctx, func() {
		err := s.handleChoice(ctx, handler)
		if err != nil {
			errs <- err
			return
		}
	}); err != nil {
		return fmt.Errorf("error adding job: %w", err)
	}
	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (s *Carousel[T]) handleChoice(ctx context.Context, handler HandlerFunc[T]) error {
	rider, err := s.chooser.Choose(ctx)
	if err != nil {
		return err
	}

	if err = handler(ctx, rider); err != nil {
		return err
	}
	return nil
}

type ReadinessChecker[T any] interface {
	IsReady(ctx context.Context, item T) (bool, error)
}

type CronCoordinator struct {
	exec     *crontab.Crontab
	schedule string
}

func NewCronCoordinator(schedule string) *CronCoordinator {
	return &CronCoordinator{
		exec:     crontab.New(),
		schedule: schedule,
	}
}

func (c *CronCoordinator) Execute(ctx context.Context, schedule string, handler func()) error {
	if err := c.exec.AddJob(schedule, handler); err != nil {
		return fmt.Errorf("error adding job: %w", err)
	}
	<-ctx.Done()
	c.exec.Shutdown()
	return ctx.Err()
}

type EveryExecutor struct {
	every time.Duration
}

func NewEveryExecutor(every time.Duration) *EveryExecutor {
	return &EveryExecutor{every: every}
}

func (e *EveryExecutor) Execute(ctx context.Context, handler func()) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			handler()
			time.Sleep(e.every)
		}
	}
}
