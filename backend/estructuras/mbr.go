package estructuras

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// total de 35*4 + 3*4 + 1 =140 +12 +1 = 153
type MBR struct {
	Size          int32        //tamaño
	CreationDate  float32      //fecha de creacion
	DiskSignature int32        //numero random
	Fit           [1]byte      //ajuste (F, W , B)
	Partitions    [4]Partition //particiones disponibles (unicamente 4)
}

// Método para obtener la primera partición disponible
func (mbr *MBR) PrimeraParticionDisponible() (*Partition, int32, int32) {
	// Calcular el offset para el start de la partición
	offset := binary.Size(mbr) // Tamaño del MBR en bytes

	// Recorrer las particiones del MBR
	for i := 0; i < len(mbr.Partitions); i++ {
		// Si el tipo de partición está inactiva y el start es -1
		if mbr.Partitions[i].Type[0] == byte('0') && mbr.Partitions[i].Start == -1 {
			// Devolver la partición, el offset y el índice
			return &mbr.Partitions[i], int32(offset), int32(i)
		} else {
			// Calcular el nuevo offset para la siguiente partición, es decir, sumar el tamaño de la partición
			offset += int(mbr.Partitions[i].Size)
		}
	}
	return nil, -1, -1
}

// Método para obtener una partición por nombre y me devuelve su indice en el arreglo
func (mbr *MBR) ParticionPorNombre(name string) (*Partition, int) {
	// Recorrer las particiones del MBR
	for i, particion := range mbr.Partitions {
		// Convertir Name a string y eliminar los caracteres nulos
		nombreParticion := strings.Trim(string(particion.Name[:]), "\x00 ")
		// Convertir el nombre de la partición a string y eliminar los caracteres nulos
		inputName := strings.Trim(name, "\x00 ")
		// Si el nombre de la partición coincide, devolver la partición y el índice
		if strings.EqualFold(nombreParticion, inputName) {
			return &particion, i
		}
	}
	return nil, -1
}

// Función para obtener una partición por ID
func (mbr *MBR) ParticionPorId(id string) (*Partition, error) {
	for i := 0; i < len(mbr.Partitions); i++ {
		// Convertir Name a string y eliminar los caracteres nulos
		partitionID := strings.Trim(string(mbr.Partitions[i].Id[:]), "\x00 ")
		// Convertir el id a string y eliminar los caracteres nulos
		inputID := strings.Trim(id, "\x00 ")
		// Si el nombre de la partición coincide, devolver la partición
		if strings.EqualFold(partitionID, inputID) {
			return &mbr.Partitions[i], nil
		}
	}
	return nil, errors.New("ERROR: partición no encontrada")
}

// me dice si existe partición extendida
func (mbr *MBR) ParticionExtendida() (*Partition, int) {
	// Recorrer las particiones del MBR
	for i, particion := range mbr.Partitions {
		if string(particion.Type[:]) == "E" {
			return &particion, i
		}
	}
	return nil, -1
}

// // Método para imprimir los valores del MBR
func (mbr *MBR) Print() {
	// Convertir CreationDate a time.Time
	creationTime := time.Unix(int64(mbr.CreationDate), 0)
	// Convertir Fit a char
	diskFit := rune(mbr.Fit[0])
	fmt.Printf("MBR Size: %d bytes\n", mbr.Size)
	fmt.Printf("Creation Date: %s\n", creationTime.Format(time.RFC3339))
	fmt.Printf("Disk Signature: %d\n", mbr.DiskSignature)
	fmt.Printf("Disk Fit: %c\n", diskFit)
	for _, particion := range mbr.Partitions {
		fmt.Println(particion.ToString())
	}
}

// // // literalmente un tostring del objeto MBR
// func (mbr *MBR) ToString() string {
// 	// Convertir CreationDate a time.Time
// 	creationTime := time.Unix(int64(mbr.CreationDate), 0)
// 	// Convertir Fit a char
// 	diskFit := rune(mbr.Fit[0])
// 	contenido := fmt.Sprintf("MBR Size: %d bytes\n", mbr.Size)
// 	contenido += fmt.Sprintf("Creation Date: %s\n", creationTime.Format(time.RFC3339))
// 	contenido += fmt.Sprintf("Disk Signature: %d\n", mbr.DiskSignature)
// 	contenido += fmt.Sprintf("Disk Fit: %c\n\n", diskFit)
// 	for _, particion := range mbr.Partitions {
// 		contenido += fmt.Sprint(particion.ToString())
// 	}
// 	return contenido
// }

// em devuelve la partición, el indice y su start para ingresar la nueva partición según el ajuste
func (mbr *MBR) AplicandoAjuste(sizeP int32, ajuste string) (*Partition, int32, int32) {
	offset := int32(binary.Size(mbr)) // Byte después del MBR

	type candidato struct {
		particion    Partition
		indice       int
		inicioOffset int32
		diferencia   int32
	}

	var candidatos []candidato

	for i, part := range mbr.Partitions {
		if part.Start == -1 && part.Type[0] == '0' {
			if part.Size == -1 {
				// Espacio libre al final del disco
				espacioLibre := mbr.Size - offset
				if espacioLibre >= sizeP {
					candidatos = append(candidatos, candidato{
						particion:    part,
						indice:       i,
						inicioOffset: offset,
						diferencia:   espacioLibre - sizeP,
					})
				}
				break
			}

			// Espacio no usado con tamaño asignado
			if part.Size >= sizeP {
				candidatos = append(candidatos, candidato{
					particion:    part,
					indice:       i,
					inicioOffset: offset,
					diferencia:   part.Size - sizeP,
				})
			}

			offset += part.Size
		} else if part.Start != -1 {
			offset = part.Start + part.Size
		}
	}

	if len(candidatos) == 0 {
		return nil, -1, -1
	}

	if ajuste == "F" {
		for _, part := range candidatos {
			if part.diferencia >= 0 {
				return &part.particion, int32(part.indice), part.inicioOffset
			}
		}
	}

	// Ordenar si es Best o Worst
	if len(candidatos) > 1 {
		sort.Slice(candidatos, func(i, j int) bool {
			if candidatos[i].diferencia == candidatos[j].diferencia {
				return candidatos[i].inicioOffset < candidatos[j].inicioOffset
			}
			return candidatos[i].diferencia < candidatos[j].diferencia
		})
	}

	if ajuste == "B" {
		if candidatos[0].diferencia >= 0 {
			return &candidatos[0].particion, int32(candidatos[0].indice), candidatos[0].inicioOffset
		}
	}

	if ajuste == "W" {
		for i := len(candidatos) - 1; i >= 0; i-- {
			if candidatos[i].diferencia >= 0 {
				return &candidatos[i].particion, int32(candidatos[i].indice), candidatos[i].inicioOffset
			}
		}
	}

	return nil, -1, -1
}

// SerializeMBR escribe la estructura MBR al inicio de un archivo binario
func (mbr *MBR) Serialize(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Serializar la estructura MBR directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

// DeserializeMBR lee la estructura MBR desde el inicio de un archivo binario
func (mbr *MBR) Deserialize(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Obtener el tamaño de la estructura MBR
	mbrSize := binary.Size(mbr)
	if mbrSize <= 0 {
		return fmt.Errorf(" ERROR: EL tanaño del mbr no es valido: %d", mbrSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura MBR
	buffer := make([]byte, mbrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura MBR
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}
