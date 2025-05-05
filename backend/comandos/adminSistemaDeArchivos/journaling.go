package adminSistemaArchivos

import (
	"fmt"
	"main/estructuras"
	"main/global"
	"strings"
)

func Journaling(parametros []string) string {

	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "JOURNALING ERROR: para crear un grupo necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	// verificar que sea el usuario root
	usuario, _, _ := global.UsuarioActual.ObtenerUsuario()
	if usuario != "root" {
		msj := "JOURNALING ERROR: solamente el usuario root puede crear grupos.\n"
		fmt.Println(msj)
		return msj
	}

	// ya sé que solamente viene un parametro
	parametro := strings.Split(parametros[0], "=")
	nombre := strings.ToLower(parametro[0])              // nombre del parametro ID
	idPart := strings.ReplaceAll(parametro[1], "\"", "") // valor del ID

	// verificando que sea el parametro id
	if nombre != "id" {
		msj := "JOURNALING ERROR: el único parametro permitido dentro del comando es ID\n"
		fmt.Println(msj)
		return msj
	}

	// verificando el idPart del journaling
	backup, ok := estructuras.Backups[idPart]

	// verificando que haya valor
	if !ok {
		msj := fmt.Sprintf("JOURNALING ERROR: no hay journaling asociado a la partición %s\n", idPart)
		fmt.Println(msj)
		return msj
	}

	// guardamos la tabla
	global.TablaHTMLJournaling = backup.ReporteJournals()
	// regreso una tabla html
	return "Tabla journaling creada correctamente"
}
