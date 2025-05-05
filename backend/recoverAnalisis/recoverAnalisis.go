package analisisRecovery

import (
	"fmt"
	archivos "main/comandos/adminArchivos"
	discos "main/comandos/adminDiscos"
	usuarios "main/comandos/adminUsuarios"
	reportes "main/comandos/reportes"
	"strings"
)

func AnalizarRecovery(cadena string) string {
	// respuesta final
	respuesta := ""
	// quitamos espacios al final
	linea := strings.TrimRight(cadena, " ")
	// obtenemos la lista de parametros y el comando: [comando, param1=valor1, param2=valor2, ...]
	parametros := strings.Split(linea, " -")

	switch strings.ToLower(parametros[0]) {
	// ADMINISTRACION DE comandos
	case "mkdisk":
		if len(parametros) > 1 {
			respuesta = discos.Mkdisk(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando mkdisk"
			fmt.Println(msj)
			respuesta += msj
		}
	case "rmdisk":
		if len(parametros) > 1 {
			respuesta = discos.Rmdisk(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando rmdisk"
			fmt.Println(msj)
			respuesta += msj
		}
	case "fdisk":
		if len(parametros) > 1 {
			respuesta = discos.Fdisk(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando fdisk"
			fmt.Println(msj)
			respuesta += msj
		}
	case "mount":
		if len(parametros) > 1 {
			respuesta = discos.Mount(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando mount"
			fmt.Println(msj)
			respuesta += msj
		}
	case "unmount":
		if len(parametros) > 1 {
			respuesta = discos.Unmount(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando unmount"
			fmt.Println(msj)
			respuesta += msj
		}
	case "mounted":
		respuesta = discos.Mounted()
	case "cat":
		if len(parametros) > 1 {
			respuesta = archivos.Cat(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando cat"
			fmt.Println(msj)
			respuesta += msj
		}
	case "login":
		if len(parametros) > 1 {
			respuesta = usuarios.Login(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando login"
			fmt.Println(msj)
			respuesta += msj
		}
	case "logout":
		respuesta = usuarios.Logout()
	case "mkgrp":
		if len(parametros) > 1 {
			respuesta = usuarios.Mkgrp(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando mkgrp"
			fmt.Println(msj)
			respuesta += msj
		}
	case "rmgrp":
		if len(parametros) > 1 {
			respuesta = usuarios.Rmgrp(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando rmgrp"
			fmt.Println(msj)
			respuesta += msj
		}
	case "mkusr":
		if len(parametros) > 1 {
			respuesta = usuarios.Mkusr(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando mkusr"
			fmt.Println(msj)
			respuesta += msj
		}
	case "rmusr":
		if len(parametros) > 1 {
			respuesta = usuarios.Rmusr(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando rmusr"
			fmt.Println(msj)
			respuesta += msj
		}
	case "chgrp":
		if len(parametros) > 1 {
			respuesta = usuarios.Chgrp(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando chgrp"
			fmt.Println(msj)
			respuesta += msj
		}
	case "mkfile":
		if len(parametros) > 1 {
			respuesta = archivos.Mkfile(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando mkfile"
			fmt.Println(msj)
			respuesta += msj
		}
	case "mkdir":
		if len(parametros) > 1 {
			respuesta = archivos.Mkdir(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando mkdir"
			fmt.Println(msj)
			respuesta += msj
		}
	case "rep":
		if len(parametros) > 1 {
			respuesta = reportes.Rep(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando rep"
			fmt.Println(msj)
			respuesta += msj
		}
	case "remove":
		if len(parametros) > 1 {
			respuesta = archivos.Remove(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando remove"
			fmt.Println(msj)
			respuesta += msj
		}
	case "edit":
		if len(parametros) > 1 {
			respuesta = archivos.Edit(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando edit"
			fmt.Println(msj)
			respuesta += msj
		}
	case "rename":
		if len(parametros) > 1 {
			respuesta = archivos.Rename(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando rename"
			fmt.Println(msj)
			respuesta += msj
		}
	case "copy":
		if len(parametros) > 1 {
			respuesta = archivos.Copy(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando copy"
			fmt.Println(msj)
			respuesta += msj
		}
	case "move":
		if len(parametros) > 1 {
			respuesta = archivos.Move(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando move"
			fmt.Println(msj)
			respuesta += msj
		}
	case "find":
		if len(parametros) > 1 {
			respuesta = archivos.Find(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando find"
			fmt.Println(msj)
			respuesta += msj
		}
	case "chown":
		if len(parametros) > 1 {
			respuesta = archivos.Chown(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando chown"
			fmt.Println(msj)
			respuesta += msj
		}
	case "chmod":
		if len(parametros) > 1 {
			respuesta = archivos.Chmod(parametros[1:])
		} else {
			msj := "ERROR: Faltan parametros en el comando chmod"
			fmt.Println(msj)
			respuesta += msj
		}
	default:
		msj := "ERROR: Comando no reconocido: " + parametros[0]
		fmt.Println(msj)
		return msj
	}

	return respuesta
}
