package ratelimiter

import (
	"context"
	"time"

	"github.com/karlseguin/ccache/v3"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	defaultCacheSize = 1024 * 1024
	defaultTTL       = 2 * time.Second

	segmentsCount = 10
)

type localLimiterKey struct {
	Key    Key
	Action Action
}

type LocalLimiter struct {
	config Config

	lm limiter

	logger  *zap.Logger
	limits  [segmentsCount]*ccache.Cache[*rate.Limiter]
	climits map[Action]*ccache.Cache[*rate.Limiter]
}

func NewLocalLimiter(
	config Config,
	logger *zap.Logger,
	customLimits CustomLimitFuncs,
) *LocalLimiter {
	lcfg := localConfig(config.Local)
	config.Local = lcfg

	ll := &LocalLimiter{
		config:  config,
		logger:  logger,
		lm:      newLimiter(config.Rate, customLimits),
		climits: make(map[Action]*ccache.Cache[*rate.Limiter], len(lcfg.Custom)),
	}

	// all limiters.
	for i := range segmentsCount {
		ll.limits[i] = ccache.New(ccache.Configure[*rate.Limiter]().MaxSize(lcfg.CacheSize))
	}

	// custom limiters.
	for name, cfg := range lcfg.Custom {
		ll.climits[Action(name)] = ccache.New(ccache.Configure[*rate.Limiter]().MaxSize(cfg.CacheSize))
	}

	return ll
}

func (l *LocalLimiter) Limit(_ context.Context, key Key, action Action) (err error) {
	defer func() {
		localTotal.WithLabelValues(
			string(action),
		).Inc()
	}()

	// first try to check custom, if not found - try to get from cache
	lk := localLimiterKey{Key: key, Action: action}
	lCache, ok := l.climits[action]
	if !ok {
		seg := lk.segmentForKey()
		lCache = l.limits[seg]
	}

	var limiter *rate.Limiter
	item := lCache.Get(lk.key())
	if item == nil {
		limiter = l.load(lk)
		lCache.Set(lk.key(), limiter, l.config.Local.TTL)
	} else {
		limiter = item.Value()
		item.Extend(l.config.Local.TTL) // extend TTL for cache
	}

	if allow := limiter.Allow(); !allow {
		localExceed.WithLabelValues(string(action)).Inc()

		return ErrLimitExceeded
	}

	return nil

}

func (l *LocalLimiter) load(lk localLimiterKey) *rate.Limiter {
	lm := l.lm.LimitForKeyAndAction(lk.Key, lk.Action)
	limit := rate.Limit(lm)

	return rate.NewLimiter(limit, int(lm))
}

func (t localLimiterKey) key() string {
	return string(t.Key) + "_" + string(t.Action)
}

func (t localLimiterKey) segmentForKey() uint32 {
	return fnv32a(t.key()) % segmentsCount
}

func fnv32a(s string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	hash := uint32(offset32)
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= prime32
	}

	return hash
}

func localConfig(lcfg LocalConfig) LocalConfig {
	cfg := lcfg
	if lcfg.CacheSize <= 0 {
		cfg.CacheSize = defaultCacheSize
	}
	if cfg.CacheSize == 1 { // плохое поведение кеша если указать 1
		cfg.CacheSize++
	}
	if lcfg.TTL <= 0 {
		cfg.TTL = defaultTTL
	}

	for name, ccfg := range cfg.Custom {
		if ccfg.CacheSize <= 0 {
			ccfg.CacheSize = defaultCacheSize
		}
		if ccfg.CacheSize == 1 {
			ccfg.CacheSize++
		}
		cfg.Custom[name] = ccfg
	}

	return cfg
}
