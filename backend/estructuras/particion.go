package estructuras

import (
	"errors"
	"fmt"
	"main/global"
)

// total de: 3*4 + 3*1 + 16 + 4 = 35 bytes
type Partition struct {
	Status [1]byte  // Estado de la partición: 0 = creada , 1 = montada
	Type   [1]byte  // Tipo de partición
	Fit    [1]byte  // Ajuste de la partición
	Start  int32    // Byte de inicio de la partición
	Size   int32    // Tamaño de la partición
	Name   [16]byte // Nombre de la partición
	// al montar
	Correlative int32   // Correlativo de la partición
	Id          [4]byte // ID de la partición
}

// Crear una partición con los parámetros proporcionados
func (particion *Partition) CrearParticion(comienzo, tamano int32, tipo, ajuste, nombre string) {

	particion.Status[0] = 0 // 0 -> creada

	if len(tipo) > 0 {
		particion.Type[0] = tipo[0]
	}

	if len(ajuste) > 0 {
		particion.Fit[0] = ajuste[0]
	}

	particion.Start = int32(comienzo)
	particion.Size = int32(tamano)
	copy(particion.Name[:], nombre)
}

// monta la partición
func (particion *Partition) MontarParticion(correlativo int32, id string) error {
	// cambiamos el estado de 0 a 1 porque ya se montó
	particion.Status = [1]byte{1}
	// Asignar correlativo a la partición
	particion.Correlative = correlativo
	// Asignar ID a la partición
	copy(particion.Id[:], id)

	return nil
}

// Imprimir los valores de la partición
func (p *Partition) Print() {
	fmt.Printf("Part_status: %c\n", p.Status[0])
	fmt.Printf("Type: %c\n", p.Type[0])
	fmt.Printf("Part_fit: %c\n", p.Fit[0])
	fmt.Printf("Start: %d\n", p.Start)
	fmt.Printf("Size: %d\n", p.Size)
	fmt.Printf("Name: %s\n", string(p.Name[:]))
	fmt.Printf("Part_correlative: %d\n", p.Correlative)
	fmt.Printf("Id: %s\n", string(p.Id[:]))
}

func (salida *Partition) ToString() string {
	contenido := fmt.Sprintf("Partición %s: ", global.BorrandoIlegibles(string(salida.Name[:])))
	contenido += fmt.Sprintf("Part_status: %d, ", int32(salida.Status[0]))
	contenido += fmt.Sprintf("Type: %s, ", string(salida.Type[0]))
	contenido += fmt.Sprintf("Part_fit: %s, ", string(salida.Fit[0]))
	contenido += fmt.Sprintf("Start: %d, ", salida.Start)
	contenido += fmt.Sprintf("Size: %d, ", salida.Size)
	contenido += fmt.Sprintf("Part_correlative: %d, ", salida.Correlative)
	contenido += fmt.Sprintf("Id: %s\n", global.BorrandoIlegibles(string(salida.Id[:])))
	return contenido
}

// me devuelve la partición el super bloque, el path del disco y un error
func ObtenerSuperBloque(id string) (*SuperBlock, *Partition, string, error) {

	// verificamos que exista el Id en las particiones montada (son las únicas que tienen ID)
	pathDisco, ok := global.ParticionesMontadas[id]
	if !ok {
		msj := "ERROR: el Id ingresado no existe"
		return nil, nil, "", errors.New(msj)
	}

	// abrimos el disco
	disco, err := global.AbrirDisco(pathDisco)
	if err != nil {
		return nil, nil, "", err
	}

	// obtenemos el MBR que está en el disco
	var mbr MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		return nil, nil, "", err
	}

	// obtenemos la partición
	particion, err := mbr.ParticionPorId(id)
	if err != nil {
		return nil, nil, "", err
	}

	// Crear una instancia de SuperBlock
	var sb SuperBlock
	if err := global.LeerEnDisco(disco, &sb, int64(particion.Start)); err != nil {
		return nil, nil, "", err
	}

	defer disco.Close()
	// retornamos
	return &sb, particion, pathDisco, nil
}
