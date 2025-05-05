package adminDiscos

import (
	"fmt"
	"main/estructuras"
	"main/global"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

/*
	mkdisk -size=NUMERO -path=RUTA
	mkdisk -size=NUMERO -path=RUTA -fit=(BF|FF|WF) -fit=(BF|FF|WF) -unit=(M|K)
*/

type parametros_mkdisk struct {
	Size int32   // tamaño del disco	- OBLIGATORIO
	Path string  // ruta del disco 	- OBLIGATORIO
	Fit  [1]byte // Tipo de ajuste (B, F, W) - por defecto F
	Unit [1]byte // Unidad de medida del tamaño (K o M) - por defecto M
}

func Mkdisk(parametros []string) string {
	// creamos el objeto mkdisk
	mkdisk := parametros_mkdisk{}

	// llenamos los valores por defecto
	copy(mkdisk.Fit[:], "F")  // ajuste por defecto
	copy(mkdisk.Unit[:], "M") // unidad por defecto

	// recorremos los parametros
	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "size":
				// el valorInicial no está convertido a las unidades
				valorInicial, err := strconv.Atoi(parametro[1])
				if err != nil {
					return "MKDISK ERROR: Error al convertir el tamaño del disco: " + parametro[1] + "\n"
				}
				if valorInicial < 0 {
					return "MKDISK ERROR: El tamaño del disco no puede ser negativo: " + parametro[1] + "\n"
				}
				mkdisk.Size = int32(valorInicial)
			case "path":
				path := strings.ReplaceAll(parametro[1], "\"", "")
				mkdisk.Path = path
			case "fit":
				ajuste := strings.ToUpper(parametro[1])
				if ajuste != "FF" && ajuste != "BF" && ajuste != "WF" {
					return "MKDISK ERROR: El ajuste del disco no es valido: " + ajuste + "\n"
				}
				copy(mkdisk.Fit[:], ajuste)
			case "unit":
				unidad := strings.ToUpper(parametro[1])
				if unidad != "K" && unidad != "M" {
					return "MKDISK ERROR: La unidad del disco no es valida: " + parametro[1] + "\n"
				}
				copy(mkdisk.Unit[:], unidad)
			default:
				return "MKDISK ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			return "MKDISK ERROR: formato invalido para parametro: " + param + "\n"
		}
	}

	// verificamos los parametros obligatorios
	if mkdisk.Size == 0 {
		return "El parametro size es obligatorio.\n"
	}
	if mkdisk.Path == "" {
		return "La ruta del disco es obligatoria.\n"
	}
	return comandoMkdisk(mkdisk)
}

func comandoMkdisk(comando parametros_mkdisk) string {
	// creando
	if err := global.CrearDisco(comando.Path); err != nil {
		msj := "MKDISK: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// abriendo
	archivo, err := global.AbrirDisco(comando.Path)
	if err != nil {
		msj := "MKDISK " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// dandole tamaño al disco
	if err := global.EscribirEnDisco(archivo, make([]byte, comando.Size), 0); err != nil {
		msj := "MKDISK " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// creando MBR
	mbr, err := crearMBR(comando)
	if err != nil {
		msj := "MKDISK " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// escribiendo el MBR en el disco
	if err := global.EscribirEnDisco(archivo, mbr, 0); err != nil {
		return "MKDISK " + err.Error()
	}

	// cerramos archivo
	defer archivo.Close()

	// impresión de retorno

	// var salida estructuras.MBR
	// if err := funciones.LeerEnDisco(archivo, &salida, 0); err != nil {
	// 	return "MKDISK " + err.Error()
	// }

	// salida.Print()	// imprimiendo para ver si se creó correctamente el MBR y el disco

	disco := strings.Split(comando.Path, "/")
	global.RutaDiscosLocales = strings.Join(disco[:len(disco)-1], "/") // guardamos la ruta de los discos locales
	return "Disco " + disco[len(disco)-1] + " creado correctamente en: " + global.RutaDiscosLocales + "\n"
}

func crearMBR(comando parametros_mkdisk) (estructuras.MBR, error) {
	fecha := time.Now()
	fechaUnix := fecha.Unix()
	fechaFloat := float32(fechaUnix)

	mbr := estructuras.MBR{
		Size:          global.ConvertirUnidades(comando.Size, comando.Unit),
		CreationDate:  float32(fechaFloat),
		DiskSignature: rand.Int32N(9999),
		Fit:           [1]byte{comando.Fit[0]},
		Partitions: [4]estructuras.Partition{
			// inicialización de particiones
			{Status: [1]byte{'0'}, Type: [1]byte{'0'}, Fit: [1]byte{'0'}, Start: -1, Size: -1, Name: [16]byte{'0'}, Correlative: -1, Id: [4]byte{'0'}},
			{Status: [1]byte{'0'}, Type: [1]byte{'0'}, Fit: [1]byte{'0'}, Start: -1, Size: -1, Name: [16]byte{'0'}, Correlative: -1, Id: [4]byte{'0'}},
			{Status: [1]byte{'0'}, Type: [1]byte{'0'}, Fit: [1]byte{'0'}, Start: -1, Size: -1, Name: [16]byte{'0'}, Correlative: -1, Id: [4]byte{'0'}},
			{Status: [1]byte{'0'}, Type: [1]byte{'0'}, Fit: [1]byte{'0'}, Start: -1, Size: -1, Name: [16]byte{'0'}, Correlative: -1, Id: [4]byte{'0'}},
		},
	}

	return mbr, nil
}
