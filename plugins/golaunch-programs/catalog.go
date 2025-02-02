package main

import (
	"encoding/json"
	// "fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	sdk "github.com/kdar/golaunch/sdk/go"
	"github.com/kdar/golaunch/sdk/go/fuzzy"
	"github.com/kdar/golaunch/sdk/go/system"

	// "time"

	"github.com/MichaelTJones/walk"
	"github.com/boltdb/bolt"
	"github.com/kardianos/osext"
	"github.com/rjeczalik/notify"
)

type QueryResults []sdk.QueryResult

func (a QueryResults) Len() int      { return len(a) }
func (a QueryResults) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a QueryResults) Less(i, j int) bool {
	// // Items with zero usage are lowest priority
	// if a[i].Usage > 0 && a[j].Usage == 0 {
	// 	return true
	// }
	// if a[i].Usage == 0 && a[j].Usage > 0 {
	// 	return false
	// }
	//
	// localEqual := a[i].LowName == a[i].Query
	// otherEqual := a[j].LowName == a[j].Query
	//
	// // Exact match between search text and item name has higest priority
	// if localEqual && !otherEqual {
	// 	return true
	// }
	// if !localEqual && otherEqual {
	// 	return false
	// }

	return a[i].Score > a[j].Score
}

type Catalog struct {
	cfg       *Config
	db        *bolt.DB
	historydb *bolt.DB
	watcher   chan notify.EventInfo
	system    *system.System
	metadata  *sdk.Metadata

	cm      sync.Mutex
	cache   map[string]sdk.Program
	hm      sync.Mutex
	history map[string]sdk.Program
	// a synchronous flag to tell if we're syncing or not
	indexing chan struct{}
}

func NewCatalog(md *sdk.Metadata, cfg *Config, sys *system.System) *Catalog {
	return &Catalog{
		cfg:      cfg,
		system:   sys,
		metadata: md,
		cache:    make(map[string]sdk.Program),
		history:  make(map[string]sdk.Program),
		indexing: make(chan struct{}, 1),
		watcher:  make(chan notify.EventInfo, 100),
	}
}

func (c *Catalog) Init() error {
	db, err := bolt.Open("programs.db", 0666, nil)
	if err != nil {
		return err
	}
	c.db = db

	historydb, err := bolt.Open("history.db", 0666, nil)
	if err != nil {
		return err
	}
	c.historydb = historydb

	for _, source := range c.cfg.Sources {
		if err := notify.Watch(filepath.Join(os.ExpandEnv(source.Path), "./..."), c.watcher, notify.FileNotifyChangeFileName); err != nil {
			log.Println(err)
		}
	}

	// cache function for loading the db into memory
	cachefn := func(cache map[string]sdk.Program) func(tx *bolt.Tx) error {
		return func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("programs"))
			if b == nil {
				return nil
			}

			cur := b.Cursor()
			for k, v := cur.First(); k != nil; k, v = cur.Next() {
				var program sdk.Program
				json.Unmarshal(v, &program)
				cache[string(k)] = program
			}
			return nil
		}
	}

	// we will always cache the history since it's not that big
	c.historydb.View(cachefn(c.history))

	// load from the database into memory if the option is set
	if c.cfg.CacheInMemory {
		c.db.View(cachefn(c.cache))
	}

	go c.watch()

	return nil
}

func (c *Catalog) Shutdown() {
	if c.db != nil {
		c.db.Close()
	}

	if c.historydb != nil {
		c.historydb.Close()
	}

	if c.watcher != nil {
		notify.Stop(c.watcher)
	}
}

func (c *Catalog) calcScore(name string, usage, fuzzScore int) int {
	score := fuzzScore
	if usage > 0 {
		score += 10 + usage*5
	}

	if strings.Contains(name, "help") || strings.Contains(name, "documentation") {
		score -= 10
	} else if strings.Contains(name, "uninstall") {
		score -= 20
	}

	return score
}

func (c *Catalog) Query(query string) []sdk.QueryResult {
	var results QueryResults

	// FIXME: not case sensitive right now
	query = strings.ToLower(query)

	// start := time.Now()

	pwd, _ := osext.ExecutableFolder()

	contextmenu := []sdk.ContextMenuItem{{
		Label:   "Copy path",
		Enabled: true,
		Icon:    filepath.Join(pwd, "images", "copy.png"),
	}, {
		Label:   "Open containing folder",
		Enabled: true,
		Icon:    filepath.Join(pwd, "images", "open-containing-folder.png"),
	}, {
		Label:   "Run as admin",
		Enabled: true,
		Icon:    filepath.Join(pwd, "images", "shield.png"),
	}}

	if c.cfg.CacheInMemory {
		for k, v := range c.cache {
			name := filepath.Base(k[:len(k)-len(filepath.Ext(k))])
			mr := fuzzy.Match(query, name)
			if mr.Success {
				if true { // if _, err := os.Stat(v.Path); err == nil {
					lowName := strings.ToLower(name)
					results = append(results, sdk.QueryResult{
						Program:  v,
						ID:       c.metadata.ID,
						Title:    name,
						Subtitle: v.Path,
						Query:    query,
						LowName:  lowName,
						Score:    c.calcScore(lowName, v.Usage, mr.Score),
						ContextMenu: append([]sdk.ContextMenuItem{{
							Label:   name,
							Enabled: false,
						}, {
							Type: "separator",
						}}, contextmenu...),
					})
				} else {
					c.removePath(v.Path)
				}
			}
		}
	} else {
		// grab from the database instead of cache
		c.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("programs"))
			if b == nil {
				return nil
			}

			cur := b.Cursor()
			for k, v := cur.First(); k != nil; k, v = cur.Next() {
				ks := string(k)
				name := filepath.Base(ks[:len(ks)-len(filepath.Ext(ks))])
				mr := fuzzy.Match(query, name)
				if mr.Success {
					var program sdk.Program
					json.Unmarshal(v, &program)

					if true { //if _, err := os.Stat(program.Path); err == nil {
						lowName := strings.ToLower(name)
						results = append(results, sdk.QueryResult{
							Program:  program,
							ID:       c.metadata.ID,
							Title:    name,
							Subtitle: program.Path,
							Query:    query,
							LowName:  lowName,
							Score:    c.calcScore(lowName, program.Usage, mr.Score),
							ContextMenu: append([]sdk.ContextMenuItem{{
								Label:   name,
								Enabled: false,
							}, {
								Type: "separator",
							}}, contextmenu...),
						})
					} else {
						c.removePath(program.Path)
					}
				}
			}
			return nil
		})
	}

	sort.Sort(results)

	for k, v := range c.history {
		if query == k {
			for i := 0; i < len(results); i++ {
				if v.Path == results[i].Path {
					if true { // if _, err := os.Stat(v.Path); err == nil {
						name := filepath.Base(v.Path[:len(v.Path)-len(filepath.Ext(v.Path))])
						copy(results[1:i+1], results[0:i])
						results[0] = sdk.QueryResult{
							Program:  v,
							ID:       c.metadata.ID,
							Title:    name,
							Subtitle: v.Path,
							Query:    query,
							LowName:  strings.ToLower(name),
							Score:    -1,
							ContextMenu: append([]sdk.ContextMenuItem{{
								Label:   name,
								Enabled: false,
							}, {
								Type: "separator",
							}}, contextmenu...),
						}
					} else {
						c.removeHistory(k)
					}
				}
			}
		}
	}

	if len(results) > c.cfg.MaxResults {
		results = results[:c.cfg.MaxResults]
	}

	//fmt.Fprintf(os.Stderr, "query elasped time: %v\n", time.Now().Sub(start))

	return results
}

func (c *Catalog) used(qr sdk.QueryResult) error {
	qr.Program.Usage += 1

	c.history[qr.Query] = qr.Program

	if c.cfg.CacheInMemory {
		c.cm.Lock()
		c.cache[qr.Program.Path] = qr.Program
		c.cm.Unlock()
	}

	c.historydb.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("programs"))
		if err != nil {
			return err
		}

		data, _ := json.Marshal(qr.Program)
		return b.Put([]byte(qr.Query), data)
	})

	return c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("programs"))
		if err != nil {
			return err
		}

		data, _ := json.Marshal(qr.Program)
		return b.Put([]byte(qr.Program.Path), data)
	})
}

func (c *Catalog) removeHistory(key string) error {
	delete(c.history, key)

	return c.historydb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("programs"))
		return b.Delete([]byte(key))
	})
}

func (c *Catalog) addPath(path string) error {
	icon, _ := c.system.EmbeddedAppIcon(path)
	program := sdk.Program{
		Path:  path,
		Icon:  icon,
		Usage: 0,
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("programs"))
		if err != nil {
			return err
		}

		// keep usage if we're merging
		if kvbytes := b.Get([]byte(path)); len(kvbytes) != 0 {
			var kvprog sdk.Program
			if err := json.Unmarshal(kvbytes, &kvprog); err != nil {
				return err
			}
			program.Usage = kvprog.Usage
		}

		if c.cfg.CacheInMemory {
			c.cm.Lock()
			c.cache[path] = program
			c.cm.Unlock()
		}

		data, _ := json.Marshal(program)
		return b.Put([]byte(path), data)
	})
}

func (c *Catalog) removePath(path string) error {
	if c.cfg.CacheInMemory {
		c.cm.Lock()
		delete(c.cache, path)
		c.cm.Unlock()
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("programs"))
		return b.Delete([]byte(path))
	})
}

func (c *Catalog) watch() {
	for ei := range c.watcher {
		doupdate := false
		for _, source := range c.cfg.Sources {
			if filepath.HasPrefix(ei.Path(), source.Path) {
				if source.containsExt(filepath.Ext(ei.Path())) {
					doupdate = true
					//log.Printf("programs: fs event: %v", ei)
					break
				}
			}
		}

		if !doupdate {
			continue
		}

		switch ei.Event() {
		case notify.FileActionAdded:
			if err := c.addPath(ei.Path()); err != nil {
				log.Println(err)
			}
		case notify.FileActionRemoved:
			if err := c.removePath(ei.Path()); err != nil {
				log.Println(err)
			}
		}
	}
}

func (c *Catalog) IsEmpty() bool {
	empty := false
	c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("programs"))
		if b == nil {
			empty = true
			return nil
		}

		empty = b.Stats().KeyN == 0
		return nil
	})

	return empty
}

func (c *Catalog) Index() {
	select {
	case v := <-c.indexing:
		//log.Println("already indexing.")
		c.indexing <- v
		return
	default:
	}

	c.indexing <- struct{}{}

	//log.Println("started indexing...")

	c.db.NoSync = true

	//start := time.Now()
	for _, source := range c.cfg.Sources {
		err := walk.Walk(os.ExpandEnv(source.Path), func(path string, f os.FileInfo, err error) error {
			return c.walkFn(&source, path, f, err)
		})
		if err != nil {
			log.Println(err)
		}
	}
	//log.Printf("indexing elasped time: %v\n", time.Now().Sub(start))

	c.db.NoSync = false

	<-c.indexing
}

func (c *Catalog) walkFn(source *Source, path string, f os.FileInfo, err error) error {
	if len(path) == 0 {
		return nil
	}

	if f.IsDir() || !source.containsExt(filepath.Ext(path)) {
		return nil
	}

	return c.addPath(path)
}
