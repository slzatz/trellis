package main

import (
	"fmt"
	"machine"
	"math/rand"
	"time"

	"tinygo.org/x/drivers/net/mqtt"
	"tinygo.org/x/drivers/trellis"
	"tinygo.org/x/drivers/wifinina"
)

var NINA_SPI = machine.SPI0

// NINA-m4 express pins
const (
	NINA_SDO    machine.Pin = machine.PB23
	NINA_SDI    machine.Pin = machine.PB22
	NINA_CS     machine.Pin = machine.PA23
	NINA_SCK    machine.Pin = machine.PA17
	NINA_GPIO0  machine.Pin = machine.PA20
	NINA_RESETN machine.Pin = machine.PA22
	NINA_ACK    machine.Pin = machine.PA21
	NINA_TX     machine.Pin = machine.PB16
	NINA_RX     machine.Pin = machine.PB17
)

var (
	bat     = machine.ADC{machine.PB01}
	spi     = NINA_SPI
	adaptor *wifinina.Device
	cl      mqtt.Client
	topic   = "trellis"

	keyMap = map[uint8]string{
		0:  "shuffle neil young",
		1:  "shuffle patty griffin",
		2:  "shuffle aimee mann",
		3:  "shuffle lucinda williams",
		4:  "shuffle jason isbell",
		5:  "shuffle radiohead",
		6:  "shuffle tom petty",
		7:  "shuffle amanda shires",
		8:  "shuffle jackson browne",
		9:  "shuffle ani difranco",
		10: "shuffle counting crows",
		11: "album after the goldrush neil young",
		12: "random_playlist",
		13: "station patty griffin",
		14: "next",
		15: "play_pause",
	}
)

func main() {
	time.Sleep(5 * time.Second)
	//i2c := machine.I2C0
	//err := i2c.Configure(machine.I2CConfig{
	err := machine.I2C0.Configure(machine.I2CConfig{
		//Frequency: machine.TWI_FREQ_100KHZ,
		Frequency: machine.TWI_FREQ_400KHZ,
		SCL:       machine.SCL_PIN,
		SDA:       machine.SDA_PIN,
	})
	//err := machine.I2C0.Configure(machine.I2CConfig{})
	if err != nil {
		println("could not configure I2C:", err)
		return
	}

	in := machine.D5
	in.Configure(machine.PinConfig{Mode: machine.PinInput})
	in.High()

	tr := trellis.New(machine.I2C0, 0x70, 10)
	fmt.Println("New")
	tr.Configure()
	fmt.Println("Configure")
	time.Sleep(1 * time.Second)

	err = machine.SPI0.Configure(machine.SPIConfig{Frequency: 2000000}) //115200 worked
	if err != nil {
		println(err)
	}

	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		SDO:       NINA_SDO, //MOSI = machine.SPIO_SDO_PIN
		SDI:       NINA_SDI, //MISO = machine.SPIO_SDI_PIN
		SCK:       NINA_SCK, //SCK = machine.SPIO_SCK_PIN
	})

	time.Sleep(5 * time.Second)

	// Init wifit
	adaptor = wifinina.New(spi,
		NINA_CS,
		NINA_ACK,
		NINA_GPIO0,
		NINA_RESETN,
	)
	//adaptor.Configure()
	adaptor.Configure2(false)   //true = reset active high
	time.Sleep(5 * time.Second) // necessary
	s, err := adaptor.GetFwVersion()
	if err != nil {
		println("GetFwVersion Error:", err)
	}
	println("firmware:", s)

	//time.Sleep(10 * time.Second) ///////

	for {
		err := connectToAP()
		if err == nil {
			break
		}
	}

	opts := mqtt.NewClientOptions()
	clientID := "tinygo-client-" + randomString(5)
	opts.AddBroker(server).SetClientID(clientID)
	println(clientID)
	//opts.AddBroker(server).SetClientID("tinygo-client-2")

	println("Connecting to MQTT broker at", server)
	cl = mqtt.NewClient(opts)
	token := cl.Connect()

	if token.Wait() && token.Error() != nil {
		failMessage("mqtt connect", token.Error().Error())
	}

	var i uint8
	for {
		trellis.Clear()
		time.Sleep(1 * time.Millisecond)
		trellis.SetLED(i)
		time.Sleep(1 * time.Millisecond)
		//fmt.Printf("WriteDisplay\r\n")
		tr.WriteDisplay()
		i += 1
		if i == 16 {
			//i = 0
			break
		}
		time.Sleep(300 * time.Millisecond)
	}

	/*
		//lastkey := uint8(16)
		lastkey := 16
		for {
			//b := tr.ReadSwitches()
			b, key := tr.ReadSingleSwitch()
			time.Sleep(10 * time.Millisecond)
			if b {
				if lastkey != key {
					trellis.Clear()
					trellis.SetLED(uint8(key))
					time.Sleep(1 * time.Millisecond)
					tr.WriteDisplay()
					lastkey = key
					fmt.Printf("key = %d\r\n", key)
				}
			}
		}
	*/

	//j := 0
	t := time.Now()
	for {
		b := tr.ReadSwitches()
		//Need some delay here 20 ms works for i2c 400MHZ but YMMV
		time.Sleep(20 * time.Millisecond)

		var k uint8
		if b {
			for k = 0; k < 16; k++ {
				if trellis.IsKeyPressed(k) {
					//if trellis.JustPressed(k) {
					trellis.Clear()
					trellis.SetLED(k)
					time.Sleep(1 * time.Millisecond)
					tr.WriteDisplay()
					fmt.Printf("%d was pressed\r\n", k)
					if action, ok := keyMap[k]; ok {
						sendMessage2(action)
					} else {
						sendMessage(k)
					}
					break
				}
			}
		}
		//j++
		//if j == 10000 {
		if time.Since(t) > time.Minute {
			token := cl.Pingreq()
			if token.Error() != nil {
				failMessage("ping", token.Error().Error())
			}
			//j = 0
			t = time.Now()
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func connectToAP() error {
	time.Sleep(2 * time.Second)
	println("Connecting to " + ssid)
	err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	if err != nil {
		println(err)
		return err
	}

	println("Connected.")

	time.Sleep(2 * time.Second)
	ip, _, _, err := adaptor.GetIP()
	for ; err != nil; ip, _, _, err = adaptor.GetIP() {
		println(err.Error())
		time.Sleep(1 * time.Second)
	}
	println(ip.String())
	return nil
}

// Returns an int >= min, < max
func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Generate a random string of A-Z chars with len = l
func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}

func failMessage(action, msg string) {
	println(action, ": ", msg)
	time.Sleep(5 * time.Second)
}

func sendMessage(key uint8) {
	println("Publishing MQTT message...")
	data := []byte(fmt.Sprintf(`{"key":%d}`, key))
	token := cl.Publish(topic, 0, false, data)
	token.Wait()
	if err := token.Error(); err != nil {
		switch t := err.(type) {
		case wifinina.Error:
			println(t.Error(), "attempting to reconnect")
			if token := cl.Connect(); token.Wait() && token.Error() != nil {
				failMessage("mqtt send", token.Error().Error())
			}
		default:
			println(err.Error())
		}
	}
}

func sendMessage2(action string) {
	println("Publishing MQTT message...")
	//data := []byte(fmt.Sprintf(`{"key":%d}`, key))
	data := []byte(fmt.Sprintf(`{"action":%q}`, action))
	token := cl.Publish(topic, 0, false, data)
	token.Wait()
	if err := token.Error(); err != nil {
		switch t := err.(type) {
		case wifinina.Error:
			println(t.Error(), "attempting to reconnect")
			if token := cl.Connect(); token.Wait() && token.Error() != nil {
				failMessage("mqtt send", token.Error().Error())
			}
		default:
			println(err.Error())
		}
	}
}
