package adminArchivos

import (
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strings"
)

type parametros_mkdir struct {
	Path string // ruta en la que se creará el archivo - OBLIGATORIO
	P    bool   // opcional (si viene se crean las carpetas padre)
}

/*
	mkdir -path=RUTA
	mkdir -path=RUTA -p
*/

func Mkdir(parametros []string) string {

	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "MKDIR ERROR: para crear un grupo necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	mkdir := parametros_mkdir{}

	// valores por defecto
	mkdir.P = false // no hace falta porque al crear el struct se inicializa con cero, pero para que quede claro

	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		//verificamos que precisamente sea [nombre,valor]
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "path":
				mkdir.Path = strings.ReplaceAll(parametro[1], "\"", "")
			default:
				return "MKDIR ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			if strings.ToLower(parametro[0]) == "p" {
				mkdir.P = true
				continue
			}
			return "MKDIR ERROR: formato invalido para el parametro: " + parametro[0] + "\n"
		}
	}

	// en cualquier caso path es obligaorios
	if mkdir.Path == "" {
		return "MKDIR ERROR: La ruta es obligatoria.\n"
	}

	return comandomkdir(mkdir)
}

func comandomkdir(comando parametros_mkdir) string {
	// el usuario loggeado actualmente tiene el idParticion
	_, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "MKDIR ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	//Validar que exista la ruta
	stepPath := strings.Split(comando.Path, "/")
	idInicial := int32(0)
	idActual := int32(0)
	crear := -1
	for i, itemPath := range stepPath[1:] {
		idActual = gestionSistema.BuscarInodo(idInicial, "/"+itemPath, *superBloque, pathDisco)
		if idInicial != idActual {
			idInicial = idActual
		} else {
			crear = i + 1 //porque estoy iniciando desde 1 e i inicia en 0
			break
		}
	}

	//crear carpetas padre si se tiene permiso
	if crear != -1 {
		if crear == len(stepPath)-1 {
			gestionSistema.CreaCarpeta(idInicial, stepPath[crear])
		} else {
			if comando.P { // utilizamos p
				for _, item := range stepPath[crear:] {
					idInicial = gestionSistema.CreaCarpeta(idInicial, item)
					if idInicial == 0 {
						msj := "MKDIR ERROR: No se pudo crear carpeta"
						fmt.Println(msj)
						return msj
					}
				}
			} else {
				msj := "MKDIR ERROR: no existen las carpetas padre y no se ah dado permisos para crearlas"
				fmt.Println(msj)
				return msj
			}
		}
	} else {
		msj := "MKDIR ERROR: La carpeta ya existe"
		fmt.Println(msj)
		return msj
	}

	if superBloque.S_filesystem_type == 3 {
		var valor string
		if comando.P {
			valor = fmt.Sprintf("mkdir -path=%s -p\n", comando.Path)
		} else {
			valor = fmt.Sprintf("mkdir -path=%s\n", comando.Path)
		}

		if err := estructuras.CrearJournalOperacion("mkdir", comando.Path, "", valor); err != nil {
			msj := "MKDIR " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	return fmt.Sprintf("la carpeta %s fue creada exitosamente", stepPath[len(stepPath)-1])
}
