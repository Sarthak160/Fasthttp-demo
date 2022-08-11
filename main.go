package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	//fast "main/fasthttp"
	"time"

	//"github.com/fasthttp/router"
	//"github.com/qiangxue/fasthttp-routing"
	"github.com/keploy/go-sdk/integrations/kfasthttp"
	"github.com/keploy/go-sdk/integrations/kmongo"
	// "github.com/keploy/go-sdk/integrations/kmongo"

	"github.com/keploy/go-sdk/keploy"
	"github.com/valyala/fasthttp"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

func index(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("Welcome!")

}

// func hello(ctx *fasthttp.RequestCtx) {
// 	switch string(ctx.Method()) {
// 	case "POST":
// 		insertPerson(ctx)
// 	default:
// 		ctx.Error("not found", fasthttp.StatusNotFound)
// 	}
// 	//fmt.Println(string(ctx.Request.URI().FullURI()))
// 	// name := string(ctx.URI().LastPathSegment())
// 	//fmt.Fprintf(ctx, "Hello, %s!\n", ctx.UserValue("name"))

// }

type Name struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	LastName  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

func insertPerson(c *fasthttp.RequestCtx) {
	//fmt.Println("method start")
	c.Response.Header.Add("content-type", "application/json")
	//fmt.Println("header ")
	person := Name{}
	//fmt.Println("name")
	g := bytes.NewReader(c.PostBody())

	if err := json.NewDecoder(g).Decode(&person); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("decoder")

	//col = kmongo.NewCollection((client.Database("anything")).Collection("persons"))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := col.InsertOne(ctx, person)
	if err != nil {
		logger.Fatal("error in db", zap.Error(err))
	}

	json.NewEncoder(c.Response.BodyWriter()).Encode(result)

}

var col *kmongo.Collection

//var client *mongo.Client
var logger *zap.Logger

func main() {

	logger, _ = zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	conn := "mongodb://localhost:27017"
	dbname := "anything"
	colName := "persons"
	client, err := mongo.NewClient(options.Client().ApplyURI(conn))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		fmt.Println("could not ping to mongo db service: %v\n", err)
		return
	}

	fmt.Println("connected to nosql database:", conn)

	db := client.Database(dbname)

	// integrate keploy with mongo
	col = kmongo.NewCollection(db.Collection(colName))
	fmt.Println("mongodb connection success")

	fmt.Println("collection instance is ready")
	k := keploy.New(keploy.Config{
		App: keploy.AppConfig{
			Name: "Super-Faast",
			Port: "8080",
		},
		Server: keploy.ServerConfig{
			URL: "http://localhost:8081/api",
		},
	})

	//keploy.SetTestMode()
	mw := kfasthttp.FastHttpMiddlware(k)
	m := func(ctx *fasthttp.RequestCtx) {
	
		switch string(ctx.Path()) {
		case "/index":
			index(ctx)
		case "/hello":
			insertPerson(ctx)
		// case "/insert":
		// 	bazHandler.HandlerFunc(ctx)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}
	log.Fatal(fasthttp.ListenAndServe(":8080", mw(m)))

}