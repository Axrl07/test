package global

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// ConvertToBytes convierte un tamaño y una unidad a bytes
func ConvertirUnidades(size int32, unit [1]byte) int32 {
	switch string(unit[:]) {
	case "K":
		return size * 1024
	case "M":
		return size * 1024 * 1024
	default:
		// cuando sea B de bytes
		return size
	}
}

func RevertirConversionUnidades(size int32, unit string) (int32, string) {
	switch strings.ToUpper(unit) {
	case "K":
		return size / 1024, "K"
	case "M":
		return size / (1024 * 1024), "M"
	default:
		// cuando sea B de bytes
		return size, "B"
	}
}

// elimina caracteres no legibles de una cadena
func BorrandoIlegibles(entrada string) string {
	transformFunc := func(r rune) rune {
		// Solo permitir ASCII imprimible (32-126) y saltos de línea comunes
		if unicode.IsPrint(r) {
			return r
		}
		return -1 // Se eliminará este carácter
	}

	// Aplicar la función de transformación a la cadena de entrada
	salida := strings.Map(transformFunc, entrada)
	return salida
}

// Metodo que anula bytes nulos para B_name
func ObtenerNombreB(nombre string) string {
	posicionNulo := strings.IndexByte(nombre, 0)

	if posicionNulo != -1 {
		if posicionNulo != 0 {
			//tiene bytes nulos
			nombre = nombre[:posicionNulo]
		} else {
			//el  nombre esta vacio
			nombre = "-"
		}

	}
	return nombre //-1 el nombre no tiene bytes nulos
}

// obtiene la fecha en el formato 15/04/2025 13:59
func ObtenerFecha(entrada time.Time) string {
	// obtenemos la fecha
	fechaYhora := strings.Split(entrada.Format(time.RFC3339), "T")
	fechaSLice := strings.Split(fechaYhora[0], "-")
	fecha := fechaSLice[2] + "/" + fechaSLice[1] + "/" + fechaSLice[0]

	// obtenemos la hora
	fechaYhora = strings.Split(fechaYhora[1], ":")
	hora := fechaYhora[0] + ":" + fechaYhora[1]
	return fecha + " " + hora
}

// compara un string con cada elemento de un slice string y retorna true si hay coincidencia
func Contiene(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// ObtenerPrimero devuelve el primer elemento de un slice
func ObtenerPrimero[T any](slice []T) (T, error) {
	if len(slice) == 0 {
		var zero T
		return zero, errors.New("el slice está vacío")
	}
	return slice[0], nil
}

// EliminarElemento elimina un elemento de un slice en el índice dado
func EliminarElemento[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice // Índice fuera de rango, devolver el slice original
	}
	return append(slice[:index], slice[index+1:]...)
}

// splitStringIntoChunks divide una cadena en partes de tamaño chunkSize y las almacena en una lista
func DividirPorChunkSize(s string) []string {
	var chunks []string
	for i := 0; i < len(s); i += 64 {
		end := i + 64
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func GetParentDirectories(path string) ([]string, string) {
	// Normalizar el path
	path = filepath.Clean(path)

	// Dividir el path en sus componentes
	components := strings.Split(path, string(filepath.Separator))

	// Lista para almacenar las rutas de las carpetas padres
	var parentDirs []string

	// Construir las rutas de las carpetas padres, excluyendo la última carpeta
	for i := 1; i < len(components)-1; i++ {
		parentDirs = append(parentDirs, components[i])
	}

	// La última carpeta es la carpeta de destino
	destDir := components[len(components)-1]

	return parentDirs, destDir
}

// splitStringIntoChunks divide una cadena en partes de tamaño chunkSize y las almacena en una lista
func SplitStringIntoChunks(s string) []string {
	var chunks []string
	for i := 0; i < len(s); i += 64 {
		end := i + 64
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}
