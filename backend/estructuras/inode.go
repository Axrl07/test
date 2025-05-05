package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

// Total de 3*4 + 3*4 +15*4 + 1 + 3 = 24 + 60 + 4 = 88 bytes
type Inode struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime float32
	I_ctime float32
	I_mtime float32
	I_block [15]int32
	I_type  [1]byte
	I_perm  [3]byte
}

// Serialize escribe la estructura Inode en un archivo binario en la posición especificada
func (inode *Inode) Serialize(path string, offset int64) error {
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

	// Serializar la estructura Inode directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, inode)
	if err != nil {
		return err
	}

	return nil
}

// limpia
func (inode *Inode) Clear(path string, offset int64) error {
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

	// Determinar el tamaño exacto de la estructura
	size := binary.Size(inode)

	// Crear un buffer de bytes nulos (0x00) del tamaño de la estructura
	nullBuffer := make([]byte, size)

	// Escribir el buffer de bytes nulos para sobrescribir el bloque
	_, err = file.Write(nullBuffer)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura Inode desde un archivo binario en la posición especificada
func (inode *Inode) Deserialize(path string, offset int64) error {
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

	// Obtener el tamaño de la estructura Inode
	inodeSize := binary.Size(inode)
	if inodeSize <= 0 {
		return fmt.Errorf("ERROR: El tamaño del Inodo es invalido: %d", inodeSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura Inode
	buffer := make([]byte, inodeSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura Inode
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, inode)
	if err != nil {
		return err
	}

	return nil
}

// Print imprime los atributos del inodo
func (inode *Inode) Print() {
	atime := time.Unix(int64(inode.I_atime), 0)
	ctime := time.Unix(int64(inode.I_ctime), 0)
	mtime := time.Unix(int64(inode.I_mtime), 0)

	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_size)
	fmt.Printf("I_atime: %s\n", atime.Format(time.RFC3339))
	fmt.Printf("I_ctime: %s\n", ctime.Format(time.RFC3339))
	fmt.Printf("I_mtime: %s\n", mtime.Format(time.RFC3339))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Printf("I_type: %s\n", string(inode.I_type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.I_perm[:]))
}
