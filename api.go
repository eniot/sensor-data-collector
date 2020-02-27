package main

import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func apiCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "api",
		Short: "Sensor data api",
		Run: func(cmd *cobra.Command, args []string) {

			flags := cmd.Flags()
			addr := flags.Lookup("addr").Value.String()
			mongoURI := flags.Lookup("mongo").Value.String()
			database := flags.Lookup("database").Value.String()

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
			if err != nil {
				log.Fatal(err)
			}
			mongoDb := mongoClient.Database(database)

			e := echo.New()
			e.GET("/", func(c echo.Context) error { return listDevices(c, mongoDb) })
			e.Logger.Fatal(e.Start(addr))

		},
	}
	cmd.Flags().StringP("addr", "a", ":8000", "Listen Address")
	viper.BindPFlag("addr", cmd.Flags().Lookup("addr"))

	cmd.Flags().StringP("mongo", "m", "mongodb://localhost:27017", "Mongo DB URI")
	viper.BindPFlag("mongo", cmd.Flags().Lookup("mongo"))

	cmd.Flags().StringP("database", "d", "sensor-events", "Mongo DB name")
	viper.BindPFlag("database", cmd.Flags().Lookup("database"))

	return cmd
}

func listDevices(c echo.Context, mongoDb *mongo.Database) error {
	devices := mongoDb.Collection("devices")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cur, err := devices.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	res := make([]interface{}, 10)
	cur.All(ctx, res)
	return c.JSON(200, res)
}
