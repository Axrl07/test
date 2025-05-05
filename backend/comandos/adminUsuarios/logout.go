package adminUsuarios

import (
	"fmt"
	"main/global"
)

func Logout() string {
	// si hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "LOGOUT ERROR: para cerrar sesión debe haber una la sesión iniciada."
		fmt.Println(msj)
		return msj
	}

	// cerramos la sesión
	nombre, _, _ := global.UsuarioActual.ObtenerUsuario()
	global.UsuarioActual.Logout()

	return "Se ha cerrado la sesión del usuario " + nombre + " exitosamente."
}
