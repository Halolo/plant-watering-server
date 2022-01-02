package relays

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/warthog618/gpiod"
)

type config struct {
	Name     string
	Path     string
	Chip     int
	Gpio     int
	Duration int
}

type Relays struct {
	conf  []config
	lines map[int]*gpiod.Line
}

func New() *Relays {
	relays := &Relays{
		lines: make(map[int]*gpiod.Line),
	}

	configFile, err := ioutil.ReadFile("api/plants.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	err = json.Unmarshal(configFile, &relays.conf)
	if err != nil {
		log.Fatal("Error during Unmarshal: ", err)
	}

	var chips []*gpiod.Chip
	for c := 0; c < 2; c++ {
		chip, err := gpiod.NewChip(fmt.Sprintf("gpiochip%d", c))
		if err != nil {
			log.Fatal("gpiod.NewChip failed: ", err)
		}

		chips = append(chips, chip)
	}

	for _, plant := range relays.conf {
		log.Printf("%s: [chip %d] gpio %d\n", plant.Name, plant.Chip, plant.Gpio)
		relays.lines[plant.Gpio], err = chips[plant.Chip].RequestLine(plant.Gpio, gpiod.AsOutput(1))
		if err != nil {
			log.Fatal("RequestLine failed: ", err)
		}

		http.HandleFunc(plant.Path, relays.handler)
	}

	return relays
}

func (r *Relays) Serve() {
	http.ListenAndServe(":8080", nil)
}

// Internal

func (r *Relays) handler(w http.ResponseWriter, req *http.Request) {
	for _, plant := range r.conf {
		if plant.Path == req.URL.Path {
			log.Printf("Activating %q for %ds\n", plant.Name, plant.Duration)
			go r.activate(plant.Gpio, time.Second*time.Duration(plant.Duration))
			break
		}
	}
}

func (r *Relays) activate(line int, duration time.Duration) {
	defer r.lines[line].SetValue(1)

	r.lines[line].SetValue(0)
	<-time.After(duration)
}
