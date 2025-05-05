package adminDiscos

import (
	"fmt"
	"main/estructuras"
	"main/global"
)

func Mounted() string {

	//verificamos que hayan particiones montada
	if len(global.ParticionesMontadas) == 0 {
		msj := "No hay particiones montadas"
		fmt.Println(msj)
		return msj
	}

	var salida string
	// itreamos sobre los discos que tinen particiones montadas
	for _, objetoMontaje := range global.Montaje {

		// abrimos el disco
		disco, err := global.AbrirDisco(objetoMontaje.Path)
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

		// cerrramos el disco porque solo queremos obtener el MBR
		defer disco.Close()

		// generamos la salida de mounted
		partMontadas := "\n\nLISTA DE PARTICIONES MONTADAS EN EL DISCO " + objetoMontaje.Path + "\n"
		for i := 0; i < 4; i++ {
			estado := int32(mbr.Partitions[i].Status[0])
			// el 1 significa que está montada
			if estado == 1 {
				partMontadas += mbr.Partitions[i].ToString()
			}
		}
		salida += partMontadas + "\n"
		//salida += mbr.ToString() //esto es para ver los mbr junto con sus particiones
	}

	return salida
}
