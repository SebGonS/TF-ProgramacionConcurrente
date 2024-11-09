package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	SERVER_ADDRESS = "localhost:9000" // Dirección y puerto del servidor
)

// Función principal del cliente
func main() {
	// Conectar al servidor
	conn, err := net.Dial("tcp", SERVER_ADDRESS)
	if err != nil {
		fmt.Println("Error al conectar con el servidor:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Conectado al servidor", SERVER_ADDRESS)

	// Recibir y almacenar los datos del servidor
	dataSlice := recibirDatos(conn)

	// Calcular similitud entre ítems o usuarios en los datos recibidos
	resultados := calcularSimilitud(dataSlice)

	//fmt.Println("Enviando resultados:", resultados)

	// Enviar los resultados parciales al servidor
	enviarResultados(conn, resultados)
}

// Recibir datos desde el servidor y almacenarlos en un slice
func recibirDatos(conn net.Conn) [][]string {
	var dataSlice [][]string
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer datos del servidor:", err)
			break
		}
		line = strings.TrimSpace(line)
		if line == "END" { // Señal de fin de transmisión de datos
			fmt.Println("Datos recibidos completamente.")
			break
		}
		row := strings.Split(line, ",")
		dataSlice = append(dataSlice, row)
	}

	return dataSlice
}

// Calcular similitud entre ítems o usuarios
func calcularSimilitud(dataSlice [][]string) map[string]float64 {
	resultados := make(map[string]float64)

	// Itera sobre cada fila de dataSlice
	for _, row := range dataSlice {
		if len(row) < 5 { // Asegurarse de que haya al menos 5 columnas para acceder a `stars`
			fmt.Println("Fila con datos insuficientes:", row)
			continue
		}

		productID := row[2] // Columna del ID del producto (índice 2)
		stars := row[4]     // Columna de las calificaciones `stars` (índice 4)

		//fmt.Println("Valor stars:", stars)

		// Convertimos 'stars' a float64 solo para almacenamiento
		if valor, err := strconv.ParseFloat(stars, 64); err == nil {
			//fmt.Println("Valor valor:", stars)
			resultados[productID] += valor
		} else {
			fmt.Println("Error al convertir 'stars' en float:", err)
		}
	}

	//fmt.Println("Resultados calculados en el cliente:", resultados) // Verifica el contenido de resultados
	return resultados
}

// Enviar resultados al servidor
func enviarResultados(conn net.Conn, resultados map[string]float64) {
	writer := bufio.NewWriter(conn)

	for productID, score := range resultados {
		line := fmt.Sprintf("%s,%.2f\n", productID, score)
		writer.WriteString(line)
		writer.Flush()
	}
	fmt.Fprintln(conn, "END") // Señal de fin de envío de resultados
	fmt.Println("Resultados enviados al servidor.")
}