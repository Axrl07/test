package adminArchivos

import (
	"encoding/binary"
	"errors"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strings"
)

type parametros_chown struct {
	Path    string // path de la carpeta o archivo desde donde comienzo a modificar
	Usuario string // nombre del nuevo propietario
	R       bool   // dice si el cambio es recursivo o no	- OPCIONAL
	uid     int32
	gid     int32
}

func Chown(parametros []string) string {
	// verificando sesión iniciada
	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "CHOWN ERROR: para poder cmabiar permisos UGO necesita iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	// verificar que sea el usuario root
	usuario, _, idParticion := global.UsuarioActual.ObtenerUsuario()
	if usuario != "root" {
		msj := "CHOWN ERROR: solamente el usuario root puede cambiar permisos UGO.\n"
		fmt.Println(msj)
		return msj
	}

	chown := parametros_chown{}

	for _, parametro := range parametros {
		param := strings.Split(parametro, "=")
		switch strings.ToLower(param[0]) {
		case "path":
			chown.Path = strings.ReplaceAll(param[1], "\"", "")
		case "usuario":
			uidNuevo, gidNuevo, err := verificarUsuario(param[1], idParticion)
			if err != nil {
				return err.Error()
			}
			chown.uid = uidNuevo
			chown.gid = gidNuevo
		case "r":
			chown.R = true
		default:
			msj := "CHOWN ERROR: el parametro " + param[0] + " es invalido.\n"
			fmt.Println(msj)
			return msj
		}
	}

	if chown.Path == "" {
		msj := "CHOWN ERROR: el parametro PATH es obligatorio.\n"
		fmt.Println(msj)
		return msj
	}

	return comandoChown(chown)
}

func verificarUsuario(nombre string, idParticion string) (int32, int32, error) {
	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "CHOWN ERROR: no ha sido posible ejecutar obtener superbloque"
		fmt.Println(msj)
		return -1, -1, errors.New(msj)
	}

	// obtenemos gid y uid
	uid, gid, err := superBloque.ObtenerGID_UID(pathDisco, nombre)
	if err != nil {
		msj := "CHOWN ERROR: no existe el usuario " + nombre
		fmt.Println(msj)
		return -1, -1, errors.New(msj)
	}

	return uid, gid, nil
}

func comandoChown(comando parametros_chown) string {
	_, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "CHOWN ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// verificamos que exista la carpeta o archivo
	idInodo := gestionSistema.BuscarInodo(0, comando.Path, *superBloque, pathDisco)
	if idInodo == -1 {
		msj := "CHOWN ERROR: no fue posible encontrar " + comando.Path
		fmt.Println(msj)
		return msj
	}

	// cambiamos permisos
	if err := cambiarPropietarioRecursivo(idInodo, comando, superBloque, pathDisco); err != nil {
		msj := "CHOWN " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// si es Ext3
	if superBloque.S_filesystem_type == 3 {
		var valor string
		if comando.R {
			valor = fmt.Sprintf("chown -path=%s -usuario=%s -r\n", comando.Path, comando.Usuario)
		} else {
			valor = fmt.Sprintf("chown -path=%s -usuario=%s\n", comando.Path, comando.Usuario)
		}
		if err := estructuras.CrearJournalOperacion("chown", comando.Path, comando.Usuario, valor); err != nil {
			msj := "CHOWN " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	if comando.R {
		return fmt.Sprintf("el propietario de %s ha sido cambiado recursivamente\n", comando.Path)
	} else {
		return fmt.Sprintf("el propietario de %s ha sido cambiado correctamente en\n", comando.Path)
	}
}

func cambiarPropietarioRecursivo(idInodo int32, comando parametros_chown, superBloque *estructuras.SuperBlock, pathDisco string) error {
	// Cargar el inodo desde el que vamos a cambiar
	var inodo estructuras.Inode
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		return fmt.Errorf("ERROR: al cargar el inodo %d: %v", idInodo, err)
	}

	// cambiamos propietario
	inodo.I_uid = comando.uid
	inodo.I_gid = comando.gid

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
						cambiarPropietarioRecursivo(apuntador, comando, superBloque, pathDisco)
					}
				}
			}
		}
	}

	return nil
}
