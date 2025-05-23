package metrics

import (
	"math"
	"sync/atomic"
)

// Gauges hold an int64 value that can be set arbitrarily.
type Gauge[T int32 | int64 | uint32 | uint64 | float32 | float64] interface {
	Snapshot() Gauge[T]
	Update(T)
	Value() T
}

// GetOrRegisterGauge returns an existing Gauge or constructs and registers a
// new StandardGauge.
func GetOrRegisterGauge[T int32 | int64 | uint32 | uint64 | float32 | float64](name string, r Registry) Gauge[T] {
	if val, ok := r.GetOrRegister(name, NewGauge[T]()).(Gauge[T]); ok {
		return val
	} else {
		panic("unsupported gauge type")
	}
}

// NewGauge constructs a new StandardGauge.
func NewGauge[T int32 | int64 | uint32 | uint64 | float32 | float64]() Gauge[T] {
	var tmp T
	switch any(tmp).(type) {
	case int32:
		if val, ok := any(&StandardGauge[int32]{0}).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case int64:
		if val, ok := any(&StandardGauge[int64]{0}).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case uint32:
		if val, ok := any(&StandardGauge[uint32]{0}).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case uint64:
		if val, ok := any(&StandardGauge[uint64]{0}).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case float32:
		if val, ok := any(&StandardGaugeFloat[float32]{0, 0}).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case float64:
		if val, ok := any(&StandardGaugeFloat[float64]{0, 0}).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	default:
		panic("unsupported gauge type")
	}
}

// NewRegisteredGauge constructs and registers a new StandardGauge.
func NewRegisteredGauge[T int32 | int64 | uint32 | uint64 | float32 | float64](name string, r Registry) Gauge[T] {
	c := NewGauge[T]()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// Int32GaugeSnapshot is a read-only copy of another Gauge.
type Int32GaugeSnapshot int32

// Snapshot returns the snapshot.
func (g Int32GaugeSnapshot) Snapshot() Gauge[int32] { return g }

// Update panics.
func (Int32GaugeSnapshot) Update(int32) {
	panic("Update called on a Int32GaugeSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g Int32GaugeSnapshot) Value() int32 { return int32(g) }

// Int64GaugeSnapshot is a read-only copy of another Gauge.
type Int64GaugeSnapshot int64

// Snapshot returns the snapshot.
func (g Int64GaugeSnapshot) Snapshot() Gauge[int64] { return g }

// Update panics.
func (Int64GaugeSnapshot) Update(int64) {
	panic("Update called on a Int64GaugeSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g Int64GaugeSnapshot) Value() int64 { return int64(g) }

// Uint32GaugeSnapshot is a read-only copy of another Gauge.
type Uint32GaugeSnapshot uint32

// Snapshot returns the snapshot.
func (g Uint32GaugeSnapshot) Snapshot() Gauge[uint32] { return g }

// Update panics.
func (Uint32GaugeSnapshot) Update(uint32) {
	panic("Update called on a Uint32GaugeSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g Uint32GaugeSnapshot) Value() uint32 { return uint32(g) }

// Uint64GaugeSnapshot is a read-only copy of another Gauge.
type Uint64GaugeSnapshot uint64

// Snapshot returns the snapshot.
func (g Uint64GaugeSnapshot) Snapshot() Gauge[uint64] { return g }

// Update panics.
func (Uint64GaugeSnapshot) Update(uint64) {
	panic("Update called on a Uint64GaugeSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g Uint64GaugeSnapshot) Value() uint64 { return uint64(g) }

// StandardGauge is the standard implementation of a Gauge and uses the
// sync/atomic package to manage a single int64 value.
type StandardGauge[T int32 | int64 | uint32 | uint64] struct {
	value T
}

// Snapshot returns a read-only copy of the gauge.
func (g *StandardGauge[T]) Snapshot() Gauge[T] {
	var tmp T
	switch any(tmp).(type) {
	case int32:
		if val, ok := any(Int32GaugeSnapshot(g.Value())).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case int64:
		if val, ok := any(Int64GaugeSnapshot(g.Value())).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case uint32:
		if val, ok := any(Uint32GaugeSnapshot(g.Value())).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case uint64:
		if val, ok := any(Uint64GaugeSnapshot(g.Value())).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	default:
		panic("unsupported gauge type")
	}

}

// Update updates the gauge's value.
func (g *StandardGauge[T]) Update(v T) {
	switch real_origin := any(&g.value).(type) {
	case *int32:
		atomic.StoreInt32(real_origin, int32(v))
	case *int64:
		atomic.StoreInt64(real_origin, int64(v))
	case *uint32:
		atomic.StoreUint32(real_origin, uint32(v))
	case *uint64:
		atomic.StoreUint64(real_origin, uint64(v))
	default:
		panic("unsupported gauge type")
	}

}

// Value returns the gauge's current value.
func (g *StandardGauge[T]) Value() T {
	return g.value
}

// Float64GaugeSnapshot is a read-only copy of another GaugeFloat64.
type Float64GaugeSnapshot float64

// Snapshot returns the snapshot.
func (g Float64GaugeSnapshot) Snapshot() Gauge[float64] { return g }

// Update panics.
func (Float64GaugeSnapshot) Update(float64) {
	panic("Update called on a Float64GaugeSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g Float64GaugeSnapshot) Value() float64 { return float64(g) }

// Float64GaugeSnapshot is a read-only copy of another GaugeFloat64.
type Float32GaugeSnapshot float32

// Snapshot returns the snapshot.
func (g Float32GaugeSnapshot) Snapshot() Gauge[float32] { return g }

// Update panics.
func (Float32GaugeSnapshot) Update(float32) {
	panic("Update called on a Float32GaugeSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g Float32GaugeSnapshot) Value() float32 { return float32(g) }

// StandardGaugeFloat64 is the standard implementation of a GaugeFloat64 and uses
// sync.Mutex to manage a single float64 value.
type StandardGaugeFloat[T float32 | float64] struct {
	short_value uint32
	long_value  uint64
}

// Snapshot returns a read-only copy of the gauge.
func (g *StandardGaugeFloat[T]) Snapshot() Gauge[T] {
	var tmp T
	switch any(tmp).(type) {
	case float32:
		if val, ok := any(Float32GaugeSnapshot(g.Value())).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case float64:
		if val, ok := any(Float64GaugeSnapshot(g.Value())).(Gauge[T]); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	default:
		panic("unsupported gauge type")
	}
}

// Update updates the gauge's value.
func (g *StandardGaugeFloat[T]) Update(v T) {
	switch real_origin := any(v).(type) {
	case float32:
		atomic.StoreUint32(&g.short_value, math.Float32bits(real_origin))
	case float64:
		atomic.StoreUint64(&g.long_value, math.Float64bits(real_origin))
	}
}

// Value returns the gauge's current value.
func (g *StandardGaugeFloat[T]) Value() T {
	var tmp T
	switch any(tmp).(type) {
	case float32:
		if val, ok := any(math.Float32frombits(atomic.LoadUint32(&g.short_value))).(T); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	case float64:
		if val, ok := any(math.Float64frombits(atomic.LoadUint64(&g.long_value))).(T); ok {
			return val
		} else {
			panic("unsupported gauge type")
		}
	}
	panic("unsupported gauge type")
}
