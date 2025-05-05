package reportes

import (
	"fmt"
	"main/global"
	"path/filepath"
	"strings"
)

type Parametros_rep struct {
	Name       string // nombre del reporte a realizar  - OBLIGATORIO
	Path       string // path donde se va guardar el archivo - OBLIGATORIO
	Id         string // id de la partición que se usará para el reporte - OBLIGATORIO
	PathFileLs string // para repotes file y ls y es el nombre del archivo o carpeta
}

/*
	rep -id=IDENTIFICADOR -path=RUTA -name=NOMBRE
	rep -id=IDENTIFICADOR -path=RUTA -name=NOMBRE - Path_file_ls=RUTA
*/

func Rep(parametros []string) string {
	// creamos el objeto rep
	rep := Parametros_rep{}

	// recorremos los parametros
	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "name":
				// de una vez hacemos a que el name esté en minusculas
				rep.Name = strings.ToLower(parametro[1])
			case "path":
				path := strings.ReplaceAll(parametro[1], "\"", "") //quitamos comillas
				rep.Path = path
			case "id":
				rep.Id = parametro[1]
			case "path_file_ls":
				pathFile := strings.ReplaceAll(parametro[1], "\"", "") //quitamos comillas
				pathFile = strings.TrimRight(pathFile, " ")            //quitamos espacios al final
				rep.PathFileLs = pathFile
			default:
				return "REP ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			return "REP ERROR: formato invalido para parametro: " + param + "\n"
		}
	}

	// verificamos los parametros obligatorios
	if rep.Name == "" {
		msj := "REP ERROR: El parametro name es obligatorio.\n"
		fmt.Println(msj)
		return msj
	}
	if rep.Path == "" {
		msj := "REP ERROR: La ruta del disco es obligatoria.\n"
		fmt.Println(msj)
		return msj
	}
	if rep.Id == "" {
		msj := "REP ERROR: EL id de la partición es obligatorio.\n"
		fmt.Println(msj)
		return msj
	}

	if rep.Name == "ls" || rep.Name == "file" {
		if rep.PathFileLs == "" {
			msj := "REP ERROR: el repote " + rep.Name + " require el parametro " + rep.PathFileLs + "\n"
			fmt.Println(msj)
			return msj
		}

	}

	return comandoRep(rep)
}

// para reportes en graphviz
func comandoRep(comando Parametros_rep) string {

	pathDisco, ok := global.ParticionesMontadas[comando.Id]

	// verificamos que exista el Id en las particiones montada (son las únicas que tienen ID)
	if !ok {
		msj := "REP ERROR: el Id ingresado no existe"
		fmt.Println(msj)
		return msj
	}

	var salida string
	tiposG := []string{"mbr", "disk", "tree", "inode", "block", "sb", "ls"}
	tiposT := []string{"bm_inode", "bm_block", "file"}

	if global.Contiene(tiposG, comando.Name) {
		salida = reporteGrafico(comando, pathDisco)
	} else if global.Contiene(tiposT, comando.Name) {
		salida = reporteTextual(comando)
	} else {
		msj := "REP ERROR: el repote de tipo " + comando.Name + " no existe\n"
		fmt.Println(msj)
		return msj
	}

	return salida
}

// ------------------------------------------- reportes con graphviz -------------------------------------------
func reporteGrafico(comando Parametros_rep, pathDisco string) string {
	// pasamos el path al tipo de reporte de interés
	var contenido string
	var errores error

	// hacemos el reporte
	switch strings.ToLower(comando.Name) {
	case "mbr":
		salida, err := reporteMBR(pathDisco)
		contenido = salida
		errores = err
	case "disk":
		salida, err := reporteDisk(pathDisco)
		contenido = salida
		errores = err
	case "inode":
		salida, err := reporteInode(comando.Id)
		contenido = salida
		errores = err
	case "block":
		salida, err := reporteBlock(comando.Id)
		contenido = salida
		errores = err
	case "tree":
		salida, err := reporteTree(comando.Id)
		contenido = salida
		errores = err
	case "sb":
		salida, err := reporteSb(comando.Id)
		contenido = salida
		errores = err
	case "ls":
		pathFIle := strings.Split(comando.PathFileLs, "/")
		salida, err := reporteLs(comando.Id, pathFIle)
		contenido = salida
		errores = err
	}

	//verificando el error
	if errores != nil {
		return "REP " + errores.Error()
	}

	//grafico, err := estructuras.ReporteMBR(pathDisco)

	// obtenemos el nombre del path
	tmp := strings.Split(comando.Path, "/")
	nombre := strings.Split(tmp[len(tmp)-1], ".")[0]

	// creando carpetas
	carpeta := filepath.Dir(comando.Path)

	//ajustando ruta con nombre
	rutaReporte := carpeta + "/" + nombre + ".dot"

	// ejecución del comando graphviz
	if err := global.RepGraphiz(rutaReporte, contenido, nombre); err != nil {
		msj := "REP " + err.Error()
		fmt.Println(msj)
		return msj
	}

	return fmt.Sprintf("El reporte %s, de tipo %s fue creado exitosamente.", nombre, comando.Name)
}

// ------------------------------------------- reportes sin graphviz -------------------------------------------
func reporteTextual(comando Parametros_rep) string {
	// pasamos el path al tipo de reporte de interés
	var contenido string
	var errores error

	switch strings.ToLower(comando.Name) {
	case "bm_inode":
		salida, err := reporteBm_Inode(comando.Id)
		contenido = salida
		errores = err
	case "bm_block":
		salida, err := reporteBm_Block(comando.Id)
		contenido = salida
		errores = err
	case "file":
		salida, err := reporteFile(comando.Id, comando.PathFileLs)
		contenido = salida
		errores = err
	}

	//verificando el error
	if errores != nil {
		return "REP " + errores.Error()
	}

	// obtenemos el nombre del path
	tmp := strings.Split(comando.Path, "/")
	nombre := strings.Split(tmp[len(tmp)-1], ".")[0]

	// creando carpetas
	carpeta := filepath.Dir(comando.Path)

	//ajustando ruta con nombre
	rutaReporte := carpeta + "/" + nombre + ".txt"

	// ejecución del comando graphviz
	if err := global.Reporte(rutaReporte, contenido); err != nil {
		msj := "REP " + err.Error()
		fmt.Println(msj)
		return msj
	}

	return fmt.Sprintf("El reporte %s, de tipo %s fue creado exitosamente.", nombre, comando.Name)
}
