package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	"samhza.com/place"
)

func main() {
	r := csv.NewReader(os.Stdin)
	r.ReuseRecord = true
	wr := bufio.NewWriter(os.Stdout)
	changes := make([][10]byte, 160353104)
	i := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		if record[0] == "timestamp" {
			continue
		}
		const layout = "2006-01-02 15:04:05.999 MST"
		t, err := time.Parse(layout, record[0])
		if err != nil {
			log.Fatalln(err)
		}
		change := place.Change{}
		change.Time = int(t.UnixMilli() - int64(place.Epoch))
		change.Color = color(record[2])
		coords := strings.Split(record[3], ",")
		change.X1 = atoi(coords[0])
		change.Y1 = atoi(coords[1])
		if len(coords) > 2 {
			change.X2 = atoi(coords[2])
			change.Y2 = atoi(coords[3])
		}
		change.Encode(changes[i][:])
		i++
	}
	sort.Slice(changes, func(i, j int) bool {
		itime := int(changes[i][0]) | int(changes[i][1])<<8 | int(changes[i][2])<<16 | int(changes[i][3]&0x1F)<<24
		jtime := int(changes[j][0]) | int(changes[j][1])<<8 | int(changes[j][2])<<16 | int(changes[j][3]&0x1F)<<24
		if itime == jtime {
			return i < j
		}
		return itime < jtime
	})
	for _, change := range changes {
		_, err := wr.Write(change[:])
		if err != nil {
			log.Fatalln(err)
		}
	}
	if err := wr.Flush(); err != nil {
		log.Fatalln(err)
	}

	f, err := os.Create("memprofile")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	runtime.GC()    // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalln(err)
	}
	return n
}

func color(s string) int {
	switch s {
	case "#000000":
		return 0
	case "#00756F":
		return 1
	case "#009EAA":
		return 2
	case "#00A368":
		return 3
	case "#00CC78":
		return 4
	case "#00CCC0":
		return 5
	case "#2450A4":
		return 6
	case "#3690EA":
		return 7
	case "#493AC1":
		return 8
	case "#515252":
		return 9
	case "#51E9F4":
		return 10
	case "#6A5CFF":
		return 11
	case "#6D001A":
		return 12
	case "#6D482F":
		return 13
	case "#7EED56":
		return 14
	case "#811E9F":
		return 15
	case "#898D90":
		return 16
	case "#94B3FF":
		return 17
	case "#9C6926":
		return 18
	case "#B44AC0":
		return 19
	case "#BE0039":
		return 20
	case "#D4D7D9":
		return 21
	case "#DE107F":
		return 22
	case "#E4ABFF":
		return 23
	case "#FF3881":
		return 24
	case "#FF4500":
		return 25
	case "#FF99AA":
		return 26
	case "#FFA800":
		return 27
	case "#FFB470":
		return 28
	case "#FFD635":
		return 29
	case "#FFF8B8":
		return 30
	case "#FFFFFF":
		return 31
	}

	return 0
}
