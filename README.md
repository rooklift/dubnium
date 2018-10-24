# 🐢 🐊 🦎 🐍

Basically-working alternate *engine* (not bot) for [Halite 3](https://github.com/HaliteChallenge/Halite-III).

Dubnium has its own mapgen (though seeds are not compatible) but you can also load the map from a (decompressed, plain-JSON) replay:

`./dubnium.exe -f replay.json bot1.exe bot2.exe bot3.exe bot4.exe`

On the same map, using deterministic bots, it should produce the exact same outcome, except that ships are sent to the bots in a different order, which will cause discrepancies for bots that don't sort their ships.
