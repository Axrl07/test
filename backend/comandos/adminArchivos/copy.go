package adminArchivos

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strings"
)

type parametros_copy struct {
	Path    string // OBLIGATORIO
	Destino string // OBLIGATORIO
}

/*
	copy -path=RUTA -destino=RUTA
*/

func Copy(parametros []string) string {

	copy := parametros_copy{}

	for _, param := range parametros {
		parametro := strings.Split(param, "=")
		switch strings.ToLower(parametro[0]) {
		case "path":
			path := strings.ReplaceAll(parametro[1], "\"", "")
			copy.Path = path
		case "destino":
			destino := strings.ReplaceAll(parametro[1], "\"", "")
			copy.Destino = destino
		default:
			msj := fmt.Sprintf("COPY ERROR: El parametro %s no es admitido dentrod el comando.\n", parametro[0])
			fmt.Println(msj)
			return msj
		}
	}

	if copy.Path == "" {
		msj := "COPY ERROR: El parametro -path es obligatorio.\n"
		fmt.Println(msj)
		return msj
	}

	if copy.Destino == "" {
		msj := "COPY ERROR: El parametro -destino es obligatorio.\n"
		fmt.Println(msj)
		return msj
	}

	return comandoCopy(copy)
}

func comandoCopy(comando parametros_copy) string {
	// obtenemos la partición, el path del disco fisico (virtual), error
	usuario, _, idParticion := global.UsuarioActual.ObtenerUsuario()
	superBloque, particion, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "COPY ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// para verificar los permisos
	uidCreador, gidGrupo, err := superBloque.ObtenerGID_UID(pathDisco, usuario)
	if err != nil {
		msj := "COPY ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// primero obtenemos el inodo donde vamos a copiar (el destino)
	idDestino := gestionSistema.BuscarInodo(0, comando.Destino, *superBloque, pathDisco)

	// verificamos que el inodo destino exista
	if idDestino == -1 {
		msj := "COPY ERROR: No se encontro el directorio destino " + comando.Destino + "\n"
		fmt.Println(msj)
		return msj
	}

	// obtenemos el directorio destino
	var inodoDestino estructuras.Inode
	inodoDestino.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idDestino*int32(binary.Size(estructuras.Inode{})))))

	//verificamos si el destino es un archivo, si lo es no se puede copiar
	if string(inodoDestino.I_type[:]) == "1" {
		msj := "COPY ERROR: No se puede copiar a un archivo, el destino debe ser un directorio.\n"
		fmt.Println(msj)
		return msj
	}

	//verificando permisos de escritura y lectura
	if uidCreador == inodoDestino.I_uid || gidGrupo == 1 || gidGrupo == inodoDestino.I_gid || uidCreador == 1 {
		// obtenemos el inodo path (la carpeta a copiar)
		idCopia := gestionSistema.BuscarInodo(0, comando.Path, *superBloque, pathDisco)

		// verificamos que el inodo a copiar exista
		if idCopia == -1 {
			msj := "COPY ERROR: No se encontro el archivo " + comando.Path + "\n"
			fmt.Println(msj)
			return msj
		}

		// ya que sabemos que si existe deserealizamos desde donde nace (el bloque que apunta a este)
		var inodoCopiar estructuras.Inode
		inodoCopiar.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idCopia*int32(binary.Size(estructuras.Inode{})))))

		// nombre del archivo o carpeta a copiar
		arregloRuta := strings.Split(comando.Path, "/")
		nombre := arregloRuta[len(arregloRuta)-1]

		// verificamos donde colocar la copia (si en un bloqueFolder que apunte a un archivo o dentro de los I_block)
		var folderBlock estructuras.FolderBlock
		var idPapa int32 // guardamos el id del papa
		for _, idBloque := range inodoDestino.I_block {
			//verificamos que exista el bloque
			if err := folderBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
				msj := "COPY ERROR: No se encontro el bloque folder " + comando.Destino + "\n"
				fmt.Println(msj)
				return msj
			}

			var guardamos bool
			var nuevo bool
			if idBloque != -1 {
				// guardamos el papa por si no existe espacio en este bloque
				idPapa = folderBlock.B_content[1].B_inodo
				// si apunta a un folder block nos intersa saber si está vacío en alguna de sus 2 posiciones
				for i := 2; i < 4; i++ {
					apuntador := folderBlock.B_content[i].B_inodo
					if apuntador == -1 {
						// el apuntador es hacia el inodo que se va a crear por eso el contador de inodos actual sin aumentarlo
						folderBlock.B_content[i].B_inodo = superBloque.S_inodes_count
						// el nombre de la carpeta o archivo copia
						copy(folderBlock.B_content[i].B_name[:], nombre)
						guardamos = true
						break
					}
				}
			} else {
				// si no apunta a nada creamos un bloque y lo conectamos aquí (el idBloque == -1)
				// Crear el bloque de la carpeta
				nuevoBloque := &estructuras.FolderBlock{
					B_content: [4]estructuras.FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: idDestino},
						{B_name: [12]byte{'.', '.'}, B_inodo: idPapa},
						{B_name: [12]byte{'-'}, B_inodo: superBloque.S_inodes_count},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}
				//llenamos la informaicón  del primer apuntador libre (usamos el contador actual de inodos SIN AUMENTARLO, eso se hace al ir creando las cocpias)
				copy(nuevoBloque.B_content[2].B_name[:], nombre)
				guardamos = true
				nuevo = true
			}

			// si se guardó la información entonces trabajamos
			if guardamos {
				// Serializamos el bloque de carpeta
				if err := folderBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
					msj := "COPY ERROR: No se pudo serializar el bloque folder " + comando.Destino + "\n"
					fmt.Println(msj)
					return msj
				}

				if nuevo {
					// Actualizamos bitmap y contadores de bloques
					if err := superBloque.ActualizarBitmapBloques(pathDisco); err != nil {
						msj := "COPY ERROR: No se pudo actualizar el bitmap de bloques " + comando.Destino + "\n"
						fmt.Println(msj)
						return msj
					}
					superBloque.S_blocks_count++
					superBloque.S_free_blocks_count--
					superBloque.S_first_blo += superBloque.S_block_size
				}
				break
			}
		}

		// hacemos la copia pero pasamos el path.Destino para que sea el padre en el primer bloque copiado (en vez del padre de la carpeta a copiar)
		if err := copiar(superBloque, pathDisco, idCopia, particion); err != nil {
			msj := "COPY ERROR: " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("copy", comando.Destino, comando.Path, fmt.Sprintf("copy -destino=%s -path=%s\n", comando.Destino, comando.Path)); err != nil {
			msj := "COPY " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	return fmt.Sprintf("se copió la carpeta %s en  %s \n", comando.Path, comando.Destino)
}

func copiar(superBloque *estructuras.SuperBlock, pathDisco string, idCopia int32, particion *estructuras.Partition) error {
	// Obtener el inodo a copiar
	var inodoCopiar estructuras.Inode
	if err := inodoCopiar.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idCopia*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		return err
	}

	// Crear nuevo inodo (copia)
	nuevoInodo := estructuras.Inode{
		I_uid:   inodoCopiar.I_uid,
		I_gid:   inodoCopiar.I_gid,
		I_size:  inodoCopiar.I_size,
		I_atime: inodoCopiar.I_atime,
		I_ctime: inodoCopiar.I_ctime,
		I_mtime: inodoCopiar.I_mtime,
		I_type:  inodoCopiar.I_type,
		I_perm:  inodoCopiar.I_perm,
	}

	// Inicializar todos los bloques a -1 (vacíos)
	for i := range nuevoInodo.I_block {
		nuevoInodo.I_block[i] = -1
	}

	// Guardar el ID del nuevo inodo
	nuevoInodoID := superBloque.S_inodes_count

	// Serializamos el nuevo inodo primero (se actualizará más adelante)
	if err := nuevoInodo.Serialize(pathDisco, int64(superBloque.S_inode_start+(nuevoInodoID*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		return err
	}

	// Actualizamos bitmap y contadores de inodos
	if err := superBloque.ActualizarBitmapInodos(pathDisco); err != nil {
		return err
	}
	superBloque.S_inodes_count++
	superBloque.S_free_inodes_count--
	superBloque.S_first_ino += superBloque.S_inode_size

	// Si es un archivo, copiamos su contenido
	if string(inodoCopiar.I_type[:]) == "1" {
		for i, idBloque := range inodoCopiar.I_block {
			// Si el bloque está vacío, terminamos
			if idBloque == -1 {
				break
			}

			// Obtenemos el bloque de archivo
			var bloqueArchivo estructuras.FileBlock
			if err := bloqueArchivo.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
				return err
			}

			// Crear nuevo bloque de archivo
			nuevoBloqueArchivo := estructuras.FileBlock{
				B_content: bloqueArchivo.B_content,
			}

			// Guardamos el ID del nuevo bloque
			nuevoBloqueID := superBloque.S_blocks_count

			// Actualizamos la referencia en el nuevo inodo
			nuevoInodo.I_block[i] = nuevoBloqueID

			// Serializamos el nuevo bloque
			if err := nuevoBloqueArchivo.Serialize(pathDisco, int64(superBloque.S_block_start+(nuevoBloqueID*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
				return err
			}

			// Actualizamos bitmap y contadores de bloques
			if err := superBloque.ActualizarBitmapBloques(pathDisco); err != nil {
				return err
			}
			superBloque.S_blocks_count++
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo += superBloque.S_block_size
		}
	} else {
		// Es una carpeta, copiamos su estructura
		for i, idBloque := range inodoCopiar.I_block {
			// Si el bloque está vacío, terminamos
			if idBloque == -1 {
				break
			}

			// Obtenemos el bloque de carpeta
			var bloqueCarpeta estructuras.FolderBlock
			if err := bloqueCarpeta.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
				return err
			}

			// Crear nuevo bloque de carpeta
			nuevoBloqueCarpeta := estructuras.FolderBlock{}

			// El nuevo bloque tiene el mismo padre pero actualiza su referencia a sí mismo
			nuevoBloqueCarpeta.B_content[0].B_name = [12]byte{'.'}
			nuevoBloqueCarpeta.B_content[0].B_inodo = nuevoInodoID // Apunta al nuevo inodo

			nuevoBloqueCarpeta.B_content[1].B_name = [12]byte{'.', '.'}
			nuevoBloqueCarpeta.B_content[1].B_inodo = bloqueCarpeta.B_content[1].B_inodo // Mantiene el mismo padre

			// Inicializar el resto de entradas como vacías
			nuevoBloqueCarpeta.B_content[2].B_inodo = -1
			nuevoBloqueCarpeta.B_content[3].B_inodo = -1

			// Guardamos el ID del nuevo bloque
			nuevoBloqueID := superBloque.S_blocks_count

			// Actualizamos la referencia en el nuevo inodo
			nuevoInodo.I_block[i] = nuevoBloqueID

			// Serializamos el nuevo bloque (inicialmente con entradas vacías)
			if err := nuevoBloqueCarpeta.Serialize(pathDisco, int64(superBloque.S_block_start+(nuevoBloqueID*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
				return err
			}

			// Actualizamos bitmap y contadores de bloques
			if err := superBloque.ActualizarBitmapBloques(pathDisco); err != nil {
				return err
			}
			superBloque.S_blocks_count++
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo += superBloque.S_block_size

			// Ahora copiamos recursivamente los contenidos de la carpeta
			for j := 2; j < 4; j++ {
				// Si la entrada está vacía, continuamos
				if bloqueCarpeta.B_content[j].B_inodo == -1 {
					continue
				}

				// Copiamos el nombre al nuevo bloque
				copy(nuevoBloqueCarpeta.B_content[j].B_name[:], bloqueCarpeta.B_content[j].B_name[:])

				// Guardamos el ID actual (antes de copiar)
				currentInodeCount := superBloque.S_inodes_count

				// Copiamos recursivamente el inodo apuntado
				idInodoHijo := bloqueCarpeta.B_content[j].B_inodo
				if err := copiar(superBloque, pathDisco, idInodoHijo, particion); err != nil {
					return err
				}

				// El nuevo ID del inodo hijo es el que se generó durante la copia recursiva
				nuevoBloqueCarpeta.B_content[j].B_inodo = currentInodeCount
			}

			// Serializamos el bloque de carpeta actualizado con todas sus entradas
			if err := nuevoBloqueCarpeta.Serialize(pathDisco, int64(superBloque.S_block_start+(nuevoBloqueID*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
				return err
			}
		}
	}

	// Serializamos el inodo actualizado con todos sus bloques
	if err := nuevoInodo.Serialize(pathDisco, int64(superBloque.S_inode_start+(nuevoInodoID*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		return err
	}

	// Guardamos el superbloque actualizado
	if err := superBloque.Serialize(pathDisco, int64(particion.Start)); err != nil {
		return err
	}

	return nil
}
