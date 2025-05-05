package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"main/analisis"
	"main/estructuras"
	"main/global"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// ------------------------------------------------ estructuras JSON ------------------------------------------------

type Respuesta struct {
	Salida string `json:"salida"`
}

type Solicitud struct {
	Entrada string `json:"entrada"`
}

type DiskJSON struct {
	Ruta          string `json:"ruta"`          // ruta de la localización del disco
	Nombre        string `json:"nombre"`        // nombre del disco
	Size          string `json:"size"`          //tamaño
	FechaCreacion string `json:"fechaCreacion"` //fecha de creacion
	Id            int32  `json:"id"`            //numero random
	Ajuste        string `json:"ajuste"`        //ajuste (F, W , B)
	Letra         string `json:"letra"`         // letra si se montó alguna partición
}

type ParticionJSON struct {
	Estado      string `json:"estado"`      // estado de la partición 1 si está montada, 0 si fue creada y -1 si solo se inicializó
	Tipo        string `json:"tipo"`        // tipo de partición P, E o L
	Ajuste      string `json:"ajuste"`      // Ajuste de la partición: B, F, W
	Inicio      string `json:"inicio"`      // byte en el que inicia la partición
	Size        string `json:"size"`        // tamaño de la partición
	Nombre      string `json:"nombre"`      // nombre de la partición
	Correlativo string `json:"correlativo"` // Correlativo de la partición (-1 si no ha sido montada)
	Id          string `json:"id"`          // identificador de la partición si fue montada
}

type Usuario struct {
	Nombre      string `json:"nombre"`
	Clave       string `json:"clave"`
	IdParticion string `json:"idParticion"`
}

// CarpetaJSON estructura para representar carpetas en el sistema de archivos
type CarpetaJSON struct {
	Nombre            string        `json:"nombre"`
	Permisos          string        `json:"permisos"`
	Propietario       string        `json:"propietario"`
	Grupo             string        `json:"grupo"`
	FechaCreacion     string        `json:"fechaCreacion"`
	FechaModificacion string        `json:"fechaModificacion"`
	FechaAcceso       string        `json:"fechaAcceso"`
	Tipo              string        `json:"tipo"`
	Hijos             []interface{} `json:"hijos"` // puede contener CarpetaJSON o ArchivoJSON
}

// ArchivoJSON estructura para representar archivos en el sistema de archivos
type ArchivoJSON struct {
	Nombre            string `json:"nombre"`
	Permisos          string `json:"permisos"`
	Propietario       string `json:"propietario"`
	Grupo             string `json:"grupo"`
	FechaCreacion     string `json:"fechaCreacion"`
	FechaModificacion string `json:"fechaModificacion"`
	FechaAcceso       string `json:"fechaAcceso"`
	Tipo              string `json:"tipo"`
	Contenido         string `json:"contenido"`
	Size              int32  `json:"size"`
}

// ------------------------------------------------ ANALIZAR COMANDOS ------------------------------------------------

func handlerAnalizar(c *fiber.Ctx) error {
	var req Solicitud
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "JSON inválido"})
	}

	// inicializamos bufio y el resultado
	analizador := bufio.NewScanner(strings.NewReader(req.Entrada))
	var resultado string

	for analizador.Scan() {
		// si la linea está vacía, la ignoro
		if analizador.Text() == "" {
			continue
		}

		/*
			basicamente esparamos la linea para que si viene:
				comando # comentario
			comando estén en la posición [0] y el comentario en la posición [1]

			en caso sea:
				# comentario
			entonces:
			la posición [0] será un string vacío y la posición [1] será el comentario
		*/
		linea := strings.Split(analizador.Text(), "#")

		if len(linea[0]) != 0 {
			msj := "--> Ejecutando: " + linea[0] + "\n"
			msj += analisis.Analizar(linea[0]) + "\n\n"
			fmt.Println(msj)
			resultado += msj
		}

		// si es un comentario, lo ignoro
		if len(linea) > 1 && linea[1] != "" {
			//fmt.Println("Comentario ignorado: " + linea[1])
			continue
		}
	}

	return c.Status(fiber.StatusOK).JSON(Respuesta{Salida: resultado})
}

// ------------------------------------------------ ARCHIVOS LOCALES ------------------------------------------------

// obteniene el nombre de los archivos de una carpeta
func InfoCarpeta(carpeta string) ([]string, error) {
	var nombres []string
	archivos, err := os.ReadDir(carpeta)
	if err != nil {
		return nil, err
	}
	for _, archivo := range archivos {
		if !archivo.IsDir() {
			info, err := archivo.Info() // Accede a la información del archivo
			if err != nil {
				return nil, err
			}
			nombres = append(nombres, info.Name())
		}
	}
	return nombres, nil
}

// ------------------------------------------------ OBTENER DISCOS, PARTICIONES Y SISTEMA ------------------------------------------------

// construirArbolJSON construye un árbol JSON con la estructura de carpetas y archivos
func construirArbolJSON(idInodo int32, nombre string, superBloque *estructuras.SuperBlock, pathDisco string) (interface{}, error) {
	var inodo estructuras.Inode
	offset := int64(superBloque.S_inode_start + (idInodo * int32(binary.Size(estructuras.Inode{}))))

	if err := inodo.Deserialize(pathDisco, offset); err != nil {
		return nil, fmt.Errorf("ERROR: al cargar el inodo %d: %v", idInodo, err)
	}

	fmt.Println(inodo.I_uid)
	// Obtener datos del propietario y grupo
	nombreUsuario, nombreGrupo, err := superBloque.ObtenerUsuario_Grupo(pathDisco, inodo.I_uid)
	if err != nil {
		return nil, fmt.Errorf("ERROR: %v", err)
	}

	// Formatear permisos
	permisos := convertirPermisosAFormato(string(inodo.I_perm[:]))

	// Formatear fechas
	fechaCreacion := time.Unix(int64(inodo.I_ctime), 0).Format(time.RFC3339)
	fechaModificacion := time.Unix(int64(inodo.I_mtime), 0).Format(time.RFC3339)
	fechaAcceso := time.Unix(int64(inodo.I_atime), 0).Format(time.RFC3339)

	// Procesar según el tipo (carpeta o archivo)
	if inodo.I_type[0] == '0' { // Es una carpeta
		carpeta := CarpetaJSON{
			Nombre:            nombre,
			Permisos:          permisos,
			Propietario:       nombreUsuario,
			Grupo:             nombreGrupo,
			FechaCreacion:     fechaCreacion,
			FechaModificacion: fechaModificacion,
			FechaAcceso:       fechaAcceso,
			Tipo:              "carpeta",
			Hijos:             []interface{}{},
		}

		// Procesar bloques directos
		for i := 0; i < 12; i++ {
			idBloque := inodo.I_block[i]
			if idBloque == -1 {
				continue
			}

			var folderBlock estructuras.FolderBlock
			blockOffset := int64(superBloque.S_block_start + (idBloque * int32(binary.Size(estructuras.FolderBlock{}))))

			if err := folderBlock.Deserialize(pathDisco, blockOffset); err != nil {
				return nil, fmt.Errorf("ERROR: al cargar el bloque de carpeta %d: %v", idBloque, err)
			}

			// Ignoramos . y .. (índices 0 y 1)
			for j := 2; j < 4; j++ {
				apuntador := folderBlock.B_content[j].B_inodo
				nombreHijo := strings.TrimRight(string(folderBlock.B_content[j].B_name[:]), "\x00")

				if apuntador != -1 && nombreHijo != "" {
					hijo, err := construirArbolJSON(apuntador, nombreHijo, superBloque, pathDisco)
					if err != nil {
						return nil, err
					}
					carpeta.Hijos = append(carpeta.Hijos, hijo)
				}
			}
		}
		return carpeta, nil

	} else { // Es un archivo
		var contenido strings.Builder

		// Procesar bloques directos
		for i := 0; i < 12; i++ {
			idBloque := inodo.I_block[i]
			if idBloque == -1 {
				continue
			}

			contenidoBloque, err := leerBloqueArchivo(idBloque, superBloque, pathDisco)
			if err != nil {
				return nil, err
			}
			contenido.WriteString(contenidoBloque)
		}

		// Procesar bloque de indirección simple
		if inodo.I_block[12] != -1 {
			contenidoIndirecto, err := leerBloquesIndirectos(inodo.I_block[12], superBloque, pathDisco, 1)
			if err != nil {
				return nil, err
			}
			contenido.WriteString(contenidoIndirecto)
		}

		// Procesar bloque de indirección doble
		if inodo.I_block[13] != -1 {
			contenidoIndirecto, err := leerBloquesIndirectos(inodo.I_block[13], superBloque, pathDisco, 2)
			if err != nil {
				return nil, err
			}
			contenido.WriteString(contenidoIndirecto)
		}

		// Procesar bloque de indirección triple
		if inodo.I_block[14] != -1 {
			contenidoIndirecto, err := leerBloquesIndirectos(inodo.I_block[14], superBloque, pathDisco, 3)
			if err != nil {
				return nil, err
			}
			contenido.WriteString(contenidoIndirecto)
		}

		archivo := ArchivoJSON{
			Nombre:            nombre,
			Permisos:          permisos,
			Propietario:       nombreUsuario,
			Grupo:             nombreGrupo,
			FechaCreacion:     fechaCreacion,
			FechaModificacion: fechaModificacion,
			FechaAcceso:       fechaAcceso,
			Tipo:              "archivo",
			Contenido:         contenido.String(),
			Size:              inodo.I_size,
		}

		return archivo, nil
	}
}

// leerBloqueArchivo lee un bloque de archivo y devuelve su contenido como string
func leerBloqueArchivo(idBloque int32, superBloque *estructuras.SuperBlock, pathDisco string) (string, error) {
	var fileBlock estructuras.FileBlock
	blockOffset := int64(superBloque.S_block_start + (idBloque * int32(binary.Size(estructuras.FileBlock{}))))

	if err := fileBlock.Deserialize(pathDisco, blockOffset); err != nil {
		return "", fmt.Errorf("ERROR: al cargar el bloque de archivo %d: %v", idBloque, err)
	}

	// Eliminar caracteres no legibles y nulos
	contenido := global.BorrandoIlegibles(string(fileBlock.B_content[:]))
	return contenido, nil
}

// leerBloquesIndirectos lee los bloques indirectos para archivos
func leerBloquesIndirectos(idBloque int32, superBloque *estructuras.SuperBlock, pathDisco string, nivel int) (string, error) {
	var resultado strings.Builder

	// Si es nivel 1, procesamos un bloque de punteros que apuntan a bloques de archivo
	if nivel == 1 {
		var pointerBlock [16]int32 // Asumiendo que un bloque de punteros contiene 16 punteros (64 bytes / 4 bytes)
		pointerBlockOffset := int64(superBloque.S_block_start + (idBloque * int32(binary.Size(estructuras.FileBlock{}))))

		file, err := os.Open(pathDisco)
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = file.Seek(pointerBlockOffset, 0)
		if err != nil {
			return "", err
		}

		err = binary.Read(file, binary.LittleEndian, &pointerBlock)
		if err != nil {
			return "", err
		}

		for _, puntero := range pointerBlock {
			if puntero == -1 {
				continue
			}

			contenido, err := leerBloqueArchivo(puntero, superBloque, pathDisco)
			if err != nil {
				return "", err
			}
			resultado.WriteString(contenido)
		}
	} else if nivel > 1 {
		// Para niveles mayores a 1, procesamos bloques de punteros que apuntan a otros bloques de punteros
		var pointerBlock [16]int32
		pointerBlockOffset := int64(superBloque.S_block_start + (idBloque * int32(binary.Size(estructuras.FileBlock{}))))

		file, err := os.Open(pathDisco)
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = file.Seek(pointerBlockOffset, 0)
		if err != nil {
			return "", err
		}

		err = binary.Read(file, binary.LittleEndian, &pointerBlock)
		if err != nil {
			return "", err
		}

		for _, puntero := range pointerBlock {
			if puntero == -1 {
				continue
			}

			contenidoIndirecto, err := leerBloquesIndirectos(puntero, superBloque, pathDisco, nivel-1)
			if err != nil {
				return "", err
			}
			resultado.WriteString(contenidoIndirecto)
		}
	}

	return resultado.String(), nil
}

// convertirPermisosAFormato convierte los permisos a formato legible tipo "rwxr-xr--"
func convertirPermisosAFormato(permisos string) string {
	if len(permisos) != 3 {
		return permisos // Si no tiene el formato esperado, devolver como está
	}

	var resultado strings.Builder

	// Convertir cada dígito a formato rwx
	for _, c := range permisos {
		num, err := strconv.Atoi(string(c))
		if err != nil {
			return permisos
		}

		r := "-"
		w := "-"
		x := "-"

		if num&4 != 0 {
			r = "r"
		}
		if num&2 != 0 {
			w = "w"
		}
		if num&1 != 0 {
			x = "x"
		}

		resultado.WriteString(r + w + x)
	}

	return resultado.String()
}

// devuelve el los discos locales
func handlerDiscos(c *fiber.Ctx) error {
	carpeta := global.RutaDiscosLocales // obtenemos la ruta de los nombres locales (se guarda al ejecutar mkdisk)
	nombres, err := InfoCarpeta(carpeta)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Respuesta{Salida: "Ha ocurrido un error al leer la carpeta donde se encuentran los discos"})
	}

	var discos []DiskJSON
	// vamos a obtener el mbr de cada disco (variable disco es el nombre)
	for _, disco := range nombres {
		ruta := carpeta + "/" + disco
		var mbr estructuras.MBR
		if err := mbr.Deserialize(ruta); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Respuesta{Salida: "Ha ocurrido un error al leer el disco"})
		}

		var valor string
		// si pasamos de 1000 Kb entonces pasamos a Mb
		sizeK, unitK := global.RevertirConversionUnidades(mbr.Size, "k")
		if sizeK >= 1000 {
			sizeM, unitM := global.RevertirConversionUnidades(mbr.Size, "m")
			valor = fmt.Sprintf("%d %s", sizeM, unitM)
		} else {
			valor = fmt.Sprintf("%d %s", sizeK, unitK)
		}

		discoMOntaje, existe := global.Montaje[ruta]

		var letra string
		if !existe {
			letra = "no asignada"
		} else {
			letra = discoMOntaje.Letra
		}

		discos = append(discos, DiskJSON{
			Ruta:          ruta,
			Nombre:        disco,
			Size:          valor,
			FechaCreacion: global.ObtenerFecha(time.Unix(int64(mbr.CreationDate), 0)),
			Id:            mbr.DiskSignature,
			Ajuste:        string(mbr.Fit[0]),
			Letra:         letra,
		})
	}

	// devolvemos la lista de discos en la carpeta
	return c.Status(fiber.StatusOK).JSON(discos)
}

// devuelve las particiones
func handlerParticiones(c *fiber.Ctx) error {
	var req Solicitud
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "JSON inválido"})
	}

	archivo := global.RutaDiscosLocales // obtenemos la ruta de los discos locales (se guarda al ejecutar mkdisk)
	archivo += "/" + req.Entrada        // agregamos el nombre del disco

	var mbr estructuras.MBR
	if err := mbr.Deserialize(archivo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Respuesta{Salida: "Ha ocurrido un error al leer el disco"})
	}

	var particiones []ParticionJSON
	for _, particion := range mbr.Partitions {
		// solo enviamos las que están inicializadas
		if particion.Start != -1 {
			var valor string
			// si pasamos de 1000 Kb entonces pasamos a Mb
			sizeK, unitK := global.RevertirConversionUnidades(particion.Size, "k")
			if sizeK >= 1000 {
				sizeM, unitM := global.RevertirConversionUnidades(particion.Size, "m")
				valor = fmt.Sprintf("%d %s", sizeM, unitM)
			} else {
				valor = fmt.Sprintf("%d %s", sizeK, unitK)
			}

			var estado string
			estado = "creada"
			if int32(particion.Status[0]) == 1 {
				estado = "montada"
			}

			particiones = append(particiones, ParticionJSON{
				Estado:      estado,
				Tipo:        string(particion.Type[0]),
				Ajuste:      string(particion.Fit[0]),
				Inicio:      fmt.Sprintf("%d B", particion.Start),
				Size:        valor,
				Nombre:      global.BorrandoIlegibles(string(particion.Name[:])),
				Correlativo: fmt.Sprintf("%d", particion.Correlative),
				Id:          global.BorrandoIlegibles(string(particion.Id[:])),
			})
		}
	}
	// devolvemos la lista de discos en la carpeta
	return c.Status(fiber.StatusOK).JSON(particiones)
}

// devuelve las particiones extendidas ( solo hay 1 por disco entonces podemos solo buscarla con el nombre del disco)
func handlerParticionesExtendidas(c *fiber.Ctx) error {
	var req Solicitud
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "JSON inválido"})
	}

	archivo := global.RutaDiscosLocales // obtenemos la ruta de los discos locales (se guarda al ejecutar mkdisk)
	archivo += "/" + req.Entrada        // agregamos el nombre del disco

	var mbr estructuras.MBR
	if err := mbr.Deserialize(archivo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Respuesta{Salida: "Ha ocurrido un error al leer el disco"})
	}

	var particionesExtendidas []ParticionJSON
	for _, particion := range mbr.Partitions {
		// solo enviamos las que están inicializadas
		if particion.Start != -1 && string(particion.Type[0]) == "E" {

			var ebr estructuras.EBR
			offset := particion.Start
			for i := 0; ebr.Next != -1; i++ {

				if err := ebr.Deserialize(archivo, int64(offset)); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(Respuesta{Salida: "Ha ocurrido un error al leer el EBR"})
				}

				if ebr.Size == 0 {
					break
				}

				estadoTextual := "Montada"
				if ebr.Status == 0 {
					estadoTextual = "Creada"
				}

				var valor string
				// si pasamos de 1000 Kb entonces pasamos a Mb
				sizeK, unitK := global.RevertirConversionUnidades(ebr.Size, "k")
				if sizeK >= 1000 {
					sizeM, unitM := global.RevertirConversionUnidades(ebr.Size, "m")
					valor = fmt.Sprintf("%d %s", sizeM, unitM)
				} else {
					valor = fmt.Sprintf("%d %s", sizeK, unitK)
				}

				particionesExtendidas = append(particionesExtendidas, ParticionJSON{
					Estado:      estadoTextual,
					Tipo:        string(ebr.Type[0]),
					Ajuste:      string(ebr.Fit[0]),
					Inicio:      fmt.Sprintf("%d B", ebr.Start),
					Size:        valor,
					Nombre:      global.BorrandoIlegibles(string(ebr.Name[:])),
					Correlativo: "",
					Id:          "",
				})

				// vamos al siguiente ebr
				offset += ebr.Size
			}
		}
	}

	// devolvemos la lista de discos en la carpeta
	return c.Status(fiber.StatusOK).JSON(particionesExtendidas)
}

// handlerSistema maneja las solicitudes para obtener la estructura completa del sistema de archivos
func handlerSistema(c *fiber.Ctx) error {
	var req Solicitud
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "JSON inválido"})
	}

	fmt.Println("Solicitando estructura del sistema para partición: " + req.Entrada)

	// Obtener el superbloque de la partición solicitada
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(req.Entrada)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "Error: El ID de partición no existe en el sistema"})
	}

	// Construir el árbol JSON del sistema de archivos
	arbol, err := construirArbolJSON(0, "/", superBloque, pathDisco)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Respuesta{
			Salida: fmt.Sprintf("Error al construir el sistema de archivos: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"sistema": arbol, // 👈 envía directamente la estructura como JSON
		"mensaje": "Sistema de archivos generado correctamente",
	})
}

// ------------------------------------------------ cerrar, iniciar y verificar sesión ------------------------------------------------

func handlerVerificar(c *fiber.Ctx) error {
	// si hay sesión iniciada entonces retornamos valores actuales
	if global.UsuarioActual.HaySesionIniciada() {
		usuario, password, idPart := global.UsuarioActual.ObtenerUsuario()
		return c.Status(fiber.StatusOK).JSON(Usuario{Nombre: usuario, Clave: password, IdParticion: idPart})
	}
	// sino entonces un JSON vacío
	return c.Status(fiber.StatusBadRequest).JSON(Usuario{Nombre: "", Clave: "", IdParticion: ""})
}

func handlerLogin(c *fiber.Ctx) error {
	// obtenemos los valores enviados
	var req Usuario
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "JSON inválido"})
	}

	// verificamos el idPart
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(req.IdParticion)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "no existe el id de la particióon en sistema"})
	}

	//contenido users.txt
	contenido, _ := superBloque.ObtenerUsuariosTxt(pathDisco)

	// separamos por "\n"
	lineas := strings.Split(contenido, "\n")
	// para saber si el user y el pass coinciden
	user := false
	pwd := false
	// iteramos en las lineas
	for _, linea := range lineas {
		// obtenemos los atributos
		atributos := strings.Split(linea, ",")
		// verificamos que sea un usuario
		if len(atributos) == 5 && atributos[1] == "U" {
			// verificamos si se encuentra el usuario y la contraseña
			// id, U , grupo , usuario , clave
			if atributos[3] == req.Nombre {
				user = true
				if atributos[4] == req.Clave {
					pwd = true
				}
			}
		}
	}

	// si los datos no coinciden
	if !user {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "no existe el nombre de usuario ingresado"})
	}
	if !pwd {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "la contraseña no coincide con el usuario ingresado"})
	}

	// loggeamos el usuario si todo coincide
	global.UsuarioActual.Login(req.Nombre, req.Clave, req.IdParticion)

	//enviamos una respuesta
	return c.Status(fiber.StatusOK).JSON(Respuesta{Salida: fmt.Sprintf("usuario loggeado : nombre: %s en Id Partición: %s", req.Nombre, req.IdParticion)})
}

func handlerLogout(c *fiber.Ctx) error {
	// si hay sesión iniciada entonces retornamos
	if !global.UsuarioActual.HaySesionIniciada() {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "No hay sesión iniciada."})
	}

	// cerramos la sesión
	nombre, _, _ := global.UsuarioActual.ObtenerUsuario()
	global.UsuarioActual.Logout()
	return c.Status(fiber.StatusOK).JSON(Respuesta{Salida: "Se ha cerrado la sesión del usuario " + nombre + " exitosamente."})
}

// ------------------------------------------------ JOURNALING ------------------------------------------------

func handlerTablaJournaling(c *fiber.Ctx) error {
	if global.TablaHTMLJournaling == "" {
		return c.Status(fiber.StatusBadRequest).JSON(Respuesta{Salida: "no hay tabla journaling"})
	}
	return c.Status(fiber.StatusOK).JSON(Respuesta{Salida: global.TablaHTMLJournaling})
}

func handlerHealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(Respuesta{Salida: "Servidor inicializador correctamente"})
}

// ------------------------------------------------ endpoints ------------------------------------------------

func main() {
	// APP
	app := fiber.New() // Inicializamos Fiber

	// CORS
	core := cors.New(cors.Config{ // configuramos CORS
		AllowOrigins: "http://localhost:5173",        // solo se permiten peticiones de este origen
		AllowMethods: "GET, POST",                    // solo se permiten peticiones GET y POST
		AllowHeaders: "Origin, Content-Type, Accept", // solo se permiten estos headers
	})
	app.Use(core) // usamos Middleware de CORS

	// Peticiones GET
	app.Get("/health", handlerHealthCheck)              // verifica la salud del servidor
	app.Get("/tablaJournaling", handlerTablaJournaling) // devuelve la tabla del journaling como html puro
	app.Get("/verificarLogin", handlerVerificar)        // verifica si hay sesión iniciada actualmente
	app.Get("/logout", handlerLogout)                   // cerrar sesión
	app.Get("/discos", handlerDiscos)                   // endpoint para obtener discos y reenviar nombres de los discos locales

	// Peticiones POST
	app.Post("/analizar", handlerAnalizar)                           // endpoint para analizar el texto
	app.Post("/login", handlerLogin)                                 // endpoint para iniciar sesión
	app.Post("/particiones", handlerParticiones)                     // endpoint para obtener particiones y reenviar nombres de las particiones del disco
	app.Post("/particionesExtendidas", handlerParticionesExtendidas) // endpoint para obtener particiones extendidas
	app.Post("/sistema", handlerSistema)                             // envia todo el sistema de archivos

	// puerto
	log.Fatal(app.Listen(":8000")) // iniciamos el servidor en el puerto 8000
}
