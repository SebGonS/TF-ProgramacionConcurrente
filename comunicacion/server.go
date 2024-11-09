package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PORT         = ":9000"             // Puerto en el que escuchará el servidor
	DATASET_PATH = "data.csv"          // Ruta al archivo del dataset limpio
	NUM_CLIENTES = 3                   // Número total de clientes esperados
	LOG_FILE     = "server_errors.log" // Archivo de log para errores
	TIMEOUT      = 10 * time.Second    // Tiempo límite para detectar desconexiones
)

// Estructura para almacenar datos parciales de cada cliente
type PartialResult struct {
	NodeID    int
	Resultado map[string]float64
}

// Mapa para almacenar los resultados recibidos de los nodos clientes
var resultados = make(map[int]PartialResult)
var mu sync.Mutex
var nodeCount = 0
var resultadosRecibidos = 0

func main() {
	// Configurar el log para escribir en un archivo
	file, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("No se pudo abrir el archivo de log:", err)
	}
	defer file.Close()
	log.SetOutput(file)

	// Configuración del servidor TCP
	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Servidor escuchando en el puerto", PORT)

	// Cargar y distribuir el dataset
	dataset, err := cargarDataset(DATASET_PATH)
	if err != nil {
		fmt.Println("Error al cargar el dataset:", err)
		return
	}

	// Manejar conexiones de clientes
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error al aceptar conexión:", err)
			continue
		}

		// Incrementa el contador de nodos y asigna un ID único al nodo
		nodeCount++
		nodeID := nodeCount

		fmt.Printf("Cliente conectado: %s, NodeID: %d\n", conn.RemoteAddr(), nodeID)

		// Asignar una goroutine para manejar la conexión del cliente
		go manejarCliente(conn, dataset, nodeID)
	}
}

// Cargar el dataset desde un archivo CSV
func cargarDataset(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Omitir la primera fila (encabezado)
	if len(data) > 0 {
		data = data[1:]
	}

	return data, nil
}

// Manejar la conexión con el cliente
func manejarCliente(conn net.Conn, dataset [][]string, nodeID int) {
	defer conn.Close()

	// Determinar la porción de datos que enviará al cliente
	dataSlice := dividirDataset(dataset, NUM_CLIENTES, nodeID-1)

	// Establecer el tiempo límite inicial para la conexión
	conn.SetDeadline(time.Now().Add(TIMEOUT))

	// Enviar datos parciales al cliente
	err := enviarDatos(conn, dataSlice)
	if err != nil {
		log.Printf("Error al enviar datos al cliente %d: %v\n", nodeID, err)
		return
	}

	// Recibir y almacenar los resultados del cliente
	err = recibirResultados(conn, nodeID)
	if err != nil {
		log.Printf("Error al recibir datos del cliente %d: %v\n", nodeID, err)
		return
	}

	// Comprobar si ya se han recibido todos los resultados
	mu.Lock()
	resultadosRecibidos++
	if resultadosRecibidos == NUM_CLIENTES {
		fmt.Println("Todos los resultados han sido recibidos.")
		fmt.Println("Resultados combinados:", combinarResultados())
	}
	mu.Unlock()
}

// Dividir el dataset en porciones para cada cliente
func dividirDataset(dataset [][]string, numClientes int, index int) [][]string {
	chunkSize := (len(dataset) + numClientes - 1) / numClientes
	start := index * chunkSize
	end := start + chunkSize

	if end > len(dataset) {
		end = len(dataset)
	}
	return dataset[start:end]
}

// Enviar porciones del dataset al cliente
func enviarDatos(conn net.Conn, dataSlice [][]string) error {
	writer := bufio.NewWriter(conn)

	for _, row := range dataSlice {
		// Actualizar el tiempo límite antes de cada operación
		conn.SetDeadline(time.Now().Add(TIMEOUT))

		line := strings.Join(row, ",") + "\n"
		_, err := writer.WriteString(line)
		if err != nil {
			return fmt.Errorf("fallo en el envío de datos: %v", err)
		}
		writer.Flush()
	}
	fmt.Fprintln(conn, "END") // Señal de fin de transmisión de datos
	return nil
}

// Recibir resultados parciales del cliente
func recibirResultados(conn net.Conn, nodeID int) error {
	reader := bufio.NewReader(conn)
	var nodeResults = make(map[string]float64)

	for {
		// Actualizar el tiempo límite antes de cada operación
		conn.SetDeadline(time.Now().Add(TIMEOUT))

		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error al leer datos del cliente %d: %v", nodeID, err)
			return err
		}

		message = strings.TrimSpace(message)
		if message == "END" {
			fmt.Printf("Resultados recibidos completamente de NodeID: %d\n", nodeID)
			break
		}

		// Procesar y almacenar el resultado parcial
		parts := strings.Split(message, ",")
		if len(parts) == 2 {
			productID := parts[0]
			similarityScore, err := strconv.ParseFloat(parts[1], 64)
			if err == nil {
				nodeResults[productID] = similarityScore
			} else {
				log.Printf("Error al convertir 'similarityScore' para el cliente %d: %v", nodeID, err)
			}
		}
	}

	mu.Lock()
	resultados[nodeID] = PartialResult{NodeID: nodeID, Resultado: nodeResults}
	mu.Unlock()
	return nil
}

// Estructura auxiliar para ordenar productos por puntaje
type Producto struct {
	ProductID string
	Puntaje   float64
}

// Función para combinar y seleccionar los 5 productos más recomendables por cliente
func combinarResultados() map[int][]Producto {
	combinedResults := make(map[int][]Producto)

	// Procesar resultados por cada cliente (NodeID)
	for nodeID, partial := range resultados {
		productos := []Producto{}

		// Convertir los resultados del mapa a un slice de Producto
		for productID, puntaje := range partial.Resultado {
			productos = append(productos, Producto{ProductID: productID, Puntaje: puntaje})
		}

		// Ordenar los productos por puntaje en orden descendente
		sort.Slice(productos, func(i, j int) bool {
			return productos[i].Puntaje > productos[j].Puntaje
		})

		// Seleccionar los 5 productos más recomendables
		if len(productos) > 5 {
			productos = productos[:5]
		}

		// Almacenar los top 5 productos en los resultados combinados
		combinedResults[nodeID] = productos
	}

	return combinedResults
}
