# Dubnium 🐢

Fully working alternate *engine* (not bot) for [Halite 3](https://github.com/HaliteChallenge/Halite-III). Build with `go build dubnium.go` or get a Windows binary from the [Releases](https://github.com/fohristiwhirl/dubnium/releases) section.

Dubnium has its own mapgen. The seeds should be compatible with Official (please report any discrepancies you see).

On the same map, using deterministic bots, Dubnium should produce the exact same outcome as Official, except that ships are sent to the bots in a different order, which will cause discrepancies for bots that don't sort their ships.
