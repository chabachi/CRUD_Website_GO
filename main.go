package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

// Item represents the structure for the 'items' collection
type Item struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func main() {
	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize PocketBase app
	app := pocketbase.New()

	// Start the PocketBase server in the background
	go func() {
		if err := app.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	// Optional: Run additional hooks or custom logic with PocketBase events
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		log.Println("PocketBase is running...")
		return nil
	})

	// CRUD Routes

	// Create a new item
	e.POST("/items", func(c echo.Context) error {
		item := new(Item)
		if err := c.Bind(item); err != nil {
			return err
		}

		// Find the collection by name
		collection, err := app.Dao().FindCollectionByNameOrId("items")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Collection not found"})
		}

		// Create a new record for the 'items' collection
		record := models.NewRecord(collection)
		record.Set("title", item.Title)
		record.Set("description", item.Description)
		record.Set("price", item.Price)

		// Save the record using PocketBase's built-in Dao
		if err := app.Dao().SaveRecord(record); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusCreated, record)
	})

	// Get an item by ID
	e.GET("/items/:id", func(c echo.Context) error {
		id := c.Param("id")

		// Find the record by ID using PocketBase's Dao
		record, err := app.Dao().FindRecordById("items", id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Item not found"})
		}

		return c.JSON(http.StatusOK, record)
	})

	// Get all items
	e.GET("/items", func(c echo.Context) error {
		// Find all records in the 'items' collection
		records, err := app.Dao().FindRecordsByExpr("items", nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusOK, records)
	})

	// Update an item by ID
	e.PUT("/items/:id", func(c echo.Context) error {
		id := c.Param("id")
		item := new(Item)
		if err := c.Bind(item); err != nil {
			return err
		}

		// Find the existing record by ID
		record, err := app.Dao().FindRecordById("items", id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Item not found"})
		}

		// Update fields
		record.Set("title", item.Title)
		record.Set("description", item.Description)
		record.Set("price", item.Price)

		// Save the updated record
		if err := app.Dao().SaveRecord(record); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusOK, record)
	})

	// Delete an item by ID
	e.DELETE("/items/:id", func(c echo.Context) error {
		id := c.Param("id")

		// Find the record by ID
		record, err := app.Dao().FindRecordById("items", id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Item not found"})
		}

		// Delete the record
		if err := app.Dao().DeleteRecord(record); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	})

	// Start the Echo server
	e.Logger.Fatal(e.Start(":8080"))
}
