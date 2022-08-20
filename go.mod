module github.com/slzatz/trellis

go 1.18

replace tinygo.org/x/drivers v0.21.0 => /home/slzatz/drivers

require (
	tinygo.org/x/drivers v0.21.0
	tinygo.org/x/tinyfont v0.2.1
)

require github.com/eclipse/paho.mqtt.golang v1.2.0 // indirect
