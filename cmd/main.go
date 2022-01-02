package main

import (
	"internal/relays"
)

func main() {
	relays := relays.New()

	relays.Serve()
}
