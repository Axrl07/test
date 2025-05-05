package gestionSistema

import (
	"encoding/binary"
	"main/estructuras"
	"main/global"
	"strings"
	"time"
)

func BuscarInodo(idInodo int32, path string, superBloque estructuras.SuperBlock, pathdisco string) int32 {
	//Dividir la ruta por cada /
	stepsPath := strings.Split(path, "/")
	//el arreglo vendra [ ,val1, val2] por lo que me corro una posicion
	tmpPath := stepsPath[1:]
	//fmt.Println("Ruta actual ", tmpPath)

	//cargo el inodo a partir del cual voy a buscar
	var Inode0 estructuras.Inode
	Inode0.Deserialize(pathdisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{})))))
	//Recorrer los bloques directos (carpetas/archivos) en la raiz
	var folderBlock estructuras.FolderBlock
	for i := 0; i < 12; i++ {
		idBloque := Inode0.I_block[i]
		if idBloque != -1 {
			folderBlock.Deserialize(pathdisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))
			//Recorrer el bloque actual buscando la carpeta/archivo en la raiz
			for j := 2; j < 4; j++ {
				//apuntador es el apuntador del bloque al inodo (carpeta/archivo), si existe es distinto a -1
				apuntador := folderBlock.B_content[j].B_inodo
				if apuntador != -1 {
					pathActual := global.ObtenerNombreB(string(folderBlock.B_content[j].B_name[:]))
					if tmpPath[0] == pathActual {
						//buscarInodo(apuntador, ruta[1:], path, superBloque, iSuperBloque, file, r)
						if len(tmpPath) > 1 {
							return buscarIrecursivo(apuntador, tmpPath[1:], superBloque.S_inode_start, superBloque.S_block_start, pathdisco)
						} else {
							return apuntador
						}
					}
				}
			}
		}
	}
	return idInodo
}

// Buscar inodo de forma recursiva
func buscarIrecursivo(idInodo int32, path []string, iStart int32, bStart int32, pathDisco string) int32 {
	//cargo el inodo actual
	var inodo estructuras.Inode
	inodo.Deserialize(pathDisco, int64(iStart+(idInodo*int32(binary.Size(estructuras.Inode{})))))

	//Nota: el inodo tiene tipo. No es necesario pero se podria validar que sea carpeta
	//recorro el inodo buscando la siguiente carpeta
	var folderBlock estructuras.FolderBlock
	for i := 0; i < 12; i++ {
		idBloque := inodo.I_block[i]
		if idBloque != -1 {
			folderBlock.Deserialize(pathDisco, int64(bStart+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))
			//Recorrer el bloque buscando la carpeta actua
			for j := 2; j < 4; j++ {
				apuntador := folderBlock.B_content[j].B_inodo
				if apuntador != -1 {
					pathActual := global.ObtenerNombreB(string(folderBlock.B_content[j].B_name[:]))
					if path[0] == pathActual {
						if len(path) > 1 {
							//sin este if path[1:] termina en un arreglo de tamaño 0 y retornaria -1
							return buscarIrecursivo(apuntador, path[1:], iStart, bStart, pathDisco)
						} else {
							//cuando el arreglo path tiene tamaño 1 esta en la carpeta que busca
							return apuntador
						}
					}
				}
			}
		}
	}
	return -1
}

// crea una carpeta en el sistema de archivos (no importa si es ext2 o ext3)
func CreaCarpeta(idInode int32, carpeta string) int32 {

	// el usuario loggeado actualmente tiene el idParticion
	_, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, particion, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		return -1
	}

	var inodo estructuras.Inode
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInode*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		return -1
	}

	//Recorrer los bloques directos del inodo para ver si hay espacio libre
	for i := 0; i < 12; i++ {
		idBloque := inodo.I_block[i]
		if idBloque != -1 {
			//Existe un folderblock con idBloque que se debe revisar si tiene espacio para la nueva carpeta
			var folderBlock estructuras.FolderBlock
			folderBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))

			//Recorrer el bloque para ver si hay espacio
			for j := 2; j < 4; j++ {
				apuntador := folderBlock.B_content[j].B_inodo
				//Hay espacio en el bloque
				if apuntador == -1 {

					//modifico el bloque actual
					copy(folderBlock.B_content[j].B_name[:], carpeta)
					ino := superBloque.S_inodes_count
					folderBlock.B_content[j].B_inodo = ino

					//ACTUALIZAR EL FOLDERBLOCK ACTUAL (idBloque) EN EL ARCHIVO
					folderBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))

					uid, gid, _ := superBloque.ObtenerGID_UID(pathDisco, global.UsuarioActual.Username)

					// Crear el inodo de la carpeta
					newInodo := &estructuras.Inode{
						I_uid:   uid,
						I_gid:   gid,
						I_size:  0,
						I_atime: float32(time.Now().Unix()),
						I_ctime: float32(time.Now().Unix()),
						I_mtime: float32(time.Now().Unix()),
						I_block: [15]int32{superBloque.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
						I_type:  [1]byte{'0'},
						I_perm:  [3]byte{'6', '6', '4'},
					}

					// Serializar el inodo de la carpeta
					newInodo.Serialize(pathDisco, int64(superBloque.S_inode_start+(ino*int32(binary.Size(estructuras.Inode{})))))

					// Actualizar el bitmap de inodos
					superBloque.ActualizarBitmapInodos(pathDisco)
					// Actualizar el superbloque
					superBloque.S_inodes_count++
					superBloque.S_free_inodes_count--
					superBloque.S_first_ino += superBloque.S_inode_size

					// Crear el bloque de la carpeta
					newFolderBlock := &estructuras.FolderBlock{
						B_content: [4]estructuras.FolderContent{
							{B_name: [12]byte{'.'}, B_inodo: ino},
							{B_name: [12]byte{'.', '.'}, B_inodo: folderBlock.B_content[0].B_inodo},
							{B_name: [12]byte{'-'}, B_inodo: -1},
							{B_name: [12]byte{'-'}, B_inodo: -1},
						},
					}

					//escribo el nuevo bloque (block)
					newFolderBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(superBloque.S_blocks_count*int32(binary.Size(estructuras.FolderBlock{})))))

					// Actualizar el bitmap de bloques
					superBloque.ActualizarBitmapBloques(pathDisco)

					// Actualizar el superbloque
					superBloque.S_blocks_count++
					superBloque.S_free_blocks_count--
					superBloque.S_first_blo += superBloque.S_block_size

					//Escribir en el archivo los cambios del superBloque
					superBloque.Serialize(pathDisco, int64(particion.Start))

					return ino
				}
			}
		} else {
			inodo.I_block[i] = superBloque.S_blocks_count

			inodo.Serialize(pathDisco, int64(superBloque.S_inode_start+(idInode*int32(binary.Size(estructuras.Inode{})))))

			//cargo el primer bloque del inodo actual para tomar los datos de actual y padre (son los mismos para el nuevo)
			var folderBlock estructuras.FolderBlock
			bloque := inodo.I_block[0]
			folderBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(bloque*int32(binary.Size(estructuras.FolderBlock{})))))

			// Crear el bloque de la carpeta
			newFolderBlock1 := &estructuras.FolderBlock{
				B_content: [4]estructuras.FolderContent{
					{B_name: [12]byte{'.'}, B_inodo: folderBlock.B_content[0].B_inodo},
					{B_name: [12]byte{'.', '.'}, B_inodo: folderBlock.B_content[1].B_inodo},
					{B_name: [12]byte{'-'}, B_inodo: -1},
					{B_name: [12]byte{'-'}, B_inodo: -1},
				},
			}

			ino := superBloque.S_inodes_count                     //primer inodo libre
			newFolderBlock1.B_content[2].B_inodo = ino            //apuntador al inodo nuevo
			copy(newFolderBlock1.B_content[2].B_name[:], carpeta) //nombre del inodo nuevo

			newFolderBlock1.Serialize(pathDisco, int64(superBloque.S_block_start+(superBloque.S_blocks_count*int32(binary.Size(estructuras.FolderBlock{})))))

			// Actualizar el bitmap de bloques
			superBloque.ActualizarBitmapBloques(pathDisco)

			// Actualizar el superbloque
			superBloque.S_blocks_count++
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo += superBloque.S_block_size

			uid, gid, _ := superBloque.ObtenerGID_UID(pathDisco, global.UsuarioActual.Username)

			// Crear el inodo de la carpeta
			newInodo := &estructuras.Inode{
				I_uid:   uid,
				I_gid:   gid,
				I_size:  0,
				I_atime: float32(time.Now().Unix()),
				I_ctime: float32(time.Now().Unix()),
				I_mtime: float32(time.Now().Unix()),
				I_block: [15]int32{superBloque.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
				I_type:  [1]byte{'0'},
				I_perm:  [3]byte{'6', '6', '4'},
			}

			newInodo.Serialize(pathDisco, int64(superBloque.S_inode_start+(ino*int32(binary.Size(estructuras.Inode{})))))

			// Actualizar el bitmap de inodos
			superBloque.ActualizarBitmapInodos(pathDisco)
			// Actualizar el superbloque
			superBloque.S_inodes_count++
			superBloque.S_free_inodes_count--
			superBloque.S_first_ino += superBloque.S_inode_size

			// Crear el bloque de la carpeta
			newFolderBlock2 := &estructuras.FolderBlock{
				B_content: [4]estructuras.FolderContent{
					{B_name: [12]byte{'.'}, B_inodo: ino},
					{B_name: [12]byte{'.', '.'}, B_inodo: newFolderBlock1.B_content[0].B_inodo},
					{B_name: [12]byte{'-'}, B_inodo: -1},
					{B_name: [12]byte{'-'}, B_inodo: -1},
				},
			}

			newFolderBlock2.Serialize(pathDisco, int64(superBloque.S_block_start+((superBloque.S_blocks_count)*int32(binary.Size(estructuras.FolderBlock{})))))

			// Actualizar el bitmap de bloques
			superBloque.ActualizarBitmapBloques(pathDisco)

			// Actualizar el superbloque
			superBloque.S_blocks_count++
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo += superBloque.S_block_size

			superBloque.Serialize(pathDisco, int64(particion.Start))
			return ino
		}
	} // Fin for bloques directos
	return 0
}

// crea un archivo en el sistema de archivos (no importa si es ext2 o ext3)
func CrearArchivo(idInodo int32, file string, size int, contenido string, pathDisco string) string {

	// el usuario loggeado actualmente tiene el idParticion
	_, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, particion, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		return "MKFILE ERROR: No fue posible obtener superbloque\n"
	}

	// cargo el inodo de la carpeta que contendra el archivo
	var inodoFile estructuras.Inode
	inodoFile.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{})))))

	//recorro el inodo de la carpeta para ver donde guardar el archivo (si hay espacio)
	for i := 0; i < 12; i++ {
		idBloque := inodoFile.I_block[i]
		if idBloque != -1 {
			var folderBlock estructuras.FolderBlock
			folderBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))

			for j := 2; j < 4; j++ {
				apuntador := folderBlock.B_content[j].B_inodo
				//Hay espacio en el bloque
				if apuntador == -1 {
					//modifico el bloque actual
					copy(folderBlock.B_content[j].B_name[:], file)
					ino := superBloque.S_inodes_count //primer inodo libre
					folderBlock.B_content[j].B_inodo = ino
					//ACTUALIZAR EL FOLDERBLOCK ACTUAL (idBloque) EN EL ARCHIVO
					folderBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))

					uid, gid, _ := superBloque.ObtenerGID_UID(pathDisco, global.UsuarioActual.Username)

					// Crear el inodo de la carpeta
					newInodo := &estructuras.Inode{
						I_uid:   uid,
						I_gid:   gid,
						I_size:  int32(size),
						I_atime: float32(time.Now().Unix()),
						I_ctime: float32(time.Now().Unix()),
						I_mtime: float32(time.Now().Unix()),
						I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
						I_type:  [1]byte{'1'},
						I_perm:  [3]byte{'6', '6', '4'},
					}

					//El apuntador a su primer bloque (el primero disponible)
					fileblock := superBloque.S_blocks_count

					//division del contenido en los fileblocks de 64 bytes
					inicio := 0
					fin := 0
					sizeContenido := len(contenido)
					if sizeContenido < 64 {
						fin = len(contenido)
					} else {
						fin = 64
					}

					//crear el/los fileblocks con el contenido del archivo0
					for i := int32(0); i < 12; i++ {
						newInodo.I_block[i] = fileblock
						//Guardar la informacion del bloque
						data := contenido[inicio:fin]
						var newFileBlock estructuras.FileBlock
						copy(newFileBlock.B_content[:], []byte(data))
						//escribo el nuevo bloque (fileblock)
						newFileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(fileblock*int32(binary.Size(estructuras.FileBlock{})))))

						//escribir el bitmap de bloques (se usa un bloque por iteracion).
						superBloque.ActualizarBitmapBloques(pathDisco)
						//modifico el superbloque (solo el bloque usado por iteracion)
						superBloque.S_blocks_count++
						superBloque.S_free_blocks_count--
						superBloque.S_first_blo += superBloque.S_block_size

						//validar si queda data que agregar al archivo para continuar con el ciclo o detenerlo
						calculo := len(contenido[fin:])
						if calculo > 64 {
							inicio = fin
							fin += 64
						} else if calculo > 0 {
							inicio = fin
							fin += calculo
						} else {
							//detener el ciclo de creacion de fileblocks
							break
						}
						//Aumento el fileblock
						fileblock++
					}

					//escribo el nuevo inodo (ino)
					newInodo.Serialize(pathDisco, int64(superBloque.S_inode_start+(ino*int32(binary.Size(estructuras.Inode{})))))

					//escribir el bitmap de inodos (se uso un inodo).
					superBloque.ActualizarBitmapInodos(pathDisco)

					//modifico el superbloque por el inodo usado
					superBloque.S_inodes_count++
					superBloque.S_free_inodes_count--
					superBloque.S_first_ino += superBloque.S_inode_size

					//Escribir en el archivo los cambios del superBloque
					superBloque.Serialize(pathDisco, int64(particion.Start))

					return "Se creó el archivo"
				}
			}
		} else {
			block := superBloque.S_blocks_count
			inodoFile.I_block[i] = block

			inodoFile.Serialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{})))))

			//cargo el primer bloque del inodo actual para tomar los datos de actual y padre (son los mismos para el nuevo)
			var folderBlock estructuras.FolderBlock
			bloque := inodoFile.I_block[0] //cargo el primer folderblock para obtener los datos del actual y su padre
			folderBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(bloque*int32(binary.Size(estructuras.FolderBlock{})))))

			// Crear el bloque de la carpeta
			newFolderBlock1 := &estructuras.FolderBlock{
				B_content: [4]estructuras.FolderContent{
					{B_name: [12]byte{'.'}, B_inodo: folderBlock.B_content[0].B_inodo},
					{B_name: [12]byte{'.', '.'}, B_inodo: folderBlock.B_content[1].B_inodo},
					{B_name: [12]byte{'-'}, B_inodo: -1},
					{B_name: [12]byte{'-'}, B_inodo: -1},
				},
			}

			ino := superBloque.S_inodes_count                  //primer inodo libre
			newFolderBlock1.B_content[2].B_inodo = ino         //apuntador al inodo nuevo
			copy(newFolderBlock1.B_content[2].B_name[:], file) //nombre del inodo nuevo

			//escribo el nuevo bloque (block)
			newFolderBlock1.Serialize(pathDisco, int64(superBloque.S_block_start+(block*int32(binary.Size(estructuras.FolderBlock{})))))

			// Actualizar el bitmap de bloques
			superBloque.ActualizarBitmapBloques(pathDisco)

			// Actualizar el superbloque
			superBloque.S_blocks_count++
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo += superBloque.S_block_size

			uid, gid, _ := superBloque.ObtenerGID_UID(pathDisco, global.UsuarioActual.Username)

			// Crear el inodo de la carpeta
			newInodo := &estructuras.Inode{
				I_uid:   uid,
				I_gid:   gid,
				I_size:  int32(size),
				I_atime: float32(time.Now().Unix()),
				I_ctime: float32(time.Now().Unix()),
				I_mtime: float32(time.Now().Unix()),
				I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
				I_type:  [1]byte{'1'},
				I_perm:  [3]byte{'6', '6', '4'},
			}

			//El apuntador a su primer bloque (el primero disponible)
			fileblock := superBloque.S_blocks_count

			//division del contenido en los fileblocks de 64 bytes
			inicio := 0
			fin := 0
			sizeContenido := len(contenido)
			if sizeContenido < 64 {
				fin = len(contenido)
			} else {
				fin = 64
			}

			//crear el/los fileblocks con el contenido del archivo0
			for i := int32(0); i < 12; i++ {
				newInodo.I_block[i] = fileblock
				//Guardar la informacion del bloque
				data := contenido[inicio:fin]
				var newFileBlock estructuras.FileBlock
				copy(newFileBlock.B_content[:], []byte(data))
				//escribo el nuevo bloque (fileblock)
				newFileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(fileblock*int32(binary.Size(estructuras.FileBlock{})))))

				// Actualizar el bitmap de bloques
				superBloque.ActualizarBitmapBloques(pathDisco)

				// Actualizar el superbloque
				superBloque.S_blocks_count++
				superBloque.S_free_blocks_count--
				superBloque.S_first_blo += superBloque.S_block_size

				//validar si queda data que agregar al archivo para continuar con el ciclo o detenerlo
				calculo := len(contenido[fin:])
				if calculo > 64 {
					inicio = fin
					fin += 64
				} else if calculo > 0 {
					inicio = fin
					fin += calculo
				} else {
					//detener el ciclo de creacion de fileblocks
					break
				}
				//Aumento el fileblock
				fileblock++
			}

			//escribo el nuevo inodo (ino)
			newInodo.Serialize(pathDisco, int64(superBloque.S_inode_start+(ino*int32(binary.Size(estructuras.Inode{})))))
			// Actualizar el bitmap de inodos
			superBloque.ActualizarBitmapInodos(pathDisco)
			// Actualizar el superbloque
			superBloque.S_inodes_count++
			superBloque.S_free_inodes_count--
			superBloque.S_first_ino += superBloque.S_inode_size

			//Escribir en el archivo los cambios del superBloque
			superBloque.Serialize(pathDisco, int64(particion.Start))

			return "Se creó el archivo"
		}
	}

	return "ERROR MKFILE: No fue posible crear el archivo"
}
