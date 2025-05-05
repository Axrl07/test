package estructuras

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Total de 4*15 + 2*4 = 68 bytes
type SuperBlock struct {
	S_filesystem_type   int32   // el número que identifica el sistema de archivos
	S_inodes_count      int32   // total de inodos
	S_blocks_count      int32   // total de bloques
	S_free_inodes_count int32   // total de inodos libre
	S_free_blocks_count int32   // total de bloques libres
	S_mtime             float32 // Última fecha en el que el sistema fue montado
	S_umtime            float32 // Última fecha en que el sistema fue desmontado
	S_mnt_count         int32   // Indica cuantas veces se ha montado el sistema
	S_magic             int32   // identifica al sistema de archivos, tendrá el valor 0xEF53
	S_inode_size        int32   // Tamaño del inodo
	S_block_size        int32   // Tamaño del bloque
	S_first_ino         int32   // dirección del primer inodo libre
	S_first_blo         int32   // dirección del primer bloque libre
	S_bm_inode_start    int32   // guarda el inicio del bitmap de inodos
	S_bm_block_start    int32   // guarda el inicio del bitmap de bloques
	S_inode_start       int32   // guarda el inicio de la tabla de inodos
	S_block_start       int32   // guarda el inicio de la tabla de bloques
	S_nvalue            int32   // guarda el valor de n para recalcular en recovery
}

// Serialize escribe la estructura SuperBlock en un archivo binario en la posición especificada
func (sb *SuperBlock) Serialize(path string, offset int64) error {
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

	// Serializar la estructura SuperBlock directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura SuperBlock desde un archivo binario en la posición especificada
func (sb *SuperBlock) Deserialize(path string, offset int64) error {
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

	// Obtener el tamaño de la estructura SuperBlock
	sbSize := binary.Size(sb)
	if sbSize <= 0 {
		return fmt.Errorf("invalid SuperBlock size: %d", sbSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura SuperBlock
	buffer := make([]byte, sbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura SuperBlock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// -------------------------------- obtener contenido de usuarios.txt completo --------------------------------

// me devuelve todo el contenido de users.txt
func (sb *SuperBlock) ObtenerUsuariosTxt(path string) (string, error) {

	inodo := &Inode{}

	// obtenemos la posición del Inodo 1 (el 0 maneja el root y el 1 el users.txt)
	if err := inodo.Deserialize(path, int64(sb.S_inode_start+(sb.S_inode_size))); err != nil {
		return "", err
	}

	var contenido string
	// Iteramos sobre cada bloque dle inodo para ir acumulando el contenido de users.txt
	for _, bloqueIndice := range inodo.I_block {
		// si es -1 entonces está vacío, paramos
		if bloqueIndice == -1 {
			break
		}

		// si es de tipo archivo lo tomamos
		if inodo.I_type[0] == '1' {
			bloque := &FileBlock{}

			if err := bloque.Deserialize(path, int64(sb.S_block_start+(sb.S_block_size*bloqueIndice))); err != nil {
				return "", err
			}

			//contenido += global.BorrandoIlegibles(string(bloque.B_content[:]))
			contenido += strings.Trim(string(bloque.B_content[:]), "\x00")

			//fmt.Printf("indice: %d , linea: %s ", i, contenido)
		}
	}

	return contenido, nil
}

// -------------------------------- obtener UID y GID apartir del nombre de usuario --------------------------------

func (sb *SuperBlock) ObtenerGID_UID(path string, username string) (int32, int32, error) {

	valorUID, grupo, err := sb.obtenerUID(path, username)
	if err != nil {
		return -1, -1, err
	}

	valorGID, err := sb.obtenerGID(path, grupo)
	if err != nil {
		return -1, -1, err
	}

	return valorUID, valorGID, nil
}

// regresa el UID usuario y el GID del grupo
func (sb *SuperBlock) obtenerUID(path string, username string) (int32, string, error) {

	contenido, err := sb.ObtenerUsuariosTxt(path)
	if err != nil {
		return -1, "", errors.New("no fue posible obtener el users.txt")
	}

	lineas := strings.Split(contenido, "\n")

	for _, linea := range lineas {
		atributos := strings.Split(linea, ",")
		if len(atributos) == 5 {
			if atributos[3] == username {
				uid, err := strconv.Atoi(atributos[0])
				if err != nil {
					return -1, "", errors.New("no fue posible convertir a int el UID")
				}
				return int32(uid), atributos[2], nil
			}
		}
	}

	return -1, "", errors.New("no se encontró el UID")
}

// regresa el UID usuario y el GID del grupo
func (sb *SuperBlock) obtenerGID(path string, groupname string) (int32, error) {

	contenido, err := sb.ObtenerUsuariosTxt(path)
	if err != nil {
		return -1, errors.New("no fue posible obtener el users.txt")
	}

	lineas := strings.Split(contenido, "\n")

	for _, linea := range lineas {
		atributos := strings.Split(linea, ",")
		if len(atributos) == 3 {
			if atributos[2] == groupname {
				gid, err := strconv.Atoi(atributos[0])
				if err != nil {
					return -1, errors.New("no fue posible convertir a int el GID")
				}
				return int32(gid), nil
			}
		}
	}

	return -1, errors.New("no se encontró el GID")
}

// -------------------------------- obtener usuario y grupo a partir del gid y uid --------------------------------

func (sb *SuperBlock) ObtenerUsuario_Grupo(path string, uid int32) (string, string, error) {

	contenido, err := sb.ObtenerUsuariosTxt(path)
	if err != nil {
		return "", "", errors.New("no fue posible obtener el users.txt")
	}

	fmt.Println(contenido)
	lineas := strings.Split(contenido, "\n")

	for _, linea := range lineas {
		atributos := strings.Split(linea, ",")
		if len(atributos) == 5 {
			id, err := strconv.Atoi(atributos[0])
			if err != nil {
				return "", "", errors.New("no fue posible obtener el id")
			}
			if int32(id) == uid {
				return atributos[3], atributos[2], nil
			}
		}
	}
	return "", "", errors.New("no fue posible obtener el usuario de users.txt")
}

// -------------------------------- actualización de bitmaps --------------------------------

// CreateBitMaps crea los Bitmaps de inodos y bloques en el archivo especificado
func (sb *SuperBlock) CreateBitMaps(path string) error {
	// Escribir Bitmaps
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Bitmap de inodos
	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(int64(sb.S_bm_inode_start), 0)
	if err != nil {
		return err
	}

	// Crear un buffer de n '0'
	buffer := make([]byte, sb.S_free_inodes_count)
	for i := range buffer {
		buffer[i] = '0'
	}

	// Escribir el buffer en el archivo
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	// Bitmap de bloques
	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(int64(sb.S_bm_block_start), 0)
	if err != nil {
		return err
	}

	// Crear un buffer de n 'O'
	buffer = make([]byte, sb.S_free_blocks_count)
	for i := range buffer {
		buffer[i] = 'O'
	}

	// Escribir el buffer en el archivo
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	return nil
}

// Actualizar Bitmap de inodos
func (sb *SuperBlock) ActualizarBitmapInodos(path string) error {
	// Abrir el archivo
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición del bitmap de inodos
	_, err = file.Seek(int64(sb.S_bm_inode_start)+int64(sb.S_inodes_count), 0)
	if err != nil {
		return err
	}

	// Escribir el bit en el archivo
	_, err = file.Write([]byte{'1'})
	if err != nil {
		return err
	}

	return nil
}

// Actualizar Bitmap de bloques
func (sb *SuperBlock) ActualizarBitmapBloques(path string) error {
	// Abrir el archivo
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición del bitmap de bloques
	_, err = file.Seek(int64(sb.S_bm_block_start)+int64(sb.S_blocks_count), 0)
	if err != nil {
		return err
	}

	// Escribir el bit en el archivo
	_, err = file.Write([]byte{'1'})
	if err != nil {
		return err
	}

	return nil
}

// -------------------------------- creación de /users.txt --------------------------------

func (sb *SuperBlock) CrearExt2(path string) error {
	// ----------- Creamos / -----------
	// Creamos el inodo raíz
	rootInode := &Inode{
		I_uid:  1,
		I_gid:  1,
		I_size: 0,
		// solo cuando se crea coinciden las 3 feechas
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		// inicialmente S_block_count = 0
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		// directamente cero porque es una carpeta
		I_type: [1]byte{'0'},
		// todos los permisos porque es el root
		I_perm: [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz
	err := rootInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos
	err = sb.ActualizarBitmapInodos(path)
	if err != nil {
		return err
	}

	// Actualizar información de INodos del superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Creamos el bloque del Inodo Raíz
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},
		},
	}

	// Actualizar el bitmap de bloques
	err = sb.ActualizarBitmapBloques(path)
	if err != nil {
		return err
	}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// ----------- por la creación de "/"

	// Actualizar información de bloques del superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// ----------- Creamos /users.txt -----------
	usersText := "1,G,root\n1,U,root,root,123\n"

	// Deserializar el inodo raíz
	err = rootInode.Deserialize(path, int64(sb.S_inode_start)) // S_inode_start porque es el inodo raiz
	if err != nil {
		return err
	}

	// Actualizamos el I_atime del inodo raíz
	rootInode.I_atime = float32(time.Now().Unix())

	// Serializar el inodo raíz
	err = rootInode.Serialize(path, int64(sb.S_inode_start)) // S_inode_start porque es el inodo raiz
	if err != nil {
		return err
	}

	// Deserializar el bloque de carpeta raíz
	err = rootBlock.Deserialize(path, int64(sb.S_block_start)) // S_block_start proque es el bloque raiz
	if err != nil {
		return err
	}

	// Actualizamos el bloque de carpeta raíz
	rootBlock.B_content[2] = FolderContent{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_block_start)) // S_block_start proque es el bloque raiz
	if err != nil {
		return err
	}

	// Creamos el inodo users.txt
	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Actualizar el bitmap de inodos
	err = sb.ActualizarBitmapInodos(path)
	if err != nil {
		return err
	}

	// Serializar el inodo users.txt
	err = usersInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// ----------- por la creación de "users.txt"
	// Actualizamos el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Creamos el bloque de users.txt
	usersBlock := &FileBlock{
		B_content: [64]byte{},
	}

	// Copiamos el texto de usuarios en el bloque
	copy(usersBlock.B_content[:], usersText)

	// Serializar el bloque de users.txt
	err = usersBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de bloques
	err = sb.ActualizarBitmapBloques(path)
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	return nil
}

// Crear users.txt en nuestro sistema de archivos
func (sb *SuperBlock) CrearExt3(path string, journauling_start int64) error {
	// ----------- Creamos / -----------

	// Creamos el inodo raíz
	rootInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz
	err := rootInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos
	err = sb.ActualizarBitmapInodos(path)
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// creando y serializando journal
	if err := CrearJournalOperacion("mkdir", "/", "", "mkdir -path=/\n"); err != nil {
		return err
	}

	// Creamos el bloque del Inodo Raíz
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},
		},
	}

	// Actualizar el bitmap de bloques
	err = sb.ActualizarBitmapBloques(path)
	if err != nil {
		return err
	}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// ----------- Creamos /users.txt -----------
	usersText := "1,G,root\n1,U,root,root,123\n"

	// Deserializar el inodo raíz
	err = rootInode.Deserialize(path, int64(sb.S_inode_start)) // start porque es el inodo raiz
	if err != nil {
		return err
	}

	// Actualizamos el inodo raíz
	rootInode.I_atime = float32(time.Now().Unix())

	// Serializar el inodo raíz
	err = rootInode.Serialize(path, int64(sb.S_inode_start)) // start porque es el inodo raiz
	if err != nil {
		return err
	}

	// Deserializar el bloque de carpeta raíz
	err = rootBlock.Deserialize(path, int64(sb.S_block_start)) // start porque es el bloque raiz
	if err != nil {
		return err
	}

	// Actualizamos el bloque de carpeta raíz
	rootBlock.B_content[2] = FolderContent{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_block_start)) // start porque es el bloque raiz
	if err != nil {
		return err
	}

	// Creamos el inodo users.txt
	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Actualizar el bitmap de inodos
	err = sb.ActualizarBitmapInodos(path)
	if err != nil {
		return err
	}

	// Serializar el inodo users.txt
	err = usersInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// creando y serializando journal
	if err := CrearJournalOperacion("mkfile", "/users.txt", usersText, "mkfile -path=/users.txt -cont=/home/angel/Escritorio/angel/users.txt\n"); err != nil {
		return err
	}

	// Creamos el bloque de users.txt
	usersBlock := &FileBlock{
		B_content: [64]byte{},
	}
	// Copiamos el texto de usuarios en el bloque
	copy(usersBlock.B_content[:], usersText)

	// Serializar el bloque de users.txt
	err = usersBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de bloques
	err = sb.ActualizarBitmapBloques(path)
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	return nil
}
