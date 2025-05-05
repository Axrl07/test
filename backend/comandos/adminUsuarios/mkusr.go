package adminUsuarios

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/global"
	"strconv"
	"strings"
)

type parametros_mkusr struct {
	User string // usuario	- OBLIGATORIO
	Pass string // contraseña - OBLIGATORIO
	Grp  string // id del grupo al que pertenece - OBLIGATORIO
}

/*
mkusr -usr=NOMBRE -pass=CLAVE -grp=GRUPO
*/
func Mkusr(parametros []string) string {
	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "MKUSR ERROR: para crear un usuario necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	// verificar que sea el usuario root
	usuario, _, partId := global.UsuarioActual.ObtenerUsuario()
	if usuario != "root" {
		msj := "MKUSR ERROR: solamente el usuario root puede crear usuarios.\n"
		fmt.Println(msj)
		return msj
	}

	mkusr := parametros_mkusr{}
	// recorremos los parametros
	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "user":
				// quitamos comillas
				usuario := strings.ReplaceAll(parametro[1], "\"", "")
				if len(usuario) > 10 {
					msj := "MKUSR ERROR: el tamaño del nombre de usuario es demasiado grande.\n"
					fmt.Println(msj)
					return msj
				}
				mkusr.User = usuario
			case "pass":
				// quitamos comillas
				contrasena := strings.ReplaceAll(parametro[1], "\"", "")
				if len(contrasena) > 10 {
					msj := "MKUSR ERROR: el tamaño de la contraseña es demasiado grande.\n"
					fmt.Println(msj)
					return msj
				}
				mkusr.Pass = contrasena
			case "grp":
				grupo := strings.ReplaceAll(parametro[1], "\"", "")
				if len(grupo) > 10 {
					msj := "MKUSR ERROR: el tamaño del nombre del grupo es demasiado grande.\n"
					fmt.Println(msj)
					return msj
				}
				mkusr.Grp = parametro[1]
			default:
				return "MKUSR ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			return "MKUSR ERROR: formato invalido para parametro: " + param + "\n"
		}
	}

	// verificamos los parametros obligatorios
	if mkusr.User == "" {
		return "MKUSR ERROR: El parametro User es obligatorio.\n"
	}
	if mkusr.Pass == "" {
		return "MKUSR ERROR: EL parametro Pass es obligatorio.\n"
	}
	if mkusr.Grp == "" {
		return "MKUSR ERROR: EL parametro Id es obligatorio.\n"
	}

	//retornamos respuesta
	return comandoMkusr(mkusr, partId)
}

func comandoMkusr(comando parametros_mkusr, idParticion string) string {
	// sb, part, path, error
	superBloque, particion, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "MKUSR " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// obtenemos el inodo de users.txt (inodo 1 con dirección sb.start + 1*inodo.size)
	var inodo estructuras.Inode
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+superBloque.S_inode_size)); err != nil {
		msj := "MKUSR " + err.Error()
		fmt.Println(msj)
		return msj
	}

	//leer los datos del user.txt
	var contenido string
	var fileBlock estructuras.FileBlock
	var idFileBlock int32 //numero de ultimo fileblock para trabajar sobre ese
	for _, item := range inodo.I_block {
		if item != -1 {
			if err := fileBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(item*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
				msj := "MKUSR " + err.Error()
				fmt.Println(msj)
				return msj
			}
			contenido += string(fileBlock.B_content[:])
			idFileBlock = item
		}
	}

	// tomemos en cuenta que al final siempre habrá un \n, por lo que nos sobra 1 linea
	lineasID := strings.Split(contenido, "\n")

	//Verificamos si el grupo y el usuario ya existen
	existeGrupo := false
	for _, registro := range lineasID[:len(lineasID)-1] {
		datos := strings.Split(registro, ",")
		if len(datos) == 3 {
			if datos[2] == comando.Grp {
				existeGrupo = true
			}
		}
	}

	// verificamos que exista el grupo exista
	if !existeGrupo {
		msj := fmt.Sprintf("MKUSR ERROR: El grupo %s no existe.", comando.Grp)
		fmt.Println(msj)
		return msj
	}

	//Verificamos si el grupo y el usuario ya existen
	existeUsuario := false
	for _, registro := range lineasID[:len(lineasID)-1] {
		datos := strings.Split(registro, ",")
		if len(datos) == 5 {
			// si ya existe el usuario entonces retornamos error
			if datos[3] == comando.User {
				existeUsuario = true
			}
		}
	}

	// verificamos que exista el grupo exista
	if existeUsuario {
		msj := fmt.Sprintf("MKUSR ERROR: El usuario %s ya existe.", comando.User)
		fmt.Println(msj)
		return msj
	}

	//buscamos el último id disponible sin tomar en cuenta los ceros (como nos sobra 1 linea empezamos desde -1)
	id := -1
	for i := len(lineasID) - 2; i >= 0; i-- {
		registro := strings.Split(lineasID[i], ",")
		// valido que sea un usuario
		if registro[1] == "U" {
			// si es diferente de 0 entonces lo tomamos en cuenta
			if registro[0] != "0" {
				//convierto el id en numero para sumarle 1 y crear el nuevo id
				id, err = strconv.Atoi(registro[0])
				if err != nil {
					msj := fmt.Sprintf("MKUSR ERROR: inconvenientes calculando id para el usuario: %s", comando.User)
					fmt.Println(msj)
					return msj
				}
				id++
				break
			}
		}
	}

	// es lo mismo que en mkgrp, pero cambiando que ahora ingresamos el usuario y no un grupo
	salida := fmt.Sprintf("%d,U,%s,%s,%s\n", id, comando.Grp, comando.User, comando.Pass)
	if id != -1 {
		contenidoActual := string(fileBlock.B_content[:]) // todo el contenido
		fmt.Println(contenido)
		posicionNulo := strings.IndexByte(contenidoActual, 0)
		//Aseguro que haya al menos un byte libre
		if posicionNulo != -1 {
			libre := 64 - (posicionNulo + len(salida))
			if libre > 0 {
				copy(fileBlock.B_content[posicionNulo:], []byte(salida))
				//Escribir el fileblock con espacio libre
				if err := fileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(idFileBlock*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
					msj := "MKUSR " + err.Error()
					fmt.Println(msj)
					return msj
				}
			} else {
				//Si es 0 (quedó exacta), entra aqui y crea un bloque vacío que podrá usarse para el proximo registro
				datosTmp := salida[:len(salida)+libre]
				fmt.Printf("contenido temporal : %s", datosTmp)
				//Ingreso lo que cabe en el bloque actual
				copy(fileBlock.B_content[posicionNulo:], []byte(datosTmp))

				if err := fileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(idFileBlock*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
					msj := "MKUSR " + err.Error()
					fmt.Println(msj)
					return msj
				}

				//creamos otro FileBlock para el resto de la informacion
				informacionGuardada := true
				for i, item := range inodo.I_block {
					if item == -1 {
						informacionGuardada = false

						// asigno numero al bloque
						inodo.I_block[i] = superBloque.S_blocks_count
						// actualizamos el bitmap de bloques
						if err := superBloque.ActualizarBitmapBloques(pathDisco); err != nil {
							msj := "MKUSR " + err.Error()
							fmt.Println(msj)
							return msj
						}

						//aumento contador de bloques
						superBloque.S_blocks_count++
						superBloque.S_free_blocks_count--
						superBloque.S_first_blo += superBloque.S_block_size

						// antes manejaba con direcciones de aountadores
						/*
							//agrego el apuntador del bloque al inodo
							inodo.I_block[i] = superBloque.S_first_blo
							//disminuyo contador de bloques libres
							superBloque.S_free_blocks_count -= 1
							superBloque.S_first_blo += 1
						*/

						datos := salida[len(salida)+libre:]
						fmt.Printf("datos dentro del for: %s", datos)
						//crear nuevo fileblock
						var newFileBlock estructuras.FileBlock
						copy(newFileBlock.B_content[:], []byte(datos))

						// serializamos el superbloque
						if err := superBloque.Serialize(pathDisco, int64(particion.Start)); err != nil {
							msj := "MKUSR " + err.Error()
							fmt.Println(msj)
							return msj
						}
						// serializamos el inodo de users.txt
						if err := inodo.Serialize(pathDisco, int64(superBloque.S_inode_start+superBloque.S_inode_size)); err != nil {
							msj := "MKUSR " + err.Error()
							fmt.Println(msj)
							return msj
						}

						// serializamos los bloques
						if newFileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(inodo.I_block[i]*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
							msj := "MKUSR " + err.Error()
							fmt.Println(msj)
							return msj
						}
						break
					}
				}

				// si no hay espacio suficiente
				if informacionGuardada {
					fmt.Println("MKUSR ERROR: Espacio insuficiente para nuevo usuario")
					return "MKUSR ERROR: Espacio insuficiente para nuevo usuario. "
				}
			}
			// for k := 0; k < len(lineasID)-1; k++ {
			// 	fmt.Println(lineasID[k])
			// }
		}
	}

	// al final creamos el journal si es ext3
	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("mkusr", "/users.txt", salida, fmt.Sprintf("mkusr -user=%s -grp=%s -pass=%s\n", comando.User, comando.Grp, comando.Pass)); err != nil {
			msj := "MKUSR " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	return fmt.Sprintf("Se ha creado y agregado al usuario %s al grupo %s exitosamente.", comando.User, comando.Grp)
}
