package adminSistemaArchivos

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/global"
	"strings"
	"time"
)

// MKFS no requiere sesión iniciada

type parametros_mkfs struct {
	Id   string //este es el id de la partición luego de ser montada (si no está montada tira error) - OBLIGATORIO
	Type string // tipo de formateo el  único valor que puede ser es full - OPCIONAL
	Fs   string // tipo de sistema de archivos al que se desea formatear, admite fs3 y fs2, por defecto fs2 - OPCIONAL
}

/*
	mkfs -id=ID
	mkfs -id=ID -type=full -fs=3fs
*/

// analizador de los parametros
func Mkfs(parametros []string) string {

	mkfs := parametros_mkfs{}

	//valores por defecto
	mkfs.Type = "full"
	mkfs.Fs = "2fs"

	for _, param := range parametros {
		// separamos el parametro en nombre y valor
		parametro := strings.Split(param, "=")
		//verificamos que precisamente sea [nombre,valor]
		if len(parametro) == 2 {
			switch strings.ToLower(parametro[0]) {
			case "id":
				mkfs.Id = parametro[1]
			case "type":
				tipo := strings.ToLower(parametro[1])
				if tipo != "full" {
					return "MKFS ERROR: el parametro Type solo puede tener como valor full, no: " + tipo + "\n"
				}
				// no es necesario almacenarlo porque ya lo hicimos antes
			case "fs":
				fs := strings.ToLower(parametro[1])
				if fs != "2fs" && fs != "3fs" {
					return "MKFS ERROR: El tipo de eliminación de la partición debe ser fast o full, no " + fs + "\n"
				}
				// este si lo sobreescribimos porque puede ser fs2 o fs3
				mkfs.Fs = fs
			default:
				return "MKFS ERROR: el parametro: " + parametro[0] + " no es valido." + "\n"
			}
		} else {
			return "MKFS ERROR: formato invalido para el parametro: " + param + "\n"
		}
	}

	// en cualquier caso path y Name son obligaorios
	if mkfs.Id == "" {
		return "MKFS ERROR: El parametro ID es obligatorio para poder ejecutar el comando.\n"
	}

	return comandoMkfs(mkfs)
}

// ejecución del coamndo
func comandoMkfs(comando parametros_mkfs) string {
	pathDisco, ok := global.ParticionesMontadas[comando.Id]
	// verificamos que exista el Id en las particiones montada (son las únicas que tienen ID)
	if !ok {
		msj := "MKFS ERROR: el Id ingresado no existe"
		fmt.Println(msj)
		return msj
	}

	// abrimos el disco
	disco, err := global.AbrirDisco(pathDisco)
	if err != nil {
		msj := "MKFS " + err.Error() + "\n"
		fmt.Println(msj)
		return msj
	}

	// obtenemos el MBR que está en el disco
	var mbr estructuras.MBR
	if err := global.LeerEnDisco(disco, &mbr, 0); err != nil {
		msj := "MKFS " + err.Error() + "\n"
		fmt.Println(msj)
		return msj
	}

	// obtenemos la partición
	particion, err := mbr.ParticionPorId(comando.Id)
	if err != nil {
		msj := "MKFS " + err.Error()
		fmt.Println(msj)
		return msj
	}

	n := calcularN(particion, comando.Fs)

	// Inicializar un nuevo superbloque
	SuperBlock := crearSuperBloque(particion, n, comando.Fs)

	if err := SuperBlock.CreateBitMaps(pathDisco); err != nil {
		msj := "MKFS ERROR: " + err.Error()
		return msj
	}

	// Validar que sistema de archivos es
	var salida string

	if SuperBlock.S_filesystem_type == 3 {
		// para que no de error en el primer journal
		global.UsuarioActual.PartitionID = comando.Id
		// Crear archivo users.txt ext3
		if err := SuperBlock.CrearExt3(pathDisco, int64(particion.Start+int32(binary.Size(estructuras.SuperBlock{})))); err != nil {
			msj := "MKFS " + err.Error()
			fmt.Println(msj)
			return msj
		}
		salida = "Formato ext3 aplicado correctamente a " + string(particion.Id[:])
	} else {
		// Crear archivo users.txt ext2
		if err := SuperBlock.CrearExt2(pathDisco); err != nil {
			msj := "MKFS " + err.Error()
			fmt.Println(msj)
			return msj
		}
		salida = "Formato ext2 aplicado correctamente a " + string(particion.Id[:])
	}

	// serializar en disco
	if err := global.EscribirEnDisco(disco, SuperBlock, int64(particion.Start)); err != nil {
		msj := "MKFS ERROR: " + err.Error()
		return msj
	}

	//cerramos el disco
	defer disco.Close()
	// cerramos la sesión actual
	global.UsuarioActual.Logout()
	return salida
}

// calculando el N para crear los bloques
func calcularN(particion *estructuras.Partition, tipo string) int32 {
	/*
		superbloque = 68 ; inodo = 88 ; FileBlock = 64
		formula: ( P - S ) / ( 4 + J + I + 3B )
	*/
	numerador := particion.Size - int32(binary.Size(estructuras.SuperBlock{}))
	denominador := int32(4 + binary.Size(estructuras.Inode{}) + 3*binary.Size(estructuras.FileBlock{}))

	// verificamos si necesita agregar el espacio del journaling
	if tipo == "3fs" {
		denominador += int32(binary.Size(estructuras.Journal{}))
	}

	// retornamos el valor
	return numerador / denominador
}

// creando superbloque
func crearSuperBloque(partition *estructuras.Partition, n int32, fs string) *estructuras.SuperBlock {
	// Calcular punteros de las estructuras
	_, bm_inode_start, bm_block_start, inode_start, block_start := calcularPosiciones(partition, fs, n)

	// fmt.Printf("Journal Start: %d\n", journal_start)
	// fmt.Printf("Bitmap Inode Start: %d\n", bm_inode_start)
	// fmt.Printf("Bitmap Block Start: %d\n", bm_block_start)
	// fmt.Printf("Inode Start: %d\n", inode_start)
	// fmt.Printf("Block Start: %d\n", block_start)

	// Tipo de sistema de archivos
	var fsType int32

	if fs == "2fs" {
		fsType = 2
	} else {
		fsType = 3
	}

	//fecha del montaje
	fecha := time.Now()
	fechaUnix := fecha.Unix()
	fechaFloat := float32(fechaUnix)

	// Crear un nuevo superbloque
	SuperBlock := &estructuras.SuperBlock{
		S_filesystem_type:   fsType,
		S_free_inodes_count: int32(n),
		S_free_blocks_count: int32(n * 3),
		S_mtime:             fechaFloat,
		S_umtime:            fechaFloat, // siempre que umtime == mtime entonces no se ha montado
		S_mnt_count:         1,          // se inicia en 1 porque se está por montar
		S_magic:             0xEF53,
		S_inodes_count:      0, // lo iniciamos en 0 por como se cuentan los bloques
		S_inode_size:        int32(binary.Size(estructuras.Inode{})),
		S_blocks_count:      0, // igual que inodes_count
		S_block_size:        int32(binary.Size(estructuras.FileBlock{})),
		S_first_ino:         inode_start,
		S_first_blo:         block_start,
		S_bm_inode_start:    bm_inode_start,
		S_bm_block_start:    bm_block_start,
		S_inode_start:       inode_start,
		S_block_start:       block_start,
		S_nvalue:            n,
	}

	return SuperBlock
}

// calculando las posiciones de los bloques para simplificar la creación de los sistemas de archivos
func calcularPosiciones(partition *estructuras.Partition, fs string, n int32) (int32, int32, int32, int32, int32) {
	SuperBlockSize := int32(binary.Size(estructuras.SuperBlock{}))
	journalSize := int32(binary.Size(estructuras.Journal{}))
	inodeSize := int32(binary.Size(estructuras.Inode{}))

	// Inicializar posiciones
	bmInodeStart := partition.Start + SuperBlockSize
	bmBlockStart := bmInodeStart + n
	inodeStart := bmBlockStart + (3 * n)
	blockStart := inodeStart + (inodeSize * n)

	// Ajustar para EXT3
	journalStart := int32(0)
	if fs == "3fs" {
		journalStart = partition.Start + SuperBlockSize
		bmInodeStart = journalStart + (journalSize * n)
		bmBlockStart = bmInodeStart + n
		inodeStart = bmBlockStart + (3 * n)
		blockStart = inodeStart + (inodeSize * n)
	}

	return journalStart, bmInodeStart, bmBlockStart, inodeStart, blockStart
}
