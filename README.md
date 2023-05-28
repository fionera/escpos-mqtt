# ESCPOS-MQTT

```
Usage of escpos-mqtt:
  -bold
        should the message be printed bold
  -broker string
        mqtt broker to connect to as IP:Port
  -client-id string
        client id for mqtt client (default "escpos-mqtt")
  -config string
        TMT20II,TMT88II,SOL802 (default "TMT20II")
  -cut-after-print
        should the paper be cut after a message
  -font-size int
        font size between 0 and 5 (default 1)
  -line-feeds-after int
        how many lines should be feeded after a message (default 1)
  -line-feeds-before int
        how many lines should be feeded before a message (default 1)
  -target string
        either IP:Port or /dev/ device file (default "/dev/usb/lp0")
  -topic string
        mqtt topic to listen on
  -topic-qos int
        quality of service. It must be 0 (at most once), 1 (at least once) or 2 (exactly one) (default 1)
```