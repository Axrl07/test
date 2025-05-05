package adminUsuarios

import (
	"fmt"
	"main/estructuras"
	"main/global"
	"strings"
)

type parametros_login struct {
	User string // usuario	- OBLIGATORIO
	Pass string // contraseña - OBLIGATORIO
	Id   string // id de la partición en la que se quiere iniciar sesión - OBLIGATORIO
}

/*
	login -user=NOMBRE -pass=CLAVE -id=IdParticion
*/

// LOGIN no requiere sesión iniciada
func Login(parametros []string) string {

	// si hay sesión iniciada entonces retornamos
	if global.UsuarioActual.HaySesionIniciada() {
		msj := "LOGIN ERROR: para iniciar sesión debe cerrar la sesión anterior."
		fmt.Println(msj)
		return msj
	}

	// creamos el objeto login
	login := parametros_login{}

	// recorremos los parametros
	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "user":
				// quitamos comillas
				usuario := strings.ReplaceAll(parametro[1], "\"", "")
				login.User = usuario
			case "pass":
				// quitamos comillas
				contrasena := strings.ReplaceAll(parametro[1], "\"", "")
				login.Pass = contrasena
			case "id":
				login.Id = parametro[1]
			default:
				return "LOGIN ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			return "LOGIN ERROR: formato invalido para parametro: " + param + "\n"
		}
	}

	// verificamos los parametros obligatorios
	if login.User == "" {
		return "LOGIN ERROR: El parametro User es obligatorio.\n"
	}
	if login.Pass == "" {
		return "LOGIN ERROR: EL parametro Pass es obligatorio.\n"
	}
	if login.Id == "" {
		return "LOGIN ERROR: EL parametro Id es obligatorio.\n"
	}

	return comandoLogin(login)
}

func comandoLogin(comando parametros_login) string {

	// sb, part, path, error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(comando.Id)
	if err != nil {
		msj := "LOGIN " + err.Error()
		fmt.Println(msj)
		return msj
	}

	//contenido users.txt, error
	contenido, err := superBloque.ObtenerUsuariosTxt(pathDisco)
	if err != nil {
		msj := "LOGIN " + err.Error()
		fmt.Println(msj)
		return msj
	}

	//fmt.Println("contenido loggin \n" + contenido + "\nfin contenido\n")
	// separamos por "\n"
	lineas := strings.Split(contenido, "\n")
	// para saber si el user y el pass coinciden
	user := false
	pwd := false
	// iteramos en las lineas
	for _, linea := range lineas {
		// obtenemos los atributos
		atributos := strings.Split(linea, ",")
		// verificamos que sea un usuario
		if len(atributos) == 5 && atributos[1] == "U" {
			// verificamos si se encuentra el usuario y la contraseña
			// id, U , grupo , usuario , clave
			if atributos[3] == comando.User {
				user = true
				if atributos[4] == comando.Pass {
					pwd = true
				}
			}
		}
	}

	// si los datos no coinciden
	if !user {
		msj := "LOGIN ERROR: Verifique el nombre de usuario."
		fmt.Println(msj)
		return msj
	}
	if !pwd {
		msj := "LOGIN ERROR: Verifique la contraseña de usuario."
		fmt.Println(msj)
		return msj
	}

	// loggeamos el usuario
	global.UsuarioActual.Login(comando.User, comando.Pass, comando.Id)

	// verificando que se guardó correctamente
	nombre, clave, id := global.UsuarioActual.ObtenerUsuario()

	return fmt.Sprintf("usuario loggeado : nombre: %s, contraseña: %s, Id Partición: %s", nombre, clave, id)
}
