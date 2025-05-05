package adminUsuarios

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/global"
	"strconv"
	"strings"
)

/*
mkgrp -name=NOMBRE
*/
func Mkgrp(parametros []string) string {

	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "MKGRP ERROR: para crear un grupo necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	// verificar que sea el usuario root
	usuario, _, id := global.UsuarioActual.ObtenerUsuario()
	if usuario != "root" {
		msj := "MKGRP ERROR: solamente el usuario root puede crear grupos.\n"
		fmt.Println(msj)
		return msj
	}

	// ya sé que solamente viene un parametro
	parametro := strings.Split(parametros[0], "=")
	nombre := strings.ToLower(parametro[0])             // nombre del parametro NAME
	valor := strings.ReplaceAll(parametro[1], "\"", "") // valor del NAME

	// verificando que sea el parametro name
	if nombre != "name" {
		msj := "MKGRP ERROR: el único parametro permitido dentro del comando es NAME\n"
		fmt.Println(msj)
		return msj
	}

	// verificando el tamaño del nombre del grupo
	if len(valor) > 10 {
		msj := "MKGRP ERROR: el tamaño del nombre del grupo es demasiado grande.\n"
		fmt.Println(msj)
		return msj
	}

	// verificando que haya valor
	if valor == "" {
		msj := "MKGRP ERROR: el parametro requiere tener un valor asociado\n"
		fmt.Println(msj)
		return msj
	}

	//retornamos respuesta
	return comandoMkgrp(id, valor)
}

func comandoMkgrp(partId, nombreGrupo string) string {
	// sb, part, path, error
	superBloque, particion, pathDisco, err := estructuras.ObtenerSuperBloque(partId)
	if err != nil {
		msj := "MKGRP " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// obtenemos el inodo de users.txt (inodo 1 con dirección sb.start + 1*inodo.size)
	var inodo estructuras.Inode
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+superBloque.S_inode_size)); err != nil {
		msj := "MKGRP " + err.Error()
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
				msj := "MKGRP " + err.Error()
				fmt.Println(msj)
				return msj
			}
			contenido += string(fileBlock.B_content[:])
			idFileBlock = item
		}
	}

	// tomemos en cuenta que al final siempre habrá un \n, por lo que nos sobra 1 linea
	lineasID := strings.Split(contenido, "\n")

	//Verificamos si el grupo ya existe
	for _, registro := range lineasID[:len(lineasID)-1] {
		datos := strings.Split(registro, ",")
		if len(datos) == 3 {
			if datos[2] == nombreGrupo {
				msj := fmt.Sprintf("MKGRP ERROR: El grupo %s ya existe", nombreGrupo)
				return msj
			}
		}
	}

	//buscamos el último id disponible sin tomar en cuenta los ceros (como nos sobra 1 linea empezamos desde -1)
	id := -1
	for i := len(lineasID) - 2; i >= 0; i-- {
		registro := strings.Split(lineasID[i], ",")
		//valido que sea un grupo
		if registro[1] == "G" {
			// si es diferente de 0 entonces lo tomamos en cuenta
			if registro[0] != "0" {
				//convierto el id en numero para sumarle 1 y crear el nuevo id
				id, err = strconv.Atoi(registro[0])
				if err != nil {
					fmt.Println("MKGRP ERROR: No se pudo obtener un nuevo id para el nuevo grupo")
					return "MKGRP ERROR: No se pudo obtener un nuevo id para el nuevo grupo"
				}
				id++
				break
			}
		}
	}

	//valido que se haya encontrado un nuevo id
	salida := fmt.Sprintf("%d,G,%s\n", id, nombreGrupo)

	if id != -1 {
		contenidoActual := string(fileBlock.B_content[:]) // todo el contenido
		posicionNulo := strings.IndexByte(contenidoActual, 0)
		//Aseguro que haya al menos un byte libre
		if posicionNulo != -1 {
			libre := 64 - (posicionNulo + len(salida))
			if libre > 0 {
				copy(fileBlock.B_content[posicionNulo:], []byte(salida))
				//Escribir el fileblock con espacio libre
				if err := fileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(idFileBlock*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
					msj := "MKGRP " + err.Error()
					fmt.Println(msj)
					return msj
				}
			} else {
				//Si es 0 (quedó exacta), entra aqui y crea un bloque vacío que podrá usarse para el proximo registro
				datosTmp := salida[:len(salida)+libre]
				//Ingreso lo que cabe en el bloque actual
				copy(fileBlock.B_content[posicionNulo:], []byte(datosTmp))

				if err := fileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(idFileBlock*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
					msj := "MKGRP " + err.Error()
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
							msj := "MKGRP " + err.Error()
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
						//crear nuevo fileblock
						var newFileBlock estructuras.FileBlock
						copy(newFileBlock.B_content[:], []byte(datos))

						// serializamos el superbloque
						if err := superBloque.Serialize(pathDisco, int64(particion.Start)); err != nil {
							msj := "MKGRP " + err.Error()
							fmt.Println(msj)
							return msj
						}
						// serializamos el inodo de users.txt
						if err := inodo.Serialize(pathDisco, int64(superBloque.S_inode_start+superBloque.S_inode_size)); err != nil {
							msj := "MKGRP " + err.Error()
							fmt.Println(msj)
							return msj
						}

						// serializamos los bloques
						if newFileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(inodo.I_block[i]*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
							msj := "MKGRP " + err.Error()
							fmt.Println(msj)
							return msj
						}
						break
					}
				}

				// si no hay espacio suficiente
				if informacionGuardada {
					fmt.Println("MKGRP ERROR: Espacio insuficiente para nuevo grupo")
					return "MKGRP ERROR: Espacio insuficiente para nuevo grupo. "
				}
			}
			//contenido
			// for k := 0; k < len(lineasID)-1; k++ {
			// 	fmt.Println(lineasID[k])
			// }
		}
	}

	// al final creamos el journal si es ext3
	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("mkgrp", "/users.txt", salida, fmt.Sprintf("mkgrp -name=%s\n", nombreGrupo)); err != nil {
			msj := "MKGRP " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	return fmt.Sprintf("Se ha agregado creado el grupo %s exitosamente.", nombreGrupo)
}
