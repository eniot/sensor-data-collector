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

var mongoDb *mongo.Database

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
			mongoDb = mongoClient.Database(database)

			e := echo.New()
			e.GET("/", listDevices)
			e.GET("/:id", getDevice)
			e.POST("/:id/events/_search", searchEvents)
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

func listDevices(c echo.Context) error {
	devices := mongoDb.Collection("devices")
	cur, err := devices.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	if err := cur.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	return c.JSON(200, results)
}

func getDevice(c echo.Context) error {
	devices := mongoDb.Collection("devices")
	result := devices.FindOne(context.TODO(), bson.M{"_id": c.Param("id")})
	var devObj bson.M
	err := result.Decode(&devObj)
	if err != nil {
		log.Fatal(err)
	}
	return c.JSON(200, devObj)
}

type searchQuery struct {
	Limit     int64  `json:"limit"`
	Skip      int64  `json:"skip"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

func searchEvents(c echo.Context) error {
	deviceEvents := mongoDb.Collection(c.Param("id"))

	var query = new(searchQuery)
	if err := c.Bind(query); err != nil {
		return err
	}

	startTime, err := time.Parse(time.RFC3339, query.StartTime)
	if err != nil {
		startTime = time.Now().AddDate(0, 0, -1)
	}
	endTime, err := time.Parse(time.RFC3339, query.EndTime)
	if err != nil {
		endTime = time.Now()
	}

	opts := options.Find().SetLimit(query.Limit).SetSkip(query.Skip)

	cur, err := deviceEvents.Find(context.TODO(), bson.M{
		"_id": bson.M{"$gte": startTime, "$lte": endTime},
	}, opts)
	if err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	if err := cur.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	return c.JSON(200, results)
}
