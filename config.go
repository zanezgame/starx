package starx

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"net/http"
	"time"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
	"path"
	"strings"
)

var VERSION = "0.0.1"

var (
	// App represents the current server process
	App = &struct {
		Master     *cluster.ServerConfig // master server config
		Config     *cluster.ServerConfig // current server information
		Name       string                // current application name
		Standalone bool                  // current server is running in standalone mode
		StartTime  time.Time             // startup time
	}{}

	// env represents the environment of the current process, includes
	// work path and config path etc.
	env = &struct {
		wd                string                      // working path
		serversConfigPath string                      // servers config path(default: $appPath/configs/servers.json)
		masterServerId    string                      // master server id
		serverId          string                      // current process server id
		settings          map[string][]ServerInitFunc // all settings
		die               chan bool                   // wait for end application

		checkOrigin func(*http.Request) bool // check origin when websocket enabled
	}{}
)

type ServerInitFunc func()

// init default configs
func init() {
	// application initialize
	App.Name = strings.TrimLeft(path.Base(os.Args[0]), "/")
	App.Standalone = true
	App.StartTime = time.Now()

	// environment initialize
	env.settings = make(map[string][]ServerInitFunc)
	env.die = make(chan bool, 1)

	if wd, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		env.wd, _ = filepath.Abs(wd)

		// config file path
		serversConfigPath := filepath.Join(wd, "configs", "servers.json")

		if fileExist(serversConfigPath) {
			env.serversConfigPath = serversConfigPath
		}
	}
}

func parseConfig() {
	// initialize servers config
	if !fileExist(env.serversConfigPath) {
		log.Fatalf("%s not found", env.serversConfigPath)
	} else {
		f, _ := os.Open(env.serversConfigPath)
		defer f.Close()

		reader := json.NewDecoder(f)
		var servers map[string][]*cluster.ServerConfig
		for {
			if err := reader.Decode(&servers); err == io.EOF {
				break
			} else if err != nil {
				log.Errorf(err.Error())
			}
		}

		for svrType, svrs := range servers {
			for _, svr := range svrs {
				svr.Type = svrType
				cluster.Register(svr)
			}
		}
		cluster.DumpServers()
	}
}

func fileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
