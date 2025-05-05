package adminSistemaArchivos

import (
	"fmt"
	"main/estructuras"
	"main/global"
	analisisRecovery "main/recoverAnalisis"
	"strings"
)

func Recovery(parametros []string) string {
	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "RECOVERY ERROR: para crear un grupo necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	// verificar que sea el usuario root
	usuario, _, _ := global.UsuarioActual.ObtenerUsuario()
	if usuario != "root" {
		msj := "RECOVERY ERROR: solamente el usuario root puede crear grupos.\n"
		fmt.Println(msj)
		return msj
	}

	// ya sé que solamente viene un parametro
	parametro := strings.Split(parametros[0], "=")
	nombre := strings.ToLower(parametro[0])              // nombre del parametro ID
	idPart := strings.ReplaceAll(parametro[1], "\"", "") // valor del ID

	// verificando que sea el parametro id
	if nombre != "id" {
		msj := "RECOVERY ERROR: el único parametro permitido dentro del comando es ID\n"
		fmt.Println(msj)
		return msj
	}

	// verificando el idPart del journaling
	Backups, ok := estructuras.Backups[idPart]

	// verificando que haya valor (sino entonces no existe)
	if !ok {
		msj := fmt.Sprintf("RECOVERY ERROR: no hay journaling asociado a la partición %s\n", idPart)
		fmt.Println(msj)
		return msj
	}

	// aqui es donde vamos a hacer al recuperación del archivo pero utilizamos el analizador del analyzer.go
	// llamamos al analizaador pero trabajamos el backup.recovery

	comandos := strings.Split(Backups.Recovery, "\n")
	fmt.Println("\n\n" + Backups.Recovery + "\n\n")
	var resultado string
	for _, linea := range comandos {
		if len(linea) != 0 {
			msj := "\t--> Ejecutando RECOVERY: " + linea + "\n"
			msj += "\t" + analisisRecovery.AnalizarRecovery(linea) + "\n\n"
			//fmt.Println(msj)
			resultado += msj
		}
	}

	return "Recuperación de datos aplicada correctamente, recorrido:\n" + resultado
}
