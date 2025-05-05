package reportes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strings"
	"time"
)

// REPORTE MBR
func reporteMBR(pathDisco string) (string, error) {

	// abrimos el disco
	disco, err := global.AbrirDisco(pathDisco)
	if err != nil {
		return "", err
	}

	//leemos el mbr
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		return "", err
	}

	// configuración del grafico
	salida := "digraph MBR {\n"
	salida += "\trankdir=LR;\n"
	salida += "\tnode [shape=plain];\n"
	salida += "\tedge [color=\"blue\"];\n\n"

	// creamos la tabla
	salida += "\tMBR [label=<\n"
	salida += "\t\t<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"

	// creando MBR
	salida += "\t\t\t<TR><TD COLSPAN=\"2\" BGCOLOR=\"lightblue\"><B>MBR</B></TD></TR>\n"

	// si pasamos de 1000 Kb entonces pasamos a Mb
	sizeK, unitK := global.RevertirConversionUnidades(mbr.Size, "k")
	if sizeK >= 1000 {
		sizeM, unitM := global.RevertirConversionUnidades(mbr.Size, "m")
		salida += fmt.Sprintf("\t\t\t<TR><TD>Size</TD><TD>%d %s</TD></TR>\n", sizeM, unitM)
	} else {
		salida += fmt.Sprintf("\t\t\t<TR><TD>Size</TD><TD>%d %s</TD></TR>\n", sizeK, unitK)
	}

	// obtenemos la fecha
	fecha := global.ObtenerFecha(time.Unix(int64(mbr.CreationDate), 0))

	// insertamos como un string
	salida += fmt.Sprintf("\t\t\t<TR><TD>Creation Date</TD><TD>%s</TD></TR>\n", fecha)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Disk Signature</TD><TD>%d</TD></TR>\n", mbr.DiskSignature)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Fit</TD><TD>%s</TD></TR>\n", string(mbr.Fit[0]))

	// creando Particiones
	for i, particion := range mbr.Partitions {
		if string(particion.Type[0]) == "P" {
			estadoTextual := "Montada"
			if particion.Status[0] == 0 {
				estadoTextual = "Creada"
			}
			salida += fmt.Sprintf("\t\t\t<TR><TD COLSPAN=\"2\" BGCOLOR=\"lightgreen\"><B>Partition %d</B></TD></TR>\n", i+1)
			salida += fmt.Sprintf("\t\t\t<TR><TD>Status</TD><TD>%s</TD></TR>\n", estadoTextual)
			salida += fmt.Sprintf("\t\t\t<TR><TD>Type</TD><TD>%s</TD></TR>\n", string(particion.Type[0]))
			salida += fmt.Sprintf("\t\t\t<TR><TD>Fit</TD><TD>%s</TD></TR>\n", string(particion.Fit[0]))
			salida += fmt.Sprintf("\t\t\t<TR><TD>Start</TD><TD>%d</TD></TR>\n", particion.Start)

			// si pasamos de 1000 Kb entonces pasamos a Mb
			sizeK, unitK := global.RevertirConversionUnidades(particion.Size, "k")
			if sizeK >= 1000 {
				sizeM, unitM := global.RevertirConversionUnidades(particion.Size, "m")
				salida += fmt.Sprintf("\t\t\t<TR><TD>Size</TD><TD>%d %s</TD></TR>\n", sizeM, unitM)
			} else {
				salida += fmt.Sprintf("\t\t\t<TR><TD>Size</TD><TD>%d %s</TD></TR>\n", sizeK, unitK)
			}

			salida += fmt.Sprintf("\t\t\t<TR><TD>Name</TD><TD>%s</TD></TR>\n", global.BorrandoIlegibles(string(particion.Name[:])))
			salida += fmt.Sprintf("\t\t\t<TR><TD>Correlative</TD><TD>%d</TD></TR>\n", particion.Correlative)
			salida += fmt.Sprintf("\t\t\t<TR><TD>Id</TD><TD>%s</TD></TR>\n", global.BorrandoIlegibles(string(particion.Id[:])))
		} else if string(particion.Type[0]) == "E" {
			// primero decimos que estamos en la extendida
			salida += fmt.Sprintf("\t\t\t<TR><TD COLSPAN=\"2\" BGCOLOR=\"lightsalmon\"><B>Partition %d</B></TD></TR>\n", i+1)

			//obtenemos la salida de los ebr
			var ebrString string
			ebrSalida, err := estructuras.ReporteEBR(disco, particion.Start, ebrString, 1)
			if err != nil {
				return "", err
			}
			salida += "\t\t\t// antes de las logicas"
			salida += ebrSalida
			salida += "\t\t\t// despues de las logicas"
		}
	}

	// cerrando tabla
	salida += "\t\t</TABLE>\n"
	salida += "\t>];\n"
	salida += "}"

	// cerramos el disco
	defer disco.Close()
	return salida, nil
}

// REPORTE DISK
func reporteDisk(pathDisco string) (string, error) {
	// abrimos el disco
	disco, err := global.AbrirDisco(pathDisco)
	if err != nil {
		return "", err
	}

	// leemos el mbr
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		return "", err
	}

	// configuración del grafico
	salida := "digraph DISK {\n"
	salida += "\tnode [shape=none];\n"

	// creamos la tabla
	salida += "\tTablaReportNodo [label=<\n"
	salida += "\t\t<TABLE border=\"1\">\n"
	salida += "\t\t\t<TR><TD BGCOLOR='gold' ROWSPAN='3'> MBR </TD>\n"

	//logica para obtener particiones
	espacioDisponible := int32(0)
	espacioLibreInicio := int32(binary.Size(mbr))

	var salidaL string
	// recorremos las particiones
	for _, particion := range mbr.Partitions {
		if particion.Size > 0 {
			espacioDisponible = particion.Start - espacioLibreInicio
			espacioLibreInicio = particion.Start + particion.Size

			if espacioDisponible > 0 {
				porcentaje := float64(espacioDisponible) * 100 / float64(mbr.Size)
				salida += fmt.Sprintf("\t\t\t<TD BGCOLOR=\"aliceblue\" ROWSPAN=\"3\"> ESPACIO LIBRE <BR/> %.2f %% </TD>\n", porcentaje)
			}
			porcentaje := float64(particion.Size) * 100 / float64(mbr.Size)
			if string(particion.Type[:]) == "P" {
				salida += fmt.Sprintf("\t\t\t<TD BGCOLOR='lightgreen' ROWSPAN='3'> PRIMARIA <br/> %.2f %% </TD>\n", porcentaje)
			} else {
				cantidad, salidaLogicas := estructuras.ReporteEBR2(mbr.Size, particion, disco)
				salida += fmt.Sprintf("\t\t\t<TD BGCOLOR='tan1' COLSPAN='%d'> EXTENDIDA </TD>\n", cantidad)
				salidaL = salidaLogicas
			}
		}
	}

	//si hay espacio despues de la 4ta particion
	espacioDisponible = mbr.Size - espacioLibreInicio
	if espacioDisponible > 0 {
		porcentaje := float64(espacioDisponible) * 100 / float64(mbr.Size)
		salida += fmt.Sprintf("\t\t\t<TD BGCOLOR=\"aliceblue\" ROWSPAN=\"3\"> ESPACIO LIBRE <BR/> %.2f %% </TD>\n", porcentaje)
	}
	salida += "\t\t\t</TR>\n"
	salida += salidaL
	salida += " \t\t</TABLE>\n"
	salida += "\t>];\n"
	salida += "}"

	// cerramos el disco
	defer disco.Close()
	return salida, nil
}

// REPORTE INODOS
func reporteInode(idPart string) (string, error) {
	// Obtenemos el superbloque
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idPart)
	if err != nil {
		return "", err
	}

	// Iniciar el contenido DOT
	salida := "digraph Inodos {"
	salida += "\tnode [shape=plaintext];\n"
	salida += "\trankdir=LR;\n"
	salida += "\tlabel=\"Inode\";\n"
	salida += "\tbgcolor=\"#E6F3FF\";\n"

	// Iterar sobre cada inodo
	for i := int32(0); i < superBloque.S_inodes_count; i++ {
		inodo := &estructuras.Inode{}
		// Deserializar el inodo
		err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(i*superBloque.S_inode_size)))
		if err != nil {
			continue // Si hay error, seguimos con el siguiente
		}

		// Verificar si el inodo está en uso (podemos validar atributos)
		if inodo.I_size <= 0 && inodo.I_type[0] == 0 {
			continue // Inodo vacío, seguimos con el siguiente
		}

		// Convertir tiempos a string usando la función de global
		atime := global.ObtenerFecha(time.Unix(int64(inodo.I_atime), 0))
		ctime := global.ObtenerFecha(time.Unix(int64(inodo.I_ctime), 0))
		mtime := global.ObtenerFecha(time.Unix(int64(inodo.I_mtime), 0))

		tipoInodo := "Carpeta"
		if inodo.I_type[0] == '1' {
			tipoInodo = "Archivo"
		}

		// Definir el contenido DOT para el inodo actual
		salida += fmt.Sprintf("\tinode%d [label=<\n", i)
		salida += "\t\t<table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n"
		salida += fmt.Sprintf("\t\t\t<tr><td colspan=\"2\" bgcolor=\"#4B92DB\"><font color=\"white\">INODO %d</font></td></tr>\n", i)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_uid</td><td>%d</td></tr>\n", inodo.I_uid)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_gid</td><td>%d</td></tr>\n", inodo.I_gid)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_size</td><td>%d</td></tr>\n", inodo.I_size)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_atime</td><td>%s</td></tr>\n", atime)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_ctime</td><td>%s</td></tr>n", ctime)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_mtime</td><td>%s</td></tr>\n", mtime)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_type</td><td>%s</td></tr>\n", tipoInodo)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_perm</td><td>%s</td></tr>`\n", string(inodo.I_perm[:]))
		salida += "\t\t\t<tr><td colspan=\"2\" bgcolor=\"#A8C6DF\">BLOQUES DIRECTOS</td></tr>\n"

		for j := 0; j < 12; j++ {
			if inodo.I_block[j] != -1 {
				salida += fmt.Sprintf("\t\t\t<tr><td>Bloque %d</td><td port=\"blk%d\">%d</td></tr>\n", j, j, inodo.I_block[j])
			} else {
				salida += fmt.Sprintf("\t\t\t<tr><td>Bloque %d</td><td>-1</td></tr>\n", j)
			}
		}

		// Agregar los bloques indirectos a la tabla
		salida += "\t\t\t<tr><td colspan=\"2\" bgcolor=\"#A8C6DF\">BLOQUES INDIRECTOS</td></tr>\n"
		salida += fmt.Sprintf("\t\t\t<tr><td>Indirecto Simple</td><td port=\"ind1\">%d</td></tr>\n", inodo.I_block[12])
		salida += fmt.Sprintf("\t\t\t<tr><td>Indirecto Doble</td><td port=\"ind2\">%d</td></tr>\n", inodo.I_block[13])
		salida += fmt.Sprintf("\t\t\t<tr><td>Indirecto Triple</td><td port=\"ind3\">%d</td></tr>\n", inodo.I_block[14])
		salida += "\t\t</table>\n"
		salida += "\t>];\n"

		// Agregar enlace al siguiente inodo si no es el último
		if i < superBloque.S_inodes_count-1 {
			salida += fmt.Sprintf("inode%d -> inode%d;\n", i, i+1)
		}
	}

	// Cerrar el contenido DOT
	salida += "}"

	return salida, nil
}

// REPORTE BLOQUES
func reporteBlock(idPart string) (string, error) {
	// Obtenemos el superbloque
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idPart)
	if err != nil {
		return "", err
	}

	// Iniciar el contenido DOT
	salida := "digraph Block {\n"
	salida += "\tnode [shape=plaintext];\n"
	salida += "\trankdir=LR;\n"
	salida += "\tlabel=\"Block\";\n"
	salida += "\tbgcolor=\"#E6F3FF\";\n"

	// Mapa para rastrear bloques ya procesados
	procesados := make(map[int32]bool)

	// Segundo paso: procesamos cada inodo y sus bloques
	for i := int32(0); i < superBloque.S_inodes_count; i++ {
		inodo := &estructuras.Inode{}
		// Deserializar el inodo
		err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(i*superBloque.S_inode_size)))
		if err != nil {
			continue
		}

		// Si el inodo está vacío, seguimos con el siguiente
		if inodo.I_size <= 0 && inodo.I_type[0] == 0 {
			continue
		}

		// Procesamos bloques directos
		for j := 0; j < 12; j++ {
			if inodo.I_block[j] != -1 {
				// Si el bloque ya fue procesado, solo agregar la conexión
				if !procesados[inodo.I_block[j]] {
					procesados[inodo.I_block[j]] = true

					// Procesamos el bloque según el tipo de inodo
					if inodo.I_type[0] == '0' { // Carpeta
						dibujarBloqueCarpeta(&salida, inodo.I_block[j], pathDisco, superBloque, true)
					} else if inodo.I_type[0] == '1' { // Archivo
						dibujarBloqueArchivo(&salida, inodo.I_block[j], pathDisco, superBloque)
					}
				}
			}
		}

		// Procesamos el bloque indirecto simple
		if inodo.I_block[12] != -1 {
			if !procesados[inodo.I_block[12]] {
				procesados[inodo.I_block[12]] = true
				dibujarBloqueApuntadores(&salida, inodo.I_block[12], pathDisco, superBloque, 1)
			}
		}

		// Procesamos el bloque indirecto doble
		if inodo.I_block[13] != -1 {
			if !procesados[inodo.I_block[13]] {
				procesados[inodo.I_block[13]] = true
				dibujarBloqueApuntadores(&salida, inodo.I_block[13], pathDisco, superBloque, 2)
			}
		}

		// Procesamos el bloque indirecto triple
		if inodo.I_block[14] != -1 {
			if !procesados[inodo.I_block[14]] {
				procesados[inodo.I_block[14]] = true
				dibujarBloqueApuntadores(&salida, inodo.I_block[14], pathDisco, superBloque, 3)
			}
		}

	}

	for i := int32(0); i < superBloque.S_blocks_count; i++ {
		if i+1 != superBloque.S_blocks_count {
			salida += fmt.Sprintf("\tblock%d -> block%d;\n", i, i+1)
		} else {
			// nada
		}
	}

	// Cerramos el diagrama
	salida += "}"

	return salida, nil
}

// REPORTE SUPER BLOQUE
func reporteSb(idPart string) (string, error) {
	// obtenemos el superbloque
	superBloque, _, _, err := estructuras.ObtenerSuperBloque(idPart)
	if err != nil {
		return "", err
	}

	// configuración del grafico
	salida := "digraph SuperBlock {\n"
	salida += "\trankdir=LR;\n"
	salida += "\tnode [shape=plain];\n"
	salida += "\tedge [color=\"blue\"];\n\n"

	// creamos la tabla
	salida += "\tSuperBlock [label=<\n"
	salida += "\t\t<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"

	// creando MBR
	salida += "\t\t\t<TR><TD COLSPAN=\"2\" BGCOLOR=\"lightblue\"><B>SuperBlock</B></TD></TR>\n"

	// obtenemos la fecha
	mtime := global.ObtenerFecha(time.Unix(int64(superBloque.S_mtime), 0))
	umtime := global.ObtenerFecha(time.Unix(int64(superBloque.S_umtime), 0))

	salida += fmt.Sprintf("\t\t\t<TR><TD>Filesystem Type</TD><TD>%d</TD></TR>\n", superBloque.S_filesystem_type)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Inodes Count</TD><TD>%d</TD></TR>\n", superBloque.S_inodes_count)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Blocks Count</TD><TD>%d</TD></TR>\n", superBloque.S_blocks_count)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Free Inodes Count</TD><TD>%d</TD></TR>\n", superBloque.S_free_inodes_count)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Free Blocks Count</TD><TD>%d</TD></TR>\n", superBloque.S_free_blocks_count)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Mount Time</TD><TD>%s</TD></TR>\n", mtime)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Unmount Time</TD><TD>%s</TD></TR>\n", umtime)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Mount Count</TD><TD>%d</TD></TR>\n", superBloque.S_mnt_count)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Magic</TD><TD>%d</TD></TR>\n", superBloque.S_magic)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Inode Size</TD><TD>%d</TD></TR>\n", superBloque.S_inode_size)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Block Size</TD><TD>%d</TD></TR>\n", superBloque.S_block_size)
	salida += fmt.Sprintf("\t\t\t<TR><TD>First Inode</TD><TD>%d</TD></TR>\n", superBloque.S_first_ino)
	salida += fmt.Sprintf("\t\t\t<TR><TD>First Block</TD><TD>%d</TD></TR>\n", superBloque.S_first_blo)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Bitmap Inode Start</TD><TD>%d</TD></TR>\n", superBloque.S_bm_inode_start)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Bitmap Block Start</TD><TD>%d</TD></TR>\n", superBloque.S_bm_block_start)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Inode Start</TD><TD>%d</TD></TR>\n", superBloque.S_inode_start)
	salida += fmt.Sprintf("\t\t\t<TR><TD>Block Start</TD><TD>%d</TD></TR>\n", superBloque.S_block_start)
	// cerrando tabla
	salida += "\t\t</TABLE>\n"
	salida += "\t>];\n"
	salida += "}"

	return salida, nil
}

// REPORTE DE ARCHIVOS O FICHEROS CON PERMISOS	- PENDIENTE
func reporteLs(idPart string, pathFile []string) (string, error) {
	// obtenemos el superbloque
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idPart)
	if err != nil {
		return "", err
	}

	salida := "digraph TablaPermisos {\n"
	salida += "\tnode [shape=plaintext fontname=\"Helvetica\"];\n"
	salida += "\ttabla [label=<\n"
	salida += "\t\t<table BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"6\" BGCOLOR=\"#fdfdfd\">\n"
	salida += "\t\t\t<tr>\n"
	salida += "\t\t\t\t<td bgcolor=\"#d9e1f2\"><b>Permisos</b></td>\n"
	salida += "\t\t\t\t<td bgcolor=\"#d9e1f2\"><b>Propietario</b></td>\n"
	salida += "\t\t\t\t<td bgcolor=\"#d9e1f2\"><b>Grupo</b></td>\n"
	salida += "\t\t\t\t<td bgcolor=\"#d9e1f2\"><b>tamaño</b></td>\n"
	salida += "\t\t\t\t<td bgcolor=\"#d9e1f2\"><b>Fecha</b></td>\n"
	salida += "\t\t\t\t<td bgcolor=\"#d9e1f2\"><b>Tipo</b></td>\n"
	salida += "\t\t\t\t<td bgcolor=\"#d9e1f2\"><b>nombre</b></td>\n"
	salida += "\t\t\t</tr>\n"

	// vamos a leer users.txt
	var FirstInodo estructuras.Inode
	FirstInodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+int32(binary.Size(estructuras.Inode{}))))

	var contUs string
	var firstFileBlock estructuras.FileBlock
	for _, item := range FirstInodo.I_block {
		if item != -1 {
			firstFileBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(item*int32(binary.Size(estructuras.FileBlock{})))))
			contUs += string(firstFileBlock.B_content[:])
		}
	}

	lineaID := strings.Split(contUs, "\n")

	rutaFile := strings.Join(pathFile, "/")
	idInodo := gestionSistema.BuscarInodo(0, rutaFile, *superBloque, pathDisco)
	var inodo estructuras.Inode
	if idInodo > 0 {
		inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{})))))
		var folderBlock estructuras.FolderBlock
		for _, idBlock := range inodo.I_block {
			if idBlock != -1 {
				folderBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBlock*int32(binary.Size(estructuras.FolderBlock{})))))
				for k := 2; k < 4; k++ {
					apuntador := folderBlock.B_content[k].B_inodo
					if apuntador != -1 {
						pathActual := global.ObtenerNombreB(string(folderBlock.B_content[k].B_name[:]))

						salida += InodoLs(pathActual, lineaID, apuntador, *superBloque, pathDisco)
					}
				}
			}
		}

	} else {
		salida = "REP ERROR NO SE ENCONTRO LA PATH INGRESADA"
	}

	salida += "\t\t</table>\n"
	salida += "\t>];"
	salida += "}"

	return salida, nil
}

func InodoLs(name string, lineaID []string, idInodo int32, superBloque estructuras.SuperBlock, pathDisco string) string {
	var contenido string

	//cargar el inodo a reportar
	var inodo estructuras.Inode
	inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{})))))

	//Busco el grupo y el usuario
	usuario := ""
	grupo := ""
	for m := 0; m < len(lineaID); m++ {
		datos := strings.Split(lineaID[m], ",")
		if len(datos) == 5 {
			us := fmt.Sprintf("%d", inodo.I_uid)
			if us == datos[0] {
				usuario = datos[3]
			}
		}
		if len(datos) == 3 {
			gr := fmt.Sprintf("%d", inodo.I_gid)
			if gr == (datos[0]) {
				grupo = datos[2]
			}
		}

	}

	tipoArchivo := "Archivo"
	var permisos string

	//Los permisos son 3 numeros porque son aplicados a: propierarios   grupos  y  otros
	//Cada numero representa los permisos de lectura, escritura y ejecucion: r w x
	// r lectura
	// w escritura
	// x ejecucion
	//Si el numero de permisos es: 764, significa que:
	//el propierario(7) tiene permisos de lectura escritura ejecucion
	//el grupo(6) tiene permisos de lectura escritura
	//otros(4) tienen permisos de lectura
	for i := 0; i < 3; i++ {
		if string(inodo.I_perm[i]) == "0" { //ninun permiso
			permisos += "---"
		} else if string(inodo.I_perm[i]) == "1" { // ejecucion
			permisos += "--x"
		} else if string(inodo.I_perm[i]) == "2" { //	escritura
			permisos += "-w-"
		} else if string(inodo.I_perm[i]) == "3" { // 	ecritura ejecucion
			permisos += "-wx"
		} else if string(inodo.I_perm[i]) == "4" { //lectura
			permisos += "r--"
		} else if string(inodo.I_perm[i]) == "5" { //lectura  	ejecucion
			permisos += "r-x"
		} else if string(inodo.I_perm[i]) == "6" { // lectura escritura
			permisos += "rw-"
		} else if string(inodo.I_perm[i]) == "7" { //lectura escritura ejecucion
			permisos += "rwx"
		}
	}

	if string(inodo.I_type[:]) == "0" {
		tipoArchivo = "Carpeta"
	}

	fecha := time.Unix(int64(inodo.I_ctime), 0)

	//permisos = "rw-rw-r--"
	contenido += "\t\t\t<tr>\n"
	contenido += fmt.Sprintf("\t\t\t\t<td align=\"center\" bgcolor=\"#ffffff\">%s</td>\n", permisos)
	contenido += fmt.Sprintf("\t\t\t\t<td align=\"center\" bgcolor=\"#ffffff\">%s</td>\n", usuario)
	contenido += fmt.Sprintf("\t\t\t\t<td align=\"center\" bgcolor=\"#ffffff\">%s</td>\n", grupo)
	contenido += fmt.Sprintf("\t\t\t\t<td align=\"center\" bgcolor=\"#ffffff\">%d</td>\n", inodo.I_size)
	contenido += fmt.Sprintf("\t\t\t\t<td align=\"center\" bgcolor=\"#ffffff\">%s</td>\n", global.ObtenerFecha(fecha))
	contenido += fmt.Sprintf("\t\t\t\t<td align=\"center\" bgcolor=\"#ffffff\">%s</td>\n", name)
	contenido += fmt.Sprintf("\t\t\t\t<td align=\"center\" bgcolor=\"#ffffff\">%s</td>\n", tipoArchivo)
	contenido += "\t\t\t</tr>\n"

	//reportar el inodo
	return contenido
}

// REPORTE ARBOL
func reporteTree(idPart string) (string, error) {
	// Obtenemos el superbloque
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idPart)
	if err != nil {
		return "", err
	}

	// Iniciar el contenido DOT
	salida := "digraph Tree {"
	salida += "\tnode [shape=plaintext];\n"
	salida += "\trankdir=LR;\n"
	salida += "\tlabel=\"Tree\";\n"
	salida += "\tbgcolor=\"#E6F3FF\";\n"

	// Mapa para rastrear bloques ya procesados
	procesados := make(map[int32]bool)

	// Iterar sobre cada inodo
	for i := int32(0); i < superBloque.S_inodes_count; i++ {
		inodo := &estructuras.Inode{}
		// Deserializar el inodo
		err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(i*superBloque.S_inode_size)))
		if err != nil {
			continue // Si hay error, seguimos con el siguiente
		}

		// Verificar si el inodo está en uso (podemos validar atributos)
		if inodo.I_size <= 0 && inodo.I_type[0] == 0 {
			continue // Inodo vacío, seguimos con el siguiente
		}

		// Convertir tiempos a string usando la función de global
		atime := global.ObtenerFecha(time.Unix(int64(inodo.I_atime), 0))
		ctime := global.ObtenerFecha(time.Unix(int64(inodo.I_ctime), 0))
		mtime := global.ObtenerFecha(time.Unix(int64(inodo.I_mtime), 0))

		tipoInodo := "Carpeta"
		if inodo.I_type[0] == '1' {
			tipoInodo = "Archivo"
		}

		// Definir el contenido DOT para el inodo actual
		salida += fmt.Sprintf("\tinode%d [label=<\n", i)
		salida += "\t\t<table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n"
		salida += fmt.Sprintf("\t\t\t<tr><td colspan=\"2\" bgcolor=\"#4B92DB\"><font color=\"white\">INODO %d</font></td></tr>\n", i)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_uid</td><td>%d</td></tr>\n", inodo.I_uid)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_gid</td><td>%d</td></tr>\n", inodo.I_gid)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_size</td><td>%d</td></tr>\n", inodo.I_size)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_atime</td><td>%s</td></tr>\n", atime)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_ctime</td><td>%s</td></tr>n", ctime)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_mtime</td><td>%s</td></tr>\n", mtime)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_type</td><td>%s</td></tr>\n", tipoInodo)
		salida += fmt.Sprintf("\t\t\t<tr><td>i_perm</td><td>%s</td></tr>`\n", string(inodo.I_perm[:]))
		salida += "\t\t\t<tr><td colspan=\"2\" bgcolor=\"#A8C6DF\">BLOQUES DIRECTOS</td></tr>\n"

		for j := 0; j < 12; j++ {
			if inodo.I_block[j] != -1 {
				salida += fmt.Sprintf("\t\t\t<tr><td>Bloque %d</td><td port=\"blk%d\">%d</td></tr>\n", j, j, inodo.I_block[j])
			} else {
				salida += fmt.Sprintf("\t\t\t<tr><td>Bloque %d</td><td>-1</td></tr>\n", j)
			}
		}

		// Agregar los bloques indirectos a la tabla
		salida += "\t\t\t<tr><td colspan=\"2\" bgcolor=\"#A8C6DF\">BLOQUES INDIRECTOS</td></tr>\n"
		salida += fmt.Sprintf("\t\t\t<tr><td>Indirecto Simple</td><td port=\"ind1\">%d</td></tr>\n", inodo.I_block[12])
		salida += fmt.Sprintf("\t\t\t<tr><td>Indirecto Doble</td><td port=\"ind2\">%d</td></tr>\n", inodo.I_block[13])
		salida += fmt.Sprintf("\t\t\t<tr><td>Indirecto Triple</td><td port=\"ind3\">%d</td></tr>\n", inodo.I_block[14])
		salida += "\t\t</table>\n"
		salida += "\t>];\n"
	}

	// Ahora procesamos los bloques
	// Iteramos nuevamente sobre cada inodo para procesar sus bloques
	for i := int32(0); i < superBloque.S_inodes_count; i++ {
		inodo := &estructuras.Inode{}
		// Deserializar el inodo
		err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(i*superBloque.S_inode_size)))
		if err != nil {
			continue
		}

		// Si el inodo está vacío, seguimos con el siguiente
		if inodo.I_size <= 0 && inodo.I_type[0] == 0 {
			continue
		}

		// Procesamos bloques directos
		for j := 0; j < 12; j++ {
			if inodo.I_block[j] != -1 {
				// Si el bloque ya fue procesado, lo saltamos
				if procesados[inodo.I_block[j]] {
					continue
				}
				procesados[inodo.I_block[j]] = true

				// Procesamos el bloque según el tipo de inodo
				if inodo.I_type[0] == '0' { // Carpeta
					// Dibujamos el bloque de carpeta
					dibujarBloqueCarpeta(&salida, inodo.I_block[j], pathDisco, superBloque, false)
					// Creamos el enlace entre inodo y bloque
					salida += fmt.Sprintf("\tinode%d:blk%d -> block%d;\n", i, j, inodo.I_block[j])
				} else if inodo.I_type[0] == '1' { // Archivo
					// Dibujamos el bloque de archivo
					dibujarBloqueArchivo(&salida, inodo.I_block[j], pathDisco, superBloque)
					// Creamos el enlace entre inodo y bloque
					salida += fmt.Sprintf("\tinode%d:blk%d -> block%d;\n", i, j, inodo.I_block[j])
				}
			}
		}

		// Procesamos el bloque indirecto simple
		if inodo.I_block[12] != -1 {
			if !procesados[inodo.I_block[12]] {
				procesados[inodo.I_block[12]] = true
				dibujarBloqueApuntadores(&salida, inodo.I_block[12], pathDisco, superBloque, 1)
				salida += fmt.Sprintf("\tinode%d:ind1 -> pblock%d;\n", i, inodo.I_block[12])
			}
		}

		// Procesamos el bloque indirecto doble
		if inodo.I_block[13] != -1 {
			if !procesados[inodo.I_block[13]] {
				procesados[inodo.I_block[13]] = true
				dibujarBloqueApuntadores(&salida, inodo.I_block[13], pathDisco, superBloque, 2)
				salida += fmt.Sprintf("\tinode%d:ind2 -> pblock%d;\n", i, inodo.I_block[13])
			}
		}

		// Procesamos el bloque indirecto triple
		if inodo.I_block[14] != -1 {
			if !procesados[inodo.I_block[14]] {
				procesados[inodo.I_block[14]] = true
				dibujarBloqueApuntadores(&salida, inodo.I_block[14], pathDisco, superBloque, 3)
				salida += fmt.Sprintf("\tinode%d:ind3 -> pblock%d;\n", i, inodo.I_block[14])
			}
		}
	}

	// Cerramos el diagrama
	salida += "}"

	return salida, nil
}

// Función para dibujar un bloque de carpeta
func dibujarBloqueCarpeta(salida *string, indice int32, pathDisco string, sb *estructuras.SuperBlock, EsbloqueRep bool) {
	bloque := &estructuras.FolderBlock{}
	err := bloque.Deserialize(pathDisco, int64(sb.S_block_start+(indice*sb.S_block_size)))
	if err != nil {
		return
	}

	*salida += fmt.Sprintf("\tblock%d [label=<\n", indice)
	*salida += "\t\t<table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n"
	*salida += fmt.Sprintf("\t\t\t<tr><td colspan=\"2\" bgcolor=\"#FFB347\"><font color=\"white\">BLOQUE %d</font></td></tr>", indice)

	// Agregar las entradas de carpeta
	for j, content := range bloque.B_content {
		nombre := global.BorrandoIlegibles(string(bytes.Trim(content.B_name[:], "\x00")))
		if nombre != "" {
			*salida += fmt.Sprintf("\t\t\t<tr><td>%s</td><td port=\"fc%d\">%d</td></tr>\n", nombre, j, content.B_inodo)
		}
	}

	*salida += "\t\t</table>\n"
	*salida += "\t>];\n"

	if !EsbloqueRep {
		// Agregar las entradas de carpeta
		for j, content := range bloque.B_content {
			nombre := global.BorrandoIlegibles(string(bytes.Trim(content.B_name[:], "\x00")))
			if nombre != "" {
				// Enlace al inodo referenciado (si no es -1 o vacío)
				if content.B_inodo != -1 && nombre != "" && nombre != "." && nombre != ".." {
					*salida += fmt.Sprintf("\tblock%d:fc%d -> inode%d;\n", indice, j, content.B_inodo)
				}
			}
		}
	}
}

// Función para dibujar un bloque de archivo
func dibujarBloqueArchivo(salida *string, indice int32, pathDisco string, sb *estructuras.SuperBlock) {
	bloque := &estructuras.FileBlock{}
	err := bloque.Deserialize(pathDisco, int64(sb.S_block_start+(indice*sb.S_block_size)))
	if err != nil {
		return
	}

	// Preparamos el contenido para mostrar usando la función global para eliminar caracteres ilegibles
	contenido := global.BorrandoIlegibles(string(bytes.Trim(bloque.B_content[:], "\x00")))
	//fmt.Println("contenido dot inti:\n\n" + contenido + "\n\n fin dot")

	// // Escapamos caracteres especiales para DOT
	contenido = strings.Replace(contenido, "<", "&lt;", -1)
	contenido = strings.Replace(contenido, ">", "&gt;", -1)
	contenido = strings.Replace(contenido, "&", "&amp;", -1)
	contenido = strings.Replace(contenido, "\"", "&quot;", -1)

	*salida += fmt.Sprintf("\tblock%d [label=<\n", indice)
	*salida += "\t\t<table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n"
	*salida += fmt.Sprintf("\t\t\t<tr><td bgcolor=\"#8CC152\"><font color=\"white\">BLOQUE %d</font></td></tr>\n", indice)
	// mostramos todo el contenido
	*salida += fmt.Sprintf("\t\t\t<tr><td>%s</td></tr>\n", contenido)
	*salida += "\t\t</table>\n"
	*salida += "\t>];\n"
}

// Función para dibujar un bloque de apuntadores
func dibujarBloqueApuntadores(salida *string, indice int32, pathDisco string, sb *estructuras.SuperBlock, nivel int) {
	bloque := &estructuras.Pointerblock{}

	// Abrimos el archivo usando las funciones de global
	archivo, err := global.AbrirDisco(pathDisco)
	if err != nil {
		return
	}
	defer archivo.Close()

	// Leer los datos usando la función de global
	offset := int64(sb.S_block_start + (indice * sb.S_block_size))
	err = global.LeerEnDisco(archivo, &bloque.B_pointers, offset)
	if err != nil {
		return
	}

	// Determinar el color según el nivel de indirección
	var color string
	switch nivel {
	case 1:
		color = "#967ADC"
	case 2:
		color = "#4B89DC"
	case 3:
		color = "#3BAFDA"
	}

	*salida += fmt.Sprintf("\tpblock%d [label=<\n", indice)
	*salida += "\t\t<table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n"
	*salida += fmt.Sprintf("\t\t\t<tr><td colspan=\"2\" bgcolor=\"%s\"><font color=\"white\">BLOQUE %d</font></td></tr>\n", color, indice)

	// Agregar las entradas de apuntadores
	for j, pointer := range bloque.B_pointers {
		if pointer != -1 {
			*salida += fmt.Sprintf("\t\t\t<tr><td>%d</td><td port=\"p%d\">%d</td></tr>\n", j, j, pointer)

			// Si es nivel 1, los apuntadores apuntan directamente a bloques de datos
			if nivel == 1 {
				// Aquí tener cuidado con el tipo de bloque al que apunta
				// Podría ser archivo o carpeta, necesitarías saber eso
				*salida += fmt.Sprintf("\tpblock%d:p%d -> block%d;\n", indice, j, pointer)
			} else {
				// Si es nivel 2 o 3, los apuntadores apuntan a otros bloques de apuntadores
				*salida += fmt.Sprintf("\tpblock%d:p%d -> pblock%d;\n", indice, j, pointer)
			}
		}
	}

	*salida += "\t\t</table>\n"
	*salida += "\t>];\n"

	// Recursivamente procesar los bloques a los que apunta según el nivel
	if nivel > 1 {
		for _, pointer := range bloque.B_pointers {
			if pointer != -1 {
				dibujarBloqueApuntadores(salida, pointer, pathDisco, sb, nivel-1)
			}
		}
	}
}
