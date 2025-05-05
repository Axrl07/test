package adminUsuarios

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/global"
	"strings"
)

type parametros_chgrp struct {
	User string // usuario	- OBLIGATORIO
	Grp  string // id del grupo al que pertenece - OBLIGATORIO
}

/*
	chgrp -usr=NOMBRE -grp=GRUPO
*/

func Chgrp(parametros []string) string {
	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "CHGRP ERROR: para crear un grupo necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	// verificar que sea el usuario root
	usuario, _, partId := global.UsuarioActual.ObtenerUsuario()
	if usuario != "root" {
		msj := "CHGRP ERROR: solamente el usuario root puede crear grupos.\n"
		fmt.Println(msj)
		return msj
	}

	chgrp := parametros_chgrp{}
	// recorremos los parametros
	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "user":
				// quitamos comillas
				usuario := strings.ReplaceAll(parametro[1], "\"", "")
				if len(usuario) >= 10 {
					msj := "CHGRP ERROR: el tamaño del nombre de usuario es demasiado grande.\n"
					fmt.Println(msj)
					return msj
				}
				chgrp.User = usuario
			case "grp":
				grupo := strings.ReplaceAll(parametro[1], "\"", "")
				if len(grupo) >= 10 {
					msj := "CHGRP ERROR: el tamaño del nombre del grupo es demasiado grande.\n"
					fmt.Println(msj)
					return msj
				}
				chgrp.Grp = parametro[1]
			default:
				return "CHGRP ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			return "CHGRP ERROR: formato invalido para parametro: " + param + "\n"
		}
	}

	// verificamos los parametros obligatorios
	if chgrp.User == "" {
		return "CHGRP ERROR: El parametro User es obligatorio.\n"
	}
	if chgrp.Grp == "" {
		return "CHGRP ERROR: EL parametro Id es obligatorio.\n"
	}

	//retornamos respuesta
	return comandoChgrp(chgrp, partId)
}

func comandoChgrp(comando parametros_chgrp, idParticion string) string {
	// sb, part, path, error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "CHGRP " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// obtenemos el inodo de users.txt (inodo 1 con dirección sb.start + 1*inodo.size)
	var inodo estructuras.Inode
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+superBloque.S_inode_size)); err != nil {
		msj := "CHGRP " + err.Error()
		fmt.Println(msj)
		return msj
	}

	//leer los datos del user.txt
	var contenido string
	var fileBlock estructuras.FileBlock
	for _, item := range inodo.I_block {
		if item != -1 {
			if err := fileBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(item*int32(binary.Size(estructuras.FileBlock{}))))); err != nil {
				msj := "CHGRP " + err.Error()
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
		if len(datos) == 3 && datos[1] == "G" {
			// verificamos si fue eliminado
			if datos[0] == "0" && datos[2] == comando.Grp {
				//fue eliminado entonces retornamos un error
				msj := fmt.Sprintf("CHGRP ERROR: El grupo %s fue eliminado.", comando.Grp)
				fmt.Println(msj)
				return msj
			}
			if datos[2] == comando.Grp {
				existeGrupo = true
			}
		}
	}

	if !existeGrupo {
		msj := fmt.Sprintf("CHGRP ERROR: El grupo %s no existe.", comando.Grp)
		fmt.Println(msj)
		return msj
	}

	//Verificamos si el usuario existe
	existeUsuario := false
	for _, registro := range lineasID[:len(lineasID)-1] {
		datos := strings.Split(registro, ",")
		// verifico que sea usuario
		if len(datos) == 5 && datos[1] == "U" {
			// verifico que sea el que se ingresó
			if datos[3] == comando.User {
				// verifico que no haya sido eliminado
				if datos[0] == "0" && datos[3] == comando.User {
					//fue eliminado entonces retornamos un error
					msj := fmt.Sprintf("CHGRP ERROR: El usuario %s fue eliminado.", comando.User)
					fmt.Println(msj)
					return msj
				}
				existeUsuario = true
			}
		}
	}

	//verificamos que existe el usuario
	if !existeUsuario {
		msj := fmt.Sprintf("CHGRP ERROR: El usuario %s no existe.", comando.User)
		fmt.Println(msj)
		return msj
	}

	// buscamos y usuario cambiarlo de grupo
	salida := ""
	for i := 0; i < len(lineasID)-1; i++ {
		atributos := strings.Split(lineasID[i], ",")
		// como ya verificamos anteriormente ahora solamente buscamos coincidencia con U , len(atributos) y usuario
		if atributos[1] == "U" && len(atributos) == 5 {
			// verifico que no haya sid
			// si es diferente de 0 entonces lo tomamos en cuenta
			if atributos[3] == comando.User {
				//seteamos el nuevo grupo
				atributos[2] = comando.Grp
				// insertamos el grupo en las lineas
				lineasID[i] = atributos[0] + "," + atributos[1] + "," + atributos[2] + "," + atributos[3] + "," + atributos[4]
				salida = atributos[0] + "," + atributos[1] + "," + atributos[2] + "," + atributos[3] + "," + atributos[4] + "\n"
			}
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
				msj := "CHGRP ERROR: inconvenientes durante la serialización del FileBlock."
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

	// al final creamos el journal si es ext3 (salida = id,U,grupoNuevo,usuario,pwd ya que solamente sería volver a enviar grupoNuevo y usuario)
	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("chgrp", "/users.txt", salida, fmt.Sprintf("chgrp -user=%s -grp=%s \n", comando.User, comando.Grp)); err != nil {
			msj := "CHGRP " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	//fmt.Println(contMod)
	return fmt.Sprintf("El usuario %s fue cambiado al grupo %s correctamente", comando.User, comando.Grp)
}
