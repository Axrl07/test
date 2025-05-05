package adminArchivos

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strings"
)

type parametros_move struct {
	Path    string // dirección actual
	Destino string // dirección destino
}

func Move(parametros []string) string {
	// verificamos que haya sesión iniciada
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "MOVE ERROR: para utilizar el comando MOVE debe iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	MOVE := parametros_move{}

	for _, param := range parametros {
		parametro := strings.Split(param, "=")
		switch strings.ToLower(parametro[0]) {
		case "path":
			MOVE.Path = strings.ReplaceAll(parametro[1], "\"", "")
		case "destino":
			MOVE.Destino = strings.ReplaceAll(parametro[1], "\"", "")
		default:
			return fmt.Sprintf("MOVE ERROR: el parametro %s no es valido", parametro[0])
		}
	}

	if MOVE.Path == "" {
		return "MOVE ERROR: el parametro path es obligatorio"
	}

	if MOVE.Destino == "" {
		return "MOVE ERROR: el parametro contenido es obligatorio"
	}

	return comandoMove(MOVE)
}

func comandoMove(comando parametros_move) string {
	usuario, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, particion, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "MOVE ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// verificamos que exista la carpeta o archivo
	//var inodo estructuras.Inode
	idInodo := gestionSistema.BuscarInodo(0, comando.Path, *superBloque, pathDisco)
	if idInodo == -1 {
		msj := "MOVE ERROR: no ha sido posible encontrar el inodo " + comando.Path
		fmt.Println(msj)
		return msj
	}
	var inodo estructuras.Inode
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		msj := "MOVE ERROR: no ha sido posible obtener la información de " + comando.Path
		fmt.Println(msj)
		return msj
	}

	// para verificar los permisos
	uidCreador, gidCreador, err := superBloque.ObtenerGID_UID(pathDisco, usuario)
	if err != nil {
		msj := "MOVE ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// verificamos si existen permisos de escritura
	if inodo.I_uid != uidCreador && uidCreador != 1 && inodo.I_gid != gidCreador && gidCreador != 1 {
		msj := "MOVE ERROR: no tiene permisos de escritura en el archivo o carpeta " + comando.Path
		fmt.Println(msj)
		return msj
	}

	// verificamos que exista la carpeta o archivo destino
	var inodoDestino estructuras.Inode
	idDestino := gestionSistema.BuscarInodo(0, comando.Destino, *superBloque, pathDisco)
	if idDestino == -1 {
		msj := "MOVE ERROR: no ha sido posible encontrar el inodo destino " + comando.Destino
		fmt.Println(msj)
		return msj
	}
	if err := inodoDestino.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idDestino*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		msj := "MOVE ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// ahora vamos a borrar el registro actual del inodo a mover
	tmp := strings.Split(comando.Path, "/")
	// obtenemos nombre del inodo
	nombre := tmp[len(tmp)-1]
	// cortamos
	tmp = tmp[:len(tmp)-1]
	ruta := strings.Join(tmp, "/")

	// ahora obtenemos el bloque que apunta al inodo
	var inodoPadre estructuras.Inode
	idInodoPadre := gestionSistema.BuscarInodo(0, ruta, *superBloque, pathDisco)
	if idInodoPadre == -1 {
		msj := "MOVE ERROR: no ha sido posible encontrar el inodo padre " + ruta
		fmt.Println(msj)
		return msj
	}
	if err := inodoPadre.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodoPadre*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		msj := "MOVE ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// vamos a eliminar la conexión entre el padre y el inodo
	folderBLock := estructuras.FolderBlock{}
	for _, idBLoque := range inodoPadre.I_block {
		if idBLoque != -1 {
			if err := folderBLock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBLoque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
				msj := "MOVE ERROR: " + err.Error()
				fmt.Println(msj)
				return msj
			}

			// buscamos si el nombre coincide con el nombre del inodo
			var salimos bool
			for i := 2; i < 4; i++ {
				// si coinciden los nombres entonces borramos registro
				nombreGuardado := global.BorrandoIlegibles(string(folderBLock.B_content[i].B_name[:]))
				if nombreGuardado == nombre {
					folderBLock.B_content[i].B_name = [12]byte{}
					folderBLock.B_content[i].B_inodo = -1

					if err := folderBLock.Serialize(pathDisco, int64(superBloque.S_block_start+(idBLoque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
						msj := "MOVE ERROR: " + err.Error()
						fmt.Println(msj)
						return msj
					}

					salimos = true
					break
				}
			}

			if salimos {
				break
			}
		}
	}

	// obtenemos el id del padre del destino por si lo necesitamos
	tmp = strings.Split(comando.Destino, "/")
	ruta = strings.Join(tmp, "/")

	idPapa := gestionSistema.BuscarInodo(0, ruta, *superBloque, pathDisco)
	if idPapa == -1 {
		msj := "MOVE ERROR: no ha sido posible encontrar el inodo padre " + ruta
		fmt.Println(msj)
		return msj
	}

	// ahora conectamos el destino con el nuevo inodo
	for indice, idBloque := range inodoDestino.I_block {
		if idBloque == -1 {
			// creamos aquí
			newFolderBlock := &estructuras.FolderBlock{
				B_content: [4]estructuras.FolderContent{
					{B_name: [12]byte{'.'}, B_inodo: idDestino},
					{B_name: [12]byte{'.', '.'}, B_inodo: idPapa},
					{B_name: [12]byte{'-'}, B_inodo: -1},
					{B_name: [12]byte{'-'}, B_inodo: -1},
				},
			}

			//escribimos valores
			copy(newFolderBlock.B_content[2].B_name[:], []byte(nombre))
			newFolderBlock.B_content[2].B_inodo = idInodo

			// serializamos el bloque
			if err := newFolderBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(superBloque.S_blocks_count*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
				msj := "MOVE ERROR: " + err.Error()
				fmt.Println(msj)
				return msj
			}

			// actualizamos el inodo (conectamos el bloque)
			inodoDestino.I_block[indice] = superBloque.S_blocks_count

			// serializamos el inodo destino
			if err := inodoDestino.Serialize(pathDisco, int64(superBloque.S_inode_start+(idDestino*int32(binary.Size(estructuras.Inode{}))))); err != nil {
				msj := "MOVE ERROR: " + err.Error()
				fmt.Println(msj)
				return msj
			}

			// actualizamos bitmap de bloques y superbloque
			if err := superBloque.ActualizarBitmapBloques(pathDisco); err != nil {
				msj := "MOVE ERROR: " + err.Error()
				fmt.Println(msj)
				return msj
			}

			superBloque.S_blocks_count++
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo += superBloque.S_block_size

			// serializamos el superbloque
			if err := superBloque.Serialize(pathDisco, int64(particion.Start)); err != nil {
				msj := "MOVE ERROR: " + err.Error()
				fmt.Println(msj)
				return msj
			}
			break

		} else {
			if err := folderBLock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
				msj := "MOVE ERROR: " + err.Error()
				fmt.Println(msj)
				return msj
			}

			// buscamos si hay espacio para conectar
			var salimos bool
			for i := 2; i < 4; i++ {
				// si hay espacio conectamos
				if folderBLock.B_content[i].B_inodo == -1 {
					copy(folderBLock.B_content[i].B_name[:], []byte(nombre))
					folderBLock.B_content[i].B_inodo = idInodo
					salimos = true
					// serializamos el bloque
					if err := folderBLock.Serialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
						msj := "MOVE ERROR: " + err.Error()
						fmt.Println(msj)
						return msj
					}
					break
				}

			}

			if salimos {
				break
			}
		}
	}

	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("move", comando.Destino, comando.Path, fmt.Sprintf("move -destino=%s -path=%s\n", comando.Destino, comando.Path)); err != nil {
			msj := "MOVE " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	return fmt.Sprintf("Se ha movido correctamente %s a %s .\n", comando.Path, comando.Destino)
}
