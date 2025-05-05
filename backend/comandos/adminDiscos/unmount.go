package adminDiscos

import (
	"fmt"
	"main/estructuras"
	"main/global"
	"strings"
)

/* parametro obligatorio
unmount -id=IDENTIFICADOR
*/

func Unmount(parametros []string) string {

	// verificando si viene un solo parametro
	if len(parametros) != 1 {
		msj := "UNMOUNT ERROR: hay más de un parametro\n"
		fmt.Println(msj)
		return msj
	}

	// obteniendo el parametro y su valor
	parametro := strings.Split(parametros[0], "=")

	//verificando que venga solo [param, valor]
	if len(parametro) != 2 {
		msj := "UNMOUNT ERROR: formato del parametro " + parametro[0] + " no es correcto\n"
		fmt.Println(msj)
		return msj
	}

	// verificando que venga id
	if parametro[0] != "id" {
		msj := "UNMOUNT ERROR: el único parametro admitido es id\n"
		fmt.Println(msj)
		return msj
	}

	return comandoUnmount(parametro[1])
}

func comandoUnmount(id string) string {

	//obtener el path del disco donde se encuentra la partición
	path, ok := global.ParticionesMontadas[id]

	// verificando que me haya devuelto un path
	if !ok {
		msj := "UNMOUNT ERROR: El id de la partición no existe\n"
		fmt.Println(msj)
		return msj
	}

	// abrimos el disco
	disco, err := global.AbrirDisco(path)
	if err != nil {
		msj := "MOUNTED " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// obtenemos el MBR que está en el disco
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		msj := "MOUNTED " + err.Error()
		fmt.Println(msj)
		return msj
	}

	//obtenemos la partición por su ID
	particion, err := mbr.ParticionPorId(id)
	if err != nil {
		msj := "MOUNTED " + err.Error() + "\n"
		fmt.Println(msj)
		return msj
	}

	// verificamos que esté montada
	if int32(particion.Status[0]) != 1 {
		msj := "UNMOUNT ERROR: EL Id corresponde a una partición que no está montada\n"
		fmt.Println(msj)
		return msj
	}

	// seteamos la partición como creada : 1 = montada -> 0 = creada
	particion.Status = [1]byte{0}
	// eliminamos el id
	particion.Id = [4]byte{'0'}
	// eliminamos el correlativo
	particion.Correlative = 0

	// escribiendo en el disco el nuevo MBR
	if err := global.EscribirEnDisco(disco, mbr, 0); err != nil {
		msj := "MOUNTED " + err.Error() + "\n"
		fmt.Println(msj)
		return msj
	}

	defer disco.Close()

	// verificando que no hubieron errores
	fmt.Println("desmontamos correctamente la partición bajo el id: "+id, " en el path: "+path)

	// quitamos de la lista de particiones montadas la del id
	delete(global.ParticionesMontadas, id)

	// quitando 1 del contador (si se terminan las particiones entonces queda en 1 no en 0)
	discoMontaje := global.Montaje[path]
	discoMontaje.Contador -= 1

	// vamos a arreglar el correlativo de cada particion
	if err := arreglandoCorrelativos(path); err != nil {
		msj := "MOUNTED " + err.Error() + "\n"
		fmt.Println(msj)
		return msj
	}

	return "La partición bajo el id: " + id + " fue desmontada exitosamente\n"
}

func arreglandoCorrelativos(path string) error {

	// abrimos el disco
	disco, err := global.AbrirDisco(path)
	if err != nil {
		return err
	}

	// obtenemos el MBR que está en el disco
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		return err
	}

	contador := 1
	for _, particion := range mbr.Partitions {
		// si está montada comenzamos a enumerarlas
		if int32(particion.Status[0]) == 1 {
			particion.Correlative = int32(contador)
			contador++
		}
	}

	defer disco.Close()
	return nil
}
