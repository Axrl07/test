package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// total de 4*16 = 64 bytes
type FolderBlock struct {
	B_content [4]FolderContent
}

// total de 12 + 4 = 16 bytes
type FolderContent struct {
	B_name  [12]byte
	B_inodo int32
}

// Serialize escribe la estructura FolderBlock en un archivo binario en la posición especificada
func (fb *FolderBlock) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar la estructura FolderBlock directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, fb)
	if err != nil {
		return err
	}

	return nil
}

// Clear sobrescribe la estructura con bytes nulos para "eliminarla" del disco
func (fb *FolderBlock) Clear(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Determinar el tamaño exacto de la estructura
	size := binary.Size(fb)

	// Crear un buffer de bytes nulos (0x00) del tamaño de la estructura
	nullBuffer := make([]byte, size)

	// Escribir el buffer de bytes nulos para sobrescribir el bloque
	_, err = file.Write(nullBuffer)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura FolderBlock desde un archivo binario en la posición especificada
func (fb *FolderBlock) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el tamaño de la estructura FolderBlock
	fbSize := binary.Size(fb)
	if fbSize <= 0 {
		return fmt.Errorf("invalid FolderBlock size: %d", fbSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura FolderBlock
	buffer := make([]byte, fbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura FolderBlock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, fb)
	if err != nil {
		return err
	}

	return nil
}

// Print imprime los atributos del bloque de carpeta
func (fb *FolderBlock) Print() {
	for i, content := range fb.B_content {
		name := string(content.B_name[:])
		fmt.Printf("Content %d:\n", i+1)
		fmt.Printf("  B_name: %s\n", name)
		fmt.Printf("  B_inodo: %d\n", content.B_inodo)
	}
}
