// Copyright 2018 The logrange Authors
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

package embed

import (
	"context"
	"fmt"
	"time"

	"github.com/jrivets/log4g"
	"github.com/logrange/linker"
	"github.com/logrange/range/pkg/cluster"
	"github.com/logrange/range/pkg/cluster/model"
	"github.com/logrange/range/pkg/kv/inmem"
	"github.com/logrange/range/pkg/records/chunk/chunkfs"
	"github.com/logrange/range/pkg/records/journal"
	"github.com/logrange/range/pkg/records/journal/ctrlr"
)

type (
	// Config defines
	JCtrlrConfig struct {

		// JournalsDir the direcotry where journals are stored
		JournalsDir string

		// MaxOpenFileDescs defines how many file descriptors could be used
		// for read operations
		MaxOpenFileDescs int

		// CheckFullScan defines that the FullCheck() of IdxChecker will be
		// called. Normally LightCheck() is called only.
		CheckFullScan bool

		// RecoverDisabled flag defines that actual recover procedure should not
		// be run when the check chunk data test is failed.
		RecoverDisabled bool

		// RecoverLostDataOk flag defines that the chunk file could be truncated
		// if the file is partually corrupted.
		RecoverLostDataOk bool

		// WriteIdleSec specifies the writer idle timeout. It will be closed if no
		// write ops happens during the timeout
		WriteIdleSec int

		// WriteFlushMs specifies data sync timeout. Buffers will be synced after
		// last write operation.
		WriteFlushMs int

		// MaxChunkSize defines the maximum chunk size.
		MaxChunkSize int64

		// MaxRecordSize defines the maimum one record size
		MaxRecordSize int64
	}

	// JCtrlr implements journal.Controller interface
	JCtrlr struct {
		inj      *linker.Injector
		delegate journal.Controller
	}
)

const (
	// DefaultMaxChunkSize 100Mb by default
	DefaultMaxChunkSize = int64(100 * 1024 * 1024)
)

// NewJCtrlr returns JCtrlr for stand-alone embedded range instance
func NewJCtrlr(cfg JCtrlrConfig) (res *JCtrlr, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("Could not initialize NewJCtrlr: err=%v", r)
		}
	}()

	if cfg.JournalsDir == "" {
		return nil, fmt.Errorf("JournalsDir in config must not be empty")
	}

	if cfg.MaxChunkSize <= 0 {
		cfg.MaxChunkSize = DefaultMaxChunkSize
	}

	injector := linker.New()
	injector.SetLogger(log4g.GetLogger("embed.injector"))
	delegate := ctrlr.NewJournalController()
	injector.Register(
		linker.Component{Name: "", Value: inmem.New()},
		linker.Component{Name: "HostRegistryConfig", Value: &cfg},
		linker.Component{Name: "JournalControllerConfig", Value: &cfg},
		linker.Component{Name: "", Value: model.NewHostRegistry()},
		linker.Component{Name: "", Value: model.NewJournalCatalog()},
		linker.Component{Name: "", Value: delegate},
	)

	injector.Init(context.Background())

	res = new(JCtrlr)
	res.inj = injector
	res.delegate = delegate
	return res, nil
}

// GetOrCreate is part of journal.Controller interface
func (jc *JCtrlr) GetOrCreate(ctx context.Context, jname string) (journal.Journal, error) {
	return jc.delegate.GetOrCreate(ctx, jname)
}

// Close closes the JCtrlr
func (jc *JCtrlr) Close() error {
	jc.inj.Shutdown()
	return nil
}

// HostId is part of model.HostRegistryConfig
func (c *JCtrlrConfig) HostId() cluster.HostId {
	return cluster.HostId(0)
}

// Localhost is part of model.HostRegistryConfig
func (c *JCtrlrConfig) Localhost() model.HostInfo {
	return model.HostInfo{cluster.LocalHostRpcAddr}
}

// LeaseTTL is part of model.HostRegistryConfig
func (c *JCtrlrConfig) LeaseTTL() time.Duration {
	return time.Minute
}

// RegisterTimeout is part of model.HostRegistryConfig
func (c *JCtrlrConfig) RegisterTimeout() time.Duration {
	return 0
}

// FdPoolSize returns size of chunk.FdPool
func (c *JCtrlrConfig) FdPoolSize() int {
	fps := 100
	if c.MaxOpenFileDescs > 0 {
		fps = c.MaxOpenFileDescs
	}
	return fps
}

// StorageDir returns path to the dir on the local File system, where journals are stored
func (c *JCtrlrConfig) StorageDir() string {
	return c.JournalsDir
}

// GetChunkConfig returns chunkfs.Config object, which will be used for constructing
// chunks
func (c *JCtrlrConfig) GetChunkConfig() chunkfs.Config {
	return chunkfs.Config{
		CheckFullScan:     c.CheckFullScan,
		RecoverDisabled:   c.RecoverDisabled,
		RecoverLostDataOk: c.RecoverLostDataOk,
		WriteIdleSec:      c.WriteIdleSec,
		WriteFlushMs:      c.WriteFlushMs,
		MaxChunkSize:      c.MaxChunkSize,
		MaxRecordSize:     c.MaxRecordSize,
	}
}
