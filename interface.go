package starx

import (
	"net/http"
	"strings"
	"time"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/component"
	"github.com/chrislonng/starx/session"
)

// Run server
func Run() {
	//welcomeMsg()
	parseConfig()
	initSetting()
	loadSettings()
	startup()
}

// Set special server initial function, starx.Set("oneServerType | anotherServerType", func(){})
func Set(svrTypes string, fn func()) {
	var types = strings.Split(strings.TrimSpace(svrTypes), "|")
	for _, t := range types {
		t = strings.TrimSpace(t)
		env.settings[t] = append(env.settings[t], fn)
	}
}

func SetRouter(svrType string, fn func(*session.Session) string) {
	cluster.Router(svrType, fn)
}

func Register(c component.Component) {
	comps = append(comps, c)
}

func SetServerID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		panic("empty server id")
	}
	env.serverId = id
}

// Set the path of servers.json
func SetServersConfig(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		panic("empty app path")
	}
	env.serversConfigPath = path
}

// Set heartbeat time internal
func SetHeartbeatInternal(d time.Duration) {
	heartbeatInternal = d
}

// SetCheckOriginFunc set the function that check `Origin` in http headers
func SetCheckOriginFunc(fn func(*http.Request) bool) {
	env.checkOrigin = fn
}

// EnableCluster enable cluster mode
func EnableCluster() {
	App.Standalone = false
}

// SetMasterServerID set master server id, config must be contained
// in servers.json master server id must be set when cluster mode
// enabled
func SetMasterServerID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		panic("empty master server id")
	}
	env.masterServerId = id
}
