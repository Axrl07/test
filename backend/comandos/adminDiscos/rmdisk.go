package adminDiscos

import (
	"fmt"
	"os"
	"strings"
)

/*
	rmdisk -path=RUTA
*/

func Rmdisk(parametros []string) string {
	// no puede venir más de un parametro
	if len(parametros) != 1 {
		msj := "MRDISK ERROR: el único parametro permitido es path\n"
		fmt.Println(msj)
		return msj
	}

	// verificando parametro
	params := strings.Split(parametros[0], "=") // [path, direccion]
	cmd, path := strings.ToLower(params[0]), params[1]
	if cmd != "path" {
		msj := fmt.Sprintf("MRDISK ERROR: parametro desconocido: %s \n", cmd)
		fmt.Println(msj)
		return msj
	}

	// configurando la ruta
	path = strings.ReplaceAll(path, "\"", "")
	disco := strings.Split(path, "/")

	//validar si existe el Disco para eliminar
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		msj := "RMDISK Error: El disco " + disco[len(disco)-1] + " no existe\n"
		fmt.Println(msj)
		return msj
	}

	//Eliminar disco
	if err := os.Remove(path); err != nil {
		fmt.Println("RMDISK Error: No pudo eliminarse el disco " + disco[len(disco)-1])
		return "RMDISK Error: No pudo eliminarse el disco " + disco[len(disco)-1]
	}

	return "Disco " + disco[len(disco)-1] + " eliminado\n"
}
