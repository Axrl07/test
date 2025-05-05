package adminArchivos

import (
	"fmt"
	"io"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"os"
	"strconv"
	"strings"
)

type parametros_mkfile struct {
	Path      string // ruta en la que se creará el archivo - OBLIGATORIO
	R         bool   // opcional (si viene se crean las carpetas padre)
	Size      int32  //  opcional por defecto 0 bytes (literalmente pa llenar de 0 - 9 pa llenar ese size)
	Cont      string // ruta opcional
	Contenido string // aquí guardamos el contenido en bytes
}

/*
	mkfile -path=RUTA
	mkfile -path=RUTA -r
	// en cualquier caso de size y cont (venga uno y no el otro)
	mkfile -path=RUTA -r -size=12 -cont=RUTAPC
*/

func Mkfile(parametros []string) string {

	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "MKFILE ERROR: para crear un archivo debe iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	mkfile := parametros_mkfile{}

	// valores por defecto
	mkfile.R = false
	mkfile.Size = 0 // no hace falta porque al crear el struct se inicializa con cero, pero para que quede claro

	var parametrosString []string
	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		//verificamos que precisamente sea [nombre,valor]
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "cont":
				mkfile.Cont = strings.ReplaceAll(parametro[1], "\"", "")
				parametrosString = append(parametrosString, "-cont="+parametro[1])
			case "path":
				mkfile.Path = strings.ReplaceAll(parametro[1], "\"", "")
				parametrosString = append(parametrosString, "-path="+parametro[1])
			case "size":
				// el valorInicial no está convertido a las unidades
				valorInicial, err := strconv.Atoi(parametro[1])
				if err != nil {
					return "MKFILE ERROR: Error al convertir el tamaño del archivo: " + parametro[1] + "\n"
				}
				if valorInicial < 0 {
					return "MKFILE ERROR: El tamaño del archivo no puede ser negativo: " + parametro[1] + "\n"
				}
				mkfile.Size = int32(valorInicial)
				parametrosString = append(parametrosString, "-size="+parametro[1])
			case "r":
				return "MKFILE ERROR: el parametro -r no puede venir con valores, se ingresó: -r=" + parametro[1]
			default:
				return "MKFILE ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			if strings.ToLower(parametro[0]) == "r" {
				mkfile.R = true
				parametrosString = append(parametrosString, "-r")
				continue
			}
			return "MKFILE ERROR: formato invalido para el parametro: " + parametro[0] + "\n"
		}
	}

	// en cualquier caso path es obligaorios
	if mkfile.Path == "" {
		return "MKFILE ERROR: La ruta es obligatoria.\n"
	}

	if mkfile.Cont == "" {
		// asignamos valor al contenido del struct
		mkfile.Contenido = calcularContenido(mkfile.Size)
	} else {
		// obtenemos el valor del contenido del archivo en Escritorio
		valor, err := ObtenerContenidoPC(mkfile.Cont)
		if err != nil {
			msj := "MKFILE " + err.Error()
			fmt.Println(msj)
			return msj
		}
		// asignamos valor al struct
		mkfile.Contenido = valor
		// actualizamos el valor de size al struct
		mkfile.Size = int32(len(valor))
	}

	parametrosStringCadena := strings.Join(parametrosString, " ")

	return comandomkfile(mkfile, parametrosStringCadena)
}

func comandomkfile(comando parametros_mkfile, parametrosString string) string {
	// el usuario loggeado actualmente tiene el idParticion
	_, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "MKFILE ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	//Validar que exista la ruta
	stepPath := strings.Split(comando.Path, "/")
	finRuta := len(stepPath) - 1 //es el archivo -> stepPath[finRuta] = archivoNuevo.txt
	idInicial := int32(0)
	idActual := int32(0)
	crear := -1
	//No incluye a finRuta, es decir, se queda en el aterior. EJ: Tamaño=5, finRuta=4. El ultimo que evalua es stepPath[3]
	for i, itemPath := range stepPath[1:finRuta] {
		idActual = gestionSistema.BuscarInodo(idInicial, "/"+itemPath, *superBloque, pathDisco)
		//si el actual y el inicial son iguales significa que no existe la carpeta
		if idInicial != idActual {
			idInicial = idActual
		} else {
			crear = i + 1 //porque estoy iniciando desde 1 e i inicia en 0
			break
		}
	}

	//crear carpetas padre si se tiene permiso
	if crear != -1 {
		if comando.R {
			for _, item := range stepPath[crear:finRuta] {
				idInicial = gestionSistema.CreaCarpeta(idInicial, item)
				if idInicial == 0 {
					msj := "MKFILE ERROR: No se pudo crear carpeta"
					fmt.Println(msj)
					return msj
				}
			}
		} else {
			msj := "MKFILE ERROR: Carpeta " + stepPath[crear] + " no existe. Sin permiso de crear carpetas padre"
			println(msj)
			return msj
		}
	}

	//verificar que no exista el archivo (recordar que BuscarInodo busca de la forma /nombreBuscar)
	idNuevo := gestionSistema.BuscarInodo(idInicial, "/"+stepPath[finRuta], *superBloque, pathDisco)
	if idNuevo == idInicial {
		gestionSistema.CrearArchivo(idInicial, stepPath[finRuta], int(comando.Size), comando.Contenido, pathDisco)
	} else {
		msj := "MKFILE ERROR: El archivo ya existe"
		fmt.Println(msj)
		return msj
	}

	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("mkfile", comando.Path, comando.Contenido, fmt.Sprintf("mkfile %s\n", parametrosString)); err != nil {
			msj := "MKFILE " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	return fmt.Sprintf("la archivo %s fue creada exitosamente", stepPath[len(stepPath)-1])
}

// obtiene el contenido completo de un archivo en el path especificado
func ObtenerContenidoPC(pathPc string) (string, error) {
	// Open file
	archivo, err := os.OpenFile(pathPc, os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("ERROR: no fue posible abrir el archivo %s: %v", pathPc, err)
	}
	defer archivo.Close()

	// Read content
	contenido, err := io.ReadAll(archivo)
	if err != nil {
		return "", fmt.Errorf("ERROR: no fue posible leer el archivo %s: %v", pathPc, err)
	}

	return string(contenido), nil
}

// se calcula un contenido lleno de "0123456789" hasta que llegue al tamaño solicitado
func calcularContenido(size int32) string {
	content := ""
	for len(content) < int(size) {
		content += "0123456789"
	}
	return content[:size] // recortando desde el byte 1 hasta el size
}
