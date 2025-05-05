package adminArchivos

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strconv"
	"strings"
)

/*
	cat file1=RUTA file2=Ruta2 ...
*/

func Cat(parametros []string) string {

	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "MKFILE ERROR: para crear utilizar el comando CAT debe iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	var filePaths []string

	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		//verificamos que precisamente sea [nombre,valor]
		if len(parametro) == 2 {
			filen := strings.ToLower(parametro[0][:4])
			if filen == "file" {
				numero := strings.ReplaceAll(strings.ToLower(parametro[0]), "file", "")
				_, errId := strconv.Atoi(numero)
				if errId != nil {
					fmt.Println("CAT ERROR: No se pudo obtener un numero de fichero")
					return "CAT ERROR: No se pudo obtener un numero de fichero"
				}
				//eliminar comillas
				tmp1 := strings.ReplaceAll(parametro[1], "\"", "")
				filePaths = append(filePaths, tmp1)
			} else {
				return "CAT ERROR: El único parametro válido es fileN, con N siendo un número entero positivo.\n"
			}
		} else {
			return "CAT ERROR: formato invalido para el parametro: " + parametro[0] + "\n"
		}
	}

	// el usuario loggeado actualmente tiene el idParticion
	usuario, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	return ComandoCat(filePaths, usuario, idParticion)
}

func ComandoCat(filen []string, usuario, idParticion string) string {

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "CAT ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// para verificar los permisos
	uidCreador, gidGrupo, err := superBloque.ObtenerGID_UID(pathDisco, usuario)
	if err != nil {
		msj := "CAT ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	var contenido string                // contenido completo del  archivo
	var fileBlock estructuras.FileBlock // lo creamos aquí porque no necesitamos más que un lugar donde alojar la nueva información
	for _, item := range filen {
		//buscar el inodo que contiene el archivo buscado
		idInodo := gestionSistema.BuscarInodo(0, item, *superBloque, pathDisco)
		var inodo estructuras.Inode

		// se asume que no se crean archivos en la raiz
		if idInodo > 0 {
			contenido += "\nContenido del archivo: '" + item + "':\n"
			if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{}))))); err != nil {
				msj := "CAT " + err.Error()
				fmt.Println(msj)
				return msj
			}

			//Verifica que el usuario actual sea del grupo root o sea el usuario que creó la carpeta
			if inodo.I_uid == uidCreador || gidGrupo == 1 {

				//recorrer los fileblocks del inodo para obtener toda su informacion
				for _, idBlock := range inodo.I_block {
					if idBlock != -1 {
						fileBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBlock*int32(binary.Size(estructuras.FileBlock{})))))
						contenido += global.BorrandoIlegibles(string(fileBlock.B_content[:])) + "\n"
					}
				}
				contenido += "\n"
			} else {
				contenido += "CAT ERROR: No tiene permisos para visualizar el archivo " + item + "\n"
			}

		} else {
			contenido += "CAT ERROR: No se encontro el archivo " + item + "\n"
		}
	}

	return contenido
}
