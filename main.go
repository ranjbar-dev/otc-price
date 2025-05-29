package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Price struct {
	Symbol string
	Price  string
	Time   int64
}

type PriceUpdate struct {
	RealPrice      float64 `json:"real_price"`
	GeneratedPrice float64 `json:"generated_price"`
	DiffPercent    float64 `json:"diff_percent"`
	Timestamp      int64   `json:"timestamp"`
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type WebSocketServer struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (s *WebSocketServer) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.mu.Unlock()
		case message := <-s.broadcast:
			s.mu.Lock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to websocket: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	s.register <- client

	// Start goroutine to read messages from client (if needed)
	go func() {
		defer func() {
			s.unregister <- client
			client.conn.Close()
		}()
		for {
			_, _, err := client.conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	// Start goroutine to write messages to client
	go func() {
		defer func() {
			client.conn.Close()
		}()
		for {
			select {
			case message, ok := <-client.send:
				if !ok {
					client.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			}
		}
	}()
}

func getPriceChannel() chan Price {
	priceChan := make(chan Price, 100)

	wsHandler := func(event *binance.WsKlineEvent) {
		priceChan <- Price{
			Symbol: event.Symbol,
			Price:  event.Kline.Close,
			Time:   event.Kline.StartTime,
		}
	}

	errHandler := func(err error) {
		fmt.Printf("WebSocket error: %v\n", err)
	}

	_, _, err := binance.WsKlineServe("BTCUSDT", "1s", wsHandler, errHandler)
	if err != nil {
		fmt.Printf("Error starting WebSocket: %v\n", err)
	}

	return priceChan
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	}

	// Create WebSocket server
	wsServer := NewWebSocketServer()
	go wsServer.Run()

	// Handle WebSocket connections
	http.HandleFunc("/ws", wsServer.HandleWebSocket)

	// Start HTTP server
	server := &http.Server{
		Addr: ":8080",
	}

	go func() {
		fmt.Println("WebSocket server started on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Handle graceful shutdown
	forever := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		server.Close()
		forever <- struct{}{}
	}()

	priceChan := getPriceChannel()
	otc := NewOtc(0, time.Now().Unix())

	file, err := os.OpenFile("data.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	for {
		select {
		case data := <-priceChan:

			price, err := strconv.ParseFloat(data.Price, 64)
			if err != nil {

				fmt.Printf("Error parsing price: %v\n", err)
				continue
			}

			otc.SetAndGeneratePrice(price)
			diffPercent := (otc.generatedPrice - otc.price) / otc.price * 100

			update := PriceUpdate{
				RealPrice:      otc.price,
				GeneratedPrice: otc.generatedPrice,
				DiffPercent:    diffPercent,
				Timestamp:      time.Now().Unix(),
			}

			// write the update to the file
			file.WriteString(fmt.Sprintf("%.2f,%.2f,%v\n", update.RealPrice, update.GeneratedPrice, update.Timestamp))

			// Convert update to JSON and broadcast to all clients
			if jsonData, err := json.Marshal(update); err == nil {
				wsServer.broadcast <- jsonData
			}

			fmt.Printf("Real: %v, Generated: %v, Diff in percent: %v\n",
				otc.price, otc.generatedPrice, diffPercent)

		case <-forever:
			fmt.Println("Program terminated")
			return
		}
	}
}
