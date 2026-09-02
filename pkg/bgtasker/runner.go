package bgtasker

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type TaskFunc func(context.Context) error

type task struct {
	Context  context.Context
	TaskFunc TaskFunc
}

type Runner struct {
	logger *zap.Logger

	config *TaskConfig

	name string

	tasksCh chan task
	wg      sync.WaitGroup
}

func NewTasker(
	logger *zap.Logger,
	config *TaskConfig,
	name string,
) *Runner {
	return &Runner{
		logger: logger,

		config: config,
		name:   name,
	}
}

func (r *Runner) Start(_ context.Context) error {
	r.tasksCh = make(chan task, r.config.AsyncBufferSize)
	r.wg = sync.WaitGroup{}

	for i := 0; i < r.config.AsyncWorkersNum; i++ {
		r.wg.Add(1)
		go r.asyncWorker()
	}

	return nil
}

func (r *Runner) Stop(_ context.Context) error {
	close(r.tasksCh)

	r.wg.Wait()

	return nil
}

func (r *Runner) Run(ctx context.Context, taskFunc TaskFunc) {
	switch r.config.RunnerType {
	case RunnerTypeBlock:
		r.blockRunner(ctx, taskFunc)
	case RunnerTypeDrop:
		r.dropRunner(ctx, taskFunc)
	default:
		r.dropRunner(ctx, taskFunc)
		r.logger.Error("unknown runner type, using drop", zap.String("type", r.config.RunnerType))
	}
}

func (r *Runner) dropRunner(ctx context.Context, taskFunc TaskFunc) {
	t := task{
		Context:  context.WithoutCancel(ctx),
		TaskFunc: taskFunc,
	}

	select {
	case r.tasksCh <- t:
	default:
		r.logger.Error("dropping event, channel is full")
	}
}

func (r *Runner) blockRunner(ctx context.Context, taskFunc TaskFunc) {
	r.tasksCh <- task{
		Context:  context.WithoutCancel(ctx),
		TaskFunc: taskFunc,
	}

}

func (r *Runner) asyncWorker() {
	defer r.wg.Done()

	processEvent := func(task task) {

		defer func(st time.Time) {
			taskDuration.WithLabelValues(r.name).Observe(time.Since(st).Seconds())
		}(time.Now())

		ctx, cancel := context.WithTimeout(
			task.Context,
			r.config.AsyncWorkerTimeout,
		)

		defer cancel()

		ctx, span := otel.Tracer("bgtasker").Start(
			ctx,
			"bgtasker.runner.processEvent",
		)

		defer span.End()

		if err := task.TaskFunc(ctx); err != nil {
			r.logger.Warn("failed to execute task", zap.Error(err))
		}

		r.logger.Debug("event processed")
	}

	for event := range r.tasksCh {
		processEvent(event)

		queueSize.WithLabelValues(r.name).Set(float64(len(r.tasksCh)))
	}

	r.logger.Info("stopping async worker")
}
