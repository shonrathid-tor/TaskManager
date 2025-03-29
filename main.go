package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` //0000000000000000000000
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("Hello ")
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("ERROR loading .env file ", err)
		}
	}
	MONGODB_URL := os.Getenv("MONGODB_URL")
	clientOption := options.Client().ApplyURI(MONGODB_URL)
	client, err := mongo.Connect(context.Background(), clientOption)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	fmt.Println("connect to MONGODB ATLAS ")
	collection = client.Database("golang_db").Collection("todos")

	app := fiber.New()
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: "http://localhost:5173",
	// 	AllowHeaders: "Origin,Content-Type,Accept",
	// }))

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Post("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "/client/dist")
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))

}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo
	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}

		todos = append(todos, todo)
	}

	return c.JSON(todos)

}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)
	//{id : 0}

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == ("") {
		return c.Status(400).JSON(fiber.Map{
			"error": " To do cannot be empty"})
	}

	insetResult, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		return err
	}

	todo.ID = insetResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)

}

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})

}

func deleteTodo(c *fiber.Ctx) error {

	id := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid to ID "})

	}
	filer := bson.M{"_id": objectID}

	_, err = collection.DeleteOne(context.Background(), filer)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{
		"success": true})

}
