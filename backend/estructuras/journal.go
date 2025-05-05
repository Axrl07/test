package estructuras

import (
	"encoding/binary"
	"errors"
	"fmt"
	"main/global"
	"os"
	"time"
)

// total de 4 + 142 = 146 bytes
type Journal struct {
	J_count   int32
	J_content Information
}

// total de 10 + 64 + 64 + 4 = 128 + 14 = 142 bytes
type Information struct {
	I_operation [10]byte
	I_path      [64]byte
	I_content   [64]byte
	I_date      float32
}

// SerializeJournal escribe la estructura Journal en un archivo binario ( el journaling_start = particion.start + superBLock.size)
func (journal *Journal) Serialize(path string, journauling_start int64) error {
	// Calcular la posición en el archivo
	offset := journauling_start + (int64(binary.Size(Journal{})) * int64(journal.J_count))

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

	// Serializar la estructura Journal directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, journal)
	if err != nil {
		return err
	}

	return nil
}

// DeserializeJournal lee la estructura Journal desde un archivo binario
func (journal *Journal) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Deserializar la estructura Journal directamente desde el archivo
	err = binary.Read(file, binary.LittleEndian, journal)
	if err != nil {
		return err
	}

	return nil
}

// PrintJournal imprime en consola la estructura Journal
func (journal *Journal) Print() {
	// Convertir el tiempo de montaje a una fecha
	date := time.Unix(int64(journal.J_content.I_date), 0)
	// obtenemos la fecha
	fecha := global.ObtenerFecha(date)

	fmt.Println("Journal:")
	fmt.Printf("J_count: %d", journal.J_count)
	fmt.Println("Information:")
	fmt.Printf("I_operation: %s", string(journal.J_content.I_operation[:]))
	fmt.Printf("I_path: %s", string(journal.J_content.I_path[:]))
	fmt.Printf("I_content: %s", string(journal.J_content.I_content[:]))
	fmt.Printf("I_date: %s", fecha)
}

// función para crear los journal necesarios , operacion es 1|0 (archivo o )
func CrearJournalOperacion(operacion string, path string, contenido string, cadena string) error {

	// creamos el journal con los datos que conocemos
	journal := &Journal{
		J_count: 1,
		J_content: Information{
			I_operation: [10]byte{},
			I_path:      [64]byte{},
			I_content:   [64]byte{},
			I_date:      float32(time.Now().Unix()),
		},
	}

	// copiamos los contenidos
	copy(journal.J_content.I_operation[:], operacion)
	copy(journal.J_content.I_path[:], path)
	copy(journal.J_content.I_content[:], contenido)

	// pasamos el id de la  partición (obligatoriamente tiene que ser de la partición del usuario actualmente)
	backup := ObtenerBackup(global.UsuarioActual.PartitionID)
	// obtenemos el último journal
	ultimoJournal := backup.ObtenerUltimoJournal()

	// si hay journals entonces me lo devolverá y usaremos su contador
	if ultimoJournal != nil {
		// cambiamos el count por count += UltimoJournal.count
		journal.J_count += ultimoJournal.J_count
	}

	// obtenemos el pathDisco // no verificamos si existe porque si lo hace luego
	//fmt.Println(global.UsuarioActual.PartitionID)

	pathDisco := global.ParticionesMontadas[global.UsuarioActual.PartitionID]
	//fmt.Println(pathDisco)

	// obtenemos el mbr
	var mbr MBR
	if err := mbr.Deserialize(pathDisco); err != nil {
		return errors.New("ERROR: no fue posible leer el mbr del path: " + pathDisco)
	}

	// obtenemos la particion
	particion, err := mbr.ParticionPorId(global.UsuarioActual.PartitionID)
	if err != nil {
		return err
	}

	// Serializar el journal ( pasamos true al final porque continua entonces no continua el journal count)
	if err := journal.Serialize(pathDisco, int64(particion.Start+int32(binary.Size(SuperBlock{})))); err != nil {
		return errors.New("ERROR: no fue posible serializar el journal de la operacióna actual")
	}

	// ya habiendo serializado guardamos el journal y luego actualizamos el backup
	backup.ActualizarRecovery(cadena)
	backup.AgregarJournal(*journal)
	backup.ActualizarBackup(global.UsuarioActual.PartitionID)
	return nil
}
