package adminSistemaArchivos

import (
	"fmt"
	"main/estructuras"
	"main/global"
	"strings"
)

func Loss(parametros []string) string {
	// si no hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "LOSS ERROR: para crear un grupo necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	// verificar que sea el usuario root
	usuario, _, _ := global.UsuarioActual.ObtenerUsuario()
	if usuario != "root" {
		msj := "LOSS ERROR: solamente el usuario root puede crear grupos.\n"
		fmt.Println(msj)
		return msj
	}

	// ya sé que solamente viene un parametro
	parametro := strings.Split(parametros[0], "=")
	nombre := strings.ToLower(parametro[0])              // nombre del parametro ID
	idPart := strings.ReplaceAll(parametro[1], "\"", "") // valor del ID

	// verificando que sea el parametro id
	if nombre != "id" {
		msj := "LOSS ERROR: el único parametro permitido dentro del comando es ID\n"
		fmt.Println(msj)
		return msj
	}

	// verificando el idPart del journaling
	_, ok := estructuras.Backups[idPart]

	// verificando que haya valor (sino entonces no existe)
	if !ok {
		msj := fmt.Sprintf("LOSS ERROR: no hay journaling asociado a la partición %s\n", idPart)
		fmt.Println(msj)
		return msj
	}

	return comandoLoss(idPart)
}

func comandoLoss(idPart string) string {
	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, particion, pathDisco, err := estructuras.ObtenerSuperBloque(idPart)
	if err != nil {
		msj := "LOSS ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// reseteamos contadores de inodos y bloques
	superBloque.S_inodes_count = 0
	superBloque.S_free_inodes_count = superBloque.S_nvalue
	superBloque.S_first_ino = superBloque.S_inode_start
	superBloque.S_blocks_count = 0
	superBloque.S_free_blocks_count = superBloque.S_nvalue * 3
	superBloque.S_first_blo = superBloque.S_block_start

	// limpiamos bm_inodes y bm_block
	superBloque.CreateBitMaps(pathDisco)

	// serializamps
	if err := superBloque.Serialize(pathDisco, int64(particion.Start)); err != nil {
		msj := "LOSS ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	return "Perdida de datos en la partición " + idPart
}
