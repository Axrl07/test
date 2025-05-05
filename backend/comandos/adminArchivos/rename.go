package adminArchivos

import (
	"encoding/binary"
	"fmt"
	"main/estructuras"
	"main/gestionSistema"
	"main/global"
	"strings"
)

type parametros_rename struct {
	Path string // dirección de la carpeta o archivo a renonmbrar
	Name string // nuevo nombre
}

func Rename(parametros []string) string {

	if !global.UsuarioActual.HaySesionIniciada() {
		msj := "RENAME ERROR: para utilizar el comando RENAME debe iniciar sesión.\n"
		fmt.Println(msj)
		return msj
	}

	rename := parametros_rename{}

	for _, param := range parametros {
		parametro := strings.Split(param, "=")
		switch strings.ToLower(parametro[0]) {
		case "path":
			rename.Path = strings.ReplaceAll(parametro[1], "\"", "")
		case "name":
			rename.Name = strings.ReplaceAll(parametro[1], "\"", "")
		default:
			return fmt.Sprint("RENAME ERROR: el parametro %s no es valido", parametro[0])
		}
	}

	if rename.Path == "" {
		return "RENAME ERROR: el parametro path es obligatorio"
	}

	if rename.Name == "" {
		return "RENAME ERROR: el parametro contenido es obligatorio"
	}

	return comandoRename(rename)
}

func comandoRename(comando parametros_rename) string {
	usuario, _, idParticion := global.UsuarioActual.ObtenerUsuario()

	// obtenemos la partición, el path del disco fisico (virtual), error
	superBloque, _, pathDisco, err := estructuras.ObtenerSuperBloque(idParticion)
	if err != nil {
		msj := "RENAME ERROR: no ha sido posible ejecutar ObtenerSuperBLoque(idParticion)"
		fmt.Println(msj)
		return msj
	}

	// para verificar los permisos
	uidCreador, gidGrupo, err := superBloque.ObtenerGID_UID(pathDisco, usuario)
	if err != nil {
		msj := "RENAME ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// verificamos que exista la carpeta o archivo
	var inodo estructuras.Inode
	idInodo := gestionSistema.BuscarInodo(0, comando.Path, *superBloque, pathDisco)
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		msj := "RENAME ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	// verificamos si existen permisos de escritura
	if inodo.I_uid != uidCreador {
		msj := "RENAME ERROR: no tiene permisos de escritura en el archivo o carpeta " + comando.Path
		fmt.Println(msj)
		return msj
	}
	if inodo.I_gid != gidGrupo {
		msj := "RENAME ERROR: no tiene permisos de escritura en el archivo o carpeta " + comando.Path
		fmt.Println(msj)
		return msj
	}

	// ahora vamos a renonmbrar el archivo o carpeta pero con el bloque que apunta al inodo
	// para ello obtenemos la tmp menos el nombre del archivo o carpeta
	tmp := strings.Split(comando.Path, "/")
	nombre := tmp[len(tmp)-1]
	// cortamos
	tmp = tmp[:len(tmp)-1]
	ruta := strings.Join(tmp, "/")

	// ahora obtenemos el bloque que apunta al inodo
	idInodoPadre := gestionSistema.BuscarInodo(0, ruta, *superBloque, pathDisco)
	if err := inodo.Deserialize(pathDisco, int64(superBloque.S_inode_start+(idInodoPadre*int32(binary.Size(estructuras.Inode{}))))); err != nil {
		msj := "RENAME ERROR: " + err.Error()
		fmt.Println(msj)
		return msj
	}

	if idInodoPadre != -1 {
		folderBLock := estructuras.FolderBlock{}
		for _, idBLoque := range inodo.I_block {
			if idBLoque != -1 {
				if err := folderBLock.Deserialize(pathDisco, int64(superBloque.S_block_start+(idBLoque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
					msj := "RENAME ERROR: " + err.Error()
					fmt.Println(msj)
					return msj
				}

				// buscamos si el nombre coincide con el nombre del inodo
				var salimos bool
				for i := 2; i < 4; i++ {
					// si coinciden los nombres entonces renombramos
					nombreGuardado := global.BorrandoIlegibles(string(folderBLock.B_content[i].B_name[:]))
					if nombreGuardado == nombre {
						folderBLock.B_content[i].B_name = [12]byte{}
						copy(folderBLock.B_content[i].B_name[:], []byte(comando.Name))
						// serializamos el bloque
						if err := folderBLock.Serialize(pathDisco, int64(superBloque.S_block_start+(idBLoque*int32(binary.Size(estructuras.FolderBlock{}))))); err != nil {
							msj := "RENAME ERROR: " + err.Error()
							fmt.Println(msj)
							return msj
						}
						salimos = true
						break
					}
				}

				if salimos {
					break
				}
			}
		}
	} else {
		msj := "RENAME ERROR: no ha sido posible renombrar " + comando.Path
		fmt.Println(msj)
		return msj
	}

	if superBloque.S_filesystem_type == 3 {
		if err := estructuras.CrearJournalOperacion("rename", comando.Path, comando.Name, fmt.Sprintf("rename -path=%s -name=%s\n", comando.Path, comando.Name)); err != nil {
			msj := "RENAME " + err.Error()
			fmt.Println(msj)
			return msj
		}
	}

	return "Se ha renombrado correctamente " + comando.Path + " a " + comando.Name + "\n"
}
