package ratelimiter

import "math"

type Limit float64

type (
	Action string
	Key    string
)

const (
	LimitZero      Limit = 0
	LimitUnlimited Limit = math.MaxFloat64
)

func (l Limit) IsValid() bool {
	return l >= 0
}

type actionKey struct {
	Action Action
	Key    Key
}

type LimitGet func(key Key, action Action) Limit

type CustomLimitFunc func() (Action, LimitGet)

type CustomLimitFuncs []CustomLimitFunc

type limiter struct {
	limits map[Action]Limit
	keys   map[actionKey]Limit
	custom map[Action]LimitGet
}

func newLimiter(
	cfg map[string]RateConfig,
	custom CustomLimitFuncs,
) limiter {
	l := limiter{
		limits: make(map[Action]Limit, len(cfg)),
		keys:   make(map[actionKey]Limit, len(cfg)),
		custom: make(map[Action]LimitGet, len(custom)),
	}

	for action, rc := range cfg {
		l.limits[Action(action)] = Limit(rc.Limit)
		if len(rc.Keys) > 0 {
			for k, v := range rc.Keys {
				l.keys[actionKey{Action: Action(action), Key: Key(k)}] = Limit(v)
			}
		}
	}

	for _, f := range custom {
		action, fun := f()
		l.custom[action] = fun
	}

	return l
}

func (l limiter) LimitForKeyAndAction(key Key, action Action) Limit {
	// first try to check custom
	if lf, ok := l.custom[action]; ok {
		if l := lf(key, action); l.IsValid() {
			return l
		}
	}

	// then check key limit
	if l, ok := l.keys[actionKey{Action: action, Key: key}]; ok && l.IsValid() {
		return l
	}

	// in the end get main limit
	if l, ok := l.limits[action]; ok && l.IsValid() {
		return l
	}

	return LimitUnlimited
}
