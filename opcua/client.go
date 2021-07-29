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

type Server struct {
	Server []ServerOption `json:"server"`
}

type ServerOption struct {
	Endpoint string   `json:"endpoint"`
	Policy   string   `json:"policy"`
	Mode     string   `json:"mode"`
	Cert     string   `json:"cert"`
	Key      string   `json:"key"`
	NodeID   []string `json:"nodeId"`
}

type OpcMsg struct {
	Event string `json:"event"`
	Data  Data   `json:"data"`
}

type Data struct {
	Time   string      `json:"time"`
	NodeID string      `json:"nodeid"`
	Value  interface{} `json:"value"`
}

func OpcuaClient(write chan []byte, read chan []byte) {
	// load Opcua server config
	byteValue, err := ioutil.ReadFile("plcConfig.json")
	if err != nil {
		log.Printf("err: %v \n", err)
	}
	oc := Server{}
	json.Unmarshal(byteValue, &oc)

	if len(oc.Server) != 1 {
		log.Fatal("Error - - - - ")
	}

	sc := ServerOption{
		Endpoint: oc.Server[0].Endpoint,
		Policy:   oc.Server[0].Policy,
		Mode:     oc.Server[0].Mode,
		Cert:     oc.Server[0].Cert,
		Key:      oc.Server[0].Key,
		NodeID:   oc.Server[0].NodeID}

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

	ed, err := opcua.GetEndpoints(sc.Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	ep := opcua.SelectEndpoint(ed, sc.Policy, ua.MessageSecurityModeFromString(sc.Mode))
	if ep == nil {
		log.Fatal("Failed to find suitable endpoint")
	}

	opts := []opcua.Option{
		opcua.SecurityPolicy(oc.Server[0].Policy),
		opcua.SecurityModeString(oc.Server[0].Mode),
		opcua.CertificateFile(oc.Server[0].Cert),
		opcua.PrivateKeyFile(oc.Server[0].Key),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	// Monitoring opc ua server
	c := opcua.NewClient(oc.Server[0].Endpoint, opts...)
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
	go startChanSub(ctx, m, subInterval, 0, wg, write, oc.Server[0].NodeID...)

	<-ctx.Done()
	wg.Wait()
}

func startChanSub(ctx context.Context, m *monitor.NodeMonitor, interval, lag time.Duration, wg *sync.WaitGroup, cStream chan []byte, nodes ...string) {
	ch := make(chan *monitor.DataChangeMessage, 16)
	sub, err := m.ChanSubscribe(ctx, &opcua.SubscriptionParameters{Interval: interval}, ch, nodes...)

	if err != nil {
		log.Fatal(err)
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
				// log.Printf("[channel ] sub=%d ts=%s node=%s value=%v", sub.SubscriptionID(), msg.SourceTimestamp.UTC().Format(time.RFC3339), msg.NodeID, msg.Value.Value())
				data := Data{
					Time:   msg.SourceTimestamp.UTC().Format(time.RFC3339),
					NodeID: msg.NodeID.String(),
					Value:  msg.Value.Value(),
				}
				opcMsg, err := json.Marshal(OpcMsg{
					Event: "opc",
					Data:  data})
				if err != nil {
					log.Printf("[err ] json error : %s", err)
				}
				cStream <- opcMsg
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
