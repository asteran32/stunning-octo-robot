package opcua

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
)

type OpcuaClientApp struct {
	cDatastream chan []byte
	cCommands   chan []byte
	configPath  string
	debug       bool
}

type Server struct {
	Server []ServerConfig `json:"server"`
}

type ServerConfig struct {
	Endpoint string   `json:"endpoint"`
	Policy   string   `json:"policy"`
	Mode     string   `json:"mode"`
	Cert     string   `json:"cert"`
	Key      string   `json:"key"`
	NodeID   []string `json:"nodeId"`
}

type opcSocketMessage struct {
	Event string  `json:"event"`
	Data  opcData `json:"data"`
}

type opcData struct {
	Time   string      `json:"time"`
	NodeID string      `json:"nodeid"`
	Value  interface{} `json:"value"`
}

func LoadClientApp(write chan []byte, read chan []byte, path string) {
	app := OpcuaClientApp{cDatastream: write, cCommands: read, configPath: path, debug: false}

	// load Opcua server config
	byteValue, err := ioutil.ReadFile(app.configPath)
	if err != nil {
		log.Printf("err: %v \n", err)
	}
	config := Server{}
	json.Unmarshal(byteValue, &config)

	for i := 0; i < len(config.Server); i++ {
		log.Printf("Start Monitoring OPC UA Server : %v .. \n", config.Server[i].Endpoint)
		sc := ServerConfig{Endpoint: config.Server[i].Endpoint,
			Policy: config.Server[i].Policy,
			Mode:   config.Server[i].Mode,
			Cert:   config.Server[i].Cert,
			Key:    config.Server[i].Key,
			NodeID: config.Server[i].NodeID}
		app.opcuaClient(&sc)
	}

}

func (app *OpcuaClientApp) opcuaClient(config *ServerConfig) {
	interval := opcua.DefaultSubscriptionInterval.String()

	subInterval, err := time.ParseDuration(interval)
	if err != nil {
		log.Fatal(err)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-signalCh
		println()
		cancel()
	}()

	endpoints, err := opcua.GetEndpoints(config.Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	ep := opcua.SelectEndpoint(endpoints, config.Policy, ua.MessageSecurityModeFromString(config.Mode))
	if ep == nil {
		log.Fatal("Failed to find suitable endpoint")
	}

	log.Print("*", ep.SecurityPolicyURI, ep.SecurityMode)

	opts := []opcua.Option{
		opcua.SecurityPolicy(config.Policy),
		opcua.SecurityModeString(config.Mode),
		opcua.CertificateFile(config.Cert),
		opcua.PrivateKeyFile(config.Key),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c := opcua.NewClient(ep.EndpointURL, opts...)
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	m, err := monitor.NewNodeMonitor(c)
	if err != nil {
		log.Fatal(err)
	}

	m.SetErrorHandler(func(_ *opcua.Client, sub *monitor.Subscription, err error) {
		log.Printf("error: sub=%d err=%s", sub.SubscriptionID(), err.Error())
	})
	wg := &sync.WaitGroup{}

	// start channel-based subscription
	wg.Add(1)
	go app.startChanSub(ctx, m, subInterval, 0, wg, config.NodeID[0], config.NodeID[1], config.NodeID[2], config.NodeID[3])

	<-ctx.Done()
	wg.Wait()
}

func (app *OpcuaClientApp) startChanSub(ctx context.Context, m *monitor.NodeMonitor, interval, lag time.Duration, wg *sync.WaitGroup, nodes ...string) {
	ch := make(chan *monitor.DataChangeMessage, 16)
	sub, err := m.ChanSubscribe(ctx, &opcua.SubscriptionParameters{Interval: interval}, ch, nodes...)

	if err != nil {
		log.Fatal(err)
		// app.cDatastream <- errMsg
	}

	defer cleanup(sub, wg)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg.Error != nil {
				log.Printf("[channel ] sub=%d error=%s", sub.SubscriptionID(), msg.Error)
			} else {
				data := opcData{Time: msg.SourceTimestamp.UTC().Format(time.RFC3339),
					NodeID: msg.NodeID.String(),
					Value:  msg.Value.Value(),
				}
				socket := opcSocketMessage{Event: "opc", Data: data}
				opcSocketMsg, err := json.Marshal(socket)
				if err != nil {
					log.Println("err: Can not convert interface to Json format.")
				}
				app.cDatastream <- opcSocketMsg
			}
			time.Sleep(lag)
		}
	}
}

func cleanup(sub *monitor.Subscription, wg *sync.WaitGroup) {
	log.Printf("stats: sub=%d delivered=%d dropped=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe()
	wg.Done()
}
