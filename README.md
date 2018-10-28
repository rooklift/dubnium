# Dubnium 🐢

Fully working alternate *engine* (not bot) for [Halite 3](https://github.com/HaliteChallenge/Halite-III). Build with `go build dubnium.go` or get a Windows build from the [Releases](https://github.com/fohristiwhirl/dubnium/releases).

Although built for fun, one of its virtues is that it works with the [Iodine](https://github.com/fohristiwhirl/iodine) realtime game viewer.

The RNG is deliberately identical to Official. On normal sized maps, a seed should give the same map as Official (please report any discrepancies you see).

On the same map, using deterministic bots, Dubnium should produce the exact same outcome as Official, except that ships are sent to the bots in a different order, and may be generated with different IDs (still consecutive), which may cause discrepancies for some bots.
