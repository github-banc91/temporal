// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package queues

import (
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"go.temporal.io/server/common/tasks"
)

const (
	fifoSchedulerDefaultWeight = 100

	fifoSchedulerTaskChannelKey = 0
)

type (
	FIFOSchedulerOptions struct {
		WorkerCount dynamicconfig.IntPropertyFn
		QueueSize   int
	}

	// FIFOScheduler is used by shard level worker pool
	// and always schedule tasks in fifo order regardless
	// which namespace the task belongs to.
	FIFOScheduler struct {
		wRRScheduler tasks.Scheduler[Executable]
	}
)

func NewFIFOScheduler(
	options FIFOSchedulerOptions,
	metricsProvider metrics.MetricsHandler,
	logger log.Logger,
) *FIFOScheduler {
	return &FIFOScheduler{
		wRRScheduler: tasks.NewInterleavedWeightedRoundRobinScheduler(
			tasks.InterleavedWeightedRoundRobinSchedulerOptions[Executable, int]{
				TaskToChannelKey:   func(_ Executable) int { return fifoSchedulerTaskChannelKey },
				ChannelKeyToWeight: func(_ int) int { return fifoSchedulerDefaultWeight },
			},
			tasks.NewParallelProcessor(
				&tasks.ParallelProcessorOptions{
					QueueSize:   options.QueueSize,
					WorkerCount: options.WorkerCount,
				},
				metricsProvider,
				logger,
			),
			metricsProvider,
			logger,
		),
	}
}

func (s *FIFOScheduler) Start() {
	s.wRRScheduler.Start()
}

func (s *FIFOScheduler) Stop() {
	s.wRRScheduler.Stop()
}

func (s *FIFOScheduler) Submit(
	executable Executable,
) error {
	s.wRRScheduler.Submit(executable)
	return nil
}

func (s *FIFOScheduler) TrySubmit(
	executable Executable,
) (bool, error) {
	return s.wRRScheduler.TrySubmit(executable), nil
}
