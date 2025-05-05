package adminArchivos

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strings"
)

type parametros_find struct {
	Path string // dirección de la carpeta o archivo desde donde empieza la busqueda
	Name int32  // identificador del tipo de archivo
}

func Find(parametros []string) string {
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "FIND ERROR: para utilizar el comando FIND debe iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	find := parametros_find{}

	// inicializamos find.Name para saber si vino entre los parametros
	find.Name = -1

	for _, param := range parametros {
		parametro := strings.Split(param, "=")
		switch strings.ToLower(parametro[0]) {
		case "path":
			find.Path = strings.ReplaceAll(parametro[1], "\"", "")
		case "name":
			nombre := strings.ReplaceAll(parametro[1], "\"", "")
			switch nombre {
			case "?.*":
				find.Name = 0
			case "*.?":
				find.Name = 1
			case "?.?":
				find.Name = 2
			case "*.*":
				find.Name = 3
			case "?":
				find.Name = 4
			case "*":
				find.Name = 5
			default:
				return fmt.Sprintf("FIND ERROR: valor %s no es valido en el parametro NAME", nombre)
			}
		default:
			return fmt.Sprintf("FIND ERROR: el parametro %s no es valido", parametro[0])
		}
	}

	if find.Path == "" {
		return "FIND ERROR: el parametro path es obligatorio"
	}
	if find.Name == -1 {
		return "FIND ERROR: El parametro Name es obligatorio"
	}

	return comandoFind(find)
}

// verifica la sintaxis del nombre carpeta o archivo según el patrón especificado
func verificarCoincidenciaNombre(nombreActual string, patronBusqueda int32) bool {
	/*
	   ?.* ->  0  (Un solo carácter seguido de una extensión de cualquier longitud)
	   *.? -> 1   (Cualquier longitud seguido de una extensión de un solo carácter)
	   ?.? ->  2  (Un solo carácter seguido de una extensión de un solo caracter)
	   *.* -> 3   (Cualquier nombre con cualquier extensión)
	   ? -> 4     (Un solo carácter sin extensión)
	   * -> 5     (Cualquier nombre sin restricciones)
	*/

	// Si tiene punto, el nombre tiene extensión
	if strings.Contains(nombreActual, ".") {
		slice := strings.Split(nombreActual, ".")
		nombreSinExt := slice[0]
		extension := slice[1]

		switch patronBusqueda {
		case 0: // ?.*
			return len(nombreSinExt) == 1
		case 1: // *.?
			return len(extension) == 1
		case 2: // ?.?
			return len(nombreSinExt) == 1 && len(extension) == 1
		case 3: // *.*
			return true // Cualquier archivo con extensión coincide
		case 4: // ?
			return false // ? solo funciona para nombres sin extensión
		case 5: // *
			return true // * coincide con cualquier nombre
		}
	} else {
		// Sin extensión
		switch patronBusqueda {
		case 0, 1, 2, 3: // Todos los patrones con extensión
			return false
		case 4: // ?
			return len(nombreActual) == 1
		case 5: // *
			return true
		}
	}

	return false
}

func comandoFind(comando parametros_find) string {
	// obtenemos la idparticion
	_, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "FIND ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// verificamos que exista la carpeta o archivo
	idInodo := gestionSistema.BuscarInodo(0, comando.Path, *superBloque, pathDisco)
	if idInodo == -1 {
		msj := "FIND ERROR: no fue posible encontrar " + comando.Path
		fmt.Println(msj)
		return msj
	}

	// Obtenemos el nombre de la carpeta/archivo de inicio
	tmp := strings.Split(comando.Path, "/")
	nombre := tmp[len(tmp)-1]
	if nombre == "" {
		nombre = "/"
	}

	resultados := "\t" + nombre + "\n"
	resultados, err = armarCamino(idInodo, comando, superBloque, pathDisco, resultados, 1)
	if err != nil {
		fmt.Println(err.Error())
		return err.Error()
	}

	return resultados
}

func armarCamino(idInodo int32, comando parametros_find, superBloque *estructuras.SuperBlock, pathDisco string, resultados string, nivel int) (string, error) {
	// Cargar el inodo desde el cual vamos a buscar
	var inodo estructuras.Inode
	err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{})))))
	if err != nil {
		return resultados, fmt.Errorf("FIND ERROR: al cargar el inodo %d: %v", idInodo, err)
	}

	// Si es una carpeta, vamos a ingresar a sus elementos
	if inodo.I_type[0] == '0' {
		// Es una carpeta, buscamos en su contenido recursivamente
		for i := 0; i < 12; i++ {
			idBloque := inodo.I_block[i]
			if idBloque != -1 {
				var folderBlock estructuras.FolderBlock
				err := folderBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))
				if err != nil {
					return resultados, fmt.Errorf("FIND ERROR: al cargar el bloque de carpeta %d: %v", idBloque, err)
				}

				// Recorrer el contenido del bloque de carpeta
				for j := 0; j < 4; j++ { // Recorremos todo el bloque incluyendo . y ..
					apuntador := folderBlock.B_content[j].B_inodo
					nombre := string(folderBlock.B_content[j].B_name[:])
					// Eliminar caracteres nulos
					nombre = strings.TrimRight(nombre, "\x00")

					if apuntador != -1 && nombre != "." && nombre != ".." {
						// Verificar si el nombre coincide con el patrón de búsqueda
						if verificarCoincidenciaNombre(nombre, comando.Name) {
							// Agregar indentación según el nivel
							indentacion := strings.Repeat("\t|_", nivel)
							resultados += indentacion + nombre + "\n"
						}

						// Cargar el inodo apuntado para continuar la búsqueda
						var inodoHijo estructuras.Inode
						err := inodoHijo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(apuntador*int32(binary.Size(estructuras.Inode{})))))
						if err != nil {
							return resultados, fmt.Errorf("FIND ERROR: al cargar el inodo hijo %d: %v", apuntador, err)
						}

						// Si es una carpeta, continuamos la búsqueda dentro de ella
						if inodoHijo.I_type[0] == '0' {
							resultados, err = armarCamino(apuntador, comando, superBloque, pathDisco, resultados, nivel+1)
							if err != nil {
								return resultados, err
							}
						}
					}
				}
			}
		}
	}

	return resultados, nil
}
