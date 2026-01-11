// Package pipeline provides a high-level abstraction for managing Go channels with
// built-in lifecycle management, error handling, and receiver callbacks.
//
// The pipeline pattern implemented here offers several advantages over raw channels:
// - Encapsulated channel management with controlled access
// - Context-aware cancellation and graceful shutdown
// - Guarded send/close paths with synchronization
// - One-to-many fan-out to registered receivers
// - Simplified receiver registration pattern
// - Built-in logging for debugging and monitoring
package pipeline

import (
	"context"
	"sync"

	"github.com/ezex-io/gopkg/logger"
)

var _ Pipeline[int] = &pipeline[int]{}

// Pipeline defines the contract for a managed channel pipeline.
// It provides type-safe channel operations with lifecycle management.
type Pipeline[T any] interface {
	// Name returns the identifier for this pipeline instance.
	Name() string

	// Close initiates a graceful shutdown of the pipeline.
	Close()

	// IsClosed reports whether the pipeline has been closed.
	IsClosed() bool

	// Send publishes a message to the pipeline (non-blocking).
	Send(T)

	// RegisterReceiver sets the handler function for incoming messages.
	RegisterReceiver(func(T))

	// UnsafeGetChannel provides direct read access to the underlying channel
	// WARNING: This bypasses pipeline management and should be used with caution.
	UnsafeGetChannel() <-chan T
}

// pipeline implements the Pipeline interface with proper synchronization
// and lifecycle management.
type pipeline[T any] struct {
	sync.RWMutex

	ctx       context.Context
	cancel    context.CancelFunc
	name      string
	closed    bool
	ch        chan T
	receivers []func(T)
}

const defaultBufferSize = 16

type options struct {
	name       string
	bufferSize int
}

// Option configures pipeline creation.
type Option func(*options)

// WithName sets the pipeline identifier used for logging and introspection.
func WithName(name string) Option {
	return func(opt *options) {
		opt.name = name
	}
}

// WithBufferSize sets the channel buffer size (0 for unbuffered).
func WithBufferSize(size int) Option {
	return func(opt *options) {
		if size < 0 {
			size = 0
		}
		opt.bufferSize = size
	}
}

// New creates and initializes a new pipeline instance.
//
// Parameters:
//   - parentCtx: The parent context for lifecycle management
//   - opts: Functional options to configure name and buffer size
//
// Returns:
//   - A new pipeline instance ready for use
func New[T any](parentCtx context.Context, opts ...Option) Pipeline[T] {
	cfg := options{
		bufferSize: defaultBufferSize,
		name:       "",
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, cancel := context.WithCancel(parentCtx)

	pipe := &pipeline[T]{
		ctx:    ctx,
		cancel: cancel,
		name:   cfg.name,
		closed: false,
		ch:     make(chan T, cfg.bufferSize),
	}

	return pipe
}

// Name returns the identifier name of the pipeline.
func (p *pipeline[T]) Name() string {
	return p.name
}

// Send writes data to the pipeline channel in a thread-safe manner.
// It handles various context cancellation scenarios and logs appropriate messages.
//
// Parameters:
//   - data: The data to send through the pipeline
func (p *pipeline[T]) Send(data T) {
	p.RLock()
	defer p.RUnlock()

	if p.closed {
		logger.Debug("send on closed channel", "name", p.name)

		return
	}

	select {
	case <-p.ctx.Done():
		err := p.ctx.Err()
		switch err {
		case context.Canceled:
			logger.Debug("pipeline draining", "name", p.name)
		case context.DeadlineExceeded:
			logger.Warn("pipeline timeout", "name", p.name)
		default:
			logger.Error("pipeline error", "name", p.name, "error", err)
		}
	case p.ch <- data:
		// Successful send
	}
}

// RegisterReceiver registers a callback to receive every message (one-to-many fan-out).
//
// Parameters:
//   - receiver: The callback function that will process received data
//
// Note: This method is NOT thread-safe; register receivers before sending.
func (p *pipeline[T]) RegisterReceiver(receiver func(T)) {
	if len(p.receivers) == 0 {
		go p.receiveLoop()
	}

	p.receivers = append(p.receivers, receiver)
}

// receiveLoop continuously listens for incoming data and fans out to all
// registered receivers until the pipeline is closed.
func (p *pipeline[T]) receiveLoop() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case data, ok := <-p.ch:
			if !ok {
				logger.Warn("channel is closed", "name", p.name)

				return
			}

			for _, handler := range p.receivers {
				handler(data)
			}
		}
	}
}

// Close shuts down the pipeline gracefully.
// It cancels the context, closes the channel, and marks the pipeline as closed.
// This method is idempotent - subsequent calls have no effect.
func (p *pipeline[T]) Close() {
	p.Lock()
	defer p.Unlock()

	if !p.closed {
		p.cancel()

		// Close the channel and mark pipeline as closed
		close(p.ch)
		p.closed = true
	}
}

// IsClosed checks if the pipeline has been closed.
//
// Returns:
//   - true if the pipeline is closed, false otherwise
func (p *pipeline[T]) IsClosed() bool {
	p.RLock()
	defer p.RUnlock()

	return p.closed
}

// UnsafeGetChannel provides direct read access to the underlying channel.
// WARNING: Bypasses all pipeline safeguards.
//
// Returns:
//   - The underlying receive channel
func (p *pipeline[T]) UnsafeGetChannel() <-chan T {
	return p.ch
}
