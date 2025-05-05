package estructuras

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"main/global"
	"os"
)

// total de 2*1 + 4*4 + 16 = 2 + 16 + 16 = 34 bytes
type EBR struct {
	Status int32    // indica si está montada: -1 = inicializada, 0 = creada
	Type   [1]byte  // guardamos el tipo de la partición que únicamente puede ser L= logica
	Fit    [1]byte  // ajuste de la partición (FF,BF,WF)
	Start  int32    //comienzo de la partición logica
	Size   int32    //tamaño de la partición logica
	Name   [16]byte //nombre de la partición logica
	Next   int32    //siguiente EBR
}

// funcion recursiva que va deserializando los ebr disponibles hasta encontrar el disponible
func ObtenerEbrDisponible(disco *os.File, inicio int) (*EBR, error) {
	// obtenemos el primer EBR que está en el inicio de la partición extendida
	var ebr EBR
	if err := global.LeerEnDisco(disco, &ebr, int64(inicio)); err != nil {
		msj := "ERROR: No fue posible Leer el EBR."
		fmt.Println(msj)
		return nil, fmt.Errorf(msj, err)
	}

	if ebr.Size == 0 && ebr.Next == -1 && ebr.Status == -1 {
		// cuando tengamos el ultimo ebr
		// fmt.Println("ebr disponible")
		// ebr.Print()
		return &ebr, nil
	} else if int(ebr.Next) != -1 {
		// obtenemos siguiente
		return ObtenerEbrDisponible(disco, int(ebr.Next))
	}

	return nil, nil
}

// funcion recursiva que va deserializando los ebr hasta encontrar el que coincida con el nombre
func LogicaPorNombre(disco *os.File, inicio int, nombre string) (*EBR, error) {
	// obtenemos el primer EBR que está en el inicio de la partición extendida
	var ebr EBR
	if err := global.LeerEnDisco(disco, &ebr, int64(inicio)); err != nil {
		msj := "ERROR: No fue posible Leer el EBR."
		fmt.Println(msj)
		return nil, fmt.Errorf(msj, err)
	}

	nombreSinIlegibles := global.BorrandoIlegibles(string(ebr.Name[:]))
	if nombreSinIlegibles == nombre {
		// si hay coincidencia
		// fmt.Println("caso en que se encuentra coincidencia")
		return &ebr, nil
	} else if int(ebr.Next) != -1 {
		// obtenemos siguiente
		// fmt.Println("siguiente ebr")
		return LogicaPorNombre(disco, int(ebr.Next), nombre)
	} else {
		// ebr.Print()
		// si no se encuentra es decir cuando next = -1
		return nil, nil
	}
}

// imprime el ebr
func (ebr *EBR) Print() {
	fmt.Println("EBR:")
	fmt.Printf("Status: %d\n", ebr.Status)
	typeEbr := string(ebr.Type[0])
	fmt.Printf("Type: %s\n", typeEbr)
	fitEbr := string(ebr.Fit[0])
	fmt.Printf("Fit: %s\n", fitEbr)
	fmt.Printf("Start: %d\n", ebr.Start)
	fmt.Printf("Size: %d\n", ebr.Size)
	nombre := global.BorrandoIlegibles(string(ebr.Name[:]))
	fmt.Printf("Name: %s\n", nombre)
	fmt.Printf("Next: %d\n", ebr.Next)
}

// tostring del objeto ebr
func (ebr *EBR) ToString() string {
	contenido := fmt.Sprintf("EBR %s: ", global.BorrandoIlegibles(string(ebr.Name[:])))
	contenido += fmt.Sprintf("Status: %d, ", ebr.Status)
	contenido += fmt.Sprintf("Type: %s, ", string(ebr.Type[0]))
	contenido += fmt.Sprintf("Fit: %s, ", string(ebr.Fit[0]))
	contenido += fmt.Sprintf("Start: %d, ", ebr.Start)
	contenido += fmt.Sprintf("Size: %d, ", ebr.Size)
	contenido += fmt.Sprintf("Next: %d\n", ebr.Next)
	return contenido
}

func (ebr *EBR) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar la estructura Inode directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, ebr)
	if err != nil {
		return err
	}

	return nil
}

// DeserializeMBR lee la estructura MBR desde el inicio de un archivo binario
func (ebr *EBR) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el tamaño de la estructura ebr
	ebrSize := binary.Size(ebr)
	if ebrSize <= 0 {
		return fmt.Errorf("ERROR: El tamaño del EBR es invalido: %d", ebrSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura EBR
	buffer := make([]byte, ebrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura ebr
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, ebr)
	if err != nil {
		return err
	}

	return nil
}

// retorna todos los EBR de una extendida para el reporte MBR
func ReporteEBR(disco *os.File, inicio int32, salida string, contador int) (string, error) {
	// obtenemos el EBR
	var ebr EBR
	if err := global.LeerEnDisco(disco, &ebr, int64(inicio)); err != nil {
		msj := "ERROR: No fue posible Leer el EBR."
		fmt.Println(msj)
		return "", errors.New(msj)
	}

	// si ebr.Next == -1 entonces llegamos al último y debemos retornar la salida
	if ebr.Next == -1 {
		return salida, nil
	} else {
		// las logicas se simula que son creadas y montadas automáticamente
		estadoTextual := "Montada"
		if ebr.Status == 0 {
			estadoTextual = "Creada"
		}
		salida += fmt.Sprintf("\t\t\t<TR><TD COLSPAN=\"2\" BGCOLOR=\"mediumpurple1\"><B>Partition Logica %d</B></TD></TR>\n", contador)
		salida += fmt.Sprintf("\t\t\t<TR><TD>Status</TD><TD>%s</TD></TR>\n", estadoTextual)
		salida += fmt.Sprintf("\t\t\t<TR><TD>Type</TD><TD>%s</TD></TR>\n", string(ebr.Type[0]))
		salida += fmt.Sprintf("\t\t\t<TR><TD>Fit</TD><TD>%s</TD></TR>\n", string(ebr.Fit[0]))
		salida += fmt.Sprintf("\t\t\t<TR><TD>Start</TD><TD>%d</TD></TR>\n", ebr.Start)

		// si pasamos de 1000 Kb entonces pasamos a Mb
		sizeK, unitK := global.RevertirConversionUnidades(ebr.Size, "k")
		if sizeK >= 1000 {
			sizeM, unitM := global.RevertirConversionUnidades(ebr.Size, "m")
			salida += fmt.Sprintf("\t\t\t<TR><TD>Size</TD><TD>%d %s</TD></TR>\n", sizeM, unitM)
		} else {
			salida += fmt.Sprintf("\t\t\t<TR><TD>Size</TD><TD>%d %s</TD></TR>\n", sizeK, unitK)
		}

		salida += fmt.Sprintf("\t\t\t<TR><TD>Name</TD><TD>%s</TD></TR>\n", global.BorrandoIlegibles(string(ebr.Name[:])))
		salida += fmt.Sprintf("\t\t\t<TR><TD>Next</TD><TD>%d</TD></TR>\n", ebr.Next)
		// obtenemos siguiente y aumentamos el contador
		contador += 1
		return ReporteEBR(disco, ebr.Next, salida, contador)
	}
}

// retorna todos los EBR de una extendida y si hay espacios vacios para el reporte Disk
func ReporteEBR2(MbrSize int32, particion Partition, disco *os.File) (int, string) {
	cantidad := 0
	salidaLogicas := "\n\n<TR> \n"
	porcentaje := 0.0

	var ebrActual EBR
	ebrSize := int32(binary.Size(ebrActual))
	finParticion := particion.Start + particion.Size

	// Función auxiliar para añadir una partición lógica al reporte
	agregarLogica := func(ebr EBR) {
		porcentaje = float64(ebr.Size+ebrSize) * 100 / float64(MbrSize)
		salidaLogicas += "\t\t\t<TD BGCOLOR=\"dodgerblue2\" ROWSPAN=\"2\"> EBR </TD>\n"
		salidaLogicas += fmt.Sprintf("\t\t\t<TD BGCOLOR=\"cornflowerblue\" ROWSPAN=\"2\"> LOGICA <br/> %.2f %% </TD>\n", porcentaje)
		cantidad += 2
	}

	// Función auxiliar para añadir espacio libre
	agregarLibre := func(tamano int32) {
		if tamano > 0 {
			porcentaje = float64(tamano) * 100 / float64(MbrSize)
			salidaLogicas += fmt.Sprintf("\t\t\t<TD BGCOLOR=\"aliceblue\" ROWSPAN=\"2\"> LIBRE <br/> %.2f %% </TD>\n", porcentaje)
			cantidad++
		}
	}

	// Leer el primer EBR
	if err := global.LeerEnDisco(disco, &ebrActual, int64(particion.Start)); err != nil {
		fmt.Println("REP ERROR: Error al leer particiones logicas")
		porcentaje = float64(particion.Size) * 100 / float64(MbrSize)
		return 1, fmt.Sprintf("\t\t\t<TD BGCOLOR=\"aliceblue\" ROWSPAN=\"2\"> LIBRE <br/> %.2f %% </TD>\n", porcentaje)
	}

	// Caso: Extendida vacía o con espacio al inicio
	if ebrActual.Size == 0 {
		if ebrActual.Next == -1 {
			// Extendida completamente vacía
			agregarLibre(particion.Size)
			salidaLogicas += "</TR>\n"
			return cantidad, salidaLogicas
		} else {
			// Espacio libre inicial antes del primer EBR válido
			agregarLibre(ebrActual.Next - particion.Start)
			if err := global.LeerEnDisco(disco, &ebrActual, int64(ebrActual.Next)); err != nil {
				fmt.Println("REP ERROR: Error al leer primer EBR válido")
				return cantidad, salidaLogicas
			}
		}
	}

	// Iterar sobre cada EBR válido
	for {
		// Verifica si es un EBR válido (Size > 0)
		if ebrActual.Size > 0 {
			agregarLogica(ebrActual)
		}

		endActual := ebrActual.Start + ebrActual.Size + ebrSize
		var siguiente int32 = ebrActual.Next

		// Último EBR
		if siguiente == -1 {
			// Solo agregar espacio libre si no fue un EBR inválido (size 0)
			if ebrActual.Size > 0 {
				agregarLibre(finParticion - endActual)
			}
			break
		}

		// Calcular espacio libre entre EBRs válidos
		if err := global.LeerEnDisco(disco, &ebrActual, int64(siguiente)); err != nil {
			fmt.Println("REP ERROR: Error al leer siguiente EBR")
			porcentaje = float64(particion.Size) * 100 / float64(MbrSize)
			return 1, fmt.Sprintf("\t\t\t<TD BGCOLOR=\"aliceblue\" ROWSPAN=\"2\"> LIBRE <br/> %.2f %% </TD>\n", porcentaje)
		}

		if ebrActual.Size > 0 {
			agregarLibre(ebrActual.Start - endActual)
		} else if ebrActual.Next == -1 {
			// Último EBR inválido: solo tomar el espacio restante
			agregarLibre(finParticion - endActual)
			break
		}
	}

	salidaLogicas += "</TR>\n"
	return cantidad, salidaLogicas
}
