module github.com/Halolo/plant-watering-server

go 1.17

require (
	github.com/warthog618/gpiod v0.7.1 // indirect
	internal/relays v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae // indirect

replace internal/relays => ./internal/relays
