package starx

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
	"github.com/gorilla/websocket"
)

func loadSettings() {
	if setting, ok := env.settings[App.Config.Type]; ok && len(setting) > 0 {
		for _, fn := range setting {
			fn()
		}
	}
}

func welcomeMsg() {
	fmt.Println(asciiLogo)
}

func initSetting() {
	// init
	if App.Standalone {
		if strings.TrimSpace(env.serverId) == "" {
			log.Fatal("server running in standalone mode, but not found server id argument")
		}

		cfg, err := cluster.Server(env.serverId)
		if err != nil {
			log.Fatal(err.Error())
		}

		App.Config = cfg
	} else {
		// if server running in cluster mode, master server config require
		// initialize master server config
		if env.masterServerId == "" {
			log.Fatalf("master server id must be set in cluster mode", env.masterServerId)
		}

		if server, err := cluster.Server(env.masterServerId); err != nil {
			log.Fatalf("wrong master server config file(%s)", env.masterServerId)
		} else {
			App.Master = server
		}

		if strings.TrimSpace(env.serverId) == "" {
			// not pass server id, running in master mode
			App.Config = App.Master
		} else {
			cfg, err := cluster.Server(env.serverId)
			if err != nil {
				log.Fatal(err.Error())
			}

			App.Config = cfg
		}
	}

	// dependencies initialization
	cluster.SetAppConfig(App.Config)
}

func startup() {
	startupComps()

	go func() {
		if App.Config.IsWebsocket {
			listenAndServeWS()
		} else {
			listenAndServe()
		}
	}()

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT)
	// stop server
	select {
	case <-env.die:
		log.Infof("The app will shutdown in a few seconds")
	case s := <-sg:
		log.Infof("got signal: %v", s)
	}
	log.Infof("server: " + App.Config.Id + " is stopping...")
	shutdownComps()
	close(env.die)
}

// Enable current server accept connection
func listenAndServe() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", App.Config.Host, App.Config.Port))
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Infof("listen at %s:%d(%s)", App.Config.Host, App.Config.Port, App.Config.String())

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf(err.Error())
			continue
		}
		if App.Config.IsFrontend {
			go handler.handle(conn)
		} else {
			go remote.handle(conn)
		}
	}
}

func listenAndServeWS() {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     env.checkOrigin,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error(err)
			return
		}

		handler.HandleWS(conn)
	})

	log.Infof("listen at %s:%d(%s)", App.Config.Host, App.Config.Port, App.Config.String())

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", App.Config.Host, App.Config.Port), nil)

	if err != nil {
		log.Fatal(err.Error())
	}
}
