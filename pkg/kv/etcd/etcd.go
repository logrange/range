// Copyright 2018-2019 The logrange Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/logrange/range/pkg/kv"

	client "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/embed"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

type clientStorage struct {
	cl     *client.Client
	lessor *lessor
}

type embeddedStorage struct {
	*clientStorage
	server *embed.Etcd
}

// Config is a set of etcd connection options
type Config struct {
	Endpoints   []string
	Dialtimeout time.Duration
	Embedded    bool
	EmbedDir    string

	EmbedListenURLs []string
	EmbedClientURLs []string
}

func toUrls(s []string) ([]url.URL, error) {
	var r []url.URL
	for _, surl := range s {
		url, err := url.Parse(surl)
		if err != nil {
			return nil, err
		}
		r = append(r, *url)
	}
	return r, nil
}

func newEmbeddedStorage(cfg *Config) (kv.Storage, error) {
	ecfg := embed.NewConfig()
	ecfg.Dir = cfg.EmbedDir

	if len(cfg.EmbedClientURLs) > 0 {
		urls, err := toUrls(cfg.EmbedClientURLs)
		if err != nil {
			return nil, err
		}
		ecfg.LCUrls = urls
	}

	if len(cfg.EmbedListenURLs) > 0 {
		urls, err := toUrls(cfg.EmbedListenURLs)
		if err != nil {
			return nil, err
		}
		ecfg.LPUrls = urls
	}

	etcd, err := embed.StartEtcd(ecfg)
	if err != nil {
		return nil, err
	}

	select {
	case <-etcd.Server.ReadyNotify():
		break
	case <-time.After(6 * time.Second):
		etcd.Server.Stop() // trigger a shutdown
		return nil, fmt.Errorf("etcd took too long to start")
	}

	client, err := newClientStorage(&Config{
		Endpoints:   []string{ecfg.LCUrls[0].String()},
		Dialtimeout: time.Second,
	})

	if err != nil {
		return nil, err
	}

	s := &embeddedStorage{
		clientStorage: client,
		server:        etcd,
	}
	return s, nil
}

func newClientStorage(cfg *Config) (*clientStorage, error) {
	cl, err := client.New(client.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.Dialtimeout,
	})
	if err != nil {
		return nil, err
	}
	s := &clientStorage{
		cl: cl,
		lessor: &lessor{
			leases: map[kv.LeaseId]*lease{},
			cl:     cl,
		},
	}
	return s, nil
}

// New returns an instance of etcd storage
func New(cfg *Config) (kv.Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("invalid config")
	}
	if cfg.Embedded {
		return newEmbeddedStorage(cfg)
	}
	return newClientStorage(cfg)
}

// lessor is an implementation of kv.Lessor for etcd
type lessor struct {
	cl *client.Client

	m      sync.Mutex
	leases map[kv.LeaseId]*lease
}

type lease struct {
	id     client.LeaseID
	ttl    time.Duration
	lessor *lessor
	cancel context.CancelFunc
}

// Id returns lease ID
func (l *lease) Id() kv.LeaseId {
	return kv.LeaseId(l.id)
}

// TTL returns lease ttl value
func (l *lease) TTL() time.Duration {
	return l.ttl
}

// Release releases the lease
func (l *lease) Release() error {
	return l.lessor.release(l)
}

func (l *lease) keepAlive(ctx context.Context, cl *client.Client) {
	tck := time.NewTicker(l.ttl / 2)
	defer tck.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tck.C:
			cl.KeepAliveOnce(ctx, l.id)
		}
	}
}

func (l *lessor) NewLease(ctx context.Context, ttl time.Duration, keepAlive bool) (kv.Lease, error) {
	if ttl <= 0 {
		return nil, fmt.Errorf("ttl must be positive, but it is %v", ttl)
	}

	log.Println("ttl", int64(ttl.Seconds()))

	grant, err := l.cl.Grant(ctx, int64(ttl.Seconds()))
	if err != nil {
		return nil, err
	}

	var cancel context.CancelFunc
	if keepAlive {
		ctx, cancel = context.WithCancel(ctx)
	}
	ls := &lease{
		id:     grant.ID,
		ttl:    ttl,
		lessor: l,
		cancel: cancel,
	}

	l.m.Lock()
	defer l.m.Unlock()
	if keepAlive {
		go ls.keepAlive(ctx, l.cl)
	}

	l.leases[kv.LeaseId(grant.ID)] = ls

	return ls, nil
}

func (l *lessor) GetLease(lid kv.LeaseId) (kv.Lease, error) {
	l.m.Lock()
	defer l.m.Unlock()
	if ls, ok := l.leases[lid]; ok {
		return ls, nil
	}
	return nil, fmt.Errorf("Lease not found: %v", lid)
}

func (l *lessor) release(ls *lease) error {
	l.m.Lock()
	defer l.m.Unlock()

	if _, ok := l.leases[ls.Id()]; !ok {
		return fmt.Errorf("Lease not found: %v", ls.Id())
	}

	if ls.cancel != nil {
		// cancel keep-alive goroutine
		ls.cancel()
	}
	_, err := l.cl.Revoke(context.Background(), ls.id)
	if err != nil {
		return err
	}
	delete(l.leases, ls.Id())
	return nil
}

func (s *clientStorage) Lessor() kv.Lessor {
	return s.lessor
}

// Create adds a new record into the storage. It returns existing record with
// ErrAlreadyExists error if it already exists in the storage.
// Create returns version of the new record with error=nil
func (s *clientStorage) Create(ctx context.Context, record kv.Record) (kv.Version, error) {
	opts := []client.OpOption{client.WithPrevKV()}
	if record.Lease != 0 {
		opts = append(opts, client.WithLease(client.LeaseID(record.Lease)))
	}
	key := string(record.Key)
	rsp, err := s.cl.Txn(ctx).
		If(client.Compare(client.CreateRevision(key), "=", 0)).
		Then(client.OpPut(key, string(record.Value), opts...)).
		Commit()

	if err != nil {
		return 0, err
	}
	if !rsp.Succeeded {
		return 0, kv.ErrAlreadyExists
	}
	return kv.Version(1), nil
}

// Get retrieves the record by its key. It will return nil and an error,
// which will indicate the reason why the operation was not succesful.
// ErrNotFound is returned if the key is not found in the storage
func (s *clientStorage) Get(ctx context.Context, key kv.Key) (kv.Record, error) {
	rsp, err := s.cl.Get(ctx, etcdKey(key))
	if err != nil {
		return kv.Record{}, err
	}
	kvs := rsp.Kvs
	if len(kvs) == 0 {
		return kv.Record{}, kv.ErrNotFound
	}
	return *etcdToKvRec(kvs[0]), nil
}

func etcdToKvRec(v *mvccpb.KeyValue) *kv.Record {
	return &kv.Record{
		Key:     kv.Key(v.Key),
		Value:   kv.Value(v.Value),
		Version: kv.Version(v.Version),
		Lease:   kv.LeaseId(v.Lease),
	}
}

func etcdKey(key kv.Key) string {
	return string(key)
}

func optsForRecord(rec kv.Record) []client.OpOption {
	opts := []client.OpOption{client.WithPrevKV()}
	if rec.Lease != 0 {
		opts = append(opts, client.WithLease(client.LeaseID(rec.Lease)))
	}
	return opts
}

// GetRange returns the list of records for the range of keys (inclusively)
func (s *clientStorage) GetRange(ctx context.Context, startKey kv.Key, endKey kv.Key) (kv.Records, error) {
	rsp, err := s.cl.Get(ctx, etcdKey(startKey), client.WithRange(etcdKey(endKey)), client.WithSort(client.SortByKey, client.SortAscend))
	if err != nil {
		return nil, err
	}
	kvs := rsp.Kvs
	if len(kvs) == 0 {
		return nil, kv.ErrNotFound
	}
	var rt kv.Records
	for _, v := range rsp.Kvs {
		rt = append(rt, *etcdToKvRec(v))
	}
	return rt, nil
}

// CasByVersion compares-and-sets the record Value if the record stored
// version is same as in the provided record. The record version will be updated
// too and returned in the result.
//
// an error will contain the reason if the operation was not successful, or
// the new version will be returned otherwise
func (s *clientStorage) CasByVersion(ctx context.Context, record kv.Record) (kv.Version, error) {

	opts := optsForRecord(record)
	key := string(record.Key)

	rsp, err := s.cl.Txn(ctx).
		If(client.Compare(client.Version(key), "=", int64(record.Version))).
		Then(client.OpPut(key, string(record.Value), opts...)).
		Commit()

	if err != nil {
		return 0, err
	}
	if !rsp.Succeeded {
		return 0, kv.ErrWrongVersion
	}
	return kv.Version(record.Version + 1), nil
}

// Delete removes the record from the storage by its key. It returns
// an error if the operation was not successful.
func (s *clientStorage) Delete(ctx context.Context, key kv.Key) error {
	rsp, err := s.cl.Delete(ctx, etcdKey(key))
	if err != nil {
		return err
	}
	if rsp.Deleted == 0 {
		return kv.ErrNotFound
	}
	return nil
}

type retval struct {
	rec *kv.Record
	err error
}

// WaitForVersionChange will wait for the record version change. The
// version param contans an expected record version. The call returns
// immediately if the record is not found together with ErrNotFound, or
// if the record version is different than expected (no error this case
// is returned).
//
// the call will block the current go-routine until one of the following
// things happens:
// - context is closed
// - the record version is changed
// - the record is deleted
//
// If the record is deleted during the call (nil, nil) is returned
func (s *clientStorage) WaitForVersionChange(ctx context.Context, key kv.Key, version kv.Version) (kv.Record, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rch := s.cl.Watch(ctx, etcdKey(key))

	retc := make(chan retval)

	go func() {
		defer close(retc)
		for wr := range rch {
			for _, e := range wr.Events {
				if e.Type == mvccpb.DELETE {
					retc <- retval{
						rec: &kv.Record{},
					}
					return
				}
				kv := e.Kv
				if kv.Version != int64(version) {
					retc <- retval{
						rec: etcdToKvRec(kv),
					}
					return
				}
			}
		}
		retc <- retval{
			rec: &kv.Record{},
			err: kv.ErrNotFound,
		}
	}()

	cur, err := s.Get(ctx, key)
	if err != nil {
		return kv.Record{}, err
	}
	if cur.Version != version {
		// cancel watch goroutine
		cancel()
		return cur, nil
	}
	r := <-retc
	return *r.rec, r.err
}
