package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/trellis"
)

func main() {
	time.Sleep(5 * time.Second)
	//i2c := machine.I2C0
	//err := i2c.Configure(machine.I2CConfig{
	err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_100KHZ,
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

	tr := trellis.New(machine.I2C0)
	fmt.Println("New")
	tr.Configure()
	fmt.Println("Configure")
	time.Sleep(1 * time.Second)
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
		/*
			time.Sleep(10 * time.Millisecond)
			//fmt.Printf("Button pushed = %t\r\n", b)
			var k uint8
			if b {
				fmt.Printf("key = %d\r\n", key)
				for k = 0; k < 16; k++ {
					if trellis.IsKeyPressed(k) {
						//if trellis.JustPressed(k) {
						trellis.Clear()
						trellis.SetLED(k)
						time.Sleep(1 * time.Millisecond)
						tr.WriteDisplay()
						if lastkey != k {
							lastkey = k
							fmt.Printf("%d was pressed\r\n", k) // seeing multiple of these
						}
						break
					}
				}
			}
		*/
		time.Sleep(1 * time.Millisecond)
	}
}
