package adminArchivos

import (
	"encoding/binary"
	"errors"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strconv"
	"strings"
)

type parametros_chmod struct {
	Path string   // path de la carpeta o archivo desde donde comienzo a modificar
	Ugo  [3]int32 // permisos ugo
	R    bool     // dice si el cambio es recursivo o no	- OPCIONAL
}

func Chmod(parametros []string) string {
	// verificando sesión iniciada
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "CHMOD ERROR: para poder cmabiar permisos UGO necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	// verificar que sea el usuario root
	usuario, _, _ := global.UsuarioActual.ObtenerUsuario()
	if usuario != "root" {
		msj := "CHMOD ERROR: solamente el usuario root puede cambiar permisos UGO.\n"
		fmt.Println(msj)
		return msj
	}

	chmod := parametros_chmod{}

	for _, parametro := range parametros {
		param := strings.Split(parametro, "=")
		switch strings.ToLower(param[0]) {
		case "path":
			chmod.Path = strings.ReplaceAll(param[1], "\"", "")
		case "ugo":
			u, g, o, err := verificarUgo(param[1])
			if err != nil {
				return err.Error()
			}
			chmod.Ugo[0] = u
			chmod.Ugo[1] = g
			chmod.Ugo[2] = o
		case "r":
			chmod.R = true
		default:
			msj := "CHMOD ERROR: el parametro " + param[0] + " es invalido.\n"
			fmt.Println(msj)
			return msj
		}
	}

	if chmod.Path == "" {
		msj := "CHMOD ERROR: el parametro PATH es obligatorio.\n"
		fmt.Println(msj)
		return msj
	}
	return comandoChmod(chmod)
}

func verificarUgo(valor string) (int32, int32, int32, error) {
	// verificamos que vengan exatamente 3 digitos
	if len(valor) != 3 {
		msj := "CHMOD ERROR: formato invalido para el valor UGO.\n"
		fmt.Println(msj)
		return -1, -1, -1, errors.New(msj)
	}

	// obtenemos los 3 digitos
	uString := valor[0:1]
	gString := valor[1:2]
	oString := valor[2:3]

	// convertimos a enteros
	u, err := strconv.Atoi(uString)
	if err != nil {
		msj := "CHMOD ERROR: no fue posible convertir a entero el valor U: " + uString + "\n"
		fmt.Println(msj)
		return -1, -1, -1, errors.New(msj)
	}
	g, err := strconv.Atoi(gString)
	if err != nil {
		msj := "CHMOD ERROR: no fue posible convertir a entero el valor U: " + gString + "\n"
		fmt.Println(msj)
		return -1, -1, -1, errors.New(msj)
	}

	o, err := strconv.Atoi(oString)
	if err != nil {
		msj := "CHMOD ERROR: no fue posible convertir a entero el valor U: " + oString + "\n"
		fmt.Println(msj)
		return -1, -1, -1, errors.New(msj)
	}

	// verificamos si son mayores de 7 (no verificamos negativos porque sino entonces el len(valor) >3)
	if u > 7 {
		msj := "CHMOD ERROR: el valor U no puede ser mayor a siete"
		fmt.Println(msj)
		return -1, -1, -1, errors.New(msj)
	}

	if g > 7 {
		msj := "CHMOD ERROR: el valor G no puede ser mayor a siete"
		fmt.Println(msj)
		return -1, -1, -1, errors.New(msj)
	}
	if g > 7 {
		msj := "CHMOD ERROR: el valor O no puede ser mayor a siete"
		fmt.Println(msj)
		return -1, -1, -1, errors.New(msj)
	}

	return int32(u), int32(g), int32(o), nil
}

func comandoChmod(comando parametros_chmod) string {
	_, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "CHMOD ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// verificamos que exista la carpeta o archivo
	idInodo := gestionSistema.BuscarInodo(0, comando.Path, *superBloque, pathDisco)
	if idInodo == -1 {
		msj := "CHMOD ERROR: no fue posible encontrar " + comando.Path
		fmt.Println(msj)
		return msj
	}

	// cambiamos permisos
	if err := cambiarUgoRecursivo(idInodo, comando, superBloque, pathDisco); err != nil {
		msj := "CHMOD " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// si es Ext3
	if superBloque.S_filesystem_type == 3 {
		ugo := fmt.Sprintf("%d%d%d", comando.Ugo[0], comando.Ugo[1], comando.Ugo[2])
		var valor string
		if comando.R {
			valor = fmt.Sprintf("chmod -path=%s -ugo=%s -r\n", comando.Path, ugo)
		} else {
			valor = fmt.Sprintf("chmod -path=%s -ugo=%s\n", comando.Path, ugo)
		}
		if err := estructuras.CrearJournalOperacion("chmod", comando.Path, ugo, valor); err != nil {
			msj := "CHMOD " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	if comando.R {
		return fmt.Sprintf("permisos UGO cambiados recursivamente desde %s \n", comando.Path)
	} else {
		return fmt.Sprintf("permisos UGO cambiados correctamente en %s \n", comando.Path)
	}
}

func cambiarUgoRecursivo(idInodo int32, comando parametros_chmod, superBloque *estructuras.SuperBlock, pathDisco string) error {
	// Cargar el inodo desde el que vamos a cambiar
	var inodo estructuras.Inode
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		return fmt.Errorf("ERROR: al cargar el inodo %d: %v", idInodo, err)
	}

	// cambiamos los permisos
	valor := fmt.Sprintf("%d%d%d", comando.Ugo[0], comando.Ugo[1], comando.Ugo[2])
	inodo.I_perm = [3]byte{}
	copy(inodo.I_perm[:], valor)

	// serializamos el inodo
	if err := inodo.Serialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		return fmt.Errorf("ERROR: al serializar el inodo %d: %v", idInodo, err)
	}

	// Si es una carpeta y se solicita recursividad
	if inodo.I_type[0] == '0' && comando.R {
		// vamos a visitar los demás inodos
		for i := 0; i < 12; i++ {
			idBloque := inodo.I_block[i]
			if idBloque != -1 {
				var folderBlock estructuras.FolderBlock
				err := folderBlock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(estructuras.FolderBlock{})))))
				if err != nil {
					return fmt.Errorf("ERROR: al cargar el bloque de carpeta %d: %v", idBloque, err)
				}

				// Recorrer el contenido del bloque de carpeta
				for j := 2; j < 4; j++ { // Empezamos desde 2 para saltar . y ..
					apuntador := folderBlock.B_content[j].B_inodo
					if apuntador != -1 {
						// no usamos el error porque solamente indica si hay permisos y si se cambio
						cambiarUgoRecursivo(apuntador, comando, superBloque, pathDisco)
					}
				}
			}
		}
	}

	return nil
}
