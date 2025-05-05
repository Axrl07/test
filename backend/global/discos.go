package global

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CrearDisco(path string) error {
	// verificando la ruta
	ruta := filepath.Dir(path)
	if err := os.MkdirAll(ruta, os.ModePerm); err != nil {
		nombre := strings.Split(path, "/")
		msj := "ERROR: la ruta no existe " + nombre[len(nombre)-1]
		fmt.Println(msj)
		return fmt.Errorf(msj, err)
	}

	// creando el disco
	if _, err := os.Stat(path); os.IsNotExist(err) {
		archivo, err := os.Create(path)
		if err != nil {
			nombre := strings.Split(path, "/")
			msj := "ERROR: no fue posible crear el disco " + nombre[len(nombre)-1]
			fmt.Println(msj)
			return fmt.Errorf(msj, err)
		}
		defer archivo.Close()
	}
	return nil
}

func AbrirDisco(path string) (*os.File, error) {
	// abriendo el disco
	archivo, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		nombre := strings.Split(path, "/")
		msj := "ERROR: no fue posible abrir el disco " + nombre[len(nombre)-1]
		fmt.Println(msj)
		return nil, fmt.Errorf(msj)
	}
	return archivo, nil
}

// escribimos desde el inicio del disco hasta posicion
func EscribirEnDisco(archivo *os.File, datos interface{}, posicion int64) error {
	// la posición es la distancia desde el 0 donde se posicionará el puntero para comenzar a escribir
	archivo.Seek(posicion, 0)
	err := binary.Write(archivo, binary.LittleEndian, datos)
	if err != nil {
		msj := "ERROR: no fue posible escribir en el disco"
		fmt.Println(msj)
		return fmt.Errorf(msj, err)
	}
	return nil
}

// leemos desde el inicio del disco hasta posicion
func LeerEnDisco(archivo *os.File, datos interface{}, posicion int64) error {
	// la posición es la distancia desde el 0 donde se posicionará el puntero para comenzar a leer
	archivo.Seek(posicion, 0)
	err := binary.Read(archivo, binary.LittleEndian, datos)
	if err != nil {
		msj := "ERROR: no fue posible leer el disco"
		fmt.Println(msj)
		return fmt.Errorf(msj, err)
	}
	return nil
}

// borra desde la posicion 1 hasta la posicion 2 en el disco proporcionado
func BorrarEnDisco(archivo *os.File, posicion int32, posicionFinal int32) error {
	// creamos los ceros
	ceros := make([]byte, posicionFinal)
	// posicionamos el puntero
	archivo.Seek(int64(posicion), 0)
	err := binary.Write(archivo, binary.LittleEndian, ceros)
	if err != nil {
		fmt.Println("ERROR: no ha sido posible aplicar el parametro DELETE con el valor full", err)
		return err
	}
	return nil
}
