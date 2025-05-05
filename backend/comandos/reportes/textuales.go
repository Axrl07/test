package reportes

import (
	"encoding/binary"
	"errors"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strings"
)

// REPORTE BITMAP INODOS
func reporteBm_Inode(idPart string) (string, error) {
	// obtenemos el superbloque
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idPart)
	if err != nil {
		return "", err
	}

	// abrimos el disco
	disco, err := global.AbrirDisco(pathDisco)
	if err != nil {
		return "", err
	}
	defer disco.Close()

	// Calcular el número total de inodos
	totalInodes := superBloque.S_inodes_count + superBloque.S_free_inodes_count

	// Obtener el contenido del bitmap de inodos
	var bitmapContent strings.Builder

	for i := int32(0); i < totalInodes; i++ {
		// Establecer el puntero
		_, err := disco.Seek(int64(superBloque.S_bm_inode_start+i), 0)
		if err != nil {
			return "", fmt.Errorf("error al establecer el puntero en el archivo: %v", err)
		}

		// Leer un byte (carácter '0' o '1')
		char := make([]byte, 1)
		_, err = disco.Read(char)
		if err != nil {
			return "", fmt.Errorf("error al leer el byte del archivo: %v", err)
		}

		// Agregar el carácter al contenido del bitmap
		bitmapContent.WriteByte(char[0])

		// Agregar un carácter de nueva línea cada 20 caracteres (20 inodos)
		if (i+1)%20 == 0 {
			bitmapContent.WriteString("\n")
		}
	}
	return bitmapContent.String(), nil
}

// REPORTE BITMAP BLOQUES
func reporteBm_Block(idPart string) (string, error) {
	// obtenemos el superbloque
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idPart)
	if err != nil {
		return "", err
	}

	// abrimos el disco
	disco, err := global.AbrirDisco(pathDisco)
	if err != nil {
		return "", err
	}
	defer disco.Close()

	// Calcular el número total de bloques
	totalBloques := superBloque.S_blocks_count + superBloque.S_free_blocks_count

	// Obtener el contenido del bitmap de bloques
	var bitmapContent strings.Builder

	for i := int32(0); i < totalBloques; i++ {
		// Establecer el puntero
		_, err := disco.Seek(int64(superBloque.S_bm_block_start+i), 0)
		if err != nil {
			return "", fmt.Errorf("error al establecer el puntero en el archivo: %v", err)
		}

		// Leer un byte (carácter '0' o '1')
		char := make([]byte, 1)
		_, err = disco.Read(char)
		if err != nil {
			return "", fmt.Errorf("error al leer el byte del archivo: %v", err)
		}

		// Agregar el carácter al contenido del bitmap
		bitmapContent.WriteByte(char[0])

		// Agregar un carácter de nueva línea cada 20 caracteres (20 inodos)
		if (i+1)%20 == 0 {
			bitmapContent.WriteString("\n")
		}
	}
	return bitmapContent.String(), nil
}

// REPORTE FILE
func reporteFile(idPart string, pathFile string) (string, error) {

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idPart)
	if err != nil {
		msj := "REP ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	var contenido string                // contenido completo del  archivo
	var fileBlock estructuras.FileBlock // solamente se refresca cada vez que deserializamos un fileblock
	var inodo estructuras.Inode         // para deserializar el unodo

	idInodo := gestionSistema.BuscarInodo(0, pathFile, *superBloque, pathDisco)
	if idInodo > 0 {
		//leemos el inodo
		if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{}))))); err != nil {
			msj := "REP " + err.Error()
			fmt.Println(msj)
			return "", errors.New(msj)
		}

		//recorrer los fileblocks del inodo para obtener toda su informacion
		for _, idBlock := range inodo.I_block {
			if idBlock != -1 {
				// leemos el fileblock
				if err := fileBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBlock*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
					msj := "REP " + err.Error()
					fmt.Println(msj)
					return "", errors.New(msj)
				}
				// Agregar el contenido del bloque al contenido total
				contenido += global.BorrandoIlegibles(string(fileBlock.B_content[:])) + "\n"
			}
		}
	} else {
		contenido += "REP ERROR: No se encontro el archivo " + pathFile + "\n"
	}
	return contenido, nil
}
