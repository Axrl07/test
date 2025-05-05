package adminDiscos

import (
	"fmt"
	"main/estructuras"
	"main/global"
	"strings"
)

type parametros_mount struct {
	// ambos son OBLIGATORIOS
	Path string
	Name string
}

/*
mount -path=RUTA -name=NOMBRE
*/
func Mount(parametros []string) string {

	mount := parametros_mount{}

	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "name":
				mount.Name = parametro[1]
			case "path":
				path := strings.ReplaceAll(parametro[1], "\"", "")
				mount.Path = path
			default:
				return "MOUNT ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			return "MOUNT ERROR: formato invalido para parametro: " + param + "\n"
		}
	}

	//verificamos que vengan los parametros
	if mount.Path == "" {
		msj := "MOUNT ERROR: el parametro " + mount.Path + " es obligatorio" + "\n"
		fmt.Println(msj)
		return msj
	}

	if mount.Name == "" {
		msj := "MOUNT ERROR: el parametro " + mount.Name + " es obligatorio" + "\n"
		fmt.Println(msj)
		return msj
	}

	return comandoMount(mount)
}

func comandoMount(comando parametros_mount) string {

	// abrimos el disco
	disco, err := global.AbrirDisco(comando.Path)
	if err != nil {
		msj := "MOUNT " + err.Error() + "\n"
		fmt.Println(msj)
		return msj
	}

	// obtenemos el MBR que está en el disco
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		msj := "MOUNT " + err.Error() + "\n"
		fmt.Println(msj)
		return msj
	}

	// obtenemos la partición por su nombre
	particion, indexP := mbr.ParticionPorNombre(comando.Name)
	if particion == nil {
		msj := "MOUNT ERROR: No existe una partición bajo el nombre " + comando.Name + "\n"
		fmt.Println(msj)
		return msj
	}

	//verificamos que sea primaria
	if string(particion.Type[:]) != "P" {
		msj := "MOUNT ERROR: la partición " + comando.Name + " no se puede montar ya que no es primaria.\n"
		fmt.Println(msj)
		return msj
	}

	// verificamos su Id, si tiene entonces ya fue montada
	if global.BorrandoIlegibles(string(particion.Id[:])) != "0" {
		msj := "MOUNT ERROR: la partición " + comando.Name + " ya fue montada\n"
		fmt.Println(msj)
		return msj
	}

	// generamos un Id para cada partición
	idParticion, indice, err := GenerarIdParticion(&comando, indexP)
	if err != nil {
		msj := "MOUNT " + err.Error() + "\n"
		fmt.Println(msj)
		return msj
	}

	//  Guardar la partición montada en la lista de montajes globales
	global.ParticionesMontadas[idParticion] = comando.Path

	// modificando particion
	particion.MontarParticion(int32(indice), idParticion)

	// guardamos la partición montada en el mbr
	mbr.Partitions[indexP] = *particion

	// escribimos en disco
	if err := global.EscribirEnDisco(disco, mbr, 0); err != nil {
		msj := "MOUNT " + err.Error() + "\n"
		fmt.Println(msj)
		return msj
	}

	//cerramos el disco
	defer disco.Close()
	particion.Print()
	return "Partición bajo el nombre " + comando.Name + " fue montada exitosamente\n"
}

func GenerarIdParticion(comando *parametros_mount, indexPartition int) (string, int, error) {
	// Asignar una letra a la partición
	discoMontado, err := global.AgregarMontaje(comando.Path)
	if err != nil {
		fmt.Println("ERROR: ha ocurrido un error obteniendo la letra:", err)
		return "", 0, err
	}

	// guardamos valores
	indice := discoMontado.Contador
	letra := discoMontado.Letra

	// actualizamos el indice
	global.ActualizarMontaje(comando.Path)

	// Crear id de partición
	idPartition := fmt.Sprintf("%d%d%s", 14, indice, letra)

	return idPartition, indice, nil
}
