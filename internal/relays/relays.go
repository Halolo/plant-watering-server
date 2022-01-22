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

type gpio struct {
	Chip string
	Gpio int
}

type plant struct {
	Name     string
	Path     string
	Duration int
	gpio
}

type config struct {
	Pump   gpio
	Plants []plant
}

type Relays struct {
	conf     config
	chips    map[string]*gpiod.Chip
	lines    map[plant]*gpiod.Line
	pump     *gpiod.Line
	requests chan plant
}

func New() *Relays {
	relays := &Relays{
		lines:    make(map[plant]*gpiod.Line),
		requests: make(chan plant),
	}

	configFile, err := ioutil.ReadFile("api/plants.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	err = json.Unmarshal(configFile, &relays.conf)
	if err != nil {
		log.Fatal("Error during Unmarshal: ", err)
	}

	chips := make(map[string]*gpiod.Chip)

	chips[relays.conf.Pump.Chip], err = gpiod.NewChip(relays.conf.Pump.Chip)
	if err != nil {
		log.Fatal("gpiod.NewChip failed: ", err)
	}

	relays.pump, err = chips[relays.conf.Pump.Chip].RequestLine(relays.conf.Pump.Gpio, gpiod.AsOutput(1))
	if err != nil {
		log.Fatal("RequestLine failed: ", err)
	}

	for _, plant := range relays.conf.Plants {
		if _, found := chips[plant.Chip]; !found {
			chips[plant.Chip], err = gpiod.NewChip(plant.Chip)
			if err != nil {
				log.Fatal("gpiod.NewChip failed: ", err)
			}
		}

		log.Printf("%s: [%s] gpio %d\n", plant.Name, plant.Chip, plant.Gpio)
		relays.lines[plant], err = chips[plant.Chip].RequestLine(plant.Gpio, gpiod.AsOutput(1))
		if err != nil {
			log.Fatal("RequestLine failed: ", err)
		}

		http.HandleFunc(plant.Path, relays.handler)
	}

	return relays
}

func (r *Relays) Serve() {
	go http.ListenAndServe(":8080", nil)

	for req := range r.requests {
		r.activate(&req)
	}
}

// Internal

func (r *Relays) handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	for _, plant := range r.conf.Plants {
		if plant.Path == req.URL.Path {
			select {
			case r.requests <- plant:
			default:
				err := fmt.Errorf("An action is already in progress")
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprint(w, err)
				log.Println(err)
			}
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (r *Relays) activate(plant *plant) {
	defer r.lines[*plant].SetValue(1)
	defer r.pump.SetValue(1)

	log.Printf("Activating %q for %ds\n", plant.Name, plant.Duration)

	r.lines[*plant].SetValue(0)
	r.pump.SetValue(0)
	<-time.After(time.Duration(plant.Duration) * time.Second)
}
