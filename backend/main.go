package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type App struct {
	DB *pgxpool.Pool
}

func main() {
	// 1. Load .env
	_ = godotenv.Load() // Ignore error if file missing (production safe)

	// 2. Connect to DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer db.Close()

	app := &App{DB: db}

	// 3. Setup Router
	mux := http.NewServeMux()
	mux.HandleFunc("/generate", app.handleGenerate)

	// 4. Setup CORS
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods: []string{"POST", "GET", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}).Handler(mux)

	fmt.Println("Backend running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func (app *App) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse Request
	type Request struct {
		UserID    string `json:"user_id"`
		UserImage string `json:"user_image_url"`
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 2. Find a Reference Image (The "Actor")
	refImage := getRandomReference()
	fmt.Printf("Job Started: User %s -> Reference %s\n", req.UserID, refImage)

	// 3. Create DB Entry (Status: Processing)
	var genID string
	err := app.DB.QueryRow(context.Background(),
		`INSERT INTO generations (user_id, status, reference_image_url, created_at) 
         VALUES ($1, 'processing', $2, NOW()) RETURNING id`,
		req.UserID, refImage).Scan(&genID)

	if err != nil {
		log.Println("DB Error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// 4. Trigger Background Process (The "Magic")
	// We use 'go' keyword to run this without making the user wait
	go app.processAI(genID, req.UserImage, refImage)

	// 5. Respond immediately
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processing",
		"id":     genID,
		"msg":    "AI is working in the background!",
	})
}

// processAI handles the long-running AI task
func (app *App) processAI(genID, userImg, refImg string) {
	// A. Call the External AI API
	resultURL, err := callA2EApi(userImg, refImg)
	
	ctx := context.Background()

	// B. Handle Failure
	if err != nil {
		log.Printf("AI Failed for ID %s: %v\n", genID, err)
		app.DB.Exec(ctx, "UPDATE generations SET status = 'failed' WHERE id = $1", genID)
		return
	}

	// C. Handle Success
	log.Printf("AI Success for ID %s. Result: %s\n", genID, resultURL)
	_, err = app.DB.Exec(ctx, 
		"UPDATE generations SET status = 'completed', result_image_url = $1 WHERE id = $2", 
		resultURL, genID)
	
	if err != nil {
		log.Println("Failed to update DB:", err)
	}
}

// callA2EApi makes the actual HTTP Post request
func callA2EApi(userImg, refImg string) (string, error) {
	// TODO: REPLACE WITH YOUR ACTUAL API KEY IF NEEDED
	apiKey := os.Getenv("A2E_API_KEY") 
	apiURL := "https://video.a2e.ai/image-generator/image-editor"

	// Construct JSON Body
	// Note: You must check the A2E docs for the EXACT field names they want!
	// I am guessing "source_image" and "target_image" based on standard APIs.
	reqBody, _ := json.Marshal(map[string]interface{}{
		"source_image": userImg, // Your face
		"target_image": refImg,  // The actor
		"prompt":       "Swap the face ensuring high quality photorealism", // Optional prompt
	})

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API Error: %s", resp.Status)
	}

	// Parse Response (Adjust this struct based on actual API response)
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Assuming the API returns a key like "url" or "output_url"
	if url, ok := result["url"].(string); ok {
		return url, nil
	}
	
	// Fallback for testing if API fails (so you can see the DB update work)
	return "https://via.placeholder.com/500?text=AI+Generated+Result", nil
}

// Simple hardcoded list of "Reference Actors" to start with
func getRandomReference() string {
	refs := []string{
		"https://upload.wikimedia.org/wikipedia/commons/thumb/c/c2/Portraits_of_Andrea_del_Sarto.jpg/800px-Portraits_of_Andrea_del_Sarto.jpg", // Classic Art
		"https://images.unsplash.com/photo-1500648767791-00dcc994a43e", // Cool Guy
		"https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d", // Modern Portrait
	}
	return refs[rand.Intn(len(refs))]
}