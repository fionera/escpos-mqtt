package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/hennedo/escpos"
)

var (
	target          string
	config          string
	clientID        string
	broker          string
	topic           string
	topicQOS        int
	cutAfterPrint   bool
	fontSize        int
	bold            bool
	lineFeedsBefore int
	lineFeedsAfter  int
)

func init() {
	flag.StringVar(&target, "target", "/dev/usb/lp0", "either IP:Port or /dev/ device file")
	flag.StringVar(&config, "config", "TMT20II", "TMT20II,TMT88II,SOL802")
	flag.StringVar(&clientID, "client-id", "escpos-mqtt", "client id for mqtt client")
	flag.StringVar(&broker, "broker", "", "mqtt broker to connect to as IP:Port")
	flag.StringVar(&topic, "topic", "", "mqtt topic to listen on")
	flag.IntVar(&topicQOS, "topic-qos", 1, "quality of service. It must be 0 (at most once), 1 (at least once) or 2 (exactly one)")
	flag.BoolVar(&cutAfterPrint, "cut-after-print", false, "should the paper be cut after a message")
	flag.IntVar(&fontSize, "font-size", 1, "font size between 0 and 5")
	flag.BoolVar(&bold, "bold", false, "should the message be printed bold")
	flag.IntVar(&lineFeedsBefore, "line-feeds-before", 1, "how many lines should be feeded before a message")
	flag.IntVar(&lineFeedsAfter, "line-feeds-after", 1, "how many lines should be feeded after a message")
}

func main() {
	flag.Parse()

	if broker == "" {
		log.Fatal("missing broker address")
	}
	if topic == "" {
		log.Fatal("missing mqtt topic")
	}

	var socket io.ReadWriteCloser
	if strings.Contains(target, "/") {
		log.Printf("got device as target: %s", target)
		open, err := os.OpenFile(target, os.O_RDWR, 0)
		if err != nil {
			log.Fatal(err)
		}
		socket = open
	} else {
		log.Printf("got address as target: %s", target)
		open, err := net.Dial("tcp", target)
		if err != nil {
			log.Fatal(err)
		}
		socket = open
	}

	defer socket.Close()

	p := escpos.New(socket)

	switch strings.ToUpper(config) {
	case "TMT20II":
		p.SetConfig(escpos.ConfigEpsonTMT20II)
	case "TMT88II":
		p.SetConfig(escpos.ConfigEpsonTMT88II)
	case "SOL802":
		p.SetConfig(escpos.ConfigSOL802)
	default:
		log.Fatalf("unknown device type: %s", config)
	}

	printFunc := p.Print
	if cutAfterPrint {
		printFunc = p.PrintAndCut
	}

	messagePubHandler := func(client mqtt.Client, message mqtt.Message) {
		for i := 0; i < lineFeedsBefore; i++ {
			_, err := p.LineFeed()
			if err != nil {
				log.Fatalf("failed calling escpos.LineFeed: %v", err)
			}
		}

		_, err := p.Bold(bold).Size(uint8(fontSize), uint8(fontSize)).Write(string(message.Payload()))
		if err != nil {
			log.Fatalf("failed calling escpos.Write: %v", err)
		}

		for i := 0; i < lineFeedsAfter; i++ {
			_, err := p.LineFeed()
			if err != nil {
				log.Fatalf("failed calling escpos.LineFeed: %v", err)
			}
		}

		if err := printFunc(); err != nil {
			log.Fatalf("failed calling printFunc: %v", err)
		}
	}

	options := mqtt.NewClientOptions()
	options.AddBroker("tcp://" + broker)
	options.SetClientID(clientID)
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = func(_ mqtt.Client) {
		log.Printf("connected to mqtt broker")
	}
	options.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Printf("lost connection to mqtt broker: %v", err)
	}

	cl := mqtt.NewClient(options)
	if token := cl.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to connect to broker: %v", token.Error())
	}

	if token := cl.Subscribe(topic, byte(topicQOS), nil); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to subscribe to topic: %v", token.Error())
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("disconnecting from broker")
	cl.Disconnect(100)
}
