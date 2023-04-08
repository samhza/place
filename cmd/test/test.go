package main

import (
	"log"

	"samhza.com/place"
)

func main() {
	chg := place.Change{
		Time:  0,
		X1:    123,
		Y1:    456,
		X2:    789,
		Y2:    1011,
		Color: 31,
	}
	out := make([]byte, 10)
	chg.Encode(out)
	chg.Decode(out)
	log.Println(chg)
}
