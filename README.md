# sensor-data-collector

MQTT sensor data collector

## API

`sensor-data-collector api -m <mongo url>`

Other options

 - "addr", "a", ":8000", "Listen Address"
 - "mongo", "m", "mongodb://localhost:27017", "Mongo DB URI"
 - "database", "d", "sensor-events", "Mongo DB name"

## Collector

`sensor-data-collector collector -m <mongo url> -b <mqtt broker url>`

Other options

 - "mongo", "m", "mongodb://localhost:27017", "Mongo DB URI"
 - "database", "d", "sensor-events", "Mongo DB name"

 - "broker", "b", "tcp://192.168.0.102:1883", "MQTT Broker URI"
 - "clientid", "i", "sensor-data-collector", "MQTT client ID"
 - "topic", "t", "res/rfbridge/device/#", "MQTT topic wildcard"