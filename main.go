package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	collectorCmd := collectorCmd()
	if err := collectorCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func collectorCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "collect",
		Short: "Sensor data collector",
		Run: func(cmd *cobra.Command, args []string) {

			flags := cmd.Flags()
			broker := flags.Lookup("broker").Value.String()
			clientid := flags.Lookup("clientid").Value.String()
			topic := flags.Lookup("topic").Value.String()
			mongoURI := flags.Lookup("mongo").Value.String()
			database := flags.Lookup("database").Value.String()

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
			if err != nil {
				log.Fatal(err)
			}
			mongoDb := mongoClient.Database(database)
			devices := mongoDb.Collection("devices")

			opts := mqtt.NewClientOptions()
			opts.AddBroker(broker)
			opts.SetClientID(clientid)

			mqttClient := mqtt.NewClient(opts)
			token := mqttClient.Connect()
			for !token.WaitTimeout(3 * time.Second) {
			}
			if err := token.Error(); err != nil {
				log.Fatal(err)
			}

			mqttClient.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {

				var event eventPayload
				json.Unmarshal(msg.Payload(), &event)

				ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)

				opts := options.Update().SetUpsert(true)
				devices.UpdateOne(ctx, bson.M{"_id": bson.M{"$eq": event.ID}},
					bson.M{"$set": bson.M{
						"_id":  event.ID,
						"type": event.Type,
						"name": event.Name}}, opts)

				deviceEvents := mongoDb.Collection(event.ID)
				loc, _ := time.LoadLocation("UTC")
				now := time.Now().In(loc)
				_, err := deviceEvents.InsertOne(ctx, bson.M{"_id": now})
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
			})

			log.Printf("Waiting for messages. To exit press CTRL+C")
			<-make(chan bool)
		},
	}
	cmd.Flags().StringP("broker", "b", "tcp://192.168.0.102:1883", "MQTT Broker URI")
	viper.BindPFlag("broker", cmd.Flags().Lookup("broker"))

	cmd.Flags().StringP("clientid", "i", "sensor-data-collector", "MQTT client ID")
	viper.BindPFlag("clientid", cmd.Flags().Lookup("clientid"))

	cmd.Flags().StringP("topic", "t", "res/rfbridge/device/#", "MQTT topic wildcard")
	viper.BindPFlag("topic", cmd.Flags().Lookup("topic"))

	cmd.Flags().StringP("mongo", "m", "mongodb://localhost:27017", "Mongo DB URI")
	viper.BindPFlag("mongo", cmd.Flags().Lookup("mongo"))

	cmd.Flags().StringP("database", "d", "sensor-events", "Mongo DB name")
	viper.BindPFlag("database", cmd.Flags().Lookup("database"))

	return cmd
}

type eventPayload struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Trigger bool   `json:"trigger"`
}
