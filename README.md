# Dubnium 🐢

Fully working alternate *engine* (not bot) for [Halite 3](https://github.com/HaliteChallenge/Halite-III). Build with `go build dubnium.go`

Although built for fun, one of its virtues is that it works with the [Iodine](https://github.com/fohristiwhirl/iodine) realtime game viewer.

Dubnium has its own mapgen. On normal sized maps, the seeds should be compatible with Official (please report any discrepancies you see).

On the same map, using deterministic bots, Dubnium should produce the exact same outcome as Official, except that ships are sent to the bots in a different order, which will cause discrepancies for bots that don't sort their ships.
