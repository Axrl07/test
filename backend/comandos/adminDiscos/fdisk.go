package adminDiscos

import (
	"errors"
	"fmt"
	"main/estructuras"
	"main/global"
	"strconv"
	"strings"
)

type parametros_fdisk struct {
	Size   int32   // Tamaño de la partición	- OBLIGAORIO AL CREAR
	Path   string  // Ruta del archivo del partición	- OBLIGATORIO
	Name   string  // Nombre de la partición	- OBLIGATORIO
	Unit   [1]byte // Unidad de medida del tamaño (B, K o M) -por defecto K
	Fit    [1]byte // Tipo de ajuste (B, F, W)	- por defecto F
	Type   [1]byte // Tipo de partición (P, E, L) - por defecto P
	Delete string  //  tipo de eliminación (fast,full) - hace obligatorios solo a NAME y PATH
	Add    int32   // aumentar o reducir espacio
}

/*
	// se toma el add si viene antes que el size, sino entonces tomamos el size
	fdisk -add=NUMERO -size=NUMERO -path=RUTA -name=NOMBRE -unit=K -fit=F -type=P
	fdisk -size=NUMERO -add=NUMERO -path=RUTA -name=NOMBRE -unit=K -fit=F -type=P
	// delete debe utilizar unicamente 3 parametros
	fdisk -delete=full -path=RUTA -name=NOMBRE
	// no puede venir add y delete al mismo tiempo
*/

func Fdisk(parametros []string) string {

	fdisk := parametros_fdisk{}

	//valores por defecto
	copy(fdisk.Unit[:], "K")
	copy(fdisk.Fit[:], "F")
	copy(fdisk.Type[:], "P")

	var importancia []string // básicamente me ayuda a saber qué vino primero si el size o el add
	// recorremos los parametros
	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		//verificamos que precisamente sea [nombre,valor]
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "size":
				importancia = append(importancia, "size")
				// el valorInicial no está convertido a las unidades
				valorInicial, err := strconv.Atoi(parametro[1])
				if err != nil {
					return "FDISK ERROR: Error al convertir el tamaño de la particion: " + parametro[1] + "\n"
				}
				if valorInicial < 0 {
					return "FDISK ERROR: El tamaño de la partición no puede ser negativo: " + parametro[1] + "\n"
				}
				fdisk.Size = int32(valorInicial)
			case "path":
				path := strings.ReplaceAll(parametro[1], "\"", "")
				fdisk.Path = path
			case "name":
				name := strings.ReplaceAll(parametro[1], "\"", "")
				fdisk.Name = name
			case "unit":
				unidad := strings.ToUpper(parametro[1])
				if unidad != "K" && unidad != "M" && unidad != "B" {
					return "FDISK ERROR: La unidad de la partición no es valida: " + parametro[1] + "\n"
				}
				copy(fdisk.Unit[:], unidad)
			case "fit":
				ajuste := strings.ToUpper(parametro[1])
				if ajuste != "FF" && ajuste != "BF" && ajuste != "WF" {
					return "FDISK ERROR: El ajuste de la partición debe ser FF, BF o WF, no: " + ajuste + "\n"
				}
				copy(fdisk.Fit[:], ajuste)
			case "type":
				tipo := strings.ToUpper(parametro[1])
				if tipo != "P" && tipo != "E" && tipo != "L" {
					return "FDISK ERROR: El ajuste de la partición debe ser P, E o L, no: " + tipo + "\n"
				}
				copy(fdisk.Type[:], tipo)
			case "delete":
				tipoD := strings.ToLower(parametro[1])
				if tipoD != "full" && tipoD != "fast" {
					return "FDISK ERROR: El tipo de eliminación de la partición debe ser fast o full, no " + tipoD + "\n"
				}
				fdisk.Delete = tipoD
			case "add":
				importancia = append(importancia, "add")
				valorInicial, err := strconv.Atoi(parametro[1])
				if err != nil {
					return "FDISK ERROR: Error al convertir el tamaño de la particion: " + parametro[1] + "\n"
				}
				if valorInicial == 0 {
					return "FDISK ERROR: EL valor asignado al parametro add, debe ser diferente de cero. \n"
				}
				fdisk.Add = int32(valorInicial)
			default:
				return "FDISK ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			return "FDISK ERROR: formato invalido para el parametro: " + param + "\n"
		}
	}

	// en cualquier caso path y Name son obligaorios
	if fdisk.Path == "" {
		return "FDISK ERROR: La ruta de la partición es obligatoria.\n"
	}
	if fdisk.Name == "" {
		return "FDISK ERROR: La partición debe tener nombre obligatoria.\n"
	}

	// verificamos si vino el delete
	if fdisk.Delete != "" {
		mensaje, err := eliminarParticion(fdisk)
		if err != nil {
			msj := "FDISK " + err.Error()
			fmt.Println(msj)
			return msj
		}
		return mensaje
	}

	// verificamos si el add vino antes que el size
	if importancia[0] == "add" {
		mensaje, err := modificarEspacioParticion(fdisk)
		if err != nil {
			msj := "FDISK " + err.Error()
			fmt.Println(msj)
			return msj
		}
		return mensaje
	}

	// solo es obligatorio si add o delete no vinieron
	if fdisk.Size == 0 {
		return "FDISK ERROR: El tamaño de la partición es obligatorio.\n"
	}

	return comandoFdisk(fdisk)
}

func comandoFdisk(comando parametros_fdisk) string {

	var salida string
	switch string(comando.Type[:]) {
	case "P":
		particionString, err := crearPrimaria(comando)
		if err != nil {
			msj := "FDISK " + err.Error() + "\n"
			fmt.Println(msj)
			return msj
		}
		salida = "Se creó correctamente " + particionString
	case "E":
		particionString, err := crearExtendida(comando)
		if err != nil {
			msj := "FDISK " + err.Error() + "\n"
			fmt.Println(msj)
			return msj
		}
		salida = "Se creó correctamente " + particionString
	case "L":
		// se devuelve el ebr actual
		ebrString, err := crearLogica(comando)
		if err != nil {
			msj := "FDISK " + err.Error() + "\n"
			fmt.Println(msj)
			return msj
		}

		// ebr.Print()
		salida = "se creó correctamente la partición logica con el " + ebrString
	}

	return salida
}

func crearPrimaria(comando parametros_fdisk) (string, error) {

	// abrimos el disco
	disco, err := global.AbrirDisco(comando.Path)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	// obtenemos el MBR que está en el disco
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		return "", err
	}

	// verificamos si ya existe una partición con ese nombre
	if particion, _ := mbr.ParticionPorNombre(comando.Name); particion != nil {
		msj := "ERROR: ya existe la partición bajo el nombre: " + comando.Name
		fmt.Println(msj)
		return "", errors.New(msj)
	}
	// ---------------------------------------------------------------------------------------------------
	// // Obtener la primera partición disponible
	// ParticionDisponible, comienzoParticion, indiceParticion := mbr.PrimeraParticionDisponible()
	// if ParticionDisponible == nil {
	// 	msj := "ERROR: No hay particiones disponibles."
	// 	fmt.Println(msj)
	// 	return "", errors.New(msj)
	// }

	// // Crear la partición con los parámetros proporcionados
	// tamano := global.ConvertirUnidades(comando.Size, comando.Unit)

	// tipo := string(comando.Type[:])
	// ajuste := string(comando.Fit[:])
	// ParticionDisponible.CrearParticion(comienzoParticion, tamano, tipo, ajuste, comando.Name)

	// // Colocar la partición en el MBR
	// if ParticionDisponible != nil {
	// 	mbr.Partitions[indiceParticion] = *ParticionDisponible
	// }

	// // escribiendo la partición en el disco
	// if err := global.EscribirEnDisco(disco, mbr, 0); err != nil {
	// 	return "", err
	// }

	// //cerramos el disco
	// defer disco.Close()

	// // retornamos error, nil
	// return ParticionDisponible.ToString(), nil
	// ---------------------------------------------------------------------------------------------------

	// Crear la partición con los parámetros proporcionados
	tamano := global.ConvertirUnidades(comando.Size, comando.Unit)
	ajuste := string(comando.Fit[:])

	// me devuelve la partición, el indice y el start para insertar la nueva particion según su tamaño y ajuste
	nuevaParticion, nuevoIndice, nuevoStart := mbr.AplicandoAjuste(tamano, ajuste)

	if nuevaParticion == nil {
		msj := "ERROR: No hay particiones disponibles."
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	tipo := string(comando.Type[:])

	// creamos la nueva partición
	nuevaParticion.CrearParticion(nuevoStart, tamano, tipo, ajuste, comando.Name)

	// colocamos en mbr la nueva particion
	mbr.Partitions[nuevoIndice] = *nuevaParticion

	// escribiendo la partición en el disco
	if err := global.EscribirEnDisco(disco, mbr, 0); err != nil {
		return "", err
	}

	//cerramos el disco
	defer disco.Close()

	// retornamos error, nil
	return nuevaParticion.ToString(), nil
}

func crearExtendida(comando parametros_fdisk) (string, error) {
	// abrimos el disco
	disco, err := global.AbrirDisco(comando.Path)
	if err != nil {
		msj := "ERROR: No se pudo leer el disco"
		fmt.Println(msj)
		return "", fmt.Errorf(msj, err)
	}

	// obtenemos el MBR que está en el disco
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		msj := "ERROR: No fue posible Leer el MBR."
		fmt.Println(msj)
		return "", fmt.Errorf(msj, err)
	}

	// verificamos si existe una partición extendida
	if particion, _ := mbr.ParticionExtendida(); particion != nil {
		msj := "ERROR: Ya existe una partición extendida."
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	// verificamos si existe bajo el nombre ingresado
	if particion, _ := mbr.ParticionPorNombre(comando.Name); particion != nil {
		msj := "ERROR: Ya existe una partición con el nombre " + comando.Name
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	// // Obtener la primera partición disponible
	// ParticionDisponible, comienzoParticion, indiceParticion := mbr.PrimeraParticionDisponible()

	// // verificar que si exista la partición
	// if ParticionDisponible == nil {
	// 	msj := "ERROR: No hay particiones disponibles."
	// 	fmt.Println(msj)
	// 	return "", errors.New(msj)
	// }

	// // Crear la partición con los parámetros proporcionados
	// tamano := global.ConvertirUnidades(comando.Size, comando.Unit)
	// tipo := string(comando.Type[:])
	// ajuste := string(comando.Fit[:])
	// ParticionDisponible.CrearParticion(comienzoParticion, tamano, tipo, ajuste, comando.Name)

	// // Colocar la partición en el MBR
	// if ParticionDisponible != nil {
	// 	mbr.Partitions[indiceParticion] = *ParticionDisponible
	// }

	// // reescribimos el MBR
	// if err := global.EscribirEnDisco(disco, mbr, 0); err != nil {
	// 	return "", err
	// }

	// // creando el primer EBR
	// ebr := estructuras.EBR{
	// 	Status: -1,
	// 	Type:   [1]byte{'L'},
	// 	Fit:    [1]byte{'F'},
	// 	Start:  comienzoParticion,
	// 	Size:   0,
	// 	Name:   [16]byte{'0'},
	// 	Next:   -1,
	// }

	// // escribir el EBR en el disco al comienzo de la partición
	// if err := global.EscribirEnDisco(disco, ebr, int64(comienzoParticion)); err != nil {
	// 	return "", err
	// }

	// //cerramos el disco
	// defer disco.Close()
	// return ParticionDisponible.ToString(), nil

	// Crear la partición con los parámetros proporcionados
	tamano := global.ConvertirUnidades(comando.Size, comando.Unit)
	ajuste := string(comando.Fit[:])

	// me devuelve la partición, el indice y el start para insertar la nueva particion según su tamaño y ajuste
	nuevaParticion, nuevoIndice, nuevoStart := mbr.AplicandoAjuste(tamano, ajuste)

	if nuevaParticion == nil {
		msj := "ERROR: No hay particiones disponibles."
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	tipo := string(comando.Type[:])

	// creamos la nueva partición
	nuevaParticion.CrearParticion(nuevoStart, tamano, tipo, ajuste, comando.Name)

	// colocamos en mbr la nueva particion
	mbr.Partitions[nuevoIndice] = *nuevaParticion

	// escribiendo la partición en el disco
	if err := global.EscribirEnDisco(disco, mbr, 0); err != nil {
		return "", err
	}

	// creando el primer EBR
	ebr := estructuras.EBR{
		Status: -1,
		Type:   [1]byte{'L'},
		Fit:    [1]byte{'F'},
		Start:  nuevoStart,
		Size:   0,
		Name:   [16]byte{'0'},
		Next:   -1,
	}

	// escribir el EBR en el disco al comienzo de la partición
	if err := global.EscribirEnDisco(disco, ebr, int64(nuevoStart)); err != nil {
		return "", err
	}

	//cerramos el disco
	defer disco.Close()

	// retornamos error, nil
	return nuevaParticion.ToString(), nil
}

func crearLogica(comando parametros_fdisk) (string, error) {
	// abrimos el disco
	disco, err := global.AbrirDisco(comando.Path)
	if err != nil {
		msj := "ERROR: No se pudo leer el disco"
		fmt.Println(msj)
		return "", fmt.Errorf(msj, err)
	}

	// obtenemos el MBR que está en el disco
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		msj := "ERROR: No fue posible Leer el MBR."
		fmt.Println(msj)
		return "", fmt.Errorf(msj, err)
	}

	// verificamos si existe una partición Logica en la extendida
	particionExt, _ := mbr.ParticionExtendida()
	if particionExt == nil {
		msj := "ERROR: No fue posible obtener la partición extendida"
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	// verificar que no exista otra partición con ese nombre
	if ebr, _ := estructuras.LogicaPorNombre(disco, int(particionExt.Start), comando.Name); ebr != nil {
		ebr.Print()
		msj := "ERROR: Ya existe una partición Logica con el nombre " + comando.Name
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	// funcion recursiva que va deserializando los ebr disponibles hasta encontrar el disponible
	ebr, _ := estructuras.ObtenerEbrDisponible(disco, int(particionExt.Start))

	if ebr == nil {
		msj := "ERROR: No fue posible obtener el EBR"
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	// configurando valores del ebr actual
	ebr.Status = 0
	ebr.Fit = comando.Fit
	ebr.Size = global.ConvertirUnidades(comando.Size, comando.Unit) // este ya es el size de la logica
	copy(ebr.Name[:], comando.Name)
	// Status, Type, Start, no se cambian pues fueron configurados anteriormente

	/*
		sigueinte ebr = (tamaño de la logica) + (ebr actual)
		| ebr              | logica          | ebr 2 |
		| binary.size(ebr) | ebr.size        |
											 | aquí empieza el nuevo ebr
	*/
	startEbrNext := ebr.Size + ebr.Start

	// verificamos que no hayamos sobre pasado el tamaño de la partición extendida
	if startEbrNext >= particionExt.Size {
		msj := "ERROR: no hay suficiente espacio para crear una nueva partición lógica"
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	ebrNext := estructuras.EBR{
		Status: -1,
		Type:   [1]byte{'L'},
		Fit:    [1]byte{'F'},
		Start:  startEbrNext,
		Size:   0,
		Name:   [16]byte{'0'},
		Next:   -1,
	}

	// linkeamos el nuevo ebr
	ebr.Next = ebrNext.Start

	// escribirmos el ebr actual
	if err := global.EscribirEnDisco(disco, ebr, int64(ebr.Start)); err != nil {
		return "", err
	}

	// escribimos el siguiente
	if err := global.EscribirEnDisco(disco, ebrNext, int64(ebrNext.Start)); err != nil {
		return "", err
	}

	//cerramos el disco
	defer disco.Close()
	return ebr.ToString(), nil
}

/*
Aumenta o disminuye el espacio de las particiones primarias
*/
func modificarEspacioParticion(comando parametros_fdisk) (string, error) {

	// abrimos el disco
	disco, err := global.AbrirDisco(comando.Path)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	// obtenemos el MBR que está en el disco
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		return "", err
	}

	// partición y su posición en el listado mbr.partitions (lo usamos al reducir y para verificar)
	particion, indice := mbr.ParticionPorNombre(comando.Name)

	// verificamos que la partición existta
	if particion == nil {
		msj := "ERROR: No existe la partición bajo el nombre: " + comando.Name
		return "", errors.New(msj)
	}

	// verificamos que la partición sea primaria
	if string(particion.Type[0]) != "P" {
		msj := "ERROR: EL comando ADD solamente puede aplicarse a particiones Primarias."
		return "", errors.New(msj)
	}

	var salida string // salida string de exito

	// reducir espacio
	if comando.Add < 0 {
		// verificamos que la reducción no sea mayor o igual a cero (se eliminaria si es negativo y no existiría si fuera 0)
		nuevoSize := particion.Size + comando.Add
		if nuevoSize <= 0 {
			msj := "ERROR: el número ingresado en el parametro ADD es demasiado grande, tome en cuenta la unidad (si no se especifica es K)"
			return "", errors.New(msj)
		}

		mbr.Partitions[indice].Size += comando.Add
		if err := global.EscribirEnDisco(disco, mbr, 0); err != nil {
			return "", err
		}

		// salida exitosa reducción

		salida = "Se redujo correctamente " + mbr.Partitions[indice].ToString()
	}

	//aumentamos el espacio
	if comando.Add > 0 {

		// número de partición a aumentar = indice + 1 (también es el indice PERO de la siguiente partición con respecto a la actual)
		numPart := indice + 1

		// tamaño del espacio libre
		var espacioLibre int32
		if numPart < 4 {
			// obtenemos el espacio libre entre la partición siguiente y la actual
			espacioLibre = mbr.Partitions[numPart].Start - (particion.Start + particion.Size)
		} else {
			// obtenemos el espacio libre entre el disco y la partición final
			espacioLibre = mbr.Size - (particion.Start + particion.Size)
		}

		// si el espacio libre es cero entonces no es posible aumentar el tamaño
		if espacioLibre == 0 {
			msj := fmt.Sprintf("ERROR: No hay espacio disponible entre la partición %s y la partición siguiente.", global.BorrandoIlegibles(string(particion.Name[:])))
			return "", errors.New(msj)
		}

		// para aumentar add <= espacioLibre
		if comando.Add > espacioLibre {
			msj := "ERROR: El valor del parametro ADD es muy grande."
			return "", errors.New(msj)
		}

		// admitimos el cambio de tamaño
		mbr.Partitions[indice].Size += comando.Add

		// reescribimos el mbr en disco
		if err := global.EscribirEnDisco(disco, mbr, 0); err != nil {
			return "", err
		}

		// salida exitosa aumento
		salida = "Se aumentó correctamente " + mbr.Partitions[indice].ToString()
	}

	// cerramos el disco
	defer disco.Close()
	return salida, nil
}

/*
elimina particiones Primarias
*/
func eliminarParticion(comando parametros_fdisk) (string, error) {

	// abrimos el disco
	disco, err := global.AbrirDisco(comando.Path)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	// obtenemos el MBR que está en el disco
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		return "", err
	}

	// partición y su posición en el listado mbr.partitions (lo usamos al eliminar y para verificar)
	particion, indice := mbr.ParticionPorNombre(comando.Name)

	// verificamos que la partición existta
	if particion == nil {
		msj := "ERROR: No existe la partición bajo el nombre: " + comando.Name
		return "", errors.New(msj)
	}

	// si fue un full entonces borramos el disco
	// y como si es extendida debemos eliminar los ebr entonces vamos a eliminar toda la extendida siempre con un full
	if comando.Delete == "full" || string(particion.Type[0]) != "E" {
		// le decimos que queremos borrar en el disco desde el start y enviamos el size
		global.BorrarEnDisco(disco, particion.Start, particion.Size)
	}

	// verificando que si pudo hacerse el full podemos sobreescribri el mbr con una partición vacía
	nuevaParticion := estructuras.Partition{Status: [1]byte{'0'}, Type: [1]byte{'0'}, Fit: [1]byte{'0'}, Start: -1, Size: -1, Name: [16]byte{'0'}, Correlative: -1, Id: [4]byte{'0'}}

	// reescribimos el mbr
	mbr.Partitions[indice] = nuevaParticion
	if err := global.EscribirEnDisco(disco, mbr, 0); err != nil {
		return "", err
	}

	// cerramos disco
	defer disco.Close()
	return fmt.Sprintf("La partición %s fue eliminada exitosamente.", global.BorrandoIlegibles(string(particion.Name[:]))), nil
}
