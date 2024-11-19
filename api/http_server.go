package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Struct para request y response
type RecommendationRequest struct {
	ProductName string `json:"product_name"`
}

type RecommendationResponse struct {
	Recommendations []string `json:"recommendations"`
}

// Connexión TCP con el nodo maestro
func fetchRecommendationsFromBackend(productName string) ([]string, error) {

	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		return nil, fmt.Errorf("error de conexión con el backend: %v", err)
	}
	defer conn.Close()

	// enviar producto al nodo maestro
	fmt.Fprintf(conn, productName+"\n")

	// leer respuesta del backend
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error al recibir respuesta del backend: %v", err)
	}

	// Parse the response (assuming it's a JSON array of recommendations)
	var recommendations []string
	err = json.Unmarshal([]byte(response), &recommendations)
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta del backend: %v", err)
	}

	return recommendations, nil
}

// manejo de los websockets
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Fallo inicialización de WebSockets", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for {
		//Leer json del websocket
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error de lectura de websocket:", err)
			break
		}
		// Parse the request
		var req RecommendationRequest
		err = json.Unmarshal(message, &req)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Formato no valido de solicitud"))
			continue
		}

		// Obtener recomendaciones del backend
		recommendations, err := fetchRecommendationsFromBackend(req.ProductName)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error de backend: %v", err)))
			continue
		}

		// crear respuesta
		response := RecommendationResponse{Recommendations: recommendations}
		responseData, err := json.Marshal(response)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error al generar respuesta"))
			continue
		}

		// enviar respuesta de vuelta al cliente
		conn.WriteMessage(websocket.TextMessage, responseData)
	}

}
