package adminUsuarios

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/global"
	"strings"
)

/*
rmgrp -name=NOMBRE
*/
func Rmgrp(parametros []string) string {

	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "RMGRP ERROR: para eliminar un grupo necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	// verificar que sea el usuario root
	usuario, _, id := global.UsuarioActual.ObtenerUsuario()
	if usuario != "root" {
		msj := "RMGRP ERROR: solamente el usuario root puede eliminar grupos.\n"
		fmt.Println(msj)
		return msj
	}

	// ya sé que solamente viene un parametro
	parametro := strings.Split(parametros[0], "=")
	nombre := strings.ToLower(parametro[0])             // nombre del parametro NAME
	valor := strings.ReplaceAll(parametro[1], "\"", "") // valor del NAME

	// verificando que sea el parametro name
	if nombre != "name" {
		msj := "RMGRP ERROR: el único parametro permitido dentro del comando es NAME\n"
		fmt.Println(msj)
		return msj
	}

	// verificando el tamaño del nombre del grupo
	if len(valor) > 10 {
		msj := "RMGRP ERROR: el tamaño del nombre del grupo es demasiado grande.\n"
		fmt.Println(msj)
		return msj
	}

	// verificando que haya valor
	if valor == "" {
		msj := "RMGRP ERROR: el parametro requiere tener un valor asociado\n"
		fmt.Println(msj)
		return msj
	}

	//retornamos respuesta
	return comandoRmgrp(id, valor)
}

func comandoRmgrp(partId, nombreGrupo string) string {
	// sb, part, path, error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(partId)
	if err != nil {
		msj := "RMGRP " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// obtenemos el inodo de users.txt (inodo 1 con dirección sb.start + 1*inodo.size)
	var inodo estructuras.Inode
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+superBloque.S_inode_size)); err != nil {
		msj := "RMGRP " + err.Error()
		fmt.Println(msj)
		return msj
	}

	//leer los datos del user.txt
	var contenido string
	var fileBlock estructuras.FileBlock
	for _, item := range inodo.I_block {
		if item != -1 {
			if err := fileBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(item*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
				msj := "RMGRP " + err.Error()
				fmt.Println(msj)
				return msj
			}
			contenido += string(fileBlock.B_content[:])
		}
	}

	// tomemos en cuenta que al final siempre habrá un \n, por lo que nos sobra 1 linea
	lineasID := strings.Split(contenido, "\n")

	//Verificamos si el grupo existe
	existeGrupo := false
	for _, registro := range lineasID[:len(lineasID)-1] {
		datos := strings.Split(registro, ",")
		if datos[1] == "G" && len(datos) == 3 {
			// verificamos si fue eliminado
			if datos[0] == "0" {
				//fue eliminado entonces retornamos un error
				msj := fmt.Sprintf("RMGRP ERROR: El grupo %s ya fue eliminado.", nombreGrupo)
				fmt.Println(msj)
				return msj
			}
			if datos[2] == nombreGrupo {
				existeGrupo = true
			}
		}
	}

	if !existeGrupo {
		msj := fmt.Sprintf("RMGRP ERROR: El grupo %s no existe.", nombreGrupo)
		fmt.Println(msj)
		return msj
	}

	// realizamos cambio ya que no se ha eliminado y si existe
	for i := 0; i < len(lineasID)-1; i++ {
		atributos := strings.Split(lineasID[i], ",")
		// eliminando el grupo
		if atributos[1] == "G" && len(atributos) == 3 {
			// realizo cambios
			if atributos[2] == nombreGrupo {
				// insertamos el 0 en el id del grupo
				atributos[0] = "0"
				lineasID[i] = atributos[0] + "," + atributos[1] + "," + atributos[2]
			}
		}

		// eliminando usuarios que pertenecen al grupo
		if atributos[1] == "U" && len(atributos) == 5 && atributos[2] == nombreGrupo {
			// insertamos el 0 en el id del grupo
			atributos[0] = "0"
			lineasID[i] = atributos[0] + "," + atributos[1] + "," + atributos[2] + "," + atributos[3] + "," + atributos[4]
		}
	}

	// creamos una nueva variable contenido para no confundirnos con la anterior y agregamos enters al final de cada linea
	var contMod string
	for _, linea := range lineasID {
		// vamos guardando el contenido actualizado
		contMod += linea + "\n"
	}

	// verificamos el tamaño total del contenidoModificado
	var fin int
	if len(contMod) > 64 {
		// el archivo no termina pero topa los 64 bits
		fin = 64
	} else {
		// el archivo termina en el tamaño total
		fin = len(contMod)
	}

	var inicio int // para ir iterando el contenido cada 64 bytes junto con el fin
	for _, item := range inodo.I_block {
		if item != -1 {
			//tomo 64 bytes de la cadena o los bytes que queden
			data := contMod[inicio:fin]
			//Modifico y guardo el bloque actual
			var newFileBlock estructuras.FileBlock
			copy(newFileBlock.B_content[:], []byte(data))
			// serializamos el actual
			if err := newFileBlock.Serialize(pathDisco, int64(superBloque.S_block_start+(item*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
				msj := "RMGRP ERROR: inconvenientes durante la serialización del FileBlock."
				fmt.Println(msj)
				return msj
			}
			// procedemos a guardar el nuevo tamaño de la cadena
			inicio = fin
			calculo := len(contMod[fin:])
			if calculo > 64 {
				fin += 64
			} else {
				fin += calculo
			}
		}
	}

	// al final creamos el journal si es ext3 (nombreGrupo porque así borramos todos los usuarios y el grupo)
	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("rmgrp", "/users.txt", nombreGrupo, fmt.Sprintf("rmgrp -name=%s\n", nombreGrupo)); err != nil {
			msj := "RMGRP " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	//fmt.Println(contMod)
	return fmt.Sprintf("Se ha eliminado el grupo %s exitosamente.", nombreGrupo)
}
