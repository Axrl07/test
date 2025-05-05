package adminArchivos

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strings"
)

type parametros_edit struct {
	Path      string // donde se va a pegar el nuevo contenido
	Contenido string // el contenido que se va a pegar
}

func Edit(parametros []string) string {
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "EDIT ERROR: para utilizar el comando EDIT debe iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	edit := parametros_edit{}

	for _, param := range parametros {
		parametro := strings.Split(param, "=")
		switch strings.ToLower(parametro[0]) {
		case "path":
			edit.Path = strings.ReplaceAll(parametro[1], "\"", "")
		case "contenido":
			edit.Contenido = strings.ReplaceAll(parametro[1], "\"", "")
		default:
			return fmt.Sprint("EDIT ERROR: el parametro %s no es valido", parametro[0])
		}
	}

	if edit.Path == "" {
		return "EDIT ERROR: el parametro path es obligatorio"
	}

	if edit.Contenido == "" {
		return "EDIT ERROR: el parametro contenido es obligatorio"
	}

	return comandoEdit(edit)
}

func comandoEdit(comando parametros_edit) string {
	usuario, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, particion, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "EDIT ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// para verificar los permisos
	uidCreador, gidCreador, err := superBloque.ObtenerGID_UID(pathDisco, usuario)
	if err != nil {
		msj := "EDIT ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// primero debemos obtener el contenido del archivo (si se necesita de un archivo cambiar contenido por var contenido string y descomentar lo comentado abajo)
	contenido, err := ObtenerContenidoPC(comando.Contenido)
	if err != nil {
		msj := "EDIT " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// //buscar el inodo que contiene el archivo buscado
	// idInodoContenido := gestionSistema.BuscarInodo(0, comando.Contenido, *superBloque, pathDisco)
	// var inodo estructuras.Inode

	// if idInodoContenido > 0 {
	// 	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodoContenido*int32(binary.Size(estructuras.Inode{}))))); err != nil {
	// 		msj := "EDIT " + err.Error()
	// 		fmt.Println(msj)
	// 		return msj
	// 	}

	// 	//Verifica que el usuario actual sea del grupo root o sea el usuario que creó la carpeta
	// 	if inodo.I_uid == uidCreador || gidGrupo == 1 {

	// 		//recorrer los fileblocks del inodo para obtener toda su informacion
	// 		for _, idBlock := range inodo.I_block {
	// 			if idBlock != -1 {
	// 				var fileBlock estructuras.FileBlock
	// 				fileBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBlock*int32(binary.Size(estructuras.FileBlock{})))))
	// 				contenido += global.BorrandoIlegibles(string(fileBlock.B_content[:]))
	// 			}
	// 		}
	// 	} else {
	// 		return "EDIT ERROR: No tiene permisos para visualizar el archivo " + comando.Contenido + "\n"
	// 	}

	// } else {
	// 	return "EDIT ERROR: No se encontro el archivo " + comando.Contenido + "\n"
	// }

	// ahora buscamos el archivo donde se va a agregar el contenido al final
	idInodoDestino := gestionSistema.BuscarInodo(0, comando.Path, *superBloque, pathDisco)
	var inodoDestino estructuras.Inode

	var terminamos bool // me dice si se terminó de editar correctamente el archivo
	if idInodoDestino > 0 {
		if err := inodoDestino.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodoDestino*int32(binary.Size(estructuras.Inode{}))))); err != nil {
			msj := "EDIT " + err.Error()
			fmt.Println(msj)
			return msj
		}

		//Verifica que el usuario actual sea del grupo root o sea el usuario que creó la carpeta
		if inodoDestino.I_uid == uidCreador || gidCreador == 1 || inodoDestino.I_gid == gidCreador || uidCreador == 1 {

			//division del contenido en los fileblocks de 64 bytes
			inicio := 0
			fin := 0
			sizeContenido := len(contenido)
			if sizeContenido < 64 {
				fin = len(contenido)
			} else {
				fin = 64
			}

			//recorrer los fileblocks del inodo para obtener toda su informacion
			for i := 0; i < 12; i++ {
				if inodoDestino.I_block[i] == -1 {
					// cortamos la data
					data := contenido[inicio:fin]
					var newFileBlock estructuras.FileBlock
					// copiamos bytes
					copy(newFileBlock.B_content[:], []byte(data))
					//escribo el nuevo bloque (fileblock)
					newFileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(superBloque.S_blocks_count*int32(binary.Size(estructuras.FileBlock{})))))

					// actualizamos el inodo
					inodoDestino.I_block[i] = superBloque.S_blocks_count

					// actualizar el bitmap de bloques
					superBloque.ActualizarBitmapBloques(pathDisco)

					// Actualizar el superbloque
					superBloque.S_blocks_count++
					superBloque.S_free_blocks_count--
					superBloque.S_first_blo += superBloque.S_block_size

					// validamos si queda data que agregar al archivo
					calculo := len(contenido[fin:])
					if calculo > 64 {
						inicio = fin
						fin += 64
					} else if calculo > 0 {
						inicio = fin
						fin += calculo
					} else {
						//detener el ciclo de creacion de fileblocks
						terminamos = true
						break
					}
				} else {
					// si el bloque no es -1, entonces lo deserializamos
					var fileBlock estructuras.FileBlock
					if err := fileBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(inodoDestino.I_block[i]*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
						msj := "EDIT " + err.Error()
						fmt.Println(msj)
						return msj
					}

					// recalculamos el contenido
					contenidoActual := global.BorrandoIlegibles(string(fileBlock.B_content[:])) + "\n"
					contenido = contenidoActual + contenido
					// fin = 64 - inicio
					inicio = 0
					fin = 0
					sizeContenido = len(contenido)
					if sizeContenido < 64 {
						fin = len(contenido)
					} else {
						fin = 64
					}

					data := contenido[inicio:fin]
					// copiamos bytes
					copy(fileBlock.B_content[:], []byte(data))
					//escribo el nuevo bloque (fileblock)
					fileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(inodoDestino.I_block[i]*int32(binary.Size(estructuras.FileBlock{})))))

					// validamos si queda data que agregar al archivo
					calculo := len(contenido[fin:])
					if calculo > 64 {
						inicio = fin
						fin += 64
					} else if calculo > 0 {
						inicio = fin
						fin += calculo
					} else {
						//detener el ciclo de creacion de fileblocks
						terminamos = true
						break
					}
				}
			}

			// seriamos el inodo
			if err := inodoDestino.Serialize(pathDisco, int64(superBloque.S_inode_start+(idInodoDestino*int32(binary.Size(estructuras.Inode{}))))); err != nil {
				msj := "EDIT " + err.Error()
				fmt.Println(msj)
				return msj
			}

			// actualizamos el superbloque
			if err := superBloque.Serialize(pathDisco, int64(particion.Start)); err != nil {
				msj := "EDIT " + err.Error()
				fmt.Println(msj)
				return msj
			}
		} else {
			return "EDIT ERROR: No tiene permisos para visualizar el archivo " + comando.Path + "\n"
		}
	} else {
		return "EDIT ERROR: No se encontro el archivo " + comando.Path + "\n"
	}

	if !terminamos {
		return "EDIT ERROR: No fue posible terminar de editar el archivo " + comando.Path + " correctamente\n"
	}

	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("edit", comando.Path, comando.Contenido, fmt.Sprintf("edit -path=%s -contenido=%s\n", comando.Path, comando.Contenido)); err != nil {
			msj := "EDIT " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	return "EL archivo " + comando.Path + " ha sido editado correctamente\n"
}
