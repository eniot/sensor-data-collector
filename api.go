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
			e.POST("/:id/events/_count", countEvents)
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
	Limit        int64  `json:"limit"`
	Skip         int64  `json:"skip"`
	StartTimeStr string `json:"startTime"`
	EndTimeStr   string `json:"endTime"`
	StartTime    time.Time
	EndTime      time.Time
}

type eventRow struct {
	ID time.Time `bson:"_id"`
}

func getSearchQuery(c echo.Context) (query *searchQuery, err error) {
	query = new(searchQuery)
	if err := c.Bind(query); err != nil {
		return nil, err
	}
	query.StartTime, err = time.Parse(time.RFC3339, query.StartTimeStr)
	if err != nil {
		query.StartTime = time.Now().AddDate(0, 0, -1)
	}
	query.EndTime, err = time.Parse(time.RFC3339, query.EndTimeStr)
	if err != nil {
		query.EndTime = time.Now()
	}
	return query, nil
}

func searchEvents(c echo.Context) error {
	query, err := getSearchQuery(c)
	if err != nil {
		return err
	}
	deviceEvents := mongoDb.Collection(c.Param("id"))
	opts := options.Find().SetLimit(query.Limit).SetSkip(query.Skip)
	cur, err := deviceEvents.Find(context.TODO(), bson.M{
		"_id": bson.M{"$gte": query.StartTime, "$lte": query.EndTime},
	}, opts)
	if err != nil {
		log.Fatal(err)
	}
	var results []time.Time
	var row eventRow
	for cur.Next(context.TODO()) {
		cur.Decode(&row)
		results = append(results, row.ID)
	}
	return c.JSON(200, results)
}

func countEvents(c echo.Context) error {
	query, err := getSearchQuery(c)
	if err != nil {
		return err
	}
	deviceEvents := mongoDb.Collection(c.Param("id"))
	count, err := deviceEvents.CountDocuments(context.TODO(), bson.M{
		"_id": bson.M{"$gte": query.StartTime, "$lte": query.EndTime},
	})
	if err != nil {
		log.Fatal(err)
	}
	return c.JSON(200, count)
}
