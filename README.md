# 🐢 🐊 🦎 🐍

Basically-working alternate *engine* (not bot) for [Halite 3](https://github.com/HaliteChallenge/Halite-III).

Dubnium has its own mapgen. The seeds are within a hair of being compatible with Official, but the generated maps tend to have slight differences in halite values (i.e. the value might be out by 1).

If you want to you can also load the map from a (decompressed, plain-JSON) replay:

`./dubnium.exe -f replay.json bot1.exe bot2.exe bot3.exe bot4.exe`

On the same map, using deterministic bots, Dubnium should produce the exact same outcome as Official, except that ships are sent to the bots in a different order, which will cause discrepancies for bots that don't sort their ships.
