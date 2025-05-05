package global

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// reportes tipo ls y file
func Reporte(path string, contenido string) error {
	//asegurar la ruta
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("ERROR: al crear el reporte, path: ", err)
		return err
	}

	// Abrir o crear un archivo para escritura
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("ERROR: al crear el archivo:", err)
		return err
	}
	defer file.Close()

	// Escribir en el archivo
	_, err = file.WriteString(contenido)
	if err != nil {
		fmt.Println("ERROR: al escribir en el archivo:", err)
		return err
	}

	return err
}

// reportes en graphviz
func RepGraphiz(path string, contenido string, nombre string) error {
	// verificamos que la ruta exista
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("ERROR: error al crear las carpetas padre: ", err)
		return err
	}

	// Abrir o crear un archivo para escritura
	grafico, err := os.Create(path)
	if err != nil {
		fmt.Println("ERROR: al crear el archivo:", err)
		return err
	}
	defer grafico.Close()

	// Escribir en el archivo
	_, err = grafico.WriteString(contenido)
	if err != nil {
		fmt.Println("ERROR: al escribir en el archivo:", err)
		return err
	}

	// ejecutamos el comando
	imagenRep := dir + "/" + nombre + ".svg"
	cmd := exec.Command("dot", "-Tsvg", path, "-o", imagenRep)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("ERROR: no fue posible generar el reporte %s.svg por: %v", nombre, err)
	}

	return err
}
