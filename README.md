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

## Templates

escpos-mqtt allows using formatting strings inside the message. 
If a setting (e.g. the font-size) is overriden inside a template, 
the default won't kick back in inside the same message.

Formatting can be prevented by adding a backslash before a formatting sequence.

A formatting sequence is build like this:
```
[CONTENT](COMMAND,...ARGUMENTS)
```

Valid commands are:
- BOLD
- REVERSE
- FONTSIZE
- BARCODE
- QRCODE
- CUT

`CONTENT` is optional and only supported for `BARCODE` and `QRCODE` types.
`BOLD` and`REVERSE` are booleans and can be used like that.
`BARCODE` has the following arguments:
- type UPCA, UPCE, EAN13, EAN8
- x-size 2-6
- y-size 2-6

`QRCODE` has the following arguments:
- model 1, 2
- size 1-16
- correction 48-51

```
[](FONTSIZE,5)Big[](BOLD)Bold[](BOLD)[](REVERSE)Reversed Colors[](REVERSE)Text
[CONTENT](BARCODE,UPCA,2,2)
[CONTENT](QRCODE,2,2)
```