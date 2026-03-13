package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ArmanAvanesyan/authsentinel/pkg/session"
	"github.com/redis/go-redis/v9"
)

// Store implements SessionStore, PKCEStore, and RefreshLockStore using Redis.
type Store struct {
	client *redis.Client
	layout session.KeyLayout
}

// New creates a Redis store. url is e.g. "redis://localhost:6379/0".
func New(ctx context.Context, url string, layout session.KeyLayout) (*Store, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("redis url: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Store{client: client, layout: layout}, nil
}

// Close closes the Redis client.
func (s *Store) Close() error {
	return s.client.Close()
}

// SessionStore returns a session.SessionStore implemented by this Store.
func (s *Store) SessionStore() session.SessionStore { return (*sessionStoreImpl)(s) }

// PKCEStore returns a session.PKCEStore implemented by this Store.
func (s *Store) PKCEStore() session.PKCEStore { return (*pkceStoreImpl)(s) }

// RefreshLockStore returns a session.RefreshLockStore implemented by this Store.
func (s *Store) RefreshLockStore() session.RefreshLockStore { return (*refreshLockStoreImpl)(s) }

type sessionStoreImpl Store

func (s *sessionStoreImpl) Get(ctx context.Context, sessionID string) (*session.Session, error) {
	return (*Store)(s).getSession(ctx, sessionID)
}
func (s *sessionStoreImpl) Set(ctx context.Context, sessionID string, sess *session.Session, ttlSeconds int) error {
	return (*Store)(s).setSession(ctx, sessionID, sess, ttlSeconds)
}
func (s *sessionStoreImpl) Delete(ctx context.Context, sessionID string) error {
	return (*Store)(s).deleteSession(ctx, sessionID)
}

func (s *Store) getSession(ctx context.Context, sessionID string) (*session.Session, error) {
	key := s.layout.SessionKey(sessionID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var sess session.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}
func (s *Store) setSession(ctx context.Context, sessionID string, sess *session.Session, ttlSeconds int) error {
	key := s.layout.SessionKey(sessionID)
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	ttl := time.Duration(ttlSeconds) * time.Second
	return s.client.Set(ctx, key, data, ttl).Err()
}
func (s *Store) deleteSession(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, s.layout.SessionKey(sessionID)).Err()
}

type pkceStoreImpl Store

func (s *pkceStoreImpl) Get(ctx context.Context, state string) (*session.PKCEState, error) {
	return (*Store)(s).getPKCE(ctx, state)
}
func (s *pkceStoreImpl) Set(ctx context.Context, state string, p *session.PKCEState, ttlSeconds int) error {
	return (*Store)(s).setPKCE(ctx, state, p, ttlSeconds)
}
func (s *pkceStoreImpl) Delete(ctx context.Context, state string) error {
	return (*Store)(s).deletePKCE(ctx, state)
}

func (s *Store) getPKCE(ctx context.Context, state string) (*session.PKCEState, error) {
	key := s.layout.PKCEKey(state)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var p session.PKCEState
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
func (s *Store) setPKCE(ctx context.Context, state string, p *session.PKCEState, ttlSeconds int) error {
	key := s.layout.PKCEKey(state)
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	ttl := time.Duration(ttlSeconds) * time.Second
	return s.client.Set(ctx, key, data, ttl).Err()
}
func (s *Store) deletePKCE(ctx context.Context, state string) error {
	return s.client.Del(ctx, s.layout.PKCEKey(state)).Err()
}

type refreshLockStoreImpl Store

func (s *refreshLockStoreImpl) Obtain(ctx context.Context, sessionID string, ttlSeconds int) (bool, error) {
	return (*Store)(s).obtainRefreshLock(ctx, sessionID, ttlSeconds)
}
func (s *refreshLockStoreImpl) Release(ctx context.Context, sessionID string) error {
	return (*Store)(s).releaseRefreshLock(ctx, sessionID)
}

func (s *Store) obtainRefreshLock(ctx context.Context, sessionID string, ttlSeconds int) (bool, error) {
	key := s.layout.RefreshLockKey(sessionID)
	ttl := time.Duration(ttlSeconds) * time.Second
	return s.client.SetNX(ctx, key, "1", ttl).Result()
}
func (s *Store) releaseRefreshLock(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, s.layout.RefreshLockKey(sessionID)).Err()
}
