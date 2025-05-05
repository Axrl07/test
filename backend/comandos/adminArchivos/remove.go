package adminArchivos

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"os"
	"strings"
)

func Remove(parametros []string) string {

	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "REMOVE ERROR: para eliminar un archivo o carpeta necesita iniciar sesión"
		fmt.Println(msj)
		return msj
	}

	// ya sé que solamente viene un parametro
	parametro := strings.Split(parametros[0], "=")
	nombreParam := strings.ToLower(parametro[0])             // nombre del parametro PATH
	valorParam := strings.ReplaceAll(parametro[1], "\"", "") // valor del PATH

	// verificando que sea el parametro name
	if nombreParam != "path" {
		msj := "RMGRP ERROR: el único parametro permitido dentro del comando es NAME\n"
		fmt.Println(msj)
		return msj
	}

	// verificando que haya valor
	if valorParam == "" {
		msj := "RMGRP ERROR: el parametro requiere tener un valor asociado\n"
		fmt.Println(msj)
		return msj
	}

	return comandoRemove(valorParam)
}

func comandoRemove(path string) string {
	usuario, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "REMOVE ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// para verificar los permisos
	uidCreador, gidGrupo, err := superBloque.ObtenerGID_UID(pathDisco, usuario)
	if err != nil {
		msj := "REMOVE ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// verificamos que exista la carpeta o archivo
	var inodo estructuras.Inode
	idInodo := gestionSistema.BuscarInodo(0, path, *superBloque, pathDisco)
	if idInodo == -1 {
		msj := "REMOVE ERROR: no fue posible encontrar " + path
		fmt.Println(msj)
		return msj
	}

	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		msj := "REMOVE ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// verificamos si existen permisos de escritura
	if inodo.I_uid != uidCreador && uidCreador != 1 && gidGrupo != 1 && inodo.I_gid != gidGrupo {
		msj := "REMOVE ERROR: no tiene permisos de eliminación en el archivo o carpeta " + path
		fmt.Println(msj)
		return msj
	}

	// ahora vamos a eliminar el archivo o carpeta pero con el bloque que apunta al inodo
	// para ello obtenemos la tmp menos el nombre del archivo o carpeta
	tmp := strings.Split(path, "/")
	nombre := tmp[len(tmp)-1]
	// cortamos
	tmp = tmp[:len(tmp)-1]
	ruta := strings.Join(tmp, "/")

	// ahora obtenemos el bloque que apunta al inodo
	idInodoPadre := gestionSistema.BuscarInodo(0, ruta, *superBloque, pathDisco)
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodoPadre*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		msj := "REMOVE ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	if idInodoPadre != -1 {
		folderBLock := estructuras.FolderBlock{}
		for indice, idBLoque := range inodo.I_block {
			if idBLoque != -1 {
				if err := folderBLock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBLoque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
					msj := "REMOVE ERROR: " + err.Error()
					fmt.Println(msj)
					return msj
				}

				// buscamos si el nombre coincide con el nombre del inodo
				var salimos bool
				for i := 2; i < 4; i++ {
					if folderBLock.B_content[i].B_inodo != -1 {
						// si coinciden los nombres entonces renombramos
						nombreGuardado := global.BorrandoIlegibles(string(folderBLock.B_content[i].B_name[:]))
						if nombreGuardado == nombre {
							if global.BorrandoIlegibles(string(folderBLock.B_content[2].B_name[:])) == nombre && folderBLock.B_content[3].B_inodo != -1 {
								folderBLock.B_content[2].B_name = [12]byte{}
								folderBLock.B_content[2].B_inodo = -1
								folderBLock.B_content[2].B_name = folderBLock.B_content[3].B_name
								folderBLock.B_content[2].B_inodo = folderBLock.B_content[3].B_inodo
								folderBLock.B_content[3].B_name = [12]byte{}
								folderBLock.B_content[3].B_inodo = -1
								if err := folderBLock.Serialize(pathDisco, int64(superBloque.S_block_start+(idBLoque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
									msj := "REMOVE ERROR: " + err.Error()
									fmt.Println(msj)
									return msj
								}
							}
							if global.BorrandoIlegibles(string(folderBLock.B_content[3].B_name[:])) == nombre && folderBLock.B_content[2].B_inodo != -1 {
								folderBLock.B_content[i].B_name = [12]byte{}
								folderBLock.B_content[i].B_inodo = -1
								if err := folderBLock.Serialize(pathDisco, int64(superBloque.S_block_start+(idBLoque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
									msj := "REMOVE ERROR: " + err.Error()
									fmt.Println(msj)
									return msj
								}
							}
							if global.BorrandoIlegibles(string(folderBLock.B_content[2].B_name[:])) == nombre && folderBLock.B_content[3].B_inodo == -1 {
								if err := folderBLock.Clear(pathDisco, int64(superBloque.S_block_start+(idBLoque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
									msj := "REMOVE ERROR: " + err.Error()
									fmt.Println(msj)
									return msj
								}
								inodo.I_block[indice] = -1
							}

							if global.BorrandoIlegibles(string(folderBLock.B_content[3].B_name[:])) == nombre && folderBLock.B_content[2].B_inodo == -1 {
								if err := folderBLock.Clear(pathDisco, int64(superBloque.S_block_start+(idBLoque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
									msj := "REMOVE ERROR: " + err.Error()
									fmt.Println(msj)
									return msj
								}
								inodo.I_block[indice] = -1
							}
							salimos = true
							break
						}
					}
				}

				if salimos {
					// serializamos el inodo padre
					if err := inodo.Serialize(pathDisco, int64(superBloque.S_inode_start+(idInodoPadre*int32(binary.Size(estructuras.Inode{}))))); err != nil {
						msj := "REMOVE " + err.Error()
						fmt.Println(msj)
						return msj
					}
					break
				}
			}
		}
	} else {
		msj := "REMOVE ERROR: no ha sido posible eliminar " + path
		fmt.Println(msj)
		return msj
	}

	//eliminamos todo lo que está desde el inodo
	if err := EliminarInodoRecursivo(idInodo, superBloque, pathDisco); err != nil {
		msj := "REMOVE ERROR: no ha sido posible borrar completamente " + path
		fmt.Println(msj)
		return msj
	}

	// si es Ext3
	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("remove", path, "", fmt.Sprintf("remove -path=%s\n", path)); err != nil {
			msj := "REMOVE " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	return "Se ha Eliminado correctamente " + path + "\n"
}

// EliminarInodoRecursivo elimina un inodo y todos sus bloques asociados (tanto FolderBlock como FileBlock)
// y recursivamente elimina todos los inodos hijos si el inodo es una carpeta
func EliminarInodoRecursivo(idInodo int32, superBloque *estructuras.SuperBlock, pathDisco string) error {
	// Cargar el inodo que vamos a eliminar
	var inodo estructuras.Inode
	err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{})))))
	if err != nil {
		return fmt.Errorf("REMOVE ERROR: al cargar el inodo %d: %v", idInodo, err)
	}

	// Si es una carpeta, primero eliminar todos sus hijos
	if inodo.I_type[0] == '0' {
		// Es una carpeta, necesitamos eliminar su contenido recursivamente
		for i := 0; i < 12; i++ {
			idBloque := inodo.I_block[i]
			if idBloque != -1 {
				var folderBlock estructuras.FolderBlock
				err := folderBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))
				if err != nil {
					return fmt.Errorf("REMOVE ERROR: al cargar el bloque de carpeta %d: %v", idBloque, err)
				}

				// Recorrer el contenido del bloque de carpeta
				for j := 2; j < 4; j++ { // Empezamos desde 2 para saltar . y ..
					apuntador := folderBlock.B_content[j].B_inodo
					if apuntador != -1 {
						// Eliminar recursivamente este hijo
						err := EliminarInodoRecursivo(apuntador, superBloque, pathDisco)
						if err != nil {
							return err
						}
					}
				}

				// Guardar los cambios en el bloque de carpeta
				err = folderBlock.Clear(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))
				if err != nil {
					return fmt.Errorf("REMOVE ERROR: al guardar el bloque de carpeta %d: %v", idBloque, err)
				}

				// Liberar el bloque en el bitmap
				err = liberarBitmapBloque(superBloque, pathDisco, idBloque)
				if err != nil {
					return err
				}
			}
		}
	} else if inodo.I_type[0] == '1' {
		// Es un archivo, solo necesitamos liberar sus bloques
		for i := 0; i < 12; i++ {
			idBloque := inodo.I_block[i]
			if idBloque != -1 {
				// Liberar el bloque en el bitmap
				fileBlock := estructuras.FileBlock{}
				err := fileBlock.Clear(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FileBlock{})))))
				if err != nil {
					return err
				}
				err = liberarBitmapBloque(superBloque, pathDisco, idBloque)
				if err != nil {
					return err
				}
			}
		}
	}

	err = inodo.Clear(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{})))))
	if err != nil {
		return err
	}

	// Finalmente, liberar el inodo en el bitmap
	err = liberarBitmapInodo(superBloque, pathDisco, idInodo)
	if err != nil {
		return err
	}

	return nil
}

// liberarBitmapInodo libera un inodo en el bitmap de inodos
func liberarBitmapInodo(superBloque *estructuras.SuperBlock, pathDisco string, idInodo int32) error {
	file, err := os.OpenFile(pathDisco, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("REMOVE ERROR: al abrir el disco para liberar bitmap de inodo: %v", err)
	}
	defer file.Close()

	// Calcular la posición en el bitmap
	bitmapPos := int64(superBloque.S_bm_inode_start) + int64(idInodo)

	// Moverse a la posición en el bitmap
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return fmt.Errorf("REMOVE ERROR: al posicionarse en el bitmap de inodos: %v", err)
	}

	// Marcar como libre ('0')
	_, err = file.Write([]byte{'0'})
	if err != nil {
		return fmt.Errorf("REMOVE ERROR: al marcar el inodo como libre: %v", err)
	}

	// Actualizar contadores en el superbloque
	superBloque.S_free_inodes_count++

	return nil
}

// liberarBitmapBloque libera un bloque en el bitmap de bloques
func liberarBitmapBloque(superBloque *estructuras.SuperBlock, pathDisco string, idBloque int32) error {
	file, err := os.OpenFile(pathDisco, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("REMOVE ERROR: al abrir el disco para liberar bitmap de bloque: %v", err)
	}
	defer file.Close()

	// Calcular la posición en el bitmap
	bitmapPos := int64(superBloque.S_bm_block_start) + int64(idBloque)

	// Moverse a la posición en el bitmap
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return fmt.Errorf("REMOVE ERROR: al posicionarse en el bitmap de bloques: %v", err)
	}

	// Marcar como libre ('O')
	_, err = file.Write([]byte{'O'})
	if err != nil {
		return fmt.Errorf("REMOVE ERROR: al marcar el bloque como libre: %v", err)
	}

	// Actualizar contadores en el superbloque
	superBloque.S_free_blocks_count++

	return nil
}
