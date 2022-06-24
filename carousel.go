package carousel

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/mileusna/crontab"
	"time"
)

var (
	ErrAlreadyMember = errors.New("presenter is already a member of the team")
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

type Coordinator interface {
	Execute(ctx context.Context, handler func()) error
}

type Carousel struct {
	members          map[*Rider]bool
	order            *list.List
	runner           Coordinator
	timeout          *time.Duration
	readinessChecker ReadinessChecker
}

func NewCarousel(exec Coordinator, readinessChecker ReadinessChecker, timeout *time.Duration) *Carousel {
	if readinessChecker == nil {
		readinessChecker = NoopReadinessChecker{}
	}
	return &Carousel{
		members:          make(map[*Rider]bool),
		order:            list.New(),
		runner:           exec,
		timeout:          timeout,
		readinessChecker: readinessChecker,
	}
}

func (s *Carousel) AddRider(p *Rider) error {
	if p == nil {
		return fmt.Errorf("presenter is nil")
	}
	if s.members[p] {
		return ErrAlreadyMember
	}
	s.members[p] = true
	s.order.PushBack(p)
	return nil
}

func (s *Carousel) RemoveRider(p *Rider) {
	delete(s.members, p)
}

type HandlerFunc func(context.Context, *Rider) error

func (s *Carousel) HandleRide(ctx context.Context, handler HandlerFunc) error {
	defer ctx.Done()
	errs := make(chan error)
	if err := s.runner.Execute(ctx, func() {
		for ready := false; !ready; {
			presenter, ok := s.order.Front().Value.(*Rider)
			if !ok {
				errs <- fmt.Errorf("cannot cast list.Element to *Rider")
			}
			ready = s.handleReadiness(ctx, presenter, s.timeout)
			s.order.MoveToBack(s.order.Front())
			if !ready {
				continue
			}
			if err := handler(ctx, presenter); err != nil {
				errs <- fmt.Errorf("handler error: %w", err)
			}
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

type ReadinessChecker interface {
	IsReady(ctx context.Context, presenter *Rider) bool
}

type NoopReadinessChecker struct {
}

func (NoopReadinessChecker) IsReady(context.Context, *Rider) bool {
	return true
}

func (s *Carousel) handleReadiness(ctx context.Context, presenter *Rider, timeout *time.Duration) bool {
	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}
	return s.readinessChecker.IsReady(ctx, presenter)
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
	select {
	case <-ctx.Done():
		c.exec.Shutdown()
		return ctx.Err()
	}
}

type EveryCoordinator struct {
	every time.Duration
}

func NewEveryCoordinator(every time.Duration) *EveryCoordinator {
	return &EveryCoordinator{every: every}
}

func (e *EveryCoordinator) Execute(ctx context.Context, handler func()) error {
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
